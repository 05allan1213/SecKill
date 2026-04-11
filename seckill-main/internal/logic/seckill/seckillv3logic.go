package seckill

import (
	"context"
	"time"

	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/data"
	"github.com/BitofferHub/seckill/internal/log"
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
	l.ctx = log.WithFields(l.ctx,
		log.Field(log.FieldAction, "seckill.v3"),
		log.Field(log.FieldUserID, req.UserID),
		log.Field(log.FieldGoodsNum, req.GoodsNum),
	)

	goods, err := l.svcCtx.GoodsRepo.GetGoodsInfoByNumWithCache(l.ctx, l.svcCtx.Data, req.GoodsNum)
	if err != nil {
		log.Error(l.ctx, "load goods failed", log.Field(log.FieldError, err.Error()))
		return buildV3Reply("", ERR_FIND_GOODS_FAILED), nil
	}

	record := newPreSecKillRecord(goods, req.UserID, "")
	secNum := record.SecNum
	alreadySecNum, err := l.svcCtx.PreStockRepo.PreDescStock(l.ctx, l.svcCtx.Data, req.UserID, goods.ID, req.Num, secNum, record)
	if err != nil {
		code := codeFromPreDescError(err)
		if code == ERR_DUPLICATE_SECKILL {
			log.WarnEvery(l.ctx, "seckill.v3.duplicate", 2*time.Second, "duplicate seckill request", log.Field(log.FieldSecNum, alreadySecNum))
			return buildV3Reply(alreadySecNum, code), nil
		}
		log.WarnEvery(l.ctx, "seckill.v3.pre_desc_failed", 2*time.Second, "pre-desc stock failed", log.Field(log.FieldError, err.Error()))
		return buildV3Reply("", code), nil
	}

	msg := &data.SeckillMessage{
		TraceID: traceIDFromContext(l.ctx),
		Goods:   goods,
		SecNum:  secNum,
		UserID:  req.UserID,
		Num:     int(req.Num),
	}
	if err := l.svcCtx.MessageRepo.SendSecKillMsg(l.ctx, l.svcCtx.Data, msg); err != nil {
		log.Error(l.ctx, "send seckill message failed", log.Field(log.FieldSecNum, secNum), log.Field(log.FieldError, err.Error()))
		if failErr := markPreSecKillFailed(l.ctx, l.svcCtx, goods, req.UserID, req.Num, secNum, ERR_SEND_SECKILL_MSG_FAILED, ""); failErr != nil {
			return nil, failErr
		}
		return buildV3Reply(secNum, ERR_SEND_SECKILL_MSG_FAILED), nil
	}

	return buildV3Reply(secNum, SUCCESS), nil
}
