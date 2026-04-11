package seckill

import (
	"context"
	"fmt"
	"time"

	"github.com/BitofferHub/pkg/utils"
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
	reply := new(pb.SecKillV2Reply)

	goods, err := l.svcCtx.GoodsRepo.GetGoodsInfoByNumWithCache(l.ctx, l.svcCtx.Data, req.GoodsNum)
	if err != nil {
		log.InfoContextf(l.ctx, "GetGoodsInfo err %s\n", err.Error())
		return nil, err
	}

	secNum := utils.NewUuid()
	now := time.Now()
	record := data.PreSecKillRecord{
		SecNum:     secNum,
		UserID:     req.UserID,
		GoodsID:    goods.ID,
		OrderNum:   "",
		Price:      goods.Price,
		Status:     int(data.SK_STATUS_BEFORE_ORDER),
		CreateTime: now,
		ModifyTime: now,
	}

	alreadySecNum, err := l.svcCtx.PreStockRepo.PreDescStock(l.ctx, l.svcCtx.Data, req.UserID, goods.ID, req.Num, secNum, &record)
	if err != nil {
		if err.Error() == data.SecKillErrSecKilling.Error() {
			reply.Message = err.Error() + ":" + fmt.Sprintf("%s", alreadySecNum)
			return reply, nil
		}
		log.ErrorContextf(l.ctx, "Desc stock err %s\n", err.Error())
		reply.Code = -10100
		reply.Message = err.Error()
		return reply, err
	}

	orderNum, code, err := secKillInStore(l.ctx, l.svcCtx, goods, secNum, req.UserID, int(req.Num))
	if err != nil {
		return reply, err
	}

	record.OrderNum = orderNum
	record.Status = int(data.SK_STATUS_BEFORE_PAY)
	record.ModifyTime = time.Now()
	if _, err := l.svcCtx.PreStockRepo.SetSuccessInPreSecKill(l.ctx, l.svcCtx.Data, req.UserID, goods.ID, secNum, &record); err != nil {
		log.ErrorContextf(l.ctx, "set pre seckill success err %s\n", err.Error())
		return reply, err
	}

	reply.Data = &pb.SecKillV2ReplyData{OrderNum: orderNum}
	reply.Code = int32(code)
	return reply, nil
}
