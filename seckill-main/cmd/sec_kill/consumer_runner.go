package main

import (
	"context"
	"sync"
	"time"

	bitlog "github.com/BitofferHub/seckill/internal/log"
	seckilllogic "github.com/BitofferHub/seckill/internal/logic/seckill"
	"github.com/BitofferHub/seckill/internal/svc"
)

type ConsumerRunner struct {
	svcCtx *svc.ServiceContext
	done   chan struct{}
	once   sync.Once
	ctx    context.Context
	cancel context.CancelFunc
}

func NewConsumerRunner(svcCtx *svc.ServiceContext) *ConsumerRunner {
	return &ConsumerRunner{svcCtx: svcCtx}
}

func (r *ConsumerRunner) Start() error {
	r.done = make(chan struct{})
	r.ctx, r.cancel = context.WithCancel(context.Background())
	go func() {
		defer close(r.done)
		defer func() {
			if recovered := recover(); recovered != nil {
				bitlog.Errorf("seckill consumer panic: %v", recovered)
			}
		}()

		consumer := r.svcCtx.Data.GetMQConsumer()
		if consumer == nil {
			return
		}
		consumer.ConsumeMessages(r.ctx, func(ctx context.Context, message []byte) error {
			return seckilllogic.HandleConsumedMessage(ctx, r.svcCtx, message)
		})
	}()

	return nil
}

func (r *ConsumerRunner) Stop(ctx context.Context) error {
	r.once.Do(func() {
		if r.cancel != nil {
			r.cancel()
		}
		if r.svcCtx != nil && r.svcCtx.Data != nil && r.svcCtx.Data.GetMQConsumer() != nil {
			r.svcCtx.Data.GetMQConsumer().Close()
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
