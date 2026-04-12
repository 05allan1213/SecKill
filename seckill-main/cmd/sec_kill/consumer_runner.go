package main

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/BitofferHub/seckill/internal/data"
	bitlog "github.com/BitofferHub/seckill/internal/log"
	"github.com/BitofferHub/seckill/internal/mq"
	seckilllogic "github.com/BitofferHub/seckill/internal/logic/seckill"
	"github.com/BitofferHub/seckill/internal/svc"
)

type ConsumerRunner struct {
	svcCtx       *svc.ServiceContext
	done         chan struct{}
	once         sync.Once
	ctx          context.Context
	cancel       context.CancelFunc
	running      atomic.Bool
	retryRunning atomic.Bool
}

func NewConsumerRunner(svcCtx *svc.ServiceContext) *ConsumerRunner {
	return &ConsumerRunner{svcCtx: svcCtx}
}

func (r *ConsumerRunner) Start() error {
	r.done = make(chan struct{})
	r.ctx, r.cancel = context.WithCancel(context.Background())

	// 启动多个主消费者
	consumers := r.svcCtx.Data.GetMQConsumers()
	for i, consumer := range consumers {
		go r.startConsumer(i, consumer)
	}

	// 启动多个 retry 消费者
	retryConsumers := r.svcCtx.Data.GetRetryConsumers()
	for i, consumer := range retryConsumers {
		go r.startRetryConsumerByIndex(i, consumer)
	}

	return nil
}

// startConsumer 启动单个消费者
func (r *ConsumerRunner) startConsumer(index int, consumer mq.Consumer) {
	if consumer == nil {
		return
	}
	r.running.Store(true)
	defer r.running.Store(false)

	defer func() {
		if recovered := recover(); recovered != nil {
			bitlog.Error(r.ctx, "seckill consumer panic",
				bitlog.Field(bitlog.FieldAction, "mq.consume"),
				bitlog.Field("consumerIndex", index),
				bitlog.Field(bitlog.FieldError, recovered),
			)
		}
	}()

	consumer.ConsumeMessages(r.ctx, func(ctx context.Context, message []byte) error {
		result := seckilllogic.HandleConsumedMessageWithRetry(ctx, r.svcCtx, message, 0, "")
		return r.handleConsumeResult(ctx, message, result, 0)
	})
}

func (r *ConsumerRunner) startRetryConsumer() {
	retryConsumers := r.svcCtx.Data.GetRetryConsumers()
	if len(retryConsumers) == 0 {
		return
	}
	for i, consumer := range retryConsumers {
		go r.startRetryConsumerByIndex(i, consumer)
	}
}

// startRetryConsumerByIndex 启动单个 retry 消费者
func (r *ConsumerRunner) startRetryConsumerByIndex(index int, retryConsumer mq.Consumer) {
	if retryConsumer == nil {
		return
	}

	r.retryRunning.Store(true)
	defer r.retryRunning.Store(false)

	retryConsumer.ConsumeMessages(r.ctx, func(ctx context.Context, message []byte) error {
		envelope, err := r.svcCtx.MessageRepo.UnmarshalEnvelope(ctx, message)
		if err != nil {
			return r.sendToDLQ(ctx, message, "failed to parse retry message: "+err.Error())
		}
		if envelope == nil {
			return r.sendToDLQ(ctx, message, "invalid retry message: nil envelope")
		}

		backoffMs := r.svcCtx.Config.Data.Kafka.Retry.BackoffMs
		if backoffMs > 0 {
			select {
			case <-time.After(time.Duration(backoffMs) * time.Millisecond):
			case <-r.ctx.Done():
				return r.ctx.Err()
			}
		}

		result := seckilllogic.HandleConsumedMessageWithRetry(ctx, r.svcCtx, message, envelope.Attempt, envelope.LastError)
		return r.handleConsumeResult(ctx, message, result, envelope.Attempt)
	})
}

