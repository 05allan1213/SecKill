package seckill

import (
	"context"

	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type SecKillV3Logic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSecKillV3Logic(ctx context.Context, svcCtx *svc.ServiceContext) *SecKillV3Logic {
	return &SecKillV3Logic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SecKillV3Logic) SecKillV3(req *pb.SecKillV3Request) (*pb.SecKillV3Reply, error) {
	return secKillV3(l.ctx, l.svcCtx, req)
}
