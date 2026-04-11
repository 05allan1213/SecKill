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
	l.ctx = log.WithFields(l.ctx,
		log.Field(log.FieldAction, "seckill.v2"),
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
	_, err = l.svcCtx.PreStockRepo.PreDescStock(l.ctx, l.svcCtx.Data, req.UserID, goods.ID, req.Num, secNum, record)
	if err != nil {
		log.WarnEvery(l.ctx, "seckill.v2.pre_desc_failed", 2*time.Second, "pre-desc stock failed", log.Field(log.FieldError, err.Error()))
		return buildV2Reply("", codeFromPreDescError(err)), nil
	}

	orderNum, code, err := secKillInStore(l.ctx, l.svcCtx, goods, secNum, req.UserID, int(req.Num))
	if err != nil || code != SUCCESS {
		if code == SUCCESS {
			code = ERR_CREATE_ORDER_FAILED
		}
		if failErr := markPreSecKillFailed(l.ctx, l.svcCtx, goods, req.UserID, req.Num, secNum, code, ""); failErr != nil {
			return nil, dependencyUnavailableError("pre-seckill rollback unavailable")
		}
		if err != nil {
			log.Error(l.ctx, "seckill v2 store failed",
				log.Field(log.FieldError, err.Error()),
				log.Field("resultCode", code),
			)
			return nil, dependencyUnavailableError("seckill storage unavailable")
		}
		return buildV2Reply("", code), nil
	}

	if err := markPreSecKillSuccess(l.ctx, l.svcCtx, goods, req.UserID, secNum, orderNum); err != nil {
		return nil, dependencyUnavailableError("pre-seckill success write unavailable")
	}

	return buildV2Reply(orderNum, code), nil
}

func (l *SecKillV2Logic) handleFallback(goods *data.Goods, req *pb.SecKillV2Request) (*pb.SecKillV2Reply, error) {
	if !l.svcCtx.Config.Fallback.Enabled {
		log.Error(l.ctx, "redis unavailable and fallback disabled")
		return nil, dependencyUnavailableError("redis unavailable and fallback disabled")
	}

	log.Warn(l.ctx, "redis unavailable, falling back to v1 database transaction",
		log.Field(log.FieldAction, "seckill.v2.fallback"),
		log.Field(log.FieldGoodsNum, req.GoodsNum),
	)

	orderNum, code, err := secKillInStore(l.ctx, l.svcCtx, goods, "", req.UserID, int(req.Num))
	if err != nil {
		log.Error(l.ctx, "seckill v2 fallback failed",
			log.Field(log.FieldError, err.Error()),
			log.Field("resultCode", code),
		)
		return nil, dependencyUnavailableError("seckill storage unavailable")
	}

	reply := buildV2Reply(orderNum, code)
	if code == SUCCESS {
		reply.Message = "success (fallback)"
	}
	return reply, nil
}
