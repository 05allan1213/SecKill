package service

import (
	"context"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/seckill/internal/biz"
	"github.com/BitofferHub/seckill/internal/data"
	"time"
)

func (s *SecKillService) Consume() {
	go func() {
		defer func() {
			if err := recover(); err == nil {
				go s.consume()
			}
		}()
		s.consume()
	}()
}

func (s *SecKillService) consume() {
	consumer := data.GetData().GetMQConsumer()
	consumer.ConsumeMessages(func(message []byte) error {
		ctx := context.Background()
		log.InfoContextf(ctx, "message is: %s", string(message))
		dt := biz.NewData(data.GetData().GetDB(), data.GetData().GetCache(), data.GetData().GetMQProducer(), data.GetData().GetMQConsumer())
		skMsg, err := s.msgUc.UnmarshalSecKillMsg(ctx, dt, message)
		if err != nil {
			log.ErrorContextf(ctx, "UnmarshalSecKillMsg err %s", err.Error())
			return err
		}
		orderNum, _, err := s.secKillInStore(ctx,
			skMsg.Goods, skMsg.SecNum, skMsg.UserID, skMsg.Num)
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
		s.preStockUc.SetSuccessInPreSecKill(ctx, dt, skMsg.UserID,
			skMsg.Goods.ID, skMsg.SecNum, record)
		return nil
	})
}
