package store

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
	"time"
)

const (
	KafkaChatMessageTopic        = "getaway_chat_message"
	KafkaChatOfflineMessageTopic = "getaway_chat_offline_message"
	KafkaChannelMessageTopic     = "gateway_channel_message"
)

var _ MessageStore = &KafkaMessageStore{}
var _ SubscriptionStore = &KafkaMessageStore{}

type KafkaMessageStore struct {
	producer sarama.AsyncProducer
}

func NewKafkaProducer(address []string) (*KafkaMessageStore, error) {

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	//config.Producer.Return.Successes = true

	producer, err := sarama.NewAsyncProducer(address, config)

	if err != nil {
		return nil, err
	}

	return &KafkaMessageStore{
		producer: producer,
	}, nil
}

func (k *KafkaMessageStore) Close() error {
	return k.producer.Close()
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

func (k *KafkaMessageStore) StoreOffline(message *messages.ChatMessage) error {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	cm := &sarama.ProducerMessage{
		Topic:     KafkaChatOfflineMessageTopic,
		Value:     &msg{data: msgBytes},
		Headers:   nil,
		Metadata:  nil,
		Offset:    0,
		Partition: 0,
		Timestamp: time.Now(),
	}
	k.producer.Input() <- cm
	return nil
}

func (k *KafkaMessageStore) NextSegmentSequence(id subscription.ChanID, info subscription.ChanInfo) (int64, int64, error) {
	//TODO implement me
	return 0, 0, nil
}

func (k *KafkaMessageStore) StoreChannelMessage(ch subscription.ChanID, m *messages.ChatMessage) error {
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	cm := &sarama.ProducerMessage{
		Topic:     KafkaChannelMessageTopic,
		Value:     &msg{data: msgBytes},
		Headers:   nil,
		Metadata:  nil,
		Offset:    0,
		Partition: 0,
		Timestamp: time.Now(),
	}
	k.producer.Input() <- cm
	return nil
}

func (k *KafkaMessageStore) StoreMessage(message *messages.ChatMessage) error {

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	cm := &sarama.ProducerMessage{
		Topic:     KafkaChatMessageTopic,
		Value:     &msg{data: msgBytes},
		Headers:   nil,
		Metadata:  nil,
		Offset:    0,
		Partition: 0,
		Timestamp: time.Now(),
	}
	k.producer.Input() <- cm
	return nil
}
