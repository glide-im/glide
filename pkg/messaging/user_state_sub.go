package messaging

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"sync"
	"time"
)

type UserStateData struct {
	Uid    string `json:"uid,omitempty"`
	Online bool   `json:"online,omitempty" json:"online,omitempty"`
}

type StateSubscribeData struct {
	Uids []string `json:"uids,omitempty"`
}

type UserState struct {
	subscribers map[string]map[string]byte
	mySubs      map[string]map[string]byte

	mu      *sync.Mutex
	gateway gate.Gateway

	logStateAt int64
}

func NewUserState(gateway gate.Gateway) *UserState {
	return &UserState{
		subscribers: map[string]map[string]byte{},
		mySubs:      map[string]map[string]byte{},
		gateway:     gateway,
		mu:          &sync.Mutex{},
	}
}

func (u *UserState) onUserOnline(id gate.ID) {
	u.mu.Lock()
	defer u.mu.Unlock()

	mySubList, ok := u.subscribers[id.UID()]
	if !ok {
		mySubList = map[string]byte{}
		u.subscribers[id.UID()] = mySubList
	}
	u.notifyOnline(id, mySubList)
}

func (u *UserState) onUserOffline(id gate.ID) {
	u.mu.Lock()
	defer u.mu.Unlock()

	mySubList, ok := u.subscribers[id.UID()]
	if !ok || len(mySubList) == 0 {
		return
	}
	u.notifyOffline(id, mySubList)

	myId := id.UID()
	sub, ok := u.mySubs[myId]
	if !ok {
		return
	}
	for uid := range sub {
		target, ok2 := u.subscribers[uid]
		if !ok2 {
			continue
		}
		delete(target, uid)
	}
	delete(u.mySubs, myId)
}

func (u *UserState) subUserStateApi(c *gate.Info, m *messages.GlideMessage) error {
	data := StateSubscribeData{}
	err := m.Data.Deserialize(&data)
	if err != nil {
		return err
	}

	myId := c.ID.UID()
	u.mu.Lock()
	defer u.mu.Unlock()

	mySubs, ok := u.mySubs[myId]
	if !ok {
		mySubs = map[string]byte{}
		u.mySubs[myId] = mySubs
	}

	for _, uid := range data.Uids {
		subscribers, ok := u.subscribers[uid]
		if !ok {
			subscribers = map[string]byte{}
			u.subscribers[uid] = subscribers
		}
		u.subscribers[uid][myId] = 0
		mySubs[uid] = 0
	}
	return nil
}

func (u *UserState) notifyOnline(src gate.ID, to map[string]byte) {
	notify := messages.NewMessage(0, messages.ActionNotifyUserState, UserStateData{
		Uid:    src.UID(),
		Online: false,
	})
	for uid := range to {
		_ = u.gateway.EnqueueMessage(gate.NewID2(uid), notify)
	}

	var s = time.Now().Unix() - u.logStateAt
	if s > 900 {
		u.logStateAt = time.Now().Unix()
		logger.D("[UserState] online users: %d, subscribes: %d", len(u.mySubs), len(u.subscribers))
	}
}

func (u *UserState) notifyOffline(src gate.ID, to map[string]byte) {
	notify := messages.NewMessage(0, messages.ActionNotifyUserState, UserStateData{
		Uid:    src.UID(),
		Online: true,
	})
	for uid := range to {
		_ = u.gateway.EnqueueMessage(gate.NewID2(uid), notify)
	}
}
