package gateway

import (
	"errors"
	"github.com/glide-im/glide/pkg/conn"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/timingwheel"
	"sync/atomic"
	"time"
)

// tw is a timer for heartbeat.
var tw = timingwheel.NewTimingWheel(time.Millisecond*500, 3, 20)

const (
	defaultServerHeartbeatDuration = time.Second * 30
	defaultHeartbeatDuration       = time.Second * 20
	defaultHeartbeatLostLimit      = 3
	defaultCloseImmediately        = false
)

// client state
const (
	_ int32 = iota
	// stateRunning client is running, can read and write message.
	stateRunning
	// stateClosing client is closing, can't read and write message, wait for send all message in queue done.
	stateClosing
	// stateClosed client is closed, cannot do anything.
	stateClosed
)

// ClientConfig client config
type ClientConfig struct {

	// ClientHeartbeatDuration is the duration of heartbeat.
	ClientHeartbeatDuration time.Duration

	// ServerHeartbeatDuration is the duration of server heartbeat.
	ServerHeartbeatDuration time.Duration

	// HeartbeatLostLimit is the max lost heartbeat count.
	HeartbeatLostLimit int

	// CloseImmediately true express when client exit, discard all message in queue, and close connection immediately,
	// otherwise client will close read, and mark as stateClosing, the client cannot receive and enqueue message,
	// after all message in queue is sent, client will close write and connection.
	CloseImmediately bool
}

// Client represent a user conn client.
type Client struct {

	// conn is the real connection
	conn conn.Connection

	// logged true if client has logged
	logged bool
	// connectAt is the time when client connected
	connectAt time.Time
	// state is the client state
	state int32

	// queuedMessage message count in the messages channel
	queuedMessage int64
	// messages is the buffered channel for message to push to client.
	messages chan *messages.GlideMessage

	// rCloseCh is the channel for read goroutine to close
	rCloseCh chan struct{}
	// wCloseCh is the channel for write goroutine to close
	wCloseCh chan struct{}
	// readClosed flag for read goroutine closed, non-zero means closed
	readClosed int32
	// writeClosed flag for write goroutine closed, non-zero means closed
	writeClosed int32

	// hbC is the timer for client heartbeat
	hbC *timingwheel.Task
	// hbS is the timer for server heartbeat
	hbS *timingwheel.Task
	// hbLost is the count of heartbeat lost
	hbLost int

	// info is the client info
	info *gate.Info

	// mgr the client manager which manage this client
	mgr gate.Gateway
	// msgHandler client message handler
	msgHandler gate.MessageHandler

	// config is the client config
	config *ClientConfig
}

func NewClientWithConfig(conn conn.Connection, mgr gate.Gateway, handler gate.MessageHandler, config *ClientConfig) *Client {
	if config == nil {
		config = &ClientConfig{
			ClientHeartbeatDuration: defaultHeartbeatDuration,
			ServerHeartbeatDuration: defaultServerHeartbeatDuration,
			HeartbeatLostLimit:      defaultHeartbeatLostLimit,
			CloseImmediately:        false,
		}
	}
	ret := new(Client)
	ret.conn = conn
	ret.state = stateRunning
	ret.messages = make(chan *messages.GlideMessage, 60)
	ret.connectAt = time.Now()
	ret.rCloseCh = make(chan struct{})
	ret.wCloseCh = make(chan struct{})
	ret.hbC = tw.After(config.ClientHeartbeatDuration)
	ret.hbS = tw.After(config.ServerHeartbeatDuration)
	ret.hbLost = 0
	ret.info = &gate.Info{
		ConnectionAt: time.Now().Unix(),
		CliAddr:      conn.GetConnInfo().Addr,
	}
	ret.mgr = mgr
	ret.msgHandler = handler
	ret.config = config
	return ret
}

func NewClient(conn conn.Connection, mgr gate.Gateway, handler gate.MessageHandler) *Client {
	return NewClientWithConfig(conn, mgr, handler, nil)
}

func (c *Client) GetInfo() gate.Info {
	return *c.info
}

// SetID set client id.
func (c *Client) SetID(id gate.ID) {
	if id == "" || id.IsTemp() {
		c.logged = false
	}
	c.info.ID = id
}

// IsRunning return true if client is running
func (c *Client) IsRunning() bool {
	return atomic.LoadInt32(&c.state) == stateRunning
}

func (c *Client) Logged() bool {
	return c.logged
}

// EnqueueMessage enqueue message to client message queue.
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
		return errors.New("client has closed")
	}
	logger.I("EnqueueMessage ID=%s msg=%v", c.info.ID, msg)
	select {
	case c.messages <- msg:
	default:
		atomic.AddInt64(&c.queuedMessage, -1)
		logger.E("msg chan is full, id=%v", c.info.ID)
	}
	return nil
}

