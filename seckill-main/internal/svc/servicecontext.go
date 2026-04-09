package svc

import (
	"github.com/BitofferHub/seckill/internal/biz"
	"github.com/BitofferHub/seckill/internal/config"
	"github.com/BitofferHub/seckill/internal/data"
	zaplog "github.com/BitofferHub/seckill/internal/log"
	"github.com/BitofferHub/seckill/internal/service"
)

type ServiceContext struct {
	Config         config.Config
	Data           *data.Data
	BizData        *biz.Data
	SecKillService *service.SecKillService
}

func NewServiceContext(c config.Config) *ServiceContext {
	zaplog.Init("./log/")

	dataLayer, err := data.NewDataFromConfig(c.Data)
	if err != nil {
		panic(err)
	}
	bizData := biz.NewData(dataLayer.GetDB(), dataLayer.GetCache(), dataLayer.GetMQProducer(), dataLayer.GetMQConsumer())

	stockRepo := data.NewSecKillStockRepo(dataLayer)
	preStockRepo := data.NewPreSecKillStockRepo(dataLayer)
	recordRepo := data.NewSecKillRecordRepo(dataLayer)
	goodsRepo := data.NewGoodsRepo(dataLayer)
	orderRepo := data.NewOrderRepo(dataLayer)
	msgRepo := data.NewSecKillMsgRepo(dataLayer)
	quotaRepo := data.NewQuotaRepo(dataLayer)
	userQuotaRepo := data.NewUserQuotaRepo(dataLayer)

	stockUsecase := biz.NewSecKillStockUsecase(stockRepo)
	preStockUsecase := biz.NewPreSecKillStockUsecase(preStockRepo)
	recordUsecase := biz.NewSecKillRecordUsecase(recordRepo)
	goodsUsecase := biz.NewGoodsUsecase(goodsRepo)
	orderUsecase := biz.NewOrderUsecase(orderRepo)
	msgUsecase := biz.NewSecKillMsgUsecase(msgRepo)
	quotaUsecase := biz.NewQuotaUsecase(quotaRepo)
	userQuotaUsecase := biz.NewUserQuotaUsecase(userQuotaRepo)

	seckillService := service.NewSecKillService(
		bizData,
		stockUsecase,
		preStockUsecase,
		recordUsecase,
		goodsUsecase,
		orderUsecase,
		msgUsecase,
		quotaUsecase,
		userQuotaUsecase,
	)

	return &ServiceContext{
		Config:         c,
		Data:           dataLayer,
		BizData:        bizData,
		SecKillService: seckillService,
	}
}
