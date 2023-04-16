package store_kafka

import (
	"github.com/Shopify/sarama"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
)

type Consumer struct {
	consumer sarama.Consumer
	cf       func(m *messages.ChatMessage)
}

func NewConsumer(address []string) (*Consumer, error) {

	consumer, err := sarama.NewConsumer(address, sarama.NewConfig())
	if err != nil {
		return nil, err
	}
	c := &Consumer{
		consumer: consumer,
	}
	if err = c.run(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Consumer) run() error {

	partitions, err2 := c.consumer.Partitions(chatMessageTopic)

	if err2 != nil {
		return err2
	}

	for _, partition := range partitions {

		consumer, err := c.consumer.ConsumePartition(chatMessageTopic, partition, sarama.OffsetNewest)
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
	}

	return nil
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}

func (c *Consumer) ConsumeMessage(cf func(m *messages.ChatMessage)) {
	c.cf = cf
}
