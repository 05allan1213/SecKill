package svc

import (
	zaplog "github.com/BitofferHub/pkg/middlewares/log"
	"github.com/BitofferHub/seckill/internal/biz"
	"github.com/BitofferHub/seckill/internal/config"
	"github.com/BitofferHub/seckill/internal/data"
	"github.com/BitofferHub/seckill/internal/service"
)

type ServiceContext struct {
	Config         config.Config
	Data           *data.Data
	SecKillService *service.SecKillService
}

func NewServiceContext(c config.Config) *ServiceContext {
	zaplog.Init(
		zaplog.WithLogPath("./log/"),
		zaplog.WithLogLevel("info"),
		zaplog.WithFileName("bitstorm.log"),
		zaplog.WithMaxBackups(100),
		zaplog.WithMaxSize(1024*1024*10),
	)

	dataLayer, err := data.NewDataFromConfig(c.Data)
	if err != nil {
		panic(err)
	}

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
		SecKillService: seckillService,
	}
}
