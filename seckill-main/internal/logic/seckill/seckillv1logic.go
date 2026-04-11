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
	goods, err := l.svcCtx.GoodsRepo.FindByNum(l.ctx, l.svcCtx.Data, req.GoodsNum)
	if err != nil {
		log.ErrorContextf(l.ctx, "GetGoodsInfo err %s\n", err.Error())
		return nil, err
	}

	orderNum, code, err := secKillInStore(l.ctx, l.svcCtx, goods, "", req.UserID, int(req.Num))
	if err != nil {
		log.ErrorContextf(l.ctx, "secKillInStore err %s\n", err.Error())
		return buildV1Reply("", code), nil
	}
	return buildV1Reply(orderNum, code), nil
}
