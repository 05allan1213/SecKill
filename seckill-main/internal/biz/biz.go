package biz

import (
	"context"

	"github.com/BitofferHub/pkg/middlewares/cache"
	"github.com/BitofferHub/seckill/internal/mq"
	"gorm.io/gorm"
)

// Data .
type Data struct {
	db         *gorm.DB
	rdb        *cache.Client
	mqProducer mq.Producer
	mqConsumer mq.Consumer
}

// NewData
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description: Get New Data
//	@param db
//	@param rdb
//	@return *Data
func NewData(db *gorm.DB, rdb *cache.Client, producer mq.Producer, consumer mq.Consumer) *Data {
	var dt = &Data{
		db:         db,
		rdb:        rdb,
		mqProducer: producer,
		mqConsumer: consumer,
	}
	return dt
}

// GetDB
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@Receiver p
//	@return *gorm.DB
func (p *Data) GetDB() *gorm.DB {
	return p.db
}

// GetCache
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@Receiver p
//	@return *cache.Client
func (p *Data) GetCache() *cache.Client {
	return p.rdb
}

func (p *Data) GetMQProducer() mq.Producer {
	return p.mqProducer
}

func (p *Data) GetMQConsumer() mq.Consumer {
	return p.mqConsumer
}

func (p *Data) CloneWithDB(db *gorm.DB) *Data {
	if db == nil {
		db = p.db
	}
	return &Data{
		db:         db,
		rdb:        p.rdb,
		mqProducer: p.mqProducer,
		mqConsumer: p.mqConsumer,
	}
}

func (p *Data) RunInTx(ctx context.Context, fn func(txData *Data) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(p.CloneWithDB(tx))
	})
}
