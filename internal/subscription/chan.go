package subscription

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
	gid int64

	msgSequence int64
	seqRemain   int64

	startup string

	mute      bool
	dissolved bool

	// messages 群消息队列
	messages chan *messages.ChatMessage
	// notify 群通知队列
	notify chan *messages.GroupNotify

	queueRunning int32
	queued       int32

	// checkActive 定时检查群活跃情况
	checkActive *timingwheel.Task

	lastMsgAt time.Time
	mu        *sync.Mutex
	members   map[int64]*memberInfo

	store store.SubscriptionStore
	gate  gate.Interface
}

func (g *Group) Update(extra interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (g *Group) GetSubscriber(id subscription.SubscriberID) (subscription.Subscriber, error) {
	//TODO implement me
	panic("implement me")
}

func (g *Group) Subscribe(id subscription.SubscriberID, extra interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (g *Group) Unsubscribe(id subscription.SubscriberID) error {
	//TODO implement me
	panic("implement me")
}

func (g *Group) UpdateSubscribe(id subscription.SubscriberID, extra interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (g *Group) UnsubscribeAll() error {
	//TODO implement me
	panic("implement me")
}

func (g *Group) Publish(msg subscription.Message) error {
	//TODO implement me
	panic("implement me")
}

func (g *Group) Close() error {
	//TODO implement me
	panic("implement me")
}

func newGroup(gid int64, seq int64) *Group {
	ret := new(Group)
	ret.mu = &sync.Mutex{}
	ret.members = map[int64]*memberInfo{}
	ret.startup = strconv.FormatInt(time.Now().Unix(), 10)
	ret.messages = make(chan *messages.ChatMessage, 100)
	ret.notify = make(chan *messages.GroupNotify, 10)
	ret.checkActive = tw.After(messageQueueSleep)
	ret.queueRunning = 0
	ret.msgSequence = seq
	ret.seqRemain = msgSeqSegmentLen
	ret.gid = gid
	return ret
}

func (g *Group) EnqueueNotify(msg *messages.GroupNotify) error {
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

func (g *Group) EnqueueMessage(msg *messages.ChatMessage, n subscription.Message) (int64, error) {

	g.mu.Lock()
	mf, exist := g.members[msg.From]
	g.mu.Unlock()

	if !exist {
		return 0, errors.New("not a group member")
	}
	if mf.muted {
		return 0, errors.New("a muted group member send message")
	}

	now := time.Now().Unix()
	seq := atomic.AddInt64(&g.msgSequence, 1)

	err := g.store.StoreMessage("", n)
	if err != nil {
		atomic.AddInt64(&g.msgSequence, -1)
		return 0, err
	}
	g.checkSeqRemain()

	err = g.store.StoreSeq("", seq)
	if err != nil {
		logger.E("Group.EnqueueMessage update group message state error, %v", err)
		return 0, err
	}

	dMsg := msg
	msg.Seq = seq
	msg.To = g.gid
	msg.SendAt = now
	if err != nil {
		return 0, err
	}

	select {
	case g.messages <- dMsg:
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
						g.SendMessage(0, messages.NewMessage(0, messages.ActionNotifyGroup, m))
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
					g.SendMessage(m.From, messages.NewMessage(-1, messages.ActionGroupMessage, m))
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

func (g *Group) SendMessage(from int64, message *messages.GlideMessage) {
	logger.D("Group.SendMessage: %s", message)
	g.mu.Lock()
	for uid, mf := range g.members {
		if !mf.online || uid == from {
			continue
		}
		err := g.gate.EnqueueMessage(gate.NewID2(uid), message)
		if err != nil {
			logger.E("%v", err)
		}
	}
	g.mu.Unlock()
}

func (g *Group) updateMember(u subscription.Update) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	return nil
}

func (g *Group) GetMember(id int64) *memberInfo {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.members[id]
}

func (g *Group) PutMember(member int64, s *memberInfo) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.members[member] = s
}

func (g *Group) RemoveMember(uid int64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.members, uid)
}

func (g *Group) HasMember(uid int64) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	_, exist := g.members[uid]
	return exist
}
