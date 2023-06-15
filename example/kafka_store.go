package main

import (
	"encoding/json"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/store"
)

func main() {

	consumer, err := store.NewKafkaConsumer([]string{"192.168.99.191:9092"})
	if err != nil {
		panic(err)
	}
	consumer.ConsumeChatMessage(func(m *messages.ChatMessage) {
		j, _ := json.Marshal(m)
		logger.D("on chat message: %s", string(j))
	})
	consumer.ConsumeOfflineMessage(func(m *messages.ChatMessage) {
		j, _ := json.Marshal(m)
		logger.D("on offline message: %s", string(j))

	})
	consumer.ConsumeChannelMessage(func(m *messages.ChatMessage) {
		j, _ := json.Marshal(m)
		logger.D("on channel message: %s", string(j))
	})
	select {}
}
