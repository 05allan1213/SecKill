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

func (c *managedKafkaConsumer) ConsumeMessages(ctx context.Context, handler func(context.Context, []byte) error) {
	if ctx == nil {
		log.Errorf("Error consuming message: nil context")
		return
	}
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if c.closed.Load() {
				return
			}
			log.Errorf("Error fetching message: %v", err)
			continue
		}

		if err := handler(ctx, message.Value); err != nil {
			log.Errorf("Error handling message: %v", err)
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
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
