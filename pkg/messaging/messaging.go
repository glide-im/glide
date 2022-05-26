package messaging

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
)

// Interface for messaging.
type Interface interface {

	// Handle handles message from gate, the entry point for the messaging.
	Handle(clientInfo *gate.Info, msg *messages.GlideMessage) error

	// PutMessageHandler registers a handler for a message type/action.
	// If the handler is existing, it will be replaced.
	PutMessageHandler(action messages.Action, i HandlerFunc)
}

// Server is the messaging server.
type Server interface {
	Interface

	SetGate(g gate.Gateway)

	SetSubscription(g subscription.Interface)

	Run() error
}
