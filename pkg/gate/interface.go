package gate

import (
	"errors"
	"github.com/glide-im/glide/pkg/conn"
	"github.com/glide-im/glide/pkg/messages"
)

var ErrClientNotExist = errors.New("client not exist")
var ErrInvalidID = errors.New("invalid id")

// Interface is the basic and common interface for all gate implementations.
//As the basic gate, it is used to provide a common gate interface for other modules to interact with the gate.
type Interface interface {

	// SetClientID sets the client id with the new id.
	SetClientID(old ID, new_ ID) error

	// ExitClient exits the client with the given id.
	ExitClient(id ID) error

	// IsOnline returns true if the client is online.
	IsOnline(id ID) bool

	// EnqueueMessage enqueues the message to the client with the given id.
	EnqueueMessage(id ID, message *messages.GlideMessage) error
}

// Server is the interface for the gateway server, which is used to handle and manager client connections.
type Server interface {
	Interface

	// SetMessageHandler sets the client message handler.
	SetMessageHandler(h MessageHandler)

	// HandleConnection handles the new client connection and returns the random and temporary id set for the connection.
	HandleConnection(c conn.Connection) ID

	Run() error
}

// MessageHandler used to handle messages from the client,
//all messages are handled by the message handler.
type MessageHandler func(cliInfo *Info, message *messages.GlideMessage)
