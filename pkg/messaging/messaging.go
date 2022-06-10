package messaging

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
)

// MessageHandler is the interface for message handler
type MessageHandler interface {
	// Handle handles the message, returns true if the message is handled,
	// otherwise the message is delegated to next handler.
	Handle(h *MessageInterfaceImpl, cliInfo *gate.Info, message *messages.GlideMessage) bool
}

// Interface for messaging.
type Interface interface {

	// Handle handles message from gate, the entry point for the messaging.
	Handle(clientInfo *gate.Info, msg *messages.GlideMessage) error

	AddHandler(i MessageHandler)
}

type Messaging interface {
	Interface

	SetSubscription(g subscription.Interface)

	SetGate(g gate.Gateway)
}

// Server is the messaging server.
type Server interface {
	Messaging

	Run() error
}
