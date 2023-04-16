package messages

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
)

func CreateKickOutMessage(c *gate.Info) *messages.GlideMessage {
	return messages.NewMessage(0, NotifyKickOut, "")
}

// ClientCustom client custom message, server does not store to database.
type ClientCustom struct {
	Type    string      `json:"type,omitempty"`
	Content interface{} `json:"content,omitempty"`
}
