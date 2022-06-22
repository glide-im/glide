package messaging

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
	"github.com/glide-im/glide/pkg/messaging/message_handler"
)

func RegisterOverAction() {
	handler, err := message_handler.NewHandler(nil, nil)
	if err != nil {
		panic(err)
	}

	// add a new handler to handle action "api.register"
	handler.AddHandler(messaging.NewActionWithReplyHandler("api.register", registerActionHandler))

	// test handler
	_ = handler.Handle(&gate.Info{
		ID: "_a_",
	}, messages.NewMessage(0, "api.register", RegisterRequest{}))
}

type RegisterRequest struct {
	Name     string
	Password string
}

func registerActionHandler(c *gate.Info, message *messages.GlideMessage) (*messages.GlideMessage, error) {

	r := &RegisterRequest{}
	err := message.Data.Deserialize(&r)
	if err != nil {
		return messages.NewMessage(0, messages.ActionApiFailed, "invalid request data"), nil
	}
	// do register

	return messages.NewMessage(0, messages.ActionApiSuccess, nil), nil
}