func (r *ConsumerRunner) handleConsumeResult(ctx context.Context, message []byte, result *seckilllogic.ConsumeResult, attempt int) error {
	if result == nil {
		return nil
	}

	if result.Success {
		return nil
	}

	envelope, _ := r.svcCtx.MessageRepo.UnmarshalEnvelope(ctx, message)
	var secNum string
	if envelope != nil && envelope.Payload != nil {
		secNum = envelope.Payload.SecNum
	}

	switch result.FailureClass {
	case data.FailureClassBusinessTerminal:
		bitlog.Warn(ctx, "business terminal failure, no retry",
			bitlog.Field(bitlog.FieldAction, "mq.consume.terminal"),
			bitlog.Field(bitlog.FieldSecNum, secNum),
			bitlog.Field("failureClass", result.FailureClass.String()),
			bitlog.Field(bitlog.FieldError, result.LastError),
		)
		return nil

	case data.FailureClassPoisonMessage:
		return r.sendToDLQ(ctx, message, result.LastError)

	case data.FailureClassTransientInfra:
		if result.ShouldRetry {
			return r.sendToRetry(ctx, message, attempt, result.LastError)
		}
		return r.sendToDLQ(ctx, message, "retry exhausted: "+result.LastError)
	}

	return nil
}

func (r *ConsumerRunner) sendToRetry(ctx context.Context, message []byte, attempt int, lastError string) error {
	envelope, err := r.svcCtx.MessageRepo.UnmarshalEnvelope(ctx, message)
	if err != nil {
		return r.sendToDLQ(ctx, message, "failed to parse for retry: "+err.Error())
	}

	var secNum string
	if envelope.Payload != nil {
		secNum = envelope.Payload.SecNum
	}

	if err := r.svcCtx.MessageRepo.SendToRetry(ctx, r.svcCtx.Data, envelope, lastError); err != nil {
		bitlog.Error(ctx, "send to retry topic failed",
			bitlog.Field(bitlog.FieldAction, "mq.retry.send"),
			bitlog.Field(bitlog.FieldSecNum, secNum),
			bitlog.Field("sourceTopic", envelope.SourceTopic),
			bitlog.Field(bitlog.FieldError, err.Error()),
		)
		return err
	}

	bitlog.Info(ctx, "message sent to retry topic",
		bitlog.Field(bitlog.FieldAction, "mq.retry.sent"),
		bitlog.Field(bitlog.FieldSecNum, secNum),
		bitlog.Field("sourceTopic", envelope.SourceTopic),
		bitlog.Field("attempt", attempt+1),
	)
	return nil
}

func (r *ConsumerRunner) sendToDLQ(ctx context.Context, message []byte, lastError string) error {
	envelope, _ := r.svcCtx.MessageRepo.UnmarshalEnvelope(ctx, message)
	if envelope == nil {
		envelope = &data.SeckillEnvelope{
			Payload:    &data.SeckillMessage{},
			LastError:  lastError,
		}
	}

	var secNum string
	if envelope.Payload != nil {
		secNum = envelope.Payload.SecNum
	}

	if err := r.svcCtx.MessageRepo.SendToDLQ(ctx, r.svcCtx.Data, envelope, lastError); err != nil {
		bitlog.Error(ctx, "send to DLQ topic failed",
			bitlog.Field(bitlog.FieldAction, "mq.dlq.send"),
			bitlog.Field(bitlog.FieldSecNum, secNum),
			bitlog.Field("dlq", true),
			bitlog.Field(bitlog.FieldError, err.Error()),
		)
		return err
	}

	bitlog.Warn(ctx, "message sent to DLQ topic",
		bitlog.Field(bitlog.FieldAction, "mq.dlq.sent"),
		bitlog.Field(bitlog.FieldSecNum, secNum),
		bitlog.Field("dlq", true),
		bitlog.Field(bitlog.FieldError, lastError),
	)
	return nil
}

func (r *ConsumerRunner) Running() bool {
	if r == nil {
		return false
	}
	return r.running.Load()
}

func (r *ConsumerRunner) RetryRunning() bool {
	if r == nil {
		return false
	}
	return r.retryRunning.Load()
}

func (r *ConsumerRunner) Stop(ctx context.Context) error {
	r.once.Do(func() {
		if r.cancel != nil {
			r.cancel()
		}
		if r.svcCtx != nil && r.svcCtx.Data != nil {
			// 关闭所有主消费者
			for _, c := range r.svcCtx.Data.GetMQConsumers() {
				if c != nil {
					c.Close()
				}
			}
			// 关闭所有 retry 消费者
			for _, c := range r.svcCtx.Data.GetRetryConsumers() {
				if c != nil {
					c.Close()
				}
			}
		}
	})

	if r.done == nil {
		return nil
	}

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	select {
	case <-r.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return context.DeadlineExceeded
	}
}
