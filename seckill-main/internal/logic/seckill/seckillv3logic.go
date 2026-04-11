package seckill

import (
	"context"
	"time"

	"github.com/BitofferHub/pkg/utils"
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

	goods, err := l.svcCtx.GoodsRepo.FindByNum(l.ctx, l.svcCtx.Data, req.GoodsNum)
	if err != nil {
		log.Error(l.ctx, "load goods failed", log.Field(log.FieldError, err.Error()))
		return nil, goodsLookupError(err)
	}

	if !l.svcCtx.Data.IsRedisAvailable(l.ctx, l.svcCtx.Config.Fallback.TimeoutMs) {
		return l.handleFallback(goods, req)
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
			return nil, dependencyUnavailableError("pre-seckill rollback unavailable")
		}
		return nil, dependencyUnavailableError("seckill message queue unavailable")
	}

	return buildV3Reply(secNum, SUCCESS), nil
}

func (l *SecKillV3Logic) handleFallback(goods *data.Goods, req *pb.SecKillV3Request) (*pb.SecKillV3Reply, error) {
	if !l.svcCtx.Config.Fallback.Enabled {
		log.Error(l.ctx, "redis unavailable and fallback disabled")
		return nil, dependencyUnavailableError("redis unavailable and fallback disabled")
	}

	log.Warn(l.ctx, "redis unavailable, falling back to v1 database transaction",
		log.Field(log.FieldAction, "seckill.v3.fallback"),
		log.Field(log.FieldGoodsNum, req.GoodsNum),
	)

	secNum := utils.NewUuid()
	orderNum, code, err := secKillInStore(l.ctx, l.svcCtx, goods, secNum, req.UserID, int(req.Num))
	if err != nil {
		log.Error(l.ctx, "seckill v3 fallback failed",
			log.Field(log.FieldError, err.Error()),
			log.Field("resultCode", code),
		)
		return nil, dependencyUnavailableError("seckill storage unavailable")
	}

	reply := buildV3Reply(secNum, code)
	if code == SUCCESS {
		reply.Message = "success (fallback)"
		log.Info(l.ctx, "seckill v3 fallback succeeded",
			log.Field(log.FieldSecNum, secNum),
			log.Field(log.FieldOrderNum, orderNum),
		)
	}
	return reply, nil
}
