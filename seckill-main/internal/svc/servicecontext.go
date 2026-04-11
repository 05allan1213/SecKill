package svc

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/BitofferHub/seckill/internal/config"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/BitofferHub/seckill/internal/model"
)

type ServiceContext struct {
	Config       config.Config
	current      atomic.Pointer[serviceBundle]
	watchContext context.CancelFunc
	watcher      *config.RuntimeWatcher
	gracePeriod  time.Duration
}

func NewServiceContext(c config.Config) *ServiceContext {
	bundle, err := newServiceBundle(c.Data)
	if err != nil {
		panic(err)
	}

	svc := &ServiceContext{
		Config:      c,
		gracePeriod: c.ConfigCenter.GracePeriod,
	}
	svc.current.Store(bundle)

	if c.ConfigCenter.Enabled && c.ConfigCenter.Watch {
		ctx, cancel := context.WithCancel(context.Background())
		svc.watchContext = cancel
		watcher, err := config.WatchRuntimeConfig(ctx, c.ConfigCenter, func(runtimeCfg config.RuntimeConfig) {
			next := c
			config.ApplyRuntimeConfig(&next, runtimeCfg)
			config.ApplyEnvOverrides(&next)
			svc.reload(next.Data)
		})
		if err != nil {
			panic(err)
		}
		svc.watcher = watcher
	}

	return svc
}

func (s *ServiceContext) Close() {
	if s.watchContext != nil {
		s.watchContext()
	}
	if s.watcher != nil {
		s.watcher.Close()
	}
	if bundle := s.current.Load(); bundle != nil {
		bundle.Close()
	}
}

func (s *ServiceContext) Store() *model.Store {
	return s.currentBundle().Store
}

func (s *ServiceContext) StockModel() *model.SecKillStockModel {
	return s.currentBundle().StockModel
}

func (s *ServiceContext) PreStockModel() *model.PreSecKillStockModel {
	return s.currentBundle().PreStockModel
}

func (s *ServiceContext) RecordModel() *model.SecKillRecordModel {
	return s.currentBundle().RecordModel
}

func (s *ServiceContext) GoodsModel() *model.GoodsModel {
	return s.currentBundle().GoodsModel
}

func (s *ServiceContext) OrderModel() *model.OrderModel {
	return s.currentBundle().OrderModel
}

func (s *ServiceContext) MsgModel() *model.SecKillMsgModel {
	return s.currentBundle().MsgModel
}

func (s *ServiceContext) QuotaModel() *model.QuotaModel {
	return s.currentBundle().QuotaModel
}

func (s *ServiceContext) UserQuotaModel() *model.UserQuotaModel {
	return s.currentBundle().UserQuotaModel
}

func (s *ServiceContext) currentBundle() *serviceBundle {
	bundle := s.current.Load()
	if bundle == nil {
		panic("service bundle is not initialized")
	}
	return bundle
}

func (s *ServiceContext) reload(data config.DataConf) {
	next, err := newServiceBundle(data)
	if err != nil {
		logx.Errorf("reload seckill runtime config failed: %v", err)
		return
	}

	prev := s.current.Swap(next)
	if prev == nil {
		return
	}

	go func(old *serviceBundle) {
		time.Sleep(s.gracePeriod)
		old.Close()
	}(prev)
}

type serviceBundle struct {
	Store          *model.Store
	StockModel     *model.SecKillStockModel
	PreStockModel  *model.PreSecKillStockModel
	RecordModel    *model.SecKillRecordModel
	GoodsModel     *model.GoodsModel
	OrderModel     *model.OrderModel
	MsgModel       *model.SecKillMsgModel
	QuotaModel     *model.QuotaModel
	UserQuotaModel *model.UserQuotaModel
}

func newServiceBundle(data config.DataConf) (*serviceBundle, error) {
	store, err := model.NewStoreFromConfig(data)
	if err != nil {
		return nil, err
	}

	return &serviceBundle{
		Store:          store,
		StockModel:     model.NewSecKillStockModel(store),
		PreStockModel:  model.NewPreSecKillStockModel(store),
		RecordModel:    model.NewSecKillRecordModel(store),
		GoodsModel:     model.NewGoodsModel(store),
		OrderModel:     model.NewOrderModel(store),
		MsgModel:       model.NewSecKillMsgModel(store),
		QuotaModel:     model.NewQuotaModel(store),
		UserQuotaModel: model.NewUserQuotaModel(store),
	}, nil
}

func (b *serviceBundle) Close() {
	if b == nil || b.Store == nil {
		return
	}
	b.Store.Close()
}
