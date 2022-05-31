package group_subscription

import (
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
)

type PublishMessage struct {
	From    subscription.SubscriberID
	Seq     int64
	Type    int
	Message *messages.GlideMessage
}
