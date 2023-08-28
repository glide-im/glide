package subscription_impl

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/store"
	"github.com/glide-im/glide/pkg/subscription"
	"github.com/glide-im/glide/pkg/timingwheel"
	errors2 "github.com/pkg/errors"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

const (
	errNotMemberOfChannel    = "not member of channel"
	errPermissionDeniedWrite = "permission denied: write"
	errChannelMuted          = "channel is muted"
	errChannelBlocked        = "channel is blocked"
)

var tw = timingwheel.NewTimingWheel(time.Second, 3, 20)

// messageQueueTimeout channel message push queue idle timeout.
const messageQueueTimeout = time.Second * 10

// ChannelSequenceStore .
type ChannelSequenceStore interface {

	// NextSegmentSequence return the next segment of specified channel, and segment length.
	NextSegmentSequence(id subscription.ChanID, info subscription.ChanInfo) (int64, int64, error)
}

// SubscriberOptions is the options for the subscriber
type SubscriberOptions struct {
	Perm   Permission
	Ticket string
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
	Perm Permission
}

func (i *SubscriberInfo) canRead() bool {
	return i.Perm.allows(MaskPermRead)
}

func (i *SubscriberInfo) canWrite() bool {
	return i.Perm.allows(MaskPermWrite)
}

func (i *SubscriberInfo) isSystem() bool {
	return i.Perm.allows(MaskPermSystem)
}

func (i *SubscriberInfo) isAdmin() bool {
	return i.Perm.allows(MaskPermAdmin)
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

	messages chan *PublishMessage

	queueRunning int32
	queued       int32

	sleepTimer *timingwheel.Task

	activeAt    time.Time
	mu          *sync.RWMutex
	subscribers map[subscription.SubscriberID]*SubscriberInfo
	info        *subscription.ChanInfo

	store    store.SubscriptionStore
	seqStore ChannelSequenceStore
	gate     gate.DefaultGateway
}

func NewChannel(chanID subscription.ChanID, gate gate.DefaultGateway,
	store store.SubscriptionStore, seqStore ChannelSequenceStore) (*Channel, error) {

	ret := &Channel{
		id:          chanID,
		messages:    make(chan *PublishMessage, 100),
		sleepTimer:  tw.After(messageQueueTimeout),
		mu:          &sync.RWMutex{},
		subscribers: map[subscription.SubscriberID]*SubscriberInfo{},
		info:        &subscription.ChanInfo{},
		store:       store,
		seqStore:    seqStore,
		gate:        gate,
	}
	err := ret.loadSeq()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (g *Channel) Update(ci *subscription.ChanInfo) error {
	g.info.Blocked = ci.Blocked
	g.info.Muted = ci.Muted
	g.info.Secret = ci.Secret
	return nil
}

func (g *Channel) GetSubscribers() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var result []string
	for id := range g.subscribers {
		result = append(result, string(id))
	}
	return result
}

func (g *Channel) Subscribe(id subscription.SubscriberID, extra interface{}) error {
	so, err := getSubscriberOptions(extra)
	if err != nil {
		return err
	}

	if g.info.Closed {
		return errors.New("channel is closed")
	}
	if g.info.Blocked {
		return errors.New("channel is blocked")
	}

	logger.I("subscriber %s subscribe channel %s", id, g.id)

	g.mu.Lock()
	sb, ok := g.subscribers[id]
	if ok {
		return sb.update(so)
	} else {
		if len(g.info.Secret) != 0 {
			if len(so.Ticket) == 0 {
				g.mu.Unlock()
				return errors.New("invalid ticket")
			}
			c := fmt.Sprintf("%d_%s_%s", so.Perm, id, g.info.Secret)
			ticket := fmt.Sprintf("%x", md5.Sum([]byte(c)))
			if ticket != so.Ticket {
				g.mu.Unlock()
				return errors.New("invalid ticket")
			}
		}
		g.subscribers[id] = NewSubscriberInfo(so)
		logger.I("subscriber %s subscribe channel %s", id, g.id)
	}
	g.mu.Unlock()

	onlineNotify := PublishMessage{
		Message: messages.NewMessage(0, messages.ActionGroupNotify, subscription.NotifyMessage{
			From: "system",
			Type: subscription.NotifyTypeOnline,
			Body: struct {
				Uid string `json:"uid"`
			}{
				string(id),
			},
		}),
	}
	_ = g.enqueueNotify(&onlineNotify)

	statusNotify := PublishMessage{
		To: []subscription.SubscriberID{id},
		Message: messages.NewMessage(0, messages.ActionGroupNotify, subscription.NotifyMessage{
			From: "system",
			Type: subscription.NotifyOnlineMembers,
			Body: struct {
				Members []string `json:"members"`
			}{
				g.GetSubscribers(),
			},
		}),
	}
	_ = g.enqueueNotify(&statusNotify)

	return nil
}

