package message_handler

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
)

type InternalHandler struct {
}

func (c *InternalHandler) Handle(h *messaging.MessageInterfaceImpl, cliInfo *gate.Info, message *messages.GlideMessage) bool {
	return message.GetAction().IsInternal()
}
