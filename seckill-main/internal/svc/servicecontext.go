package svc

import (
	"github.com/BitofferHub/seckill/internal/config"
	"github.com/BitofferHub/seckill/internal/data"
	projectlog "github.com/BitofferHub/seckill/internal/log"
)

type ServiceContext struct {
	Config        config.Config
	Data          *data.Data
	StockRepo     *data.SecKillStockRepo
	PreStockRepo  *data.PreSecKillStockRepo
	RecordRepo    *data.SecKillRecordRepo
	GoodsRepo     *data.GoodsRepo
	OrderRepo     *data.OrderRepo
	MessageRepo   *data.SecKillMsgRepo
	QuotaRepo     *data.QuotaRepo
	UserQuotaRepo *data.UserQuotaRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	projectlog.Init("./log/")

	dataLayer, err := data.NewDataFromConfig(c.Data, c.Log)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:        c,
		Data:          dataLayer,
		StockRepo:     data.NewSecKillStockRepo(dataLayer),
		PreStockRepo:  data.NewPreSecKillStockRepo(dataLayer),
		RecordRepo:    data.NewSecKillRecordRepo(dataLayer),
		GoodsRepo:     data.NewGoodsRepo(dataLayer),
		OrderRepo:     data.NewOrderRepo(dataLayer),
		MessageRepo:   data.NewSecKillMsgRepo(dataLayer),
		QuotaRepo:     data.NewQuotaRepo(dataLayer),
		UserQuotaRepo: data.NewUserQuotaRepo(dataLayer),
	}
}
