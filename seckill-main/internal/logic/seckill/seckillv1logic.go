package seckill

import (
	"context"

	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/log"
	"github.com/BitofferHub/seckill/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type SecKillV1Logic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSecKillV1Logic(ctx context.Context, svcCtx *svc.ServiceContext) *SecKillV1Logic {
	return &SecKillV1Logic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SecKillV1Logic) SecKillV1(req *pb.SecKillV1Request) (*pb.SecKillV1Reply, error) {
	l.ctx = log.WithFields(l.ctx,
		log.Field(log.FieldAction, "seckill.v1"),
		log.Field(log.FieldUserID, req.UserID),
		log.Field(log.FieldGoodsNum, req.GoodsNum),
	)
	goods, err := l.svcCtx.GoodsRepo.FindByNum(l.ctx, l.svcCtx.Data, req.GoodsNum)
	if err != nil {
		log.Error(l.ctx, "load goods failed", log.Field(log.FieldError, err.Error()))
		return buildV1Reply("", ERR_FIND_GOODS_FAILED), nil
	}

	orderNum, code, err := secKillInStore(l.ctx, l.svcCtx, goods, "", req.UserID, int(req.Num))
	if err != nil {
		log.Error(l.ctx, "seckill v1 store failed", log.Field(log.FieldError, err.Error()))
		return buildV1Reply("", code), nil
	}
	return buildV1Reply(orderNum, code), nil
}
