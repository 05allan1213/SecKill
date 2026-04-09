package service

import (
	"context"
	"errors"
	"time"

	"github.com/BitofferHub/pkg/constant"
	"github.com/BitofferHub/seckill/internal/biz"
	"github.com/BitofferHub/seckill/internal/log"
)

func (s *SecKillService) HandleConsumedMessage(ctx context.Context, message []byte) error {
	if ctx == nil {
		return errors.New("nil consume context")
	}

	skMsg, err := s.msgUc.UnmarshalSecKillMsg(ctx, s.data, message)
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
	record, err := s.preStockUc.GetSecKillInfo(ctx, s.data, skMsg.SecNum)
	if err != nil {
		log.ErrorContextf(ctx, "GetSecKillInfo err %s", err.Error())
		return err
	}
	record.OrderNum = orderNum
	record.Status = int(biz.SK_STATUS_BEFORE_PAY)
	record.ModifyTime = time.Now()
	if _, err := s.preStockUc.SetSuccessInPreSecKill(ctx, s.data, skMsg.UserID, skMsg.Goods.ID, skMsg.SecNum, record); err != nil {
		log.ErrorContextf(ctx, "SetSuccessInPreSecKill err %s", err.Error())
		return err
	}
	return nil
}
