package world_channel

import (
	"encoding/json"
	"fmt"
	messages2 "github.com/glide-im/glide/im_service/messages"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
	"github.com/glide-im/glide/pkg/subscription/subscription_impl"
	"time"
)

var sub subscription_impl.SubscribeWrap
var chanId = subscription.ChanID("the_world_channel")

func EnableWorldChannel(subscribe subscription_impl.SubscribeWrap) error {
	sub = subscribe
	err := sub.CreateChannel(chanId, &subscription.ChanInfo{
		ID: chanId,
	})
	if err != nil {
		return err
	}
	err = sub.Subscribe(chanId, "system", &subscription_impl.SubscriberOptions{Perm: subscription_impl.PermWrite})
	return err
}

func OnUserOnline(id gate.ID) {
	if id.IsTemp() {
		return
	}
	myId := subscription.SubscriberID(gate.NewID2(id.UID()))
	err := sub.Subscribe(chanId, myId,
		&subscription_impl.SubscriberOptions{Perm: subscription_impl.PermRead | subscription_impl.PermWrite})
	if err == nil {

		time.Sleep(time.Second)
		b := &messages.ChatMessage{
			Mid:     time.Now().UnixNano(),
			Seq:     0,
			From:    "system",
			To:      string(chanId),
			Type:    100,
			Content: id.UID(),
			SendAt:  time.Now().Unix(),
		}
		_ = sub.Publish(chanId, &subscription_impl.PublishMessage{
			From:    "system",
			Type:    subscription_impl.TypeMessage,
			Message: messages.NewMessage(0, messages2.ActionGroupMessage, b),
		})

		time.Sleep(time.Millisecond * 100)
		b.Mid = time.Now().UnixNano()
		b.SendAt = time.Now().Unix()
		b.Type = 1
		b.Content = fmt.Sprintf("欢迎来到世界频道, 在这个频道, 你可以与服务器所有用户聊天, 你的 UID 为: %s", id.UID())
		_ = sub.Publish(chanId, &subscription_impl.PublishMessage{
			From:    "system",
			Type:    subscription_impl.TypeMessage,
			To:      []subscription.SubscriberID{myId},
			Message: messages.NewMessage(0, messages2.ActionGroupMessage, b),
		})
	} else {
		logger.E("$v", err)
	}
}

func OnUserOffline(id gate.ID) {
	if id.IsTemp() {
		return
	}
	_ = sub.UnSubscribe(chanId, subscription.SubscriberID(gate.NewID2(id.UID())))
	b, _ := json.Marshal(&messages.ChatMessage{
		Mid:     time.Now().UnixNano(),
		Seq:     0,
		From:    "system",
		To:      string(chanId),
		Type:    101,
		Content: id.UID(),
		SendAt:  time.Now().Unix(),
	})
	_ = sub.Publish(chanId, &subscription_impl.PublishMessage{
		From:    "system",
		Type:    subscription_impl.TypeMessage,
		Message: messages.NewMessage(0, messages2.ActionGroupMessage, b),
	})
}
