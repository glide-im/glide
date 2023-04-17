package store_kafka

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/store"
	"time"
)

const (
	chatMessageTopic = "getaway_chat_message"
)

var _ store.MessageStore = &kafkaMessageStore{}

type kafkaMessageStore struct {
	producer sarama.AsyncProducer
}

func (m *kafkaMessageStore) StoreOffline(message *messages.ChatMessage) error {
	//TODO implement me
	panic("implement me")
}

func NewProducer(address []string) (*kafkaMessageStore, error) {

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	//config.Producer.Return.Successes = true

	producer, err := sarama.NewAsyncProducer(address, config)

	if err != nil {
		return nil, err
	}

	return &kafkaMessageStore{
		producer: producer,
	}, nil
}

func (m *kafkaMessageStore) Close() error {
	return m.producer.Close()
}

type msg struct {
	data []byte
}

func (m *msg) Encode() ([]byte, error) {
	return m.data, nil
}

func (m *msg) Length() int {
	return len(m.data)
}

func (m *kafkaMessageStore) StoreMessage(message *messages.ChatMessage) error {

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	cm := &sarama.ProducerMessage{
		Topic:     chatMessageTopic,
		Value:     &msg{data: msgBytes},
		Headers:   nil,
		Metadata:  nil,
		Offset:    0,
		Partition: 0,
		Timestamp: time.Now(),
	}
	m.producer.Input() <- cm
	return nil
}