func (g *Channel) Unsubscribe(id subscription.SubscriberID) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	logger.I("subscriber %s unsubscribe channel %s", id, g.id)
	_, ok := g.subscribers[id]
	if !ok {
		return errors.New(subscription.ErrNotSubscribed)
	}
	delete(g.subscribers, id)

	onlineNotify := PublishMessage{
		Message: messages.NewMessage(0, messages.ActionGroupNotify, subscription.NotifyMessage{
			From: "system",
			Type: subscription.NotifyTypeOffline,
			Body: struct {
				Uid string `json:"uid"`
			}{
				string(id),
			},
		}),
	}
	_ = g.enqueueNotify(&onlineNotify)
	return nil
}

func (g *Channel) Publish(msg subscription.Message) error {
	if g.info.Closed {
		return errors.New("channel closed")
	}
	message, ok := msg.(*PublishMessage)
	if !ok {
		return errors.New("unexpected message type, expect: *subscription.PublishMessage, actual:" + reflect.TypeOf(msg).String())
	}

	if !isValidMessageType(message.Type) {
		return errors.New(errUnknownMessageType)
	}

	g.mu.RLock()
	s, exist := g.subscribers[message.From]
	g.mu.RUnlock()

	if !exist {
		return errors.New(errNotMemberOfChannel)
	}
	if !s.canWrite() {
		return errors.New(errPermissionDeniedWrite)
	}
	if g.info.Muted {
		if !s.isSystem() || !s.isAdmin() {
			return errors.New(errChannelMuted)
		}
	}
	if g.info.Blocked {
		if !s.isSystem() {
			return errors.New(errChannelBlocked)
		}
	}

	switch message.Type {
	case TypeNotify:
		return g.enqueueNotify(message)
	case TypeMessage:
		err := g.enqueue(message)
		return err
	default:
		return errors.New(errUnknownMessageType)
	}
}

func (g *Channel) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.info.Closed = true

	close(g.messages)
	g.subscribers = map[subscription.SubscriberID]*SubscriberInfo{}

	if g.queued > 0 {
		logger.D("chan %s closed, %d messages dropped", g.id, g.queued)
	}
	return nil
}

func (g *Channel) enqueueNotify(msg *PublishMessage) error {
	select {
	case g.messages <- msg:
		atomic.AddInt32(&g.queued, 1)
	default:
		return errors.New("notify message queue is full")
	}
	return g.checkMsgQueue()
}

func (g *Channel) enqueue(m *PublishMessage) error {

	cm, err := m.GetChatMessage()
	if err != nil {
		return errors2.Wrap(err, "enqueue message deserialize body error")
	}
	m.Seq, err = g.nextSeq()
	if err != nil {
		return err
	}
	cm.Seq = m.Seq
	m.Message.Data = messages.NewData(&cm)

	if m.Type == TypeMessage {
		err = g.store.StoreChannelMessage(g.id, cm)
		if err != nil {
			return errors2.Wrap(err, "store channel message error")
		}
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

	if atomic.LoadInt32(&g.queueRunning) != 0 {
		return nil
	}

	atomic.StoreInt32(&g.queueRunning, 1)
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				atomic.StoreInt32(&g.queueRunning, 0)
				logger.E("message queue panic: %v", err)
			}
		}()

		g.sleepTimer = tw.After(messageQueueTimeout)
		for {
			select {
			case <-g.sleepTimer.C:
				g.sleepTimer.Cancel()
				if g.activeAt.Add(messageQueueTimeout).Before(time.Now()) {
					goto REST
				} else {
					g.sleepTimer = tw.After(messageQueueTimeout)
				}
			case m := <-g.messages:
				if m == nil {
					goto REST
				}
				atomic.AddInt32(&g.queued, -1)
				g.activeAt = time.Now()
				g.push(m)
			}
		}
	REST:
		dropped := 0
		if atomic.LoadInt32(&g.queued) > 0 {
			for {
				_, ok := <-g.messages
				if !ok {
					break
				}
				dropped++
				atomic.StoreInt32(&g.queued, -1)
			}
		}
		if dropped > 0 {
			logger.W("chan %s message queue stopped, %d message(s) have been dropped", g.id, dropped)
		} else {
			logger.D("chan %s message queue stopped", g.id)
		}
		atomic.StoreInt32(&g.queued, 0)
		atomic.StoreInt32(&g.queueRunning, 0)
	}()
	return nil
}

func (g *Channel) push(message *PublishMessage) {
	logger.I("chan %s push message: %v", g.id, message.Message)

	g.mu.RLock()
	defer g.mu.RUnlock()

	// TODO recycler use
	var received = map[subscription.SubscriberID]interface{}{}

	if message.To != nil {
		for _, id := range message.To {
			received[id] = nil
		}
	}

	for subscriberID, sInfo := range g.subscribers {
		if received != nil && len(received) > 0 {
			_, contained := received[subscriberID]
			if !contained {
				continue
			}
		}
		if !sInfo.canRead() {
			continue
		}
		err := g.gate.EnqueueMessage(gate.NewID2(string(subscriberID)), message.Message)
		if err != nil {
			logger.E("chan %s push message to subscribe %s error: %v", g.id, subscriberID, err)
		}
	}
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
	seq, length, err := g.seqStore.NextSegmentSequence(g.id, *g.info)
	if err != nil {
		return err
	}
	atomic.StoreInt64(&g.seqRemain, length)
	// because seq increment before set to message
	atomic.StoreInt64(&g.seq, seq)
	return nil
}
