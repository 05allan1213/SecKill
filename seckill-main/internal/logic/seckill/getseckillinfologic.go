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
	var pendingRecord *data.PreSecKillRecord

	record, err := l.svcCtx.PreStockRepo.GetSecKillInfo(l.ctx, l.svcCtx.Data, req.SecNum)
	if err == nil && record.Status != int(data.SK_STATUS_BEFORE_ORDER) {
		reply.Data = &pb.GetSecKillInfoReplyData{
			Status:   int32(record.Status),
			OrderNum: record.OrderNum,
			SecNum:   record.SecNum,
			GoodsNum: record.GoodsNum,
			Reason:   record.Reason,
		}
		return reply, nil
	}
	if err == nil && record.Status == int(data.SK_STATUS_BEFORE_ORDER) {
		pendingRecord = record
	}

	if err != nil && !errors.Is(err, data.ErrPreSecKillInfoNotFound) {
		log.Error(l.ctx, "get seckill info from redis failed",
			log.Field(log.FieldAction, "GetSecKillInfo.redis"),
			log.Field(log.FieldSecNum, req.SecNum),
			log.Field(log.FieldError, err.Error()),
		)
	}

	asyncResult, err := l.svcCtx.AsyncResultRepo.FindBySecNum(l.ctx, l.svcCtx.Data.GetDB(), req.SecNum)
	if err != nil {
		log.Error(l.ctx, "get seckill info from async result table failed",
			log.Field(log.FieldAction, "GetSecKillInfo.db_async"),
			log.Field(log.FieldSecNum, req.SecNum),
			log.Field(log.FieldError, err.Error()),
		)
		return nil, dependencyUnavailableError("seckill query unavailable")
	}

	if asyncResult != nil {
		reply.Data = &pb.GetSecKillInfoReplyData{
			Status:   int32(asyncResult.Status),
			OrderNum: asyncResult.OrderNum,
			SecNum:   asyncResult.SecNum,
			GoodsNum: asyncResult.GoodsNum,
			Reason:   asyncResult.Reason,
		}
		return reply, nil
	}

	historicalRecord, err := l.svcCtx.RecordRepo.FindBySecNum(l.ctx, l.svcCtx.Data, req.SecNum)
	if err == nil && historicalRecord != nil {
		reply.Data = &pb.GetSecKillInfoReplyData{
			Status:   int32(historicalRecord.Status),
			OrderNum: historicalRecord.OrderNum,
			SecNum:   historicalRecord.SecNum,
		}
		return reply, nil
	}

	if pendingRecord != nil {
		reply.Data = &pb.GetSecKillInfoReplyData{
			Status:   int32(pendingRecord.Status),
			OrderNum: pendingRecord.OrderNum,
			SecNum:   pendingRecord.SecNum,
			GoodsNum: pendingRecord.GoodsNum,
			Reason:   pendingRecord.Reason,
		}
		return reply, nil
	}

	reply.Code = ERR_GET_SECKILL_INFO_NOT_FOUND
	reply.Message = getErrMsg(ERR_GET_SECKILL_INFO_NOT_FOUND)
	return reply, nil
}
