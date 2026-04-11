package seckill

import (
	"context"
	"errors"
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
		log.Error(ctx, "mq message unmarshal failed",
			log.Field(log.FieldAction, "seckill.mq.consume"),
			log.Field(log.FieldError, err.Error()),
		)
		return err
	}
	if skMsg.TraceID != "" {
		ctx = context.WithValue(ctx, constant.TraceID, skMsg.TraceID)
		ctx = log.WithTrace(ctx, skMsg.TraceID)
	}
	ctx = log.WithFields(ctx,
		log.Field(log.FieldUserID, skMsg.UserID),
		log.Field(log.FieldGoodsID, skMsg.Goods.ID),
		log.Field(log.FieldGoodsNum, skMsg.Goods.GoodsNum),
		log.Field(log.FieldSecNum, skMsg.SecNum),
	)

	log.InfoEvery(ctx, "seckill.mq.consume.start", 2*time.Second, "consume seckill message", log.Field(log.FieldAction, "seckill.mq.consume"))
	orderNum, code, err := secKillInStore(ctx, svcCtx, skMsg.Goods, skMsg.SecNum, skMsg.UserID, skMsg.Num)
	if err != nil || code != SUCCESS {
		if code == SUCCESS {
			code = ERR_CREATE_ORDER_FAILED
		}
		if failErr := markPreSecKillFailed(ctx, svcCtx, skMsg.Goods, skMsg.UserID, int32(skMsg.Num), skMsg.SecNum, code, ""); failErr != nil {
			return failErr
		}
		if err != nil {
			log.Error(ctx, "seckill store flow failed",
				log.Field(log.FieldAction, "seckill.mq.consume"),
				log.Field(log.FieldError, err.Error()),
				log.Field("resultCode", code),
			)
		} else {
			log.WarnEvery(ctx, "seckill.mq.consume.reject", 2*time.Second, "seckill store rejected",
				log.Field(log.FieldAction, "seckill.mq.consume"),
				log.Field("resultCode", code),
			)
		}
		return nil
	}
	if err := markPreSecKillSuccess(ctx, svcCtx, skMsg.Goods, skMsg.UserID, skMsg.SecNum, orderNum); err != nil {
		return err
	}
	log.InfoEvery(ctx, "seckill.mq.consume.finish", 2*time.Second, "consume seckill message finished",
		log.Field(log.FieldAction, "seckill.mq.consume"),
		log.Field(log.FieldOrderNum, orderNum),
	)
	return nil
}

