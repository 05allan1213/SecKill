package seckill

import (
	"context"
	"fmt"
	"time"

	"github.com/BitofferHub/pkg/constant"
	"github.com/BitofferHub/pkg/utils"
	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/log"
	"github.com/BitofferHub/seckill/internal/model"
	"github.com/BitofferHub/seckill/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

func secKillV1(ctx context.Context, svcCtx *svc.ServiceContext, req *pb.SecKillV1Request) (*pb.SecKillV1Reply, error) {
	var reply = new(pb.SecKillV1Reply)
	goods, err := svcCtx.GoodsModel().GetGoodsInfoByNumWithCache(ctx, req.GoodsNum)
	if err != nil {
		log.ErrorContextw(ctx, "load goods for seckill v1 failed",
			logx.Field("goods_num", req.GoodsNum),
			logx.Field("user_id", req.UserID),
			logx.Field("err", err))
		return nil, err
	}

	secNum := utils.NewUuid()
	orderNum, code, err := secKillInStore(ctx, svcCtx, goods, secNum, req.UserID, int(req.Num))
	if err != nil {
		log.ErrorContextw(ctx, "seckill v1 transaction failed",
			logx.Field("sec_num", secNum),
			logx.Field("goods_id", goods.ID),
			logx.Field("user_id", req.UserID),
			logx.Field("num", req.Num),
			logx.Field("err", err))
		return reply, nil
	}

	reply.Data = &pb.SecKillV1ReplyData{OrderNum: orderNum}
	reply.Message = ""
	reply.Code = int32(code)
	return reply, nil
}

func secKillV2(ctx context.Context, svcCtx *svc.ServiceContext, req *pb.SecKillV2Request) (*pb.SecKillV2Reply, error) {
	var reply = new(pb.SecKillV2Reply)
	goods, err := svcCtx.GoodsModel().GetGoodsInfoByNumWithCache(ctx, req.GoodsNum)
	if err != nil {
		log.ErrorContextw(ctx, "load goods for seckill v2 failed",
			logx.Field("goods_num", req.GoodsNum),
			logx.Field("user_id", req.UserID),
			logx.Field("err", err))
		return nil, err
	}

	secNum := utils.NewUuid()
	record := newPreSeckillRecord(secNum, req.UserID, goods)
	alreadySecNum, err := svcCtx.PreStockModel().PreDescStock(ctx, req.UserID,
		goods.ID, req.Num, secNum, &record)
	if err != nil {
		if err.Error() == model.SecKillErrSecKilling.Error() {
			reply.Message = err.Error() + ":" + fmt.Sprintf("%s", alreadySecNum)
			return reply, nil
		}
		log.ErrorContextw(ctx, "pre deduct stock for seckill v2 failed",
			logx.Field("sec_num", secNum),
			logx.Field("goods_id", goods.ID),
			logx.Field("user_id", req.UserID),
			logx.Field("num", req.Num),
			logx.Field("err", err))
		reply.Code = -10100
		reply.Message = err.Error()
		return reply, err
	}

	orderNum, code, err := secKillInStore(ctx, svcCtx, goods, secNum, req.UserID, int(req.Num))
	if err != nil {
		return reply, err
	}

	record.OrderNum = orderNum
	record.Status = int(model.SK_STATUS_BEFORE_PAY)
	record.ModifyTime = time.Now()
	if _, err := svcCtx.PreStockModel().SetSuccessInPreSecKill(ctx, req.UserID, goods.ID, secNum, &record); err != nil {
		log.ErrorContextw(ctx, "mark seckill v2 pre stock success failed",
			logx.Field("sec_num", secNum),
			logx.Field("order_num", orderNum),
			logx.Field("goods_id", goods.ID),
			logx.Field("user_id", req.UserID),
			logx.Field("err", err))
		return reply, err
	}

	reply.Data = &pb.SecKillV2ReplyData{OrderNum: orderNum}
	reply.Code = int32(code)
	return reply, nil
}

func secKillV3(ctx context.Context, svcCtx *svc.ServiceContext, req *pb.SecKillV3Request) (*pb.SecKillV3Reply, error) {
	var reply = new(pb.SecKillV3Reply)
	goods, err := svcCtx.GoodsModel().GetGoodsInfoByNumWithCache(ctx, req.GoodsNum)
	if err != nil {
		log.ErrorContextw(ctx, "load goods for seckill v3 failed",
			logx.Field("goods_num", req.GoodsNum),
			logx.Field("user_id", req.UserID),
			logx.Field("err", err))
		return nil, err
	}

	secNum := utils.NewUuid()
	record := newPreSeckillRecord(secNum, req.UserID, goods)
	alreadySecNum, err := svcCtx.PreStockModel().PreDescStock(ctx, req.UserID,
		goods.ID, req.Num, secNum, &record)
	if err != nil {
		if err.Error() == model.SecKillErrSecKilling.Error() {
			reply.Message = err.Error() + ":" + fmt.Sprintf("%s", alreadySecNum)
			return reply, nil
		}

		log.ErrorContextw(ctx, "pre deduct stock for seckill v3 failed",
			logx.Field("sec_num", secNum),
			logx.Field("goods_id", goods.ID),
			logx.Field("user_id", req.UserID),
			logx.Field("num", req.Num),
			logx.Field("err", err))
		return nil, err
	}

	msg := &model.SeckillMessage{
		TraceID: traceIDFromContext(ctx),
		Goods:   goods,
		SecNum:  secNum,
		UserID:  req.UserID,
		Num:     int(req.Num),
	}
	if err := svcCtx.MsgModel().SendSecKillMsg(ctx, msg); err != nil {
		log.ErrorContextw(ctx, "send seckill mq message failed",
			logx.Field("sec_num", secNum),
			logx.Field("goods_id", goods.ID),
			logx.Field("user_id", req.UserID),
			logx.Field("err", err))
		return nil, err
	}

	reply.Data = &pb.SecKillV3ReplyData{SecNum: secNum}
	return reply, nil
}

func getSecKillInfo(ctx context.Context, svcCtx *svc.ServiceContext, req *pb.GetSecKillInfoRequest) (*pb.GetSecKillInfoReply, error) {
	var reply = new(pb.GetSecKillInfoReply)
	record, err := svcCtx.PreStockModel().GetSecKillInfo(ctx, req.SecNum)
	if err != nil {
		log.ErrorContextw(ctx, "get seckill info failed",
			logx.Field("sec_num", req.SecNum),
			logx.Field("err", err))
		return nil, err
	}

	reply.Data = &pb.GetSecKillInfoReplyData{
		Status:   int32(record.Status),
		OrderNum: record.OrderNum,
		SecNum:   record.SecNum,
	}
	return reply, nil
}

func getGoodsList(ctx context.Context, svcCtx *svc.ServiceContext, req *pb.GetGoodsListRequest) (*pb.GetGoodsListReply, error) {
	var reply = new(pb.GetGoodsListReply)
	goodsList, err := svcCtx.GoodsModel().GetGoodsList(ctx, int(req.Offset), int(req.Limit))
	if err != nil {
		log.ErrorContextw(ctx, "get goods list failed",
			logx.Field("offset", req.Offset),
			logx.Field("limit", req.Limit),
			logx.Field("err", err))
		return nil, err
	}

	reply.Data = &pb.GetGoodsListReplyData{GoodsList: make([]*pb.GoodInfo, 0, len(goodsList))}
	for _, bizGoods := range goodsList {
		info := new(pb.GoodInfo)
		convertBizGoodsToPbGoods(bizGoods, info)
		reply.Data.GoodsList = append(reply.Data.GoodsList, info)
	}

	return reply, nil
}

func handleConsumedMessage(ctx context.Context, svcCtx *svc.ServiceContext, message []byte) error {
	if ctx == nil {
		return fmt.Errorf("nil consume context")
	}

	skMsg, err := svcCtx.MsgModel().UnmarshalSecKillMsg(ctx, message)
	if err != nil {
		log.ErrorContextw(ctx, "unmarshal seckill message failed",
			logx.Field("err", err),
			logx.Field("message_size", len(message)))
		return err
	}
	if skMsg.TraceID != "" {
		ctx = context.WithValue(ctx, constant.TraceID, skMsg.TraceID)
	}

	log.InfoContextw(ctx, "handling seckill message",
		logx.Field("sec_num", skMsg.SecNum),
		logx.Field("user_id", skMsg.UserID),
		logx.Field("goods_id", skMsg.Goods.ID),
		logx.Field("message_size", len(message)))

	orderNum, _, err := secKillInStore(ctx, svcCtx, skMsg.Goods, skMsg.SecNum, skMsg.UserID, skMsg.Num)
	if err != nil {
		log.ErrorContextw(ctx, "consume seckill message failed",
			logx.Field("sec_num", skMsg.SecNum),
			logx.Field("user_id", skMsg.UserID),
			logx.Field("goods_id", skMsg.Goods.ID),
			logx.Field("err", err))
		return err
	}

	record, err := svcCtx.PreStockModel().GetSecKillInfo(ctx, skMsg.SecNum)
	if err != nil {
		log.ErrorContextw(ctx, "load pre seckill record failed",
			logx.Field("sec_num", skMsg.SecNum),
			logx.Field("err", err))
		return err
	}

	record.OrderNum = orderNum
	record.Status = int(model.SK_STATUS_BEFORE_PAY)
	record.ModifyTime = time.Now()
	if _, err := svcCtx.PreStockModel().SetSuccessInPreSecKill(ctx, skMsg.UserID, skMsg.Goods.ID, skMsg.SecNum, record); err != nil {
		log.ErrorContextw(ctx, "mark pre seckill success failed",
			logx.Field("sec_num", skMsg.SecNum),
			logx.Field("order_num", orderNum),
			logx.Field("err", err))
		return err
	}

	return nil
}

func secKillInStore(ctx context.Context, svcCtx *svc.ServiceContext, goods *model.Goods,
	secNum string, userID int64, num int) (string, int, error) {
	orderNum := utils.NewUuid()
	if secNum == "" {
		secNum = utils.NewUuid()
	}

	code := 0
	err := svcCtx.Store().RunInTx(ctx, func(txStore *model.Store) error {
		var (
			rowAffected    int64
			globalQuota    = new(model.Quota)
			userQuota      = new(model.UserQuota)
			err            error
			userKilledNum  int64
			userQuotaNum   int64
			userQuotaExist = true
		)

		userQuota, err = svcCtx.UserQuotaModel().WithStore(txStore).FindUserGoodsQuota(ctx, userID, goods.ID)
		if err != nil {
			if err.Error() == gorm.ErrRecordNotFound.Error() {
				userQuotaExist = false
			} else {
				log.ErrorContextw(ctx, "load user quota failed",
					logx.Field("user_id", userID),
					logx.Field("goods_id", goods.ID),
					logx.Field("err", err))
				code = ERR_FIND_USER_QUOTA_FAILED
				return err
			}
		} else {
			userQuotaNum = userQuota.Num
			userKilledNum = userQuota.KilledNum
		}

		if userQuotaNum == 0 {
			globalQuota, err = svcCtx.QuotaModel().WithStore(txStore).FindByGoodsID(ctx, goods.ID)
			if err != nil {
				if err.Error() != gorm.ErrRecordNotFound.Error() {
					log.ErrorContextw(ctx, "load goods quota failed",
						logx.Field("goods_id", goods.ID),
						logx.Field("err", err))
					code = ERR_FIND_GOODS_FAILED
					return err
				}
			} else {
				userQuotaNum = globalQuota.Num
			}
		}

		leftQuota := userQuotaNum - userKilledNum
		if int(leftQuota) < num {
			log.InfoContextw(ctx, "user quota not enough",
				logx.Field("user_id", userID),
				logx.Field("goods_id", goods.ID),
				logx.Field("left_quota", leftQuota),
				logx.Field("requested_num", num))
			code = ERR_USER_QUOTA_NOT_ENOUGH
			return nil
		}

		if !userQuotaExist {
			_, err = svcCtx.UserQuotaModel().WithStore(txStore).CreateUserQuota(ctx, &model.UserQuota{
				UserID:    userID,
				GoodsID:   goods.ID,
				KilledNum: int64(num),
			})
			if err != nil {
				log.ErrorContextw(ctx, "create user quota failed",
					logx.Field("user_id", userID),
					logx.Field("goods_id", goods.ID),
					logx.Field("num", num),
					logx.Field("err", err))
				code = ERR_CREATER_USER_QUOTA_FAILED
				return err
			}
		} else {
			_, err = svcCtx.UserQuotaModel().WithStore(txStore).IncrKilledNum(ctx, userID, goods.ID, int64(num))
			if err != nil {
				log.ErrorContextw(ctx, "increase user quota counter failed",
					logx.Field("user_id", userID),
					logx.Field("goods_id", goods.ID),
					logx.Field("num", num),
					logx.Field("err", err))
				code = ERR_RECORD_USER_KILLED_NUM_FAILED
				return err
			}
		}

		rowAffected, err = svcCtx.StockModel().WithStore(txStore).DescStock(ctx, goods.ID, int32(num))
		if err != nil {
			log.ErrorContextw(ctx, "deduct stock failed",
				logx.Field("goods_id", goods.ID),
				logx.Field("num", num),
				logx.Field("err", err))
			code = ERR_DESC_STOCK_FAILED
			return err
		}
		if rowAffected == 0 {
			log.InfoContextw(ctx, "goods stock not enough",
				logx.Field("goods_id", goods.ID),
				logx.Field("num", num))
			code = ERR_GOODS_STOCK_NOT_ENOUGH
			return nil
		}

		_, err = svcCtx.OrderModel().WithStore(txStore).CreateOrder(ctx, &model.Order{
			OrderNum: orderNum,
			GoodsID:  goods.ID,
			Price:    goods.Price,
			Buyer:    userID,
			Seller:   goods.Seller,
			Status:   int(model.SK_STATUS_BEFORE_PAY),
		})
		if err != nil {
			log.ErrorContextw(ctx, "create order failed",
				logx.Field("order_num", orderNum),
				logx.Field("goods_id", goods.ID),
				logx.Field("user_id", userID),
				logx.Field("err", err))
			code = ERR_CREATE_ORDER_FAILED
			return err
		}

		_, err = svcCtx.RecordModel().WithStore(txStore).CreateSecKillRecord(ctx, &model.SecKillRecord{
			UserID:   userID,
			GoodsID:  goods.ID,
			SecNum:   secNum,
			OrderNum: orderNum,
			Price:    goods.Price,
			Status:   int(model.SK_STATUS_BEFORE_PAY),
		})
		if err != nil {
			log.ErrorContextw(ctx, "create seckill record failed",
				logx.Field("sec_num", secNum),
				logx.Field("order_num", orderNum),
				logx.Field("goods_id", goods.ID),
				logx.Field("user_id", userID),
				logx.Field("err", err))
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

func newPreSeckillRecord(secNum string, userID int64, goods *model.Goods) model.PreSecKillRecord {
	now := time.Now()
	return model.PreSecKillRecord{
		SecNum:     secNum,
		UserID:     userID,
		GoodsID:    goods.ID,
		OrderNum:   "",
		Price:      goods.Price,
		Status:     int(model.SK_STATUS_BEFORE_ORDER),
		CreateTime: now,
		ModifyTime: now,
	}
}

func convertBizGoodsToPbGoods(goods *model.Goods, info *pb.GoodInfo) {
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
