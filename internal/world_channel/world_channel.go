package world_channel

import (
	"encoding/json"
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
	myId := subscription.SubscriberID(id.UID())
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
			Message: messages.NewMessage(0, messages.ActionGroupMessage, b),
		})

		time.Sleep(time.Millisecond * 400)
		_ = sub.Publish(chanId, &subscription_impl.PublishMessage{
			From:    "system",
			Type:    subscription_impl.TypeMessage,
			To:      []subscription.SubscriberID{myId},
			Message: messages.NewMessage(0, messages.ActionGroupMessage, b),
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
		Message: messages.NewMessage(0, messages.ActionGroupMessage, b),
	})
}
