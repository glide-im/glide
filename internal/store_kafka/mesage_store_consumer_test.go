package store_kafka

import (
	"encoding/json"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"testing"
	"time"
)

func TestNewConsumer(t *testing.T) {

	consumer, err := NewConsumer([]string{"localhost:9092"})
	if err != nil {
		t.Error(consumer)
	}
	defer consumer.Close()

	consumer.ConsumeMessage(func(m *messages.ChatMessage) {
		b, _ := json.Marshal(m)
		logger.D("message: %s", string(b))
	})

	time.Sleep(time.Second * 5)
}
