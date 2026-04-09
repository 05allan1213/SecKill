package seckill

import (
	"context"

	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetSecKillInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSecKillInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSecKillInfoLogic {
	return &GetSecKillInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSecKillInfoLogic) GetSecKillInfo(req *pb.GetSecKillInfoRequest) (*pb.GetSecKillInfoReply, error) {
	return l.svcCtx.SecKillService.GetSecKillInfo(l.ctx, req)
}