func secKillInStore(ctx context.Context, svcCtx *svc.ServiceContext, goods *data.Goods, secNum string, userID int64, num int) (string, int, error) {
	orderNum := utils.NewUuid()
	if secNum == "" {
		secNum = utils.NewUuid()
	}
	ctx = log.WithFields(ctx,
		log.Field(log.FieldAction, "seckill.store"),
		log.Field(log.FieldUserID, userID),
		log.Field(log.FieldGoodsID, goods.ID),
		log.Field(log.FieldGoodsNum, goods.GoodsNum),
		log.Field(log.FieldSecNum, secNum),
		log.Field("num", num),
	)

	code := SUCCESS
	err := svcCtx.Data.RunInTx(ctx, func(txData *data.Data) error {
		var (
			rowAffected    int64
			userQuota      = new(data.UserQuota)
			err            error
			userKilledNum  int64
			userQuotaNum   int64
			userQuotaExist = true
			quotaEnabled   bool
		)

		userQuota, err = svcCtx.UserQuotaRepo.FindUserGoodsQuota(ctx, txData, userID, goods.ID)
		if err != nil {
			if err.Error() == gorm.ErrRecordNotFound.Error() {
				userQuotaExist = false
			} else {
				log.Error(ctx, "find user quota failed", log.Field(log.FieldError, err.Error()))
				code = ERR_FIND_USER_QUOTA_FAILED
				return err
			}
		} else {
			userQuotaNum = userQuota.Num
			userKilledNum = userQuota.KilledNum
			if userQuotaNum > 0 {
				quotaEnabled = true
			}
		}

		if !quotaEnabled {
			globalQuota, err := svcCtx.QuotaRepo.FindByGoodsID(ctx, txData, goods.ID)
			if err != nil {
				if err.Error() != gorm.ErrRecordNotFound.Error() {
					log.Error(ctx, "find goods quota failed", log.Field(log.FieldError, err.Error()))
					code = ERR_FIND_GOODS_FAILED
					return err
				}
			} else if globalQuota.Num > 0 {
				userQuotaNum = globalQuota.Num
				quotaEnabled = true
			}
		}

		if quotaEnabled {
			leftQuota := userQuotaNum - userKilledNum
			if int(leftQuota) < num {
				log.WarnEvery(ctx, "seckill.store.quota_not_enough", 2*time.Second, "user quota not enough", log.Field("leftQuota", leftQuota))
				code = ERR_USER_QUOTA_NOT_ENOUGH
				return nil
			}
		}

		if !userQuotaExist {
			_, err = svcCtx.UserQuotaRepo.Save(ctx, txData, &data.UserQuota{
				UserID:    userID,
				GoodsID:   goods.ID,
				Num:       userQuotaNum,
				KilledNum: int64(num),
			})
			if err != nil {
				log.Error(ctx, "create user quota failed", log.Field(log.FieldError, err.Error()))
				code = ERR_CREATER_USER_QUOTA_FAILED
				return err
			}
		} else {
			_, err = svcCtx.UserQuotaRepo.IncrKilledNum(ctx, txData, userID, goods.ID, int64(num))
			if err != nil {
				log.Error(ctx, "increase killed num failed", log.Field(log.FieldError, err.Error()))
				code = ERR_RECORD_USER_KILLED_NUM_FAILED
				return err
			}
		}

		rowAffected, err = svcCtx.StockRepo.DescStock(ctx, txData, goods.ID, int32(num))
		if err != nil {
			log.Error(ctx, "decrease stock failed", log.Field(log.FieldError, err.Error()))
			code = ERR_DESC_STOCK_FAILED
			return err
		}
		if rowAffected == 0 {
			log.WarnEvery(ctx, "seckill.store.stock_not_enough", 2*time.Second, "goods stock not enough")
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
			log.Error(ctx, "create order failed", log.Field(log.FieldOrderNum, orderNum), log.Field(log.FieldError, err.Error()))
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
			log.Error(ctx, "create seckill record failed", log.Field(log.FieldOrderNum, orderNum), log.Field(log.FieldError, err.Error()))
			code = ERR_CREATE_SECKILL_RECORD_FAILED
			return err
		}
		return nil
	})
	if err != nil {
		return orderNum, code, err
	}
	if code == SUCCESS {
		log.InfoEvery(ctx, "seckill.store.succeeded", 2*time.Second, "seckill store succeeded", log.Field(log.FieldOrderNum, orderNum))
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

func newPreSecKillRecord(goods *data.Goods, userID int64, secNum string) *data.PreSecKillRecord {
	if secNum == "" {
		secNum = utils.NewUuid()
	}
	now := time.Now()
	return &data.PreSecKillRecord{
		SecNum:     secNum,
		UserID:     userID,
		GoodsID:    goods.ID,
		GoodsNum:   goods.GoodsNum,
		OrderNum:   "",
		Price:      goods.Price,
		Status:     int(data.SK_STATUS_BEFORE_ORDER),
		Reason:     "",
		CreateTime: now,
		ModifyTime: now,
	}
}

func loadPreSecKillRecord(ctx context.Context, svcCtx *svc.ServiceContext, goods *data.Goods, userID int64, secNum string) (*data.PreSecKillRecord, error) {
	record, err := svcCtx.PreStockRepo.GetSecKillInfo(ctx, svcCtx.Data, secNum)
	if err == nil {
		return record, nil
	}
	if errors.Is(err, data.ErrPreSecKillInfoNotFound) {
		return newPreSecKillRecord(goods, userID, secNum), nil
	}
	log.Error(ctx, "load pre-seckill info failed",
		log.Field(log.FieldAction, "seckill.pre_load"),
		log.Field(log.FieldSecNum, secNum),
		log.Field(log.FieldError, err.Error()),
	)
	return nil, err
}

func markPreSecKillSuccess(ctx context.Context, svcCtx *svc.ServiceContext, goods *data.Goods, userID int64, secNum, orderNum string) error {
	record, err := loadPreSecKillRecord(ctx, svcCtx, goods, userID, secNum)
	if err != nil {
		return err
	}
	record.OrderNum = orderNum
	record.Status = int(data.SK_STATUS_BEFORE_PAY)
	record.Reason = ""
	record.ModifyTime = time.Now()
	if _, err := svcCtx.PreStockRepo.SetSuccessInPreSecKill(ctx, svcCtx.Data, userID, goods.ID, secNum, record); err != nil {
		log.Error(ctx, "update pre-seckill success failed",
			log.Field(log.FieldAction, "seckill.set_success"),
			log.Field(log.FieldOrderNum, orderNum),
			log.Field(log.FieldError, err.Error()),
		)
		return err
	}
	return nil
}

func markPreSecKillFailed(ctx context.Context, svcCtx *svc.ServiceContext, goods *data.Goods, userID int64, num int32, secNum string, code int, reason string) error {
	record, err := loadPreSecKillRecord(ctx, svcCtx, goods, userID, secNum)
	if err != nil {
		return err
	}
	record.Status = int(data.SK_STATUS_FAILED)
	record.OrderNum = ""
	record.Reason = failureReason(code, reason)
	record.ModifyTime = time.Now()
	if _, err := svcCtx.PreStockRepo.SetFailedInPreSecKill(ctx, svcCtx.Data, userID, goods.ID, num, secNum, record); err != nil {
		log.Error(ctx, "update pre-seckill failed state failed",
			log.Field(log.FieldAction, "seckill.set_failed"),
			log.Field(log.FieldError, err.Error()),
		)
		return err
	}
	return nil
}

func failureReason(code int, fallback string) string {
	if fallback != "" {
		return fallback
	}
	return getErrMsg(code)
}

func codeFromPreDescError(err error) int {
	switch {
	case err == nil:
		return SUCCESS
	case errors.Is(err, data.SecKillErrSecKilling):
		return ERR_DUPLICATE_SECKILL
	case errors.Is(err, data.SecKillErrUserGoodsOutLimit):
		return ERR_USER_QUOTA_NOT_ENOUGH
	case errors.Is(err, data.SecKillErrNotEnough), errors.Is(err, data.SecKillErrSelledOut):
		return ERR_GOODS_STOCK_NOT_ENOUGH
	default:
		return ERR_PRE_DESC_STOCK_FAILED
	}
}

func buildV2Reply(orderNum string, code int) *pb.SecKillV2Reply {
	reply := &pb.SecKillV2Reply{
		Code:    int32(code),
		Message: getErrMsg(code),
	}
	if orderNum != "" {
		reply.Data = &pb.SecKillV2ReplyData{OrderNum: orderNum}
	}
	if code == SUCCESS {
		reply.Message = ""
	}
	return reply
}

func buildV3Reply(secNum string, code int) *pb.SecKillV3Reply {
	reply := &pb.SecKillV3Reply{
		Code:    int32(code),
		Message: getErrMsg(code),
	}
	if secNum != "" {
		reply.Data = &pb.SecKillV3ReplyData{SecNum: secNum}
	}
	if code == SUCCESS {
		reply.Message = ""
	}
	return reply
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
