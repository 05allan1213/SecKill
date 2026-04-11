package seckill

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BitofferHub/pkg/constant"
	"github.com/BitofferHub/pkg/utils"
	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/data"
	"github.com/BitofferHub/seckill/internal/log"
	"github.com/BitofferHub/seckill/internal/svc"
	"gorm.io/gorm"
)

func HandleConsumedMessage(ctx context.Context, svcCtx *svc.ServiceContext, message []byte) error {
	if ctx == nil {
		return errors.New("nil consume context")
	}

	skMsg, err := svcCtx.MessageRepo.UnmarshalSecKillMsg(ctx, svcCtx.Data, message)
	if err != nil {
		log.ErrorContextf(ctx, "UnmarshalSecKillMsg err %s", err.Error())
		return err
	}
	if skMsg.TraceID != "" {
		ctx = context.WithValue(ctx, constant.TraceID, skMsg.TraceID)
	}

	log.InfoContextf(ctx, "message is: %s", string(message))
	orderNum, _, err := secKillInStore(ctx, svcCtx, skMsg.Goods, skMsg.SecNum, skMsg.UserID, skMsg.Num)
	if err != nil {
		log.ErrorContextf(ctx, "secKillInStore err %s", err.Error())
		return err
	}
	record, err := svcCtx.PreStockRepo.GetSecKillInfo(ctx, svcCtx.Data, skMsg.SecNum)
	if err != nil {
		log.ErrorContextf(ctx, "GetSecKillInfo err %s", err.Error())
		return err
	}
	record.OrderNum = orderNum
	record.Status = int(data.SK_STATUS_BEFORE_PAY)
	record.ModifyTime = time.Now()
	if _, err := svcCtx.PreStockRepo.SetSuccessInPreSecKill(ctx, svcCtx.Data, skMsg.UserID, skMsg.Goods.ID, skMsg.SecNum, record); err != nil {
		log.ErrorContextf(ctx, "SetSuccessInPreSecKill err %s", err.Error())
		return err
	}
	return nil
}

func secKillInStore(ctx context.Context, svcCtx *svc.ServiceContext, goods *data.Goods, secNum string, userID int64, num int) (string, int, error) {
	orderNum := utils.NewUuid()
	if secNum == "" {
		secNum = utils.NewUuid()
	}

	code := SUCCESS
	err := svcCtx.Data.RunInTx(ctx, func(txData *data.Data) error {
		var (
			rowAffected    int64
			globalQuota    = new(data.Quota)
			userQuota      = new(data.UserQuota)
			err            error
			userKilledNum  int64
			userQuotaNum   int64
			userQuotaExist = true
		)

		userQuota, err = svcCtx.UserQuotaRepo.FindUserGoodsQuota(ctx, txData, userID, goods.ID)
		if err != nil {
			if err.Error() == gorm.ErrRecordNotFound.Error() {
				userQuotaExist = false
			} else {
				log.ErrorContextf(ctx, "FindUserGoodsQuota err %s\n", err.Error())
				code = ERR_FIND_USER_QUOTA_FAILED
				return err
			}
		} else {
			userQuotaNum = userQuota.Num
			userKilledNum = userQuota.KilledNum
		}

		if userQuotaNum == 0 {
			globalQuota, err = svcCtx.QuotaRepo.FindByGoodsID(ctx, txData, goods.ID)
			if err != nil {
				if err.Error() != gorm.ErrRecordNotFound.Error() {
					log.ErrorContextf(ctx, "FindByGoodsID err %s\n", err.Error())
					code = ERR_FIND_GOODS_FAILED
					return err
				}
			} else {
				userQuotaNum = globalQuota.Num
			}
		}

		leftQuota := userQuotaNum - userKilledNum
		if int(leftQuota) < num {
			log.InfoContextf(ctx, "user %d, goods %d, quota limit %d", userID, goods.ID, leftQuota)
			code = ERR_USER_QUOTA_NOT_ENOUGH
			return nil
		}

		if !userQuotaExist {
			_, err = svcCtx.UserQuotaRepo.Save(ctx, txData, &data.UserQuota{
				UserID:    userID,
				GoodsID:   goods.ID,
				KilledNum: int64(num),
			})
			if err != nil {
				log.ErrorContextf(ctx, "CreateUserQuota err %s\n", err.Error())
				code = ERR_CREATER_USER_QUOTA_FAILED
				return err
			}
		} else {
			_, err = svcCtx.UserQuotaRepo.IncrKilledNum(ctx, txData, userID, goods.ID, int64(num))
			if err != nil {
				log.ErrorContextf(ctx, "IncrKilledNum err %s\n", err.Error())
				code = ERR_RECORD_USER_KILLED_NUM_FAILED
				return err
			}
		}

		rowAffected, err = svcCtx.StockRepo.DescStock(ctx, txData, goods.ID, int32(num))
		if err != nil {
			log.ErrorContextf(ctx, "Desc stock err %s\n", err.Error())
			code = ERR_DESC_STOCK_FAILED
			return err
		}
		if rowAffected == 0 {
			log.InfoContextf(ctx, "goods %d stock not enough", goods.ID)
			code = ERR_GOODS_STOCK_NOT_ENOUGH
			return nil
		}

		_, err = svcCtx.OrderRepo.Save(ctx, txData, &data.Order{
			OrderNum: orderNum,
			GoodsID:  goods.ID,
			Price:    goods.Price,
			Buyer:    userID,
			Seller:   goods.Seller,
			Status:   int(data.SK_STATUS_BEFORE_PAY),
		})
		if err != nil {
			log.ErrorContextf(ctx, "create order err %s\n", err.Error())
			code = ERR_CREATE_ORDER_FAILED
			return err
		}

		_, err = svcCtx.RecordRepo.Save(ctx, txData, &data.SecKillRecord{
			UserID:   userID,
			GoodsID:  goods.ID,
			SecNum:   secNum,
			OrderNum: orderNum,
			Price:    goods.Price,
			Status:   int(data.SK_STATUS_BEFORE_PAY),
		})
		if err != nil {
			log.ErrorContextf(ctx, "create seckill record err %s\n", err.Error())
			code = ERR_CREATE_SECKILL_RECORD_FAILED
			return err
		}
		return nil
	})
	if err != nil {
		return orderNum, code, err
	}
	return orderNum, code, nil
}

func convertDataGoodsToPbGoods(goods *data.Goods, info *pb.GoodInfo) {
	info.GoodsNum = goods.GoodsNum
	info.GoodsName = goods.GoodsName
	info.Price = float32(goods.Price)
	info.PicUrl = goods.PicUrl
	info.Seller = goods.Seller
}

func traceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	traceID, _ := ctx.Value(constant.TraceID).(string)
	return traceID
}

func buildV1Reply(orderNum string, code int) *pb.SecKillV1Reply {
	reply := &pb.SecKillV1Reply{
		Code:    int32(code),
		Message: getErrMsg(code),
	}
	if orderNum != "" {
		reply.Data = &pb.SecKillV1ReplyData{OrderNum: orderNum}
	}
	if code == SUCCESS {
		reply.Message = ""
	}
	return reply
}

func duplicateSecKillMessage(err error, secNum string) string {
	if err == nil || err.Error() != data.SecKillErrSecKilling.Error() {
		return ""
	}
	return err.Error() + ":" + fmt.Sprintf("%s", secNum)
}
