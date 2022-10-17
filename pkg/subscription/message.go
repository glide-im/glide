package subscription

import "github.com/glide-im/glide/pkg/messages"

// Message is a message that can be publishing to channel.
type Message interface {
	GetFrom() SubscriberID

	// GetChatMessage convert message body to *messages.ChatMessage
	GetChatMessage() (*messages.ChatMessage, error)
}
