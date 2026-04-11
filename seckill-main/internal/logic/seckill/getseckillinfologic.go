package seckill

import (
	"context"

	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/log"
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
	reply := new(pb.GetSecKillInfoReply)
	record, err := l.svcCtx.PreStockRepo.GetSecKillInfo(l.ctx, l.svcCtx.Data, req.SecNum)
	if err != nil {
		log.ErrorContextf(l.ctx, "get secinfo by secnum err %s\n", err.Error())
		return nil, err
	}

	reply.Data = &pb.GetSecKillInfoReplyData{
		Status:   int32(record.Status),
		OrderNum: record.OrderNum,
		SecNum:   record.SecNum,
	}
	return reply, nil
}
