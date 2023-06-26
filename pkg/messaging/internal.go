package messaging

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
)

type InternalHandler struct {
}

func (c *InternalHandler) Handle(h *MessageInterfaceImpl, cliInfo *gate.Info, message *messages.GlideMessage) bool {
	return message.GetAction().IsInternal()
}
