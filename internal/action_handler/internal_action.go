package action_handler

import (
	"github.com/glide-im/glide/internal/config"
	"github.com/glide-im/glide/internal/message_handler"
	"github.com/glide-im/glide/internal/world_channel"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
	"time"
)

type InternalActionHandler struct {
}

func (o *InternalActionHandler) Handle(h *messaging.MessageInterfaceImpl, cliInfo *gate.Info, m *messages.GlideMessage) bool {
	if m.GetAction().IsInternal() {
		if !cliInfo.ID.IsTemp() {

			switch m.GetAction() {
			case messages.ActionInternalOffline:
				go world_channel.OnUserOffline(gate.ID(m.Data.String()))
			case messages.ActionInternalOnline:
				go func() {
					defer func() {
						err, ok := recover().(error)
						if err != nil && ok {
							logger.ErrE("push offline message error", err)
						}
					}()
					go func() {
						time.Sleep(time.Second * 1)
						world_channel.OnUserOnline(gate.ID(m.Data.String()))
					}()

					if config.Common.StoreOfflineMessage {
						message_handler.PushOfflineMessage(h, cliInfo.ID.UID())
					}
				}()
			}
		}
		return true
	}
	return false
}
