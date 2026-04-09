package mq

import "context"

type Producer interface {
	SendMessage(ctx context.Context, message []byte) error
	Close()
}

type Consumer interface {
	ConsumeMessages(ctx context.Context, handler func(context.Context, []byte) error)
	Close()
}

type Options struct {
	brokers []string
	async   bool
	ack     int8
	topic   string

	groupID   string
	partition int
	offset    int64
}

type Option func(*Options)

func WithBrokers(brokers []string) Option {
	return func(o *Options) {
		o.brokers = brokers
	}
}

func WithAsync() Option {
	return func(o *Options) {
		o.async = true
	}
}

func WithAck(ack int8) Option {
	return func(o *Options) {
		o.ack = ack
	}
}

func WithTopic(topic string) Option {
	return func(o *Options) {
		o.topic = topic
	}
}

func WithOffset(offset int64) Option {
	return func(o *Options) {
		o.offset = offset
	}
}

func newOptions(opts ...Option) Options {
	options := Options{
		brokers: []string{"127.0.0.1:9092"},
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}