// read message from client.
func (c *Client) read() {
	readChan, done := messageReader.ReadCh(c.conn)

	defer func() {
		err := recover()
		if err != nil {
			logger.E("read message error", err)
		}
	}()

	var closeReason string
	atomic.StoreInt32(&c.readClosed, 0)
	for {
		select {
		case <-c.rCloseCh:
			closeReason = "closed initiative"
			goto STOP
		case <-c.hbC.C:
			c.hbLost++
			if c.hbLost > c.config.HeartbeatLostLimit {
				closeReason = "heartbeat lost"
				goto STOP
			}
			// reset client heartbeat
			c.hbC.Cancel()
			c.hbC = tw.After(c.config.ClientHeartbeatDuration)
			_ = c.EnqueueMessage(messages.NewMessage(0, messages.ActionHeartbeat, nil))
		case msg := <-readChan:
			if msg.err != nil {
				if !c.IsRunning() || c.handleError(msg.err) {
					closeReason = "read error, " + msg.err.Error()
					goto STOP
				}
				continue
			}
			if c.info.ID == "" {
				continue
			}
			c.hbLost = 0
			c.hbC.Cancel()
			c.hbC = tw.After(c.config.ClientHeartbeatDuration)

			if msg.m.GetAction() == messages.ActionHello {
				data := msg.m.Data
				hello := messages.Hello{}
				err := data.Deserialize(&hello)
				if err != nil {
					_ = c.EnqueueMessage(messages.NewMessage(0, messages.ActionNotifyError, "invalid hello message"))
				} else {
					c.info.Version = hello.ClientVersion
				}
			} else {
				c.msgHandler(c.info, msg.m)
			}
			msg.Recycle()
		}
	}
STOP:
	c.hbC.Cancel()
	atomic.StoreInt32(&c.readClosed, 1)
	close(done)
	close(c.rCloseCh)
	logger.I("read message goroutine closed, reason=%s", closeReason)
}

// write message to client.
func (c *Client) write() {
	defer func() {
		err := recover()
		if err != nil {
			logger.D("write message error, exit client: %v", err)
			atomic.StoreInt32(&c.state, stateClosed)
			close(c.messages)
			_ = c.conn.Close()
		}
	}()

	var closeReason string
	for {
		select {
		case <-c.wCloseCh:
			closeReason = "closed initiative"
			goto STOP
		case <-c.hbS.C:
			if !c.IsRunning() {
				closeReason = "client is not active"
				goto STOP
			}
			_ = c.EnqueueMessage(messages.NewMessage(0, messages.ActionHeartbeat, ""))
			c.hbS.Cancel()
			c.hbS = tw.After(c.config.ServerHeartbeatDuration)
		case m := <-c.messages:
			b, err := codec.Encode(m)
			if err != nil {
				logger.E("serialize output message", err)
				continue
			}
			err = c.conn.Write(b)
			atomic.AddInt64(&c.queuedMessage, -1)

			c.hbS.Cancel()
			c.hbS = tw.After(c.config.ServerHeartbeatDuration)
			if err != nil {
				if !c.IsRunning() || c.handleError(err) {
					closeReason = "write error, " + err.Error()
					goto STOP
				}
			}
		}
	}
STOP:
	atomic.StoreInt32(&c.state, stateClosed)
	atomic.StoreInt32(&c.writeClosed, 1)
	close(c.wCloseCh)

	if !c.config.CloseImmediately {
		close(c.messages)
		_ = c.conn.Close()
	}

	logger.D("client closed, addr=%s, reason:%s", c.info.CliAddr, closeReason)
}

// handleError handle error, return true if client should exit.
func (c *Client) handleError(err error) bool {
	if conn.ErrClosed != err {
		logger.E("handle message error: %s", err.Error())
	}
	return true
}

// Exit client, note: exit client will not close conn right now, but will close when message chan is empty.
// It's close read right now, and close write when all message in queue is sent.
func (c *Client) Exit() {
	if c.logged {
		c.logged = false
	}
	if c.mgr != nil {
		mgr := c.mgr
		c.mgr = nil
		_ = mgr.ExitClient(c.info.ID)
	}

	c.SetID("")

	// discard all message in queue and close connection immediately
	if c.config.CloseImmediately && atomic.LoadInt32(&c.state) != stateClosed {
		atomic.StoreInt32(&c.state, stateClosed)
		if atomic.LoadInt32(&c.readClosed) == 0 {
			c.rCloseCh <- struct{}{}
		}
		if atomic.LoadInt32(&c.writeClosed) == 0 {
			c.wCloseCh <- struct{}{}
		}
		close(c.messages)
		_ = c.conn.Close()
	}

	if atomic.LoadInt32(&c.state) == stateClosed || atomic.LoadInt32(&c.state) == stateClosing {
		return
	}
	atomic.StoreInt32(&c.state, stateClosing)

	if atomic.LoadInt32(&c.readClosed) == 0 {
		c.rCloseCh <- struct{}{}
	}
}

func (c *Client) Run() {
	logger.I("new client running addr:%s id:%s", c.conn.GetConnInfo().Addr, c.info.ID)
	go c.read()
	go c.write()
}
