package message_handler

import (
	"github.com/glide-im/glide/pkg/auth/jwt_auth"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
	"github.com/glide-im/glide/pkg/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestName(t *testing.T) {
	impl := jwt_auth.NewAuthorizeImpl("secret")
	token, err := impl.GetToken(&jwt_auth.JwtAuthInfo{
		UID:    "1233123",
		Device: "2",
	})
	t.Log(token, err)
}

func TestMessageHandler_InitDefaultHandler(t *testing.T) {

	handler, err := NewHandlerWithOptions(&Options{
		MessageStore:           &store.IdleMessageStore{},
		DontInitDefaultHandler: true,
	})
	assert.NoError(t, err)

	handler.InitDefaultHandler(func(action messages.Action, fn messaging.HandlerFunc) messaging.HandlerFunc {
		t.Log(action)
		return fn
	})
}
