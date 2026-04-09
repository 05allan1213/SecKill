package service

import (
	"context"
	"time"

	"github.com/BitofferHub/pkg/constant"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/seckill/internal/biz"
	"github.com/BitofferHub/seckill/internal/data"
)

func (s *SecKillService) HandleConsumedMessage(ctx context.Context, message []byte) error {
	if ctx == nil {
		ctx = context.Background()
	}

	dt := biz.NewData(data.GetData().GetDB(), data.GetData().GetCache(), data.GetData().GetMQProducer(), data.GetData().GetMQConsumer())
	skMsg, err := s.msgUc.UnmarshalSecKillMsg(ctx, dt, message)
	if err != nil {
		log.ErrorContextf(ctx, "UnmarshalSecKillMsg err %s", err.Error())
		return err
	}
	if skMsg.TraceID != "" {
		ctx = context.WithValue(ctx, constant.TraceID, skMsg.TraceID)
	}

	log.InfoContextf(ctx, "message is: %s", string(message))
	orderNum, _, err := s.secKillInStore(ctx, skMsg.Goods, skMsg.SecNum, skMsg.UserID, skMsg.Num)
	if err != nil {
		log.ErrorContextf(ctx, "secKillInStore err %s", err.Error())
		return err
	}
	record, err := s.preStockUc.GetSecKillInfo(ctx, dt, skMsg.SecNum)
	if err != nil {
		log.ErrorContextf(ctx, "GetSecKillInfo err %s", err.Error())
		return err
	}
	record.OrderNum = orderNum
	record.Status = int(biz.SK_STATUS_BEFORE_PAY)
	record.ModifyTime = time.Now()
	s.preStockUc.SetSuccessInPreSecKill(ctx, dt, skMsg.UserID, skMsg.Goods.ID, skMsg.SecNum, record)
	return nil
}
