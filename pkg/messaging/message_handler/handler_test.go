package message_handler

import (
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
	"github.com/glide-im/glide/pkg/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
