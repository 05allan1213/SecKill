package service

import (
	"context"
	"fmt"
	"github.com/BitofferHub/pkg/constant"
	"github.com/BitofferHub/pkg/utils"
	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/biz"
	"github.com/BitofferHub/seckill/internal/data"
	"github.com/BitofferHub/seckill/internal/log"
	"gorm.io/gorm"
	"time"
)

type SecKillService struct {
	pb.UnimplementedSecKillServer
	data        *biz.Data
	stockUc     *biz.SecKillStockUsecase
	preStockUc  *biz.PreSecKillStockUsecase
	recordUc    *biz.SecKillRecordUsecase
	goodsUc     *biz.GoodsUsecase
	orderUc     *biz.OrderUsecase
	msgUc       *biz.SecKillMsgUsecase
	quotaUc     *biz.QuotaUsecase
	userQuotaUc *biz.UserQuotaUsecase
}

// NewSecKillService
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@param uc
//	@return *SecKillService
func NewSecKillService(data *biz.Data, stockUc *biz.SecKillStockUsecase, preStockUc *biz.PreSecKillStockUsecase, recordUc *biz.SecKillRecordUsecase,
	goodsUc *biz.GoodsUsecase, orderUc *biz.OrderUsecase, msgUc *biz.SecKillMsgUsecase,
	quotaUc *biz.QuotaUsecase, userQuotaUc *biz.UserQuotaUsecase) *SecKillService {
	return &SecKillService{data: data, stockUc: stockUc, preStockUc: preStockUc, recordUc: recordUc, goodsUc: goodsUc,
		orderUc: orderUc, msgUc: msgUc, quotaUc: quotaUc, userQuotaUc: userQuotaUc}
}

// SecKillV1
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@Receiver s
//	@param ctx
//	@param req
//	@return *pb.SecKillV1Reply
//	@return error
func (s *SecKillService) SecKillV1(ctx context.Context, req *pb.SecKillV1Request) (*pb.SecKillV1Reply, error) {
	var reply = new(pb.SecKillV1Reply)
	goods, e := s.goodsUc.GetGoodsInfoByNum(ctx, s.data, req.GoodsNum)
	if e != nil {
		log.ErrorContextf(ctx, "GetGoodsInfo err %s\n", e.Error())
		return nil, e
	}
	secNum := utils.NewUuid()
	orderNum, code, err := s.secKillInStore(ctx, goods, secNum, req.UserID, int(req.Num))
	if err != nil {
		log.ErrorContextf(ctx, "secKillInStore err %s\n", err.Error())
		return reply, nil
	}
	reply.Data = new(pb.SecKillV1ReplyData)
	reply.Data.OrderNum = orderNum
	reply.Message = ""
	reply.Code = int32(code)
	return reply, nil
}

func (s *SecKillService) SecKillV2(ctx context.Context, req *pb.SecKillV2Request) (*pb.SecKillV2Reply, error) {
	var reply = new(pb.SecKillV2Reply)
	goods, e := s.goodsUc.GetGoodsInfoByNumWithCache(ctx, s.data, req.GoodsNum)
	if e != nil {
		log.InfoContextf(ctx, "GetGoodsInfo err %s\n", e.Error())
		return nil, e
	}
	secNum := utils.NewUuid()
	now := time.Now()
	record := biz.PreSecKillRecord{
		SecNum:     secNum,
		UserID:     req.UserID,
		GoodsID:    goods.ID,
		OrderNum:   "",
		Price:      goods.Price,
		Status:     int(biz.SK_STATUS_BEFORE_ORDER),
		CreateTime: now,
		ModifyTime: now,
	}
	var alreadySecNum string
	alreadySecNum, e = s.preStockUc.PreDescStock(ctx, s.data, req.UserID,
		goods.ID, req.Num, secNum, &record)
	if e != nil {
		if e.Error() == data.SecKillErrSecKilling.Error() {
			reply.Message = e.Error() + ":" + fmt.Sprintf("%s", alreadySecNum)
			return reply, nil
		}
		log.ErrorContextf(ctx, "Desc stock err %s\n", e.Error())
		reply.Code = -10100
		reply.Message = e.Error()
		return reply, e
	}
	orderNum, code, err := s.secKillInStore(ctx, goods, secNum, req.UserID, int(req.Num))
	if err != nil {
		return reply, err
	}
	record.OrderNum = orderNum
	record.Status = int(biz.SK_STATUS_BEFORE_PAY)
	record.ModifyTime = time.Now()
	if _, err := s.preStockUc.SetSuccessInPreSecKill(ctx, s.data, req.UserID, goods.ID, secNum, &record); err != nil {
		log.ErrorContextf(ctx, "set pre seckill success err %s\n", err.Error())
		return reply, err
	}
	reply.Data = new(pb.SecKillV2ReplyData)
	reply.Data.OrderNum = orderNum
	reply.Code = int32(code)

	return reply, nil
}

