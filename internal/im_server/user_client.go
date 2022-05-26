package im_server

import (
	"github.com/glide-im/glide/pkg/conn"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/timingwheel"
	"sync/atomic"
	"time"
)

// tw is a timer
var tw = timingwheel.NewTimingWheel(time.Millisecond*500, 3, 20)

// heartbeatDuration 心跳间隔
const heartbeatDuration = time.Second * 20

const (
	_ = iota
	stateRunning
	stateClosing
	stateClosed
)

// Client represent a user conn conn
type Client struct {
	conn conn.Connection

	logged    bool
	connectAt time.Time
	// state client 状态
	state int32

	// queuedMessage messages in the queue
	queuedMessage int64
	// messages 带缓冲的下行消息管道, 缓冲大小40
	messages chan *messages.GlideMessage
	// rCloseCh 关闭或写入则停止读
	rCloseCh   chan struct{}
	readClosed int32

	// hbR 心跳倒计时
	hbR    *timingwheel.Task
	hbLost int

	hbW *timingwheel.Task

	info *gate.Info

	// seq 服务器下行消息递增序列号
	seq int64

	mgr gate.Manager

	msgHandler gate.MessageHandler
}

func NewClient(conn conn.Connection, mgr gate.Manager, handler gate.MessageHandler) *Client {
	ret := new(Client)
	ret.conn = conn
	ret.state = stateRunning
	// 大小为 40 的缓冲管道, 防止短时间消息过多如果网络连接 output 不及时会造成程序阻塞, 可以适当调整
	ret.messages = make(chan *messages.GlideMessage, 60)
	ret.connectAt = time.Now()
	ret.rCloseCh = make(chan struct{})
	ret.seq = 0
	ret.hbR = tw.After(heartbeatDuration)
	ret.hbW = tw.After(heartbeatDuration)
	ret.info = &gate.Info{
		ConnectionAt: time.Now().Unix(),
		CliAddr:      conn.GetConnInfo().Addr,
	}
	ret.mgr = mgr
	ret.msgHandler = handler
	return ret
}

func (c *Client) GetInfo() gate.Info {
	return *c.info
}

// SetID 设置 id 标识及设备标识
func (c *Client) SetID(id gate.ID) {

	if id == "" {
		c.logged = false
	}
	c.info.ID = id
}

func (c *Client) IsRunning() bool {
	return atomic.LoadInt32(&c.state) == stateRunning
}

func (c *Client) Logged() bool {
	return c.logged
}

// EnqueueMessage 放入下行消息队列
func (c *Client) EnqueueMessage(msg *messages.GlideMessage) error {
	atomic.AddInt64(&c.queuedMessage, 1)
	defer func() {
		e := recover()
		if e != nil {
			atomic.AddInt64(&c.queuedMessage, -1)
			logger.E("%v", e)
		}
	}()
	s := atomic.LoadInt32(&c.state)
	if s == stateClosed {
		logger.D("client has closed, enqueue msg failed")
		return ErrClientClosed
	}
	logger.I("Interface(id=%v, %s): %v", c.info.ID, msg.GetAction(), msg)
	if msg.GetSeq() < 0 {
		// 服务端主动发送的消息使用服务端的序列号
		msg.SetSeq(c.getNextSeq())
	}
	select {
	case c.messages <- msg:
	default:
		atomic.AddInt64(&c.queuedMessage, -1)
		// 消息 chan 缓冲溢出, 这条消息将被丢弃
		logger.E("msg chan is full, id=%v", c.info.ID)
	}
	return nil
}

// readMessage 开始从 Connection 中读取消息
func (c *Client) readMessage() {
	readChan, done := messageReader.ReadCh(c.conn)

	defer func() {
		err := recover()
		if err != nil {
			logger.E("read message error", err)
		}
	}()

	atomic.StoreInt32(&c.readClosed, 0)
	for {
		select {
		case <-c.rCloseCh:
			close(c.rCloseCh)
			goto STOP
		case <-c.hbR.C:
			c.hbLost++
			if c.hbLost > 3 {
				goto STOP
			}
			// reset client heartbeat
			c.hbR.Cancel()
			c.hbR = tw.After(heartbeatDuration)
			_ = c.EnqueueMessage(messages.NewMessage(0, messages.ActionHeartbeat, ""))
		case msg := <-readChan:
			if msg.err != nil {
				if !c.IsRunning() || c.handleError(msg.err) {
					// 连接断开或致命错误中断读消息
					goto STOP
				}
				continue
			}
			c.hbLost = 0
			c.hbR.Cancel()
			c.hbR = tw.After(heartbeatDuration)

			// 统一处理消息函数
			c.msgHandler(c.info, msg.m)
			msg.Recycle()
		}
	}
STOP:
	c.hbR.Cancel()
	atomic.StoreInt32(&c.readClosed, 1)
	close(done)
	logger.D("client read closed, id=%v", c.info.ID)
}

// writeMessage 开始向 Connection 中写入消息队列中的消息
func (c *Client) writeMessage() {
	defer func() {
		err := recover()
		if err != nil {
			logger.D("Client write message error: %v", err)
		}
	}()

	for {
		select {
		case <-c.hbW.C:
			if !c.IsRunning() {
				logger.D("read closed, down msg queue timeout, close write now, id=%v", c.info.ID)
				goto STOP
			}
			_ = c.EnqueueMessage(messages.NewMessage(c.getNextSeq(), messages.ActionHeartbeat, ""))
			c.hbW.Cancel()
			c.hbW = tw.After(heartbeatDuration)
		case m := <-c.messages:
			b, err := codec.Encode(m)
			if err != nil {
				logger.E("serialize output message", err)
				continue
			}
			err = c.conn.Write(b)
			atomic.AddInt64(&c.queuedMessage, -1)

			c.hbW.Cancel()
			c.hbW = tw.After(heartbeatDuration)
			if err != nil {
				if !c.IsRunning() || c.handleError(err) {
					// 连接断开或致命错误中断写消息
					goto STOP
				}
			}
		}
	}
STOP:
	c.Exit()
	atomic.StoreInt32(&c.state, stateClosed)
	close(c.messages)
	_ = c.conn.Close()
	logger.D("client write closed, uid=%v", c.info.ID)
}

// handleError 处理上下行消息过程中的错误, 如果是致命错误, 则返回 true
func (c *Client) handleError(err error) bool {
	if conn.ErrClosed != err {
		logger.E("handle message error: %s", err.Error())
	}
	if c.logged {
		err = c.mgr.ExitClient(c.info.ID)
		if err != nil {
			logger.E("%v", err)
		}
	}
	return true
}

// Exit 退出客户端
func (c *Client) Exit() {
	c.SetID("")

	s := atomic.LoadInt32(&c.state)
	if s == stateClosed || s == stateClosing {
		return
	}
	atomic.StoreInt32(&c.state, stateClosing)

	if atomic.LoadInt32(&c.readClosed) != 1 {
		c.rCloseCh <- struct{}{}
	}

	_ = c.mgr.ExitClient(c.info.ID)
}

// getNextSeq 获取下一个下行消息序列号 sequence
func (c *Client) getNextSeq() int64 {
	return atomic.AddInt64(&c.seq, 1)
}

func (c *Client) Run() {
	logger.I(">>>> client %s running, id=%v", c.conn.GetConnInfo().Addr, c.info.ID)
	go c.readMessage()
	go c.writeMessage()
}
