package group_subscription

import (
	"errors"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/store"
	"github.com/glide-im/glide/pkg/subscription"
	"github.com/glide-im/glide/pkg/timingwheel"
	"github.com/panjf2000/ants/v2"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type memberInfo struct {
	online    bool
	muted     bool
	admin     bool
	deletedAt int64
}

func newMemberInfo() *memberInfo {
	return &memberInfo{
		online:    false,
		muted:     false,
		admin:     false,
		deletedAt: 0,
	}
}

var tw = timingwheel.NewTimingWheel(time.Second, 3, 20)
var queueExec *ants.Pool

// group message sequence segment length
const msgSeqSegmentLen = 200

const messageQueueSleep = time.Second * 10

func init() {
	var e error
	queueExec, e = ants.NewPool(200000,
		ants.WithNonblocking(true),
		ants.WithPreAlloc(true),
		ants.WithPanicHandler(onQueueExecutorPanic),
	)
	if e != nil {
		panic(e)
	}
}

func onQueueExecutorPanic(i interface{}) {
	logger.E("message queue goroutine pool handle message queue panic %v", i)
}

type Group struct {
	id subscription.ChanID

	msgSequence int64
	seqRemain   int64

	startup string

	mute      bool
	dissolved bool

	// messages 群消息队列
	messages chan *PublishMessage
	// notify 群通知队列
	notify chan *PublishMessage

	queueRunning int32
	queued       int32

	// checkActive 定时检查群活跃情况
	checkActive *timingwheel.Task

	lastMsgAt time.Time
	mu        *sync.Mutex
	members   map[subscription.SubscriberID]*memberInfo

	store store.SubscriptionStore
	gate  gate.Interface
}

func newGroup(chanID subscription.ChanID, seq int64) *Group {
	ret := new(Group)
	ret.mu = &sync.Mutex{}
	ret.members = map[subscription.SubscriberID]*memberInfo{}
	ret.startup = strconv.FormatInt(time.Now().Unix(), 10)
	ret.messages = make(chan *PublishMessage, 100)
	ret.notify = make(chan *PublishMessage, 10)
	ret.checkActive = tw.After(messageQueueSleep)
	ret.queueRunning = 0
	ret.msgSequence = seq
	ret.seqRemain = msgSeqSegmentLen
	ret.id = chanID
	return ret
}

func (g *Group) Update(extra interface{}) error {

	return nil
}

func (g *Group) Subscribe(id subscription.SubscriberID, extra interface{}) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	_, ok := g.members[id]
	if ok {
		return errors.New("already subscribed")
	}
	g.members[id] = newMemberInfo()
	return nil
}

func (g *Group) Unsubscribe(id subscription.SubscriberID) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	_, ok := g.members[id]
	if !ok {
		return errors.New("not subscribed")
	}
	delete(g.members, id)
	return nil
}

func (g *Group) UpdateSubscribe(id subscription.SubscriberID, extra interface{}) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	info, ok := g.members[id]
	if !ok {
		return errors.New("not subscribed")
	}
	g.members[id] = info
	return nil
}

func (g *Group) Publish(msg subscription.Message) error {

	message, ok := msg.(*PublishMessage)
	if !ok {
		return errors.New("invalid message type, expect *subscription.PublishMessage")
	}

	if message.From == "" {
		return g.EnqueueNotify(message)
	}

	_, err := g.EnqueueMessage(message)
	return err
}

func (g *Group) Close() error {

	return nil
}

func (g *Group) EnqueueNotify(msg *PublishMessage) error {
	seq := atomic.AddInt64(&g.msgSequence, 1)
	msg.Seq = seq
	select {
	case g.notify <- msg:
		atomic.AddInt32(&g.queued, 1)
	default:
		return errors.New("notify message queue is full")
	}
	return g.checkMsgQueue()
}

func (g *Group) EnqueueMessage(m *PublishMessage) (int64, error) {

	g.mu.Lock()
	mf, exist := g.members[m.From]
	g.mu.Unlock()

	if !exist {
		return 0, errors.New("not a group member")
	}
	if mf.muted {
		return 0, errors.New("a muted group member send message")
	}

	seq := atomic.AddInt64(&g.msgSequence, 1)

	cm := messages.ChatMessage{}
	err := m.Message.Data.Deserialize(&cm)
	if err != nil {
		return 0, err
	}
	cm.Seq = seq
	m.Message.Data = messages.NewData(&cm)

	err = g.store.StoreMessage(g.id, m)
	if err != nil {
		atomic.AddInt64(&g.msgSequence, -1)
		return 0, err
	}
	g.checkSeqRemain()

	err = g.store.StoreSeq(g.id, seq)
	if err != nil {
		logger.E("Group.EnqueueMessage update group message state error, %v", err)
		return 0, err
	}

	select {
	case g.messages <- m:
		atomic.AddInt32(&g.queued, 1)
	default:
		return 0, errors.New("too many messages,the group message queue is full")
	}
	if err := g.checkMsgQueue(); err != nil {
		if err == ants.ErrPoolOverload {
			logger.E("group message queue handle goroutine pool is overload")
		}
		return 0, err
	}
	return seq, nil
}

func (g *Group) checkSeqRemain() {
	g.seqRemain--
	remain := atomic.AddInt64(&g.seqRemain, -1)
	if remain <= 0 {
		// TODO load a new segment
	}
}

func (g *Group) checkMsgQueue() error {
	if atomic.LoadInt32(&g.queueRunning) == 1 {
		return nil
	}
	err := queueExec.Submit(
		func() {
			atomic.StoreInt32(&g.queueRunning, 1)
			logger.D("run a message queue reader goroutine")
			g.checkActive = tw.After(messageQueueSleep)
			for {
				select {
				case m := <-g.notify:
					g.lastMsgAt = time.Now()
					atomic.AddInt32(&g.queued, -1)
					switch m.Type {
					default:
						g.sendMessage(m)
					}
					// 优先派送群通知消息
					continue
				case <-g.checkActive.C:
					g.checkActive.Cancel()
					if g.lastMsgAt.Add(messageQueueSleep).Before(time.Now()) {
						q := atomic.LoadInt32(&g.queued)
						if q != 0 {
							logger.W("group message queue blocked, size=" + strconv.FormatInt(int64(q), 10))
							return
						}
						// 超过三十分钟没有发消息了, 停止消息下行任务
						goto REST
					} else {
						g.checkActive = tw.After(messageQueueSleep)
					}
				case m := <-g.messages:
					atomic.AddInt32(&g.queued, -1)
					g.lastMsgAt = time.Now()
					g.sendMessage(m)
				}
			}
		REST:
			logger.D("message queue read goroutine exit")
			atomic.StoreInt32(&g.queueRunning, 0)
		},
	)
	if err != nil {
		atomic.StoreInt32(&g.queueRunning, 0)
	}
	return err
}

func (g *Group) sendMessage(message *PublishMessage) {
	logger.D("Group.sendMessage: %s", message)

	g.mu.Lock()
	for subscriberID, mf := range g.members {
		if !mf.online || subscriberID == message.From {
			continue
		}
		err := g.gate.EnqueueMessage(gate.ID(subscriberID), message.Message)
		if err != nil {
			logger.E("%v", err)
		}
	}
	g.mu.Unlock()
}
