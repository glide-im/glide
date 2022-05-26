package message_handler

import (
	"github.com/glide-im/glide/pkg/auth"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
)

func (d *MessageHandler) handleAuth(c *gate.Info, msg *messages.GlideMessage) error {

	t := auth.Token{}
	e := msg.DeserializeData(&t)
	if e != nil {
		resp := messages.NewMessage(0, messages.ActionApiFailed, "invalid token")
		d.enqueueMessage(c.ID, resp)
		return nil
	}
	err := d.auth.Auth(c, &t)

	if err != nil {
		resp := messages.NewMessage(0, messages.ActionApiSuccess, "")
		id := gate.NewID("", "", "")
		_ = d.def.GetClientInterface().SetClientID(c.ID, id)
		d.enqueueMessage(c.ID, resp)
	} else {
		resp := messages.NewMessage(0, messages.ActionApiFailed, err.Error())
		d.enqueueMessage(c.ID, resp)
	}
	return nil
}
