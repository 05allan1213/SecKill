package service

import (
	"context"
	"fmt"
	"github.com/BitofferHub/pkg/constant"
	"github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/pkg/utils"
	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/biz"
	"github.com/BitofferHub/seckill/internal/data"
	"gorm.io/gorm"
	"time"
)

type SecKillService struct {
	pb.UnimplementedSecKillServer
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
func NewSecKillService(stockUc *biz.SecKillStockUsecase, preStockUc *biz.PreSecKillStockUsecase, recordUc *biz.SecKillRecordUsecase,
	goodsUc *biz.GoodsUsecase, orderUc *biz.OrderUsecase, msgUc *biz.SecKillMsgUsecase,
	quotaUc *biz.QuotaUsecase, userQuotaUc *biz.UserQuotaUsecase) *SecKillService {
	return &SecKillService{stockUc: stockUc, preStockUc: preStockUc, recordUc: recordUc, goodsUc: goodsUc,
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
	dt := biz.NewData(data.GetData().GetDB(), data.GetData().GetCache(), data.GetData().GetMQProducer(), data.GetData().GetMQConsumer())
	goods, e := s.goodsUc.GetGoodsInfoByNum(ctx, dt, req.GoodsNum)
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
	dt := biz.NewData(data.GetData().GetDB(), data.GetData().GetCache(), data.GetData().GetMQProducer(), data.GetData().GetMQConsumer())
	goods, e := s.goodsUc.GetGoodsInfoByNumWithCache(ctx, dt, req.GoodsNum)
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
	alreadySecNum, e = s.preStockUc.PreDescStock(ctx, dt, req.UserID,
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
	s.preStockUc.SetSuccessInPreSecKill(ctx, dt, req.UserID,
		goods.ID, secNum, &record)
	reply.Data = new(pb.SecKillV2ReplyData)
	reply.Data.OrderNum = orderNum
	reply.Code = int32(code)

	return reply, nil
}

func (s *SecKillService) SecKillV3(ctx context.Context, req *pb.SecKillV3Request) (*pb.SecKillV3Reply, error) {
	var reply = new(pb.SecKillV3Reply)
	dt := biz.NewData(data.GetData().GetDB(), data.GetData().GetCache(), data.GetData().GetMQProducer(), data.GetData().GetMQConsumer())
	goods, e := s.goodsUc.GetGoodsInfoByNum(ctx, dt, req.GoodsNum)
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
	alreadySecNum, e = s.preStockUc.PreDescStock(ctx, dt, req.UserID,
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
	traceID := ctx.Value(constant.TraceID)
	var msg = &biz.SeckillMessage{
		TraceID: traceID.(string),
		Goods:   goods,
		SecNum:  secNum,
		UserID:  req.UserID,
		Num:     int(req.Num),
	}
	s.msgUc.SendSecKillMsg(ctx, dt, msg)
	reply.Data = new(pb.SecKillV3ReplyData)
	reply.Data.SecNum = secNum
	return reply, nil
}

func (s *SecKillService) secKillInStore(ctx context.Context, goods *biz.Goods,
	secNum string, userID int64, num int) (string, int, error) {
	dt := biz.NewData(data.GetData().GetDB().Begin(), data.GetData().GetCache(), data.GetData().GetMQProducer(), data.GetData().GetMQConsumer())
	var err error
	defer func() {
		if err != nil {
			dt.GetDB().Rollback()
			return
		}
		dt.GetDB().Commit()
	}()
	orderNum := utils.NewUuid()
	var rowAffected int64
	var globalQuota = new(biz.Quota)
	var userQuota = new(biz.UserQuota)

	var userKilledNum int64
	var userQuotaNum int64
	var userQuotaExist = true
	// 获得用户级别得限额配置，如果Num为0，就是不限额
	userQuota, err = s.userQuotaUc.FindUserGoodsQuota(ctx, dt, userID, goods.ID)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			userQuotaExist = false
		} else {
			log.ErrorContextf(ctx, "userQuotaUc.FindUserGoodsQuota err %s\n", err.Error())
			return orderNum, ERR_FIND_USER_QUOTA_FAILED, err
		}
	} else {
		userQuotaNum = userQuota.Num
		userKilledNum = userQuota.KilledNum
	}
	// 如果用户级别得限额数没有设置，就看看全局限额
	if userQuotaNum == 0 {
		globalQuota, err = s.quotaUc.FindByGoodsID(ctx, dt, goods.ID)
		if err != nil {
			if err.Error() != gorm.ErrRecordNotFound.Error() {
				log.ErrorContextf(ctx, "quotaUc.FindByGoodsID err %s\n", err.Error())
				return orderNum, ERR_FIND_GOODS_FAILED, err
			}
		} else {
			userQuotaNum = globalQuota.Num
		}
	}
	// 算出剩余额度
	leftQuota := userQuotaNum - userKilledNum
	if int(leftQuota) < num {
		log.InfoContextf(ctx, "user %d, goods %d, quota limit %d", userID, goods.ID, leftQuota)
		return "", ERR_USER_QUOTA_NOT_ENOUGH, nil
	}
	// 如果之前没有创建过用户级别额度信息，就创建一下，主要为了记录这个用户已经Killed多少
	// 否则增加KilledNum即可
	if !userQuotaExist {
		_, err = s.userQuotaUc.CreateUserQuota(ctx, dt, &biz.UserQuota{
			UserID:    userID,
			GoodsID:   goods.ID,
			KilledNum: int64(num),
		})
		if err != nil {
			log.ErrorContextf(ctx, "userQuotaUc.CreateUserQuota err %s\n", err.Error())
			return orderNum, ERR_CREATER_USER_QUOTA_FAILED, err
		}
	} else {
		_, err = s.userQuotaUc.IncrKilledNum(ctx, dt, userID, goods.ID, int64(num))
		if err != nil {
			log.ErrorContextf(ctx, "userQuotaUc.IncrKilledNum err %s\n", err.Error())
			return orderNum, ERR_RECORD_USER_KILLED_NUM_FAILED, err
		}
	}

	rowAffected, err = s.stockUc.DescStock(ctx, dt, goods.ID, int32(num))
	if err != nil {
		log.ErrorContextf(ctx, "Desc stock err %s\n", err.Error())
		return orderNum, ERR_DESC_STOCK_FAILED, err
	}
	if rowAffected == 0 {
		log.InfoContextf(ctx, "goods %d stock not enough", goods.ID)
		return "", ERR_GOODS_STOCK_NOT_ENOUGH, nil
	}
	_, err = s.orderUc.CreateOrder(ctx, dt, &biz.Order{
		OrderNum: orderNum,
		GoodsID:  goods.ID,
		Price:    goods.Price,
		Buyer:    userID,
		Seller:   goods.Seller,
		Status:   int(biz.SK_STATUS_BEFORE_PAY),
	})
	if err != nil {
		log.ErrorContextf(ctx, "create order err %s\n", err.Error())
		return orderNum, ERR_CREATE_ORDER_FAILED, err
	}
	if secNum == "" {
		secNum = utils.NewUuid()
	}
	_, err = s.recordUc.CreateSecKillRecord(ctx, dt, &biz.SecKillRecord{
		UserID:   userID,
		GoodsID:  goods.ID,
		SecNum:   secNum,
		OrderNum: orderNum,
		Price:    goods.Price,
		Status:   int(biz.SK_STATUS_BEFORE_PAY),
	})
	if err != nil {
		log.ErrorContextf(ctx, "create seckill record err %s\n", err.Error())
		return orderNum, ERR_CREATE_SECKILL_RECORD_FAILED, err
	}
	return orderNum, 0, nil
}

func (s *SecKillService) GetSecKillInfo(ctx context.Context, req *pb.GetSecKillInfoRequest) (*pb.GetSecKillInfoReply, error) {
	var reply = new(pb.GetSecKillInfoReply)
	dt := biz.NewData(data.GetData().GetDB(), data.GetData().GetCache(), data.GetData().GetMQProducer(), data.GetData().GetMQConsumer())
	record, err := s.preStockUc.GetSecKillInfo(ctx, dt, req.SecNum)
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
	dt := biz.NewData(data.GetData().GetDB(), data.GetData().GetCache(), data.GetData().GetMQProducer(), data.GetData().GetMQConsumer())
	goodsList, err := s.goodsUc.GetGoodsList(ctx, dt, int(req.Offset), int(req.Limit))
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
	dt := biz.NewData(data.GetData().GetDB(), data.GetData().GetCache(), data.GetData().GetMQProducer(), data.GetData().GetMQConsumer())
	bizGoods, err := s.goodsUc.GetGoodsInfoByNum(ctx, dt, req.GoodsNum)
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
