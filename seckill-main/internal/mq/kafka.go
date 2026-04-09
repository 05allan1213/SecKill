package mq

import (
	"context"
	"errors"

	"github.com/BitofferHub/seckill/internal/log"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func (kp *KafkaProducer) SendMessage(ctx context.Context, message []byte) error {
	if ctx == nil {
		return errors.New("nil context")
	}
	return kp.writer.WriteMessages(ctx, kafka.Message{
		Value: message,
	})
}

func (kp *KafkaProducer) Close() {
	if err := kp.writer.Close(); err != nil {
		log.Errorf("Error closing producer: %v", err)
	}
}

func NewKafkaProducer(options ...Option) Producer {
	opts := newOptions(options...)

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(opts.brokers...),
		Topic:                  opts.topic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequiredAcks(opts.ack),
		Async:                  opts.async,
		AllowAutoTopicCreation: true,
	}

	return &KafkaProducer{writer: writer}
}
