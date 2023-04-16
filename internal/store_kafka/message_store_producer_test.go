package store_kafka

import (
	"github.com/glide-im/glide/pkg/messages"
	"testing"
	"time"
)

func TestNewProducer(t *testing.T) {

	producer, err := NewProducer([]string{"localhost:9092"})
	defer producer.Close()
	if err != nil {
		t.Error(err)
	}

	err = producer.StoreMessage(&messages.ChatMessage{
		CliMid:  "1",
		Mid:     1,
		Seq:     1,
		From:    "2",
		To:      "2",
		Type:    2,
		Content: "2",
		SendAt:  2,
	})

	if err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second)

}
