package data

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/BitofferHub/seckill/internal/config"
	"github.com/BitofferHub/seckill/internal/log"
	"github.com/BitofferHub/seckill/internal/mq"
	"github.com/segmentio/kafka-go"
)

type managedKafkaConsumer struct {
	reader    *kafka.Reader
	closed    atomic.Bool
	closeOnce sync.Once
}

func newManagedKafkaConsumer(conf config.KafkaConsumerConf) mq.Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: conf.Brokers,
		Topic:   conf.Topic,
	})
	reader.SetOffset(conf.Offset)

	return &managedKafkaConsumer{reader: reader}
}

func (c *managedKafkaConsumer) ConsumeMessages(handler func([]byte) error) {
	for {
		message, err := c.reader.FetchMessage(context.Background())
		if err != nil {
			if c.closed.Load() {
				return
			}
			log.Errorf("Error fetching message: %v", err)
			continue
		}

		if err := handler(message.Value); err != nil {
			log.Errorf("Error handling message: %v", err)
		}

		if err := c.reader.CommitMessages(context.Background(), message); err != nil {
			if c.closed.Load() {
				return
			}
			log.Errorf("Error committing message: %v", err)
		}
	}
}

func (c *managedKafkaConsumer) Close() {
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		if err := c.reader.Close(); err != nil {
			log.Errorf("Error closing consumer: %v", err)
		}
	})
}
