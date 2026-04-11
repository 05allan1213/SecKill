package data

import (
	"context"
	"fmt"
	"github.com/BitofferHub/pkg/middlewares/cache"
	cfg "github.com/BitofferHub/seckill/internal/config"
	bitlog "github.com/BitofferHub/seckill/internal/log"
	"github.com/BitofferHub/seckill/internal/mq"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"sync"
	"time"
)

// Data .
type Data struct {
	db         *gorm.DB
	rdb        *cache.Client
	mqProducer mq.Producer
	mqConsumer mq.Consumer
	closeOnce  sync.Once
}

var data *Data

func GetData() *Data {
	return data
}
func (p *Data) GetDB() *gorm.DB {
	return p.db
}

func (p *Data) GetCache() *cache.Client {
	return p.rdb
}

func (p *Data) GetMQProducer() mq.Producer {
	return p.mqProducer
}

func (p *Data) GetMQConsumer() mq.Consumer {
	return p.mqConsumer
}

func (p *Data) Close() {
	p.closeOnce.Do(func() {
		if p.mqConsumer != nil {
			p.mqConsumer.Close()
		}
		if p.mqProducer != nil {
			p.mqProducer.Close()
		}
	})
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

func NewDataFromConfig(dt cfg.DataConf, logConf cfg.LogConf) (*Data, error) {
	db, err := openDB(dt.Database, logConf)
	if err != nil {
		return nil, err
	}
	cache.Init(
		cache.WithAddr(dt.Redis.Addr),
		cache.WithPassWord(dt.Redis.PassWord),
		cache.WithDB(int(dt.Redis.Db)),
		cache.WithPoolSize(int(dt.Redis.PoolSize)),
	)
	producer := mq.NewKafkaProducer(
		mq.WithBrokers(dt.Kafka.Producer.Brokers),
		mq.WithTopic(dt.Kafka.Producer.Topic),
		mq.WithAck(dt.Kafka.Producer.Ack),
	)
	if producer == nil {
		panic("nil producer")
	}
	consumer := newManagedKafkaConsumer(dt.Kafka.Consumer)
	dta := &Data{db: db, rdb: cache.GetRedisCli(), mqProducer: producer, mqConsumer: consumer}
	data = dta
	return dta, nil
}

func openDB(conf cfg.DatabaseConf, logConf cfg.LogConf) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.User, conf.Password, conf.Addr, conf.DataBase)

	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}
	gormCfg.Logger = bitlog.NewGormLogger(logConf.SQLMode, logConf.SQLSlowThresholdMs)

	db, err := gorm.Open(mysql.Open(dsn), gormCfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(int(conf.MaxIdleConn))
	sqlDB.SetMaxOpenConns(int(conf.MaxOpenConn))
	sqlDB.SetConnMaxLifetime(time.Duration(conf.MaxIdleTime) * time.Second)
	return db, nil
}
