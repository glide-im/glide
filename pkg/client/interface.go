package client

import (
	"github.com/glide-im/glide/pkg/messages"
)

type ServerInfo struct {
	Online      int64
	MaxOnline   int64
	MessageSent int64
	StartAt     int64

	OnlineCli []Info
}

type Interface interface {
	SigIn(old ID, new_ ID) error

	Logout(id ID) error

	IsOnline(id ID) bool

	EnqueueMessage(id ID, message *messages.GlideMessage) error
}

type MessageHandler func(cliInfo *Info, message *messages.GlideMessage) error
