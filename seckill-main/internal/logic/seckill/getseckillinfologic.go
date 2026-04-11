package seckill

import (
	"context"
	"errors"

	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/data"
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
	l.ctx = log.WithAction(l.ctx, "GetSecKillInfo")
	l.ctx = log.WithFields(l.ctx, log.Field(log.FieldSecNum, req.SecNum))
	reply := new(pb.GetSecKillInfoReply)
	record, err := l.svcCtx.PreStockRepo.GetSecKillInfo(l.ctx, l.svcCtx.Data, req.SecNum)
	if err != nil {
		if errors.Is(err, data.ErrPreSecKillInfoNotFound) {
			reply.Code = ERR_GET_SECKILL_INFO_NOT_FOUND
			reply.Message = getErrMsg(ERR_GET_SECKILL_INFO_NOT_FOUND)
			return reply, nil
		}
		log.Error(l.ctx, "get seckill info failed",
			log.Field(log.FieldAction, "GetSecKillInfo"),
			log.Field(log.FieldSecNum, req.SecNum),
			log.Field(log.FieldError, err.Error()),
		)
		return nil, err
	}

	reply.Data = &pb.GetSecKillInfoReplyData{
		Status:   int32(record.Status),
		OrderNum: record.OrderNum,
		SecNum:   record.SecNum,
		GoodsNum: record.GoodsNum,
		Reason:   record.Reason,
	}
	return reply, nil
}
