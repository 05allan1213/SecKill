package data

import (
	"context"
	"errors"
	"fmt"
	"github.com/BitofferHub/pkg/middlewares/cache"
	cfg "github.com/BitofferHub/seckill/internal/config"
	bitlog "github.com/BitofferHub/seckill/internal/log"
	"github.com/BitofferHub/seckill/internal/mq"
	_ "github.com/go-sql-driver/mysql"
	"github.com/segmentio/kafka-go"
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
	brokers    []string
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
		brokers:    p.brokers,
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

func (p *Data) PingDB(ctx context.Context) error {
	if p == nil || p.db == nil {
		return errors.New("database not configured")
	}
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (p *Data) PingRedis(ctx context.Context) error {
	if p == nil || p.rdb == nil {
		return errors.New("redis not configured")
	}
	_, _, err := p.rdb.Get(ctx, "health:probe")
	return err
}

func (p *Data) PingKafkaProducer(ctx context.Context) error {
	if p == nil || len(p.brokers) == 0 {
		return errors.New("kafka not configured")
	}
	conn, err := (&kafka.Dialer{}).DialContext(ctx, "tcp", p.brokers[0])
	if err != nil {
		return err
	}
	return conn.Close()
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
	dta := &Data{
		db:         db,
		rdb:        cache.GetRedisCli(),
		mqProducer: producer,
		mqConsumer: consumer,
		brokers:    append([]string(nil), dt.Kafka.Producer.Brokers...),
	}
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
