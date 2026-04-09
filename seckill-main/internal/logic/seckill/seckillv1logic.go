package seckill

import (
	"context"

	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
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
	return l.svcCtx.SecKillService.SecKillV1(l.ctx, req)
}
