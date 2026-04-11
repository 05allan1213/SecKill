package model

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/BitofferHub/pkg/middlewares/cache"
	cfg "github.com/BitofferHub/seckill/internal/config"
	bitlog "github.com/BitofferHub/seckill/internal/log"
	"github.com/BitofferHub/seckill/internal/mq"
	_ "github.com/go-sql-driver/mysql"
	"github.com/segmentio/kafka-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Store struct {
	db         *gorm.DB
	rdb        *cache.Client
	mqProducer mq.Producer
	mqConsumer mq.Consumer
	closeOnce  sync.Once
}

func NewStoreFromConfig(dt cfg.DataConf) (*Store, error) {
	db, err := openDB(dt.Database, 10)
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
		mq.WithAsync(),
	)
	if producer == nil {
		panic("nil producer")
	}

	return &Store{
		db:         db,
		rdb:        cache.GetRedisCli(),
		mqProducer: producer,
		mqConsumer: newManagedKafkaConsumer(dt.Kafka.Consumer),
	}, nil
}

func (s *Store) GetDB() *gorm.DB {
	return s.db
}

func (s *Store) GetCache() *cache.Client {
	return s.rdb
}

func (s *Store) GetMQProducer() mq.Producer {
	return s.mqProducer
}

func (s *Store) GetMQConsumer() mq.Consumer {
	return s.mqConsumer
}

func (s *Store) CloneWithDB(db *gorm.DB) *Store {
	if db == nil {
		db = s.db
	}
	return &Store{
		db:         db,
		rdb:        s.rdb,
		mqProducer: s.mqProducer,
		mqConsumer: s.mqConsumer,
	}
}

func (s *Store) RunInTx(ctx context.Context, fn func(txStore *Store) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(s.CloneWithDB(tx))
	})
}

func (s *Store) Close() {
	s.closeOnce.Do(func() {
		if s.mqConsumer != nil {
			s.mqConsumer.Close()
		}
		if s.mqProducer != nil {
			s.mqProducer.Close()
		}
		if s.db != nil {
			sqlDB, err := s.db.DB()
			if err == nil {
				_ = sqlDB.Close()
			}
		}
	})
}

func openDB(conf cfg.DatabaseConf, slowThresholdMillisecond int64) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.User, conf.Password, conf.Addr, conf.DataBase)

	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}
	if slowThresholdMillisecond > 0 {
		gormCfg.Logger = bitlog.NewGormLogger(slowThresholdMillisecond)
	}

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
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)
	return db, nil
}

type managedKafkaConsumer struct {
	reader    *kafka.Reader
	closed    atomic.Bool
	closeOnce sync.Once
}

func newManagedKafkaConsumer(conf cfg.KafkaConsumerConf) mq.Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: conf.Brokers,
		Topic:   conf.Topic,
	})
	reader.SetOffset(conf.Offset)
	return &managedKafkaConsumer{reader: reader}
}

func (c *managedKafkaConsumer) ConsumeMessages(ctx context.Context, handler func(context.Context, []byte) error) {
	if ctx == nil {
		bitlog.Errorf("consume kafka messages failed: nil context")
		return
	}
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if c.closed.Load() {
				return
			}
			bitlog.ErrorContextw(ctx, "fetch kafka message failed", bitlogField("err", err))
			continue
		}

		if err := handler(ctx, message.Value); err != nil {
			bitlog.ErrorContextw(ctx, "handle kafka message failed",
				bitlogField("topic", message.Topic),
				bitlogField("partition", message.Partition),
				bitlogField("offset", message.Offset),
				bitlogField("err", err))
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			if c.closed.Load() {
				return
			}
			bitlog.ErrorContextw(ctx, "commit kafka message failed",
				bitlogField("topic", message.Topic),
				bitlogField("partition", message.Partition),
				bitlogField("offset", message.Offset),
				bitlogField("err", err))
		}
	}
}

func (c *managedKafkaConsumer) Close() {
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		if err := c.reader.Close(); err != nil {
			bitlog.Errorf("close kafka consumer failed: %v", err)
		}
	})
}