func (s *SecKillService) SecKillV3(ctx context.Context, req *pb.SecKillV3Request) (*pb.SecKillV3Reply, error) {
	var reply = new(pb.SecKillV3Reply)
	goods, e := s.goodsUc.GetGoodsInfoByNum(ctx, s.data, req.GoodsNum)
	if e != nil {
		log.InfoContextf(ctx, "GetGoodsInfo err %s\n", e.Error())
		return nil, e
	}
	secNum := utils.NewUuid()
	now := time.Now()
	record := biz.PreSecKillRecord{
		SecNum:     secNum,
		UserID:     req.UserID,
		GoodsID:    goods.ID,
		OrderNum:   "",
		Price:      goods.Price,
		Status:     int(biz.SK_STATUS_BEFORE_ORDER),
		CreateTime: now,
		ModifyTime: now,
	}
	var alreadySecNum string
	alreadySecNum, e = s.preStockUc.PreDescStock(ctx, s.data, req.UserID,
		goods.ID, req.Num, secNum, &record)
	if e != nil {
		if e.Error() == data.SecKillErrSecKilling.Error() {
			reply.Message = e.Error() + ":" + fmt.Sprintf("%s", alreadySecNum)
			return reply, nil
		} else {
			log.ErrorContextf(ctx, "Desc stock err %s\n", e.Error())
		}
		return nil, e
	}
	// send to mq
	var msg = &biz.SeckillMessage{
		TraceID: traceIDFromContext(ctx),
		Goods:   goods,
		SecNum:  secNum,
		UserID:  req.UserID,
		Num:     int(req.Num),
	}
	if err := s.msgUc.SendSecKillMsg(ctx, s.data, msg); err != nil {
		log.ErrorContextf(ctx, "send seckill mq msg err %s\n", err.Error())
		return nil, err
	}
	reply.Data = new(pb.SecKillV3ReplyData)
	reply.Data.SecNum = secNum
	return reply, nil
}

