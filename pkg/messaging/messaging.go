package messaging

import (
	"github.com/glide-im/glide/pkg/client"
	"github.com/glide-im/glide/pkg/messages"
)

// Interface for messaging.
type Interface interface {

	// Handle handles message from client, the entry point for the messaging.
	Handle(clientInfo *client.Info, msg *messages.GlideMessage) error

	// PutMessageHandler registers a handler for a message type/action.
	// If the handler is existing, it will be replaced.
	PutMessageHandler(action messages.Action, i client.MessageHandler)
}
