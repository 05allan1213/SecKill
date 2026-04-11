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
	reply := new(pb.SecKillV3Reply)

	goods, err := l.svcCtx.GoodsRepo.FindByNum(l.ctx, l.svcCtx.Data, req.GoodsNum)
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
		return nil, err
	}

	msg := &data.SeckillMessage{
		TraceID: traceIDFromContext(l.ctx),
		Goods:   goods,
		SecNum:  secNum,
		UserID:  req.UserID,
		Num:     int(req.Num),
	}
	if err := l.svcCtx.MessageRepo.SendSecKillMsg(l.ctx, l.svcCtx.Data, msg); err != nil {
		log.ErrorContextf(l.ctx, "send seckill mq msg err %s\n", err.Error())
		return nil, err
	}

	reply.Data = &pb.SecKillV3ReplyData{SecNum: secNum}
	return reply, nil
}
