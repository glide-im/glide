package subscription_impl

import (
	"errors"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/store"
	"github.com/glide-im/glide/pkg/subscription"
	"github.com/glide-im/glide/pkg/timingwheel"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var tw = timingwheel.NewTimingWheel(time.Second, 3, 20)

// group message sequence segment length
const defaultMsgSeqSegmentLen = 200

const messageQueueSleep = time.Second * 10

// ChannelSequenceStore .
type ChannelSequenceStore interface {

	// NextSegmentSequence return the next segment of specified channel, and segment length.
	NextSegmentSequence(id subscription.ChanID, info subscription.ChanInfo) (int64, int64, error)
}

//SubscriberOptions is the options for the subscriber
type SubscriberOptions struct {
	Perm Permission
}

// getSubscriberOptions assertion type of `i` is *SubscribeOptions
func getSubscriberOptions(i interface{}) (*SubscriberOptions, error) {
	so, ok1 := i.(*SubscriberOptions)
	if !ok1 {
		return nil, errors.New("extra expect type: *SubscriberOptions, actual: " + reflect.TypeOf(i).String())
	}
	return so, nil
}

type SubscriberInfo struct {
	ActiveAt int64
	Perm     Permission
}

func (i *SubscriberInfo) update(options *SubscriberOptions) error {
	i.Perm = options.Perm
	return nil
}

func NewSubscriberInfo(so *SubscriberOptions) *SubscriberInfo {
	return &SubscriberInfo{
		Perm: so.Perm,
	}
}

type Channel struct {
	id subscription.ChanID

	seq       int64
	seqRemain int64

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

	lastMsgAt   time.Time
	mu          *sync.Mutex
	subscribers map[subscription.SubscriberID]*SubscriberInfo
	info        subscription.ChanInfo

	store    store.SubscriptionStore
	seqStore ChannelSequenceStore
	gate     gate.Interface
}

func NewChannel(chanID subscription.ChanID, gate gate.Interface,
	store store.SubscriptionStore, seqStore ChannelSequenceStore) (*Channel, error) {

	ret := new(Channel)
	ret.gate = gate
	ret.store = store
	ret.seqStore = seqStore

	ret.mu = &sync.Mutex{}
	ret.subscribers = map[subscription.SubscriberID]*SubscriberInfo{}
	ret.messages = make(chan *PublishMessage, 100)
	ret.notify = make(chan *PublishMessage, 10)
	ret.checkActive = tw.After(messageQueueSleep)
	ret.queueRunning = 0
	ret.id = chanID
	err := ret.loadSeq()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (g *Channel) nextSeq() (int64, error) {
	if atomic.AddInt64(&g.seqRemain, -1) <= 0 {
		err := g.loadSeq()
		if err != nil {
			return 0, err
		}
		atomic.AddInt64(&g.seq, -1)
	}
	return atomic.AddInt64(&g.seq, 1), nil
}

func (g *Channel) loadSeq() error {
	seq, length, err := g.seqStore.NextSegmentSequence(g.id, g.info)
	if err != nil {
		return err
	}
	atomic.StoreInt64(&g.seqRemain, length)
	// because seq increment before set to message
	atomic.StoreInt64(&g.seq, seq)
	return nil
}

func (g *Channel) Update(ci *subscription.ChanInfo) error {
	// TODO
	return nil
}

func (g *Channel) Subscribe(id subscription.SubscriberID, extra interface{}) error {
	so, err := getSubscriberOptions(extra)
	if err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	_, ok := g.subscribers[id]
	if ok {
		return errors.New(subscription.ErrAlreadySubscribed)
	}
	g.subscribers[id] = NewSubscriberInfo(so)
	return nil
}

func (g *Channel) Unsubscribe(id subscription.SubscriberID) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	_, ok := g.subscribers[id]
	if !ok {
		return errors.New(subscription.ErrNotSubscribed)
	}
	delete(g.subscribers, id)
	return nil
}

func (g *Channel) UpdateSubscribe(id subscription.SubscriberID, extra interface{}) error {
	so, err := getSubscriberOptions(extra)
	if err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	info, ok := g.subscribers[id]
	if !ok {
		return errors.New(subscription.ErrNotSubscribed)
	}
	return info.update(so)
}

func (g *Channel) Publish(msg subscription.Message) error {

	message, ok := msg.(*PublishMessage)
	if !ok {
		return errors.New("unexpected message type, expect: *subscription.PublishMessage, actual:" + reflect.TypeOf(msg).String())
	}

	if !isValidMessageType(message.Type) {
		return errors.New(errUnknownMessageType)
	}

	switch message.Type {
	case TypeNotify:
		return g.EnqueueNotify(message)
	case TypeMessage:
		err := g.EnqueueMessage(message)
		return err
	default:
		return errors.New(errUnknownMessageType)
	}
}

func (g *Channel) Close() error {

	return nil
}

func (g *Channel) EnqueueNotify(msg *PublishMessage) error {
	select {
	case g.notify <- msg:
		atomic.AddInt32(&g.queued, 1)
	default:
		return errors.New("notify message queue is full")
	}
	return g.checkMsgQueue()
}

func (g *Channel) EnqueueMessage(m *PublishMessage) error {

	g.mu.Lock()
	s, exist := g.subscribers[m.From]
	g.mu.Unlock()

	if !exist {
		return errors.New("not a group member")
	}
	if exist {
		if !s.Perm.allows(MaskPermWrite) {
			return errors.New("permission denied: write")
		}
	}
	cm := messages.ChatMessage{}
	err := m.Message.Data.Deserialize(&cm)
	if err != nil {
		return err
	}
	m.Seq, err = g.nextSeq()
	if err != nil {
		return err
	}
	cm.Seq = m.Seq
	m.Message.Data = messages.NewData(&cm)
	err = g.store.StoreMessage(g.id, m)
	if err != nil {
		return err
	}

	select {
	case g.messages <- m:
		atomic.AddInt32(&g.queued, 1)
	default:
		return errors.New("too many messages,the group message queue is full")
	}
	if err = g.checkMsgQueue(); err != nil {
		return err
	}
	return nil
}

func (g *Channel) checkMsgQueue() error {
	if atomic.LoadInt32(&g.queueRunning) == 1 {
		return nil
	}

	go func() {
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
	}()
	return nil
}

func (g *Channel) sendMessage(message *PublishMessage) {
	logger.D("Channel.sendMessage: %s", message)

	g.mu.Lock()
	for subscriberID, mf := range g.subscribers {
		if mf.Perm.allows(MaskPermRead) {
			continue
		}

		err := g.gate.EnqueueMessage(gate.ID(subscriberID), message.Message)

		if err != nil {
			logger.E("%v", err)
		}
	}
	g.mu.Unlock()
}
