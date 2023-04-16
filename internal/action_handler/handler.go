package action_handler

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
)

func Setup(handler messaging.Interface) {
	handler.AddHandler(&ClientCustomMessageHandler{})
	handler.AddHandler(&InternalActionHandler{})
	handler.AddHandler(messaging.NewActionHandler(messages.ActionHeartbeat, handleHeartbeat))
}

func dispatch2AllDevice(h *messaging.MessageInterfaceImpl, uid string, m *messages.GlideMessage) bool {
	devices := []string{"", "1", "2", "3"}
	ok := false
	for _, device := range devices {
		id := gate.NewID("", uid, device)
		if h.GetClientInterface().IsOnline(id) {
			err := h.GetClientInterface().EnqueueMessage(id, m)
			if err != nil {
				logger.E("dispatch message error %v", err)
			} else {
				ok = true
			}
		}
	}
	return ok
}
