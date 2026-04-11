package seckill

import (
	"context"

	"github.com/BitofferHub/seckill/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ConsumeMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConsumeMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConsumeMessageLogic {
	return &ConsumeMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConsumeMessageLogic) HandleConsumedMessage(message []byte) error {
	return handleConsumedMessage(l.ctx, l.svcCtx, message)
}
