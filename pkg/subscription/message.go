package subscription

import "github.com/glide-im/glide/pkg/messages"

// Message is a message that can be publishing to channel.
type Message interface {
	GetFrom() SubscriberID

	// GetChatMessage convert message body to *messages.ChatMessage
	GetChatMessage() (*messages.ChatMessage, error)
}

const (
	NotifyTypeOffline = 1
	NotifyTypeOnline  = 2
	NotifyTypeJoin    = 3
	NotifyTypeLeave   = 4

	NotifyOnlineMembers = 5
)

type NotifyMessage struct {
	From string      `json:"from"`
	Type int         `json:"type"`
	Body interface{} `json:"body"`
}