func (s *SecKillService) secKillInStore(ctx context.Context, goods *biz.Goods,
	secNum string, userID int64, num int) (string, int, error) {
	orderNum := utils.NewUuid()
	if secNum == "" {
		secNum = utils.NewUuid()
	}

	code := 0
	err := s.data.RunInTx(ctx, func(txData *biz.Data) error {
		var rowAffected int64
		var globalQuota = new(biz.Quota)
		var userQuota = new(biz.UserQuota)

		var err error
		var userKilledNum int64
		var userQuotaNum int64
		var userQuotaExist = true

		userQuota, err = s.userQuotaUc.FindUserGoodsQuota(ctx, txData, userID, goods.ID)
		if err != nil {
			if err.Error() == gorm.ErrRecordNotFound.Error() {
				userQuotaExist = false
			} else {
				log.ErrorContextf(ctx, "userQuotaUc.FindUserGoodsQuota err %s\n", err.Error())
				code = ERR_FIND_USER_QUOTA_FAILED
				return err
			}
		} else {
			userQuotaNum = userQuota.Num
			userKilledNum = userQuota.KilledNum
		}

		if userQuotaNum == 0 {
			globalQuota, err = s.quotaUc.FindByGoodsID(ctx, txData, goods.ID)
			if err != nil {
				if err.Error() != gorm.ErrRecordNotFound.Error() {
					log.ErrorContextf(ctx, "quotaUc.FindByGoodsID err %s\n", err.Error())
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
			_, err = s.userQuotaUc.CreateUserQuota(ctx, txData, &biz.UserQuota{
				UserID:    userID,
				GoodsID:   goods.ID,
				KilledNum: int64(num),
			})
			if err != nil {
				log.ErrorContextf(ctx, "userQuotaUc.CreateUserQuota err %s\n", err.Error())
				code = ERR_CREATER_USER_QUOTA_FAILED
				return err
			}
		} else {
			_, err = s.userQuotaUc.IncrKilledNum(ctx, txData, userID, goods.ID, int64(num))
			if err != nil {
				log.ErrorContextf(ctx, "userQuotaUc.IncrKilledNum err %s\n", err.Error())
				code = ERR_RECORD_USER_KILLED_NUM_FAILED
				return err
			}
		}

		rowAffected, err = s.stockUc.DescStock(ctx, txData, goods.ID, int32(num))
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

		_, err = s.orderUc.CreateOrder(ctx, txData, &biz.Order{
			OrderNum: orderNum,
			GoodsID:  goods.ID,
			Price:    goods.Price,
			Buyer:    userID,
			Seller:   goods.Seller,
			Status:   int(biz.SK_STATUS_BEFORE_PAY),
		})
		if err != nil {
			log.ErrorContextf(ctx, "create order err %s\n", err.Error())
			code = ERR_CREATE_ORDER_FAILED
			return err
		}

		_, err = s.recordUc.CreateSecKillRecord(ctx, txData, &biz.SecKillRecord{
			UserID:   userID,
			GoodsID:  goods.ID,
			SecNum:   secNum,
			OrderNum: orderNum,
			Price:    goods.Price,
			Status:   int(biz.SK_STATUS_BEFORE_PAY),
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

func (s *SecKillService) GetSecKillInfo(ctx context.Context, req *pb.GetSecKillInfoRequest) (*pb.GetSecKillInfoReply, error) {
	var reply = new(pb.GetSecKillInfoReply)
	record, err := s.preStockUc.GetSecKillInfo(ctx, s.data, req.SecNum)
	if err != nil {
		log.ErrorContextf(ctx, "get secinfo by secnum err %s\n", err.Error())
		return nil, err
	}
	reply.Data = new(pb.GetSecKillInfoReplyData)
	reply.Data.Status = int32(record.Status)
	reply.Data.OrderNum = record.OrderNum
	reply.Data.SecNum = record.SecNum
	return reply, nil
}

func (s *SecKillService) GetGoodsList(ctx context.Context, req *pb.GetGoodsListRequest) (*pb.GetGoodsListReply, error) {
	var reply = new(pb.GetGoodsListReply)
	goodsList, err := s.goodsUc.GetGoodsList(ctx, s.data, int(req.Offset), int(req.Limit))
	if err != nil {
		log.ErrorContextf(ctx, "get secinfo by secnum err %s\n", err.Error())
		return nil, err
	}
	reply.Data = new(pb.GetGoodsListReplyData)
	reply.Data.GoodsList = make([]*pb.GoodInfo, 0)

	for _, bizGoods := range goodsList {
		var info = new(pb.GoodInfo)
		convertBizGoodsToPbGoods(bizGoods, info)
		reply.Data.GoodsList = append(reply.Data.GoodsList, info)
	}
	return reply, nil
}

func (s *SecKillService) GetGoodsInfo(ctx context.Context, req *pb.GetGoodsInfoRequest) (*pb.GetGoodsInfoReply, error) {
	var reply = new(pb.GetGoodsInfoReply)
	bizGoods, err := s.goodsUc.GetGoodsInfoByNum(ctx, s.data, req.GoodsNum)
	if err != nil {
		log.ErrorContextf(ctx, "get secinfo by secnum err %s\n", err.Error())
		return nil, err
	}
	reply.Data = new(pb.GetGoodsInfoReplyData)
	reply.Data.GoodsInfo = new(pb.GoodInfo)
	convertBizGoodsToPbGoods(bizGoods, reply.Data.GoodsInfo)
	return reply, nil
}

func convertBizGoodsToPbGoods(bizGoods *biz.Goods, info *pb.GoodInfo) {
	info.GoodsNum = bizGoods.GoodsNum
	info.GoodsName = bizGoods.GoodsName
	info.Price = float32(bizGoods.Price)
	info.PicUrl = bizGoods.PicUrl
	info.Seller = bizGoods.Seller
}

func traceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	traceID, _ := ctx.Value(constant.TraceID).(string)
	return traceID
}
