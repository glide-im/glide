package store

import (
	"github.com/Shopify/sarama"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
)

type KafkaConsumer struct {
	consumer  sarama.Consumer
	cf        func(m *messages.ChatMessage)
	offlineCf func(m *messages.ChatMessage)
	channelCf func(m *messages.ChatMessage)
}

func NewKafkaConsumer(address []string) (*KafkaConsumer, error) {

	consumer, err := sarama.NewConsumer(address, sarama.NewConfig())
	if err != nil {
		return nil, err
	}
	c := &KafkaConsumer{
		consumer: consumer,
	}
	if err = c.run(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *KafkaConsumer) run() error {

	partitions, err2 := c.consumer.Partitions(KafkaChatMessageTopic)

	if err2 != nil {
		return err2
	}

	for _, partition := range partitions {

		consumer, err := c.consumer.ConsumePartition(KafkaChatMessageTopic, partition, sarama.OffsetNewest)
		if err != nil {
			return err
		}

		go func(pc sarama.PartitionConsumer) {
			for m := range pc.Messages() {
				var cm = messages.ChatMessage{}
				err2 := messages.JsonCodec.Decode(m.Value, &cm)
				if err2 != nil {
					logger.E("message decode error %v", err2)
					continue
				}
				if c.cf != nil {
					c.cf(&cm)
				}
			}
		}(consumer)

		consumer2, err := c.consumer.ConsumePartition(KafkaChannelMessageTopic, partition, sarama.OffsetNewest)
		if err != nil {
			return err
		}

		go func(pc sarama.PartitionConsumer) {
			for m := range pc.Messages() {
				var cm = messages.ChatMessage{}
				err2 := messages.JsonCodec.Decode(m.Value, &cm)
				if err2 != nil {
					logger.E("message decode error %v", err2)
					continue
				}
				if c.channelCf != nil {
					c.channelCf(&cm)
				}
			}
		}(consumer2)

		consumer3, err := c.consumer.ConsumePartition(KafkaChatOfflineMessageTopic, partition, sarama.OffsetNewest)
		if err != nil {
			return err
		}

		go func(pc sarama.PartitionConsumer) {
			for m := range pc.Messages() {
				var cm = messages.ChatMessage{}
				err2 := messages.JsonCodec.Decode(m.Value, &cm)
				if err2 != nil {
					logger.E("message decode error %v", err2)
					continue
				}
				if c.offlineCf != nil {
					c.offlineCf(&cm)
				}
			}
		}(consumer3)
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.consumer.Close()
}

func (c *KafkaConsumer) ConsumeChatMessage(cf func(m *messages.ChatMessage)) {
	c.cf = cf
}

func (c *KafkaConsumer) ConsumeChannelMessage(cf func(m *messages.ChatMessage)) {
	c.channelCf = cf
}

func (c *KafkaConsumer) ConsumeOfflineMessage(cf func(m *messages.ChatMessage)) {
	c.offlineCf = cf
}
