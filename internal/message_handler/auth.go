package message_handler

import (
	"github.com/glide-im/glide/pkg/auth"
	"github.com/glide-im/glide/pkg/client"
	"github.com/glide-im/glide/pkg/messages"
)

func (d *MessageHandler) handleAuth(c *client.Info, msg *messages.GlideMessage) error {

	t := auth.Token{}
	e := msg.DeserializeData(&t)
	if e != nil {
		resp := messages.NewMessage(0, messages.ActionApiFailed, "invalid token")
		d.enqueueMessage(c.ID, resp)
		return nil
	}
	result, err := auth.Auth(c, &t)

	if err != nil {
		resp := messages.NewMessage(0, messages.ActionApiSuccess, result)
		id := client.NewID("", "", "")
		_ = d.def.GetClientInterface().SigIn(c.ID, id)
		d.enqueueMessage(c.ID, resp)
	} else {
		resp := messages.NewMessage(0, messages.ActionApiFailed, err.Error())
		d.enqueueMessage(c.ID, resp)
	}
	return nil
}
