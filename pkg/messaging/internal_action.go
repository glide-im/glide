package messaging

import (
	"github.com/glide-im/glide/config"
	"github.com/glide-im/glide/internal/world_channel"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"time"
)

func (d *MessageHandlerImpl) handleInternalOffline(c *gate.Info, m *messages.GlideMessage) error {
	go world_channel.OnUserOffline(c.ID)

	d.userState.onUserOffline(c.ID)

	return nil
}

func (d *MessageHandlerImpl) handleInternalOnline(c *gate.Info, m *messages.GlideMessage) error {

	d.userState.onUserOnline(c.ID)

	go func() {
		defer func() {
			err, ok := recover().(error)
			if err != nil && ok {
				logger.ErrE("push offline message error", err)
			}
		}()
		go func() {
			time.Sleep(time.Second * 1)
			world_channel.OnUserOnline(c.ID)
		}()

		if config.Common.StoreOfflineMessage {
			// message_handler.PushOfflineMessage(h, cliInfo.ID.UID())
		}
	}()
	return nil
}
