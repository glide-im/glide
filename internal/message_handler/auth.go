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
	r, err := d.auth.Auth(c, &t)

	if err == nil {
		resp := messages.NewMessage(msg.Seq, messages.ActionApiSuccess, r.Response)
		_ = d.def.GetClientInterface().SetClientID(c.ID, r.ID)
		d.enqueueMessage(r.ID, resp)
	} else {
		resp := messages.NewMessage(msg.Seq, messages.ActionApiFailed, r.Response)
		d.enqueueMessage(c.ID, resp)
	}
	return nil
}
