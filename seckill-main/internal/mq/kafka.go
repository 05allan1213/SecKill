package mq

import (
	"context"

	"github.com/BitofferHub/seckill/internal/log"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func (kp *KafkaProducer) SendMessage(message []byte) error {
	return kp.writer.WriteMessages(context.Background(), kafka.Message{
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
