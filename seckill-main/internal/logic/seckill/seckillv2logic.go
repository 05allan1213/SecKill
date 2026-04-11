package seckill

import (
	"context"

	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type SecKillV2Logic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSecKillV2Logic(ctx context.Context, svcCtx *svc.ServiceContext) *SecKillV2Logic {
	return &SecKillV2Logic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SecKillV2Logic) SecKillV2(req *pb.SecKillV2Request) (*pb.SecKillV2Reply, error) {
	return secKillV2(l.ctx, l.svcCtx, req)
}
