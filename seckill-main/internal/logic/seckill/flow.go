package seckill

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/BitofferHub/pkg/constant"
	"github.com/BitofferHub/pkg/utils"
	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/data"
	"github.com/BitofferHub/seckill/internal/log"
	"github.com/BitofferHub/seckill/internal/svc"
	"gorm.io/gorm"
)

type ConsumeResult struct {
	Success      bool
	FailureClass data.FailureClass
	ShouldRetry  bool
	LastError    string
}

func HandleConsumedMessage(ctx context.Context, svcCtx *svc.ServiceContext, message []byte) error {
	result := HandleConsumedMessageWithRetry(ctx, svcCtx, message, 0, "")
	if result.FailureClass == data.FailureClassPoisonMessage {
		return errors.New("poison message")
	}
	return nil
}

func HandleConsumedMessageWithRetry(ctx context.Context, svcCtx *svc.ServiceContext, message []byte, attempt int, lastError string) *ConsumeResult {
	result := &ConsumeResult{}

	if data.IsPoisonMessage(message) {
		result.FailureClass = data.FailureClassPoisonMessage
		result.LastError = "empty or oversized message"
		return result
	}

	envelope, err := svcCtx.MessageRepo.UnmarshalEnvelope(ctx, message)
	if err != nil {
		var legacyMsg *data.SeckillMessage
		legacyMsg, legacyErr := svcCtx.MessageRepo.UnmarshalSecKillMsg(ctx, svcCtx.Data, message)
		if legacyErr != nil {
			result.FailureClass = data.FailureClassPoisonMessage
			result.LastError = "failed to parse message: " + err.Error()
			return result
		}
		envelope = data.NewSeckillEnvelope(legacyMsg)
	}

	if data.IsPoisonEnvelope(envelope) {
		result.FailureClass = data.FailureClassPoisonMessage
		result.LastError = "invalid envelope payload"
		return result
	}

	skMsg := envelope.Payload
	currentAttempt := envelope.Attempt
	if attempt > 0 {
		currentAttempt = attempt
	}
	if lastError != "" {
		envelope.LastError = lastError
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
		log.Field("attempt", currentAttempt),
		log.Field("sourceTopic", envelope.SourceTopic),
	)

	if err := svcCtx.AsyncResultRepo.UpsertPending(ctx, svcCtx.Data.GetDB(), skMsg.SecNum, skMsg.UserID, skMsg.Goods.ID, skMsg.Goods.GoodsNum, currentAttempt); err != nil {
		log.Warn(ctx, "upsert pending async result failed",
			log.Field(log.FieldAction, "seckill.mq.upsert_pending"),
			log.Field(log.FieldError, err.Error()),
		)
	}

	log.InfoEvery(ctx, "seckill.mq.consume.start", 2*time.Second, "consume seckill message", log.Field(log.FieldAction, "seckill.mq.consume"))

	existingRecord, _ := svcCtx.RecordRepo.FindBySecNum(ctx, svcCtx.Data, skMsg.SecNum)
	if existingRecord != nil && isSuccessStatus(existingRecord.Status) {
		log.Info(ctx, "seckill record already exists with success status, skipping processing",
			log.Field(log.FieldAction, "seckill.mq.idempotent"),
			log.Field(log.FieldSecNum, skMsg.SecNum),
			log.Field(log.FieldOrderNum, existingRecord.OrderNum),
			log.Field("recordStatus", existingRecord.Status),
		)
		if err := markPreSecKillCompleted(ctx, svcCtx, skMsg.Goods, skMsg.UserID, skMsg.SecNum, existingRecord.OrderNum, existingRecord.Status); err != nil {
			log.Warn(ctx, "mark pre-seckill success failed for existing record",
				log.Field(log.FieldAction, "seckill.mq.redis_success"),
				log.Field(log.FieldOrderNum, existingRecord.OrderNum),
				log.Field(log.FieldError, err.Error()),
			)
		}
		if upsertErr := svcCtx.AsyncResultRepo.UpsertSuccess(ctx, svcCtx.Data.GetDB(), skMsg.SecNum, existingRecord.OrderNum, existingRecord.Status); upsertErr != nil {
			log.Warn(ctx, "upsert success async result failed for existing record", log.Field(log.FieldError, upsertErr.Error()))
		}
		result.Success = true
		return result
	}

	orderNum, code, err := secKillInStore(ctx, svcCtx, skMsg.Goods, skMsg.SecNum, skMsg.UserID, skMsg.Num)

	if err != nil {
		if isSecNumDuplicateError(err) {
			existingRecord, recordErr := svcCtx.RecordRepo.FindBySecNum(ctx, svcCtx.Data, skMsg.SecNum)
			if recordErr == nil && existingRecord != nil && isSuccessStatus(existingRecord.Status) {
				log.Info(ctx, "seckill record already exists with success status, treating as success",
					log.Field(log.FieldAction, "seckill.mq.duplicate"),
					log.Field(log.FieldSecNum, skMsg.SecNum),
					log.Field(log.FieldOrderNum, existingRecord.OrderNum),
					log.Field("recordStatus", existingRecord.Status),
				)
				if err := markPreSecKillCompleted(ctx, svcCtx, skMsg.Goods, skMsg.UserID, skMsg.SecNum, existingRecord.OrderNum, existingRecord.Status); err != nil {
					log.Warn(ctx, "mark pre-seckill success failed for duplicate record",
						log.Field(log.FieldAction, "seckill.mq.redis_success"),
						log.Field(log.FieldOrderNum, existingRecord.OrderNum),
						log.Field(log.FieldError, err.Error()),
					)
				}
				if upsertErr := svcCtx.AsyncResultRepo.UpsertSuccess(ctx, svcCtx.Data.GetDB(), skMsg.SecNum, existingRecord.OrderNum, existingRecord.Status); upsertErr != nil {
					log.Warn(ctx, "upsert success async result failed for duplicate", log.Field(log.FieldError, upsertErr.Error()))
				}
				result.Success = true
				return result
			}
		}

		result.FailureClass = data.ClassifyError(err)
		result.LastError = err.Error()

		if result.FailureClass == data.FailureClassBusinessTerminal {
			if failErr := markPreSecKillFailed(ctx, svcCtx, skMsg.Goods, skMsg.UserID, int32(skMsg.Num), skMsg.SecNum, code, ""); failErr != nil {
				result.FailureClass = data.FailureClassTransientInfra
				result.LastError = "failed to mark pre-seckill failed: " + failErr.Error()
			}
			if dbErr := svcCtx.AsyncResultRepo.UpsertFailure(ctx, svcCtx.Data.GetDB(), skMsg.SecNum, getErrMsg(code), err.Error(), currentAttempt); dbErr != nil {
				log.Warn(ctx, "upsert failure async result failed",
					log.Field(log.FieldError, dbErr.Error()),
					log.Field("failureClass", result.FailureClass.String()),
				)
			}
			log.WarnEvery(ctx, "seckill.mq.consume.reject", 2*time.Second, "seckill rejected",
				log.Field(log.FieldAction, "seckill.mq.consume"),
				log.Field("resultCode", code),
				log.Field("failureClass", result.FailureClass.String()),
			)
			return result
		}

		if result.FailureClass == data.FailureClassTransientInfra {
			maxAttempts := svcCtx.Config.Data.Kafka.Retry.MaxAttempts
			if currentAttempt < maxAttempts {
				result.ShouldRetry = true
				log.Warn(ctx, "transient error, will retry",
					log.Field(log.FieldAction, "seckill.mq.retry"),
					log.Field("attempt", currentAttempt),
					log.Field("maxAttempts", maxAttempts),
					log.Field(log.FieldError, err.Error()),
					log.Field("failureClass", result.FailureClass.String()),
				)
				return result
			}
			if failErr := markPreSecKillFailed(ctx, svcCtx, skMsg.Goods, skMsg.UserID, int32(skMsg.Num), skMsg.SecNum, code, ""); failErr != nil {
				log.Warn(ctx, "mark pre-seckill failed error",
					log.Field(log.FieldError, failErr.Error()),
					log.Field("failureClass", result.FailureClass.String()),
				)
			}
			if dbErr := svcCtx.AsyncResultRepo.UpsertFailure(ctx, svcCtx.Data.GetDB(), skMsg.SecNum, "retry_exhausted", err.Error(), currentAttempt); dbErr != nil {
				log.Warn(ctx, "upsert failure async result failed",
					log.Field(log.FieldError, dbErr.Error()),
					log.Field("failureClass", result.FailureClass.String()),
				)
			}
			log.Error(ctx, "seckill retry exhausted",
				log.Field(log.FieldAction, "seckill.mq.dlq"),
				log.Field("attempt", currentAttempt),
				log.Field(log.FieldError, err.Error()),
				log.Field("failureClass", result.FailureClass.String()),
				log.Field("dlq", true),
			)
			return result
		}
		return result
	}

	if code != SUCCESS {
		if code == SUCCESS {
			code = ERR_CREATE_ORDER_FAILED
		}
		result.FailureClass = data.FailureClassBusinessTerminal
		result.LastError = getErrMsg(code)
		if failErr := markPreSecKillFailed(ctx, svcCtx, skMsg.Goods, skMsg.UserID, int32(skMsg.Num), skMsg.SecNum, code, ""); failErr != nil {
			result.FailureClass = data.FailureClassTransientInfra
			result.LastError = "failed to mark pre-seckill failed: " + failErr.Error()
		}
		if dbErr := svcCtx.AsyncResultRepo.UpsertFailure(ctx, svcCtx.Data.GetDB(), skMsg.SecNum, getErrMsg(code), "", currentAttempt); dbErr != nil {
			log.Warn(ctx, "upsert failure async result failed",
				log.Field(log.FieldError, dbErr.Error()),
				log.Field("failureClass", result.FailureClass.String()),
			)
		}
		log.WarnEvery(ctx, "seckill.mq.consume.reject", 2*time.Second, "seckill store rejected",
			log.Field(log.FieldAction, "seckill.mq.consume"),
			log.Field("resultCode", code),
			log.Field("failureClass", result.FailureClass.String()),
		)
		return result
	}

	if err := markPreSecKillSuccess(ctx, svcCtx, skMsg.Goods, skMsg.UserID, skMsg.SecNum, orderNum); err != nil {
		log.Warn(ctx, "mark pre-seckill success failed, but order already created",
			log.Field(log.FieldAction, "seckill.mq.redis_success"),
			log.Field(log.FieldOrderNum, orderNum),
			log.Field(log.FieldError, err.Error()),
		)
	}

	if err := svcCtx.AsyncResultRepo.UpsertSuccess(ctx, svcCtx.Data.GetDB(), skMsg.SecNum, orderNum, int(data.SK_STATUS_BEFORE_PAY)); err != nil {
		log.Warn(ctx, "upsert success async result failed", log.Field(log.FieldError, err.Error()))
	}

	result.Success = true
	log.InfoEvery(ctx, "seckill.mq.consume.finish", 2*time.Second, "consume seckill message finished",
		log.Field(log.FieldAction, "seckill.mq.consume"),
		log.Field(log.FieldOrderNum, orderNum),
	)
	return result
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
	return markPreSecKillCompleted(ctx, svcCtx, goods, userID, secNum, orderNum, int(data.SK_STATUS_BEFORE_PAY))
}

func markPreSecKillCompleted(ctx context.Context, svcCtx *svc.ServiceContext, goods *data.Goods, userID int64, secNum, orderNum string, status int) error {
	record, err := loadPreSecKillRecord(ctx, svcCtx, goods, userID, secNum)
	if err != nil {
		return err
	}
	record.OrderNum = orderNum
	record.Status = status
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

func isSecNumDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "duplicate entry") {
		return false
	}
	return strings.Contains(errStr, "idx_secnum") || strings.Contains(errStr, "sec_num")
}

func isSuccessStatus(status int) bool {
	switch data.SecKillStatusEnum(status) {
	case data.SK_STATUS_BEFORE_PAY, data.SK_STATUS_PAYED, data.SK_STATUS_OOT, data.SK_STATUS_CANCEL:
		return true
	default:
		return false
	}
}
