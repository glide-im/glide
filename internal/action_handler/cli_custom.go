package action_handler

import (
	m2 "github.com/glide-im/glide/im_service/messages"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
)

type ClientCustomMessageHandler struct {
}

func (c *ClientCustomMessageHandler) Handle(h *messaging.MessageInterfaceImpl, ci *gate.Info, m *messages.GlideMessage) bool {
	if m.Action != m2.ActionClientCustom {
		return false
	}
	dispatch2AllDevice(h, m.To, m)
	return true
}
