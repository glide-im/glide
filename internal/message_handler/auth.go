package message_handler

import (
	"github.com/glide-im/glide/internal/jwt_auth"
	"github.com/glide-im/glide/pkg/auth"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"strconv"
)

type AuthRequest struct {
}

func (d *MessageHandler) handleAuth(c *gate.Info, msg *messages.GlideMessage) error {

	t := auth.Token{}
	e := msg.DeserializeData(&t)
	if e != nil {
		resp := messages.NewMessage(0, messages.ActionApiFailed, "invalid token")
		d.enqueueMessage(c.ID, resp)
		return nil
	}

	info := jwt_auth.JwtAuthInfo{
		UID:    strconv.FormatInt(c.ID.UID(), 10),
		Device: strconv.FormatInt(c.ID.Device(), 10),
	}
	r, err := d.auth.Auth(&info, &t)

	if err == nil && r.Success {
		respMsg := messages.NewMessage(msg.Seq, messages.ActionApiSuccess, r.Response)
		jwtResp := r.Response.(jwt_auth.Response)
		id2 := gate.NewID("", jwtResp.Uid, jwtResp.Device)
		_ = d.def.GetClientInterface().SetClientID(c.ID, id2)
		d.enqueueMessage(id2, respMsg)
	} else {
		resp := messages.NewMessage(msg.Seq, messages.ActionApiFailed, r.Response)
		d.enqueueMessage(c.ID, resp)
	}
	return nil
}
