package store

import (
	"encoding/json"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"testing"
	"time"
)

func TestNewConsumer(t *testing.T) {

	consumer, err := NewKafkaConsumer([]string{"localhost:9092"})
	if err != nil {
		t.Error(consumer)
	}
	defer consumer.Close()

	consumer.ConsumeChatMessage(func(m *messages.ChatMessage) {
		b, _ := json.Marshal(m)
		logger.D("message: %s", string(b))
	})
	consumer.ConsumeChannelMessage(func(m *messages.ChatMessage) {
		b, _ := json.Marshal(m)
		logger.D("message: %s", string(b))
	})
	consumer.ConsumeOfflineMessage(func(m *messages.ChatMessage) {
		b, _ := json.Marshal(m)
		logger.D("message: %s", string(b))
	})

	time.Sleep(time.Second * 5)
}
