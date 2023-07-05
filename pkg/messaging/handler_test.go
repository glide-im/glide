package messaging

import (
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessageHandler_InitDefaultHandler(t *testing.T) {

	handler, err := NewHandlerWithOptions(nil, &MessageHandlerOptions{
		MessageStore:           &store.IdleMessageStore{},
		DontInitDefaultHandler: true,
	})
	assert.NoError(t, err)

	handler.InitDefaultHandler(func(action messages.Action, fn HandlerFunc) HandlerFunc {
		t.Log(action)
		return fn
	})
}
