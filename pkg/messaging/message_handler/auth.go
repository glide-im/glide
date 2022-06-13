package message_handler

import (
	"errors"
	"github.com/glide-im/glide/pkg/auth"
	"github.com/glide-im/glide/pkg/auth/jwt_auth"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/gate/gateway"
	"github.com/glide-im/glide/pkg/messages"
)

type AuthRequest struct {
}

func (d *MessageHandler) handleAuth(c *gate.Info, msg *messages.GlideMessage) error {

	t := auth.Token{}
	e := msg.Data.Deserialize(&t)
	if e != nil {
		resp := messages.NewMessage(0, messages.ActionApiFailed, "invalid token")
		d.enqueueMessage(c.ID, resp)
		return nil
	}

	info := jwt_auth.JwtAuthInfo{
		UID:    msg.From,
		Device: c.ID.Device(),
	}
	r, err := d.auth.Auth(&info, &t)

	if err == nil && r.Success {
		respMsg := messages.NewMessage(msg.Seq, messages.ActionApiSuccess, r.Response)
		jwtResp, ok := r.Response.(*jwt_auth.Response)
		if !ok {
			resp := messages.NewMessage(msg.Seq, messages.ActionApiFailed, nil)
			d.enqueueMessage(c.ID, resp)
			return errors.New("invalid response type: expected *jwt_auth.Response")
		}
		id2 := gate.NewID("", jwtResp.Uid, jwtResp.Device)
		err = d.def.GetClientInterface().SetClientID(c.ID, id2)
		if gateway.IsAlreadyExist(err) {
			_ = d.def.GetClientInterface().ExitClient(id2)
			err = d.def.GetClientInterface().SetClientID(c.ID, id2)
		}
		if gateway.IsClientNotExist(err) {
			return errors.New("auth client not exist")
		}
		d.enqueueMessage(id2, respMsg)
	} else {
		resp := messages.NewMessage(msg.Seq, messages.ActionApiFailed, r.Response)
		d.enqueueMessage(c.ID, resp)
	}
	return nil
}
