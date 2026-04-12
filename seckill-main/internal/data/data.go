package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BitofferHub/pkg/middlewares/cache"
	cfg "github.com/BitofferHub/seckill/internal/config"
	bitlog "github.com/BitofferHub/seckill/internal/log"
	"github.com/BitofferHub/seckill/internal/mq"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"sync"
)

// Data .
type Data struct {
	db            *gorm.DB
	rdb           *cache.Client      // cache 包封装的 Redis 客户端
	redisClient   *redis.Client       // 原生 Redis 客户端，用于高性能场景
	mqProducer    mq.Producer
	mqConsumer    []mq.Consumer        // 支持多个消费者
	retryProducer mq.Producer
	dlqProducer   mq.Producer
	retryConsumer []mq.Consumer       // 支持多个 retry 消费者
	brokers       []string
	conf          cfg.DataConf
	closeOnce     sync.Once
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

// GetRedisClient 获取原生 Redis 客户端，用于高性能场景
func (p *Data) GetRedisClient() *redis.Client {
	return p.redisClient
}

func (p *Data) GetMQProducer() mq.Producer {
	return p.mqProducer
}

// GetMQConsumers 获取所有消费者
func (p *Data) GetMQConsumers() []mq.Consumer {
	return p.mqConsumer
}

func (p *Data) GetRetryProducer() mq.Producer {
	return p.retryProducer
}

func (p *Data) GetDLQProducer() mq.Producer {
	return p.dlqProducer
}

// GetRetryConsumers 获取所有 retry 消费者
func (p *Data) GetRetryConsumers() []mq.Consumer {
	return p.retryConsumer
}

func (p *Data) Close() {
	p.closeOnce.Do(func() {
		// 关闭所有主消费者
		for _, c := range p.mqConsumer {
			if c != nil {
				c.Close()
			}
		}
		// 关闭所有 retry 消费者
		for _, c := range p.retryConsumer {
			if c != nil {
				c.Close()
			}
		}
		if p.mqProducer != nil {
			p.mqProducer.Close()
		}
		if p.retryProducer != nil {
			p.retryProducer.Close()
		}
		if p.dlqProducer != nil {
			p.dlqProducer.Close()
		}
		// 关闭原生 Redis 客户端
		if p.redisClient != nil {
			p.redisClient.Close()
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

func (p *Data) IsRedisAvailable(parentCtx context.Context, timeoutMs int) bool {
	if p == nil || p.rdb == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(parentCtx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()
	_, _, err := p.rdb.Get(ctx, "health:probe")
	return err == nil
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

func (p *Data) PingRetryProducer(ctx context.Context) error {
	if p == nil || p.retryProducer == nil {
		return errors.New("retry producer not configured")
	}
	conn, err := (&kafka.Dialer{}).DialContext(ctx, "tcp", p.brokers[0])
	if err != nil {
		return err
	}
	return conn.Close()
}

func (p *Data) PingDLQProducer(ctx context.Context) error {
	if p == nil || p.dlqProducer == nil {
		return errors.New("dlq producer not configured")
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

	// 初始化 cache 包（保持兼容性）
	cache.Init(
		cache.WithAddr(dt.Redis.Addr),
		cache.WithPassWord(dt.Redis.PassWord),
		cache.WithDB(int(dt.Redis.Db)),
		cache.WithPoolSize(int(dt.Redis.PoolSize)),
	)

	// 初始化原生 Redis 客户端，支持完整配置
	redisClient, err := initRedis(dt.Redis)
	if err != nil {
		return nil, fmt.Errorf("init redis client failed: %w", err)
	}

	producer := mq.NewKafkaProducer(
		mq.WithBrokers(dt.Kafka.Producer.Brokers),
		mq.WithTopic(dt.Kafka.Producer.Topic),
		mq.WithAck(dt.Kafka.Producer.Ack),
	)
	if producer == nil {
		panic("nil producer")
	}

	var retryProducer, dlqProducer mq.Producer
	var retryConsumers []mq.Consumer
	if dt.Kafka.Retry.Topic != "" {
		retryProducer = mq.NewKafkaProducer(
			mq.WithBrokers(dt.Kafka.Producer.Brokers),
			mq.WithTopic(dt.Kafka.Retry.Topic),
			mq.WithAck(1),
		)
		retryConsumerConf := dt.Kafka.Consumer
		retryConsumerConf.Topic = dt.Kafka.Retry.Topic
		retryConsumerConf.GroupID = dt.Kafka.Consumer.GroupID + "-retry"
		// 创建多个 retry 消费者
		numRetryConsumers := dt.Kafka.Consumer.NumConsumers
		if numRetryConsumers <= 0 {
			numRetryConsumers = 1
		}
		for i := int32(0); i < numRetryConsumers; i++ {
			retryConsumers = append(retryConsumers, newManagedKafkaConsumer(retryConsumerConf))
		}
	}
	if dt.Kafka.DLQ.Topic != "" {
		dlqProducer = mq.NewKafkaProducer(
			mq.WithBrokers(dt.Kafka.Producer.Brokers),
			mq.WithTopic(dt.Kafka.DLQ.Topic),
			mq.WithAck(1),
		)
	}

	// 创建多个主消费者，数量与分区数匹配
	numConsumers := dt.Kafka.Consumer.NumConsumers
	if numConsumers <= 0 {
		numConsumers = 1
	}
	var consumers []mq.Consumer
	for i := int32(0); i < numConsumers; i++ {
		consumers = append(consumers, newManagedKafkaConsumer(dt.Kafka.Consumer))
	}

	dta := &Data{
		db:            db,
		rdb:           cache.GetRedisCli(),
		redisClient:   redisClient,
		mqProducer:    producer,
		mqConsumer:    consumers,
		retryProducer: retryProducer,
		dlqProducer:   dlqProducer,
		retryConsumer: retryConsumers,
		brokers:       append([]string(nil), dt.Kafka.Producer.Brokers...),
		conf:          dt,
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
	// MaxIdleTime: 空闲连接存活时间
	sqlDB.SetConnMaxLifetime(time.Duration(conf.MaxIdleTime) * time.Second)
	// ConnMaxLifetime: 连接最大存活时间，避免长时间使用同一连接
	if conf.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second)
	}
	return db, nil
}

// initRedis 初始化 Redis 客户端，支持完整配置选项
func initRedis(conf cfg.RedisConf) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:         conf.Addr,
		Password:     conf.PassWord,
		DB:           int(conf.Db),
		PoolSize:     int(conf.PoolSize),
		MinIdleConns: int(conf.MinIdleConns),
		MaxRetries:   int(conf.MaxRetries),
	}

	// 设置超时参数（配置中为毫秒，需要转换为 Duration）
	if conf.DialTimeout > 0 {
		opts.DialTimeout = time.Duration(conf.DialTimeout) * time.Millisecond
	}
	if conf.ReadTimeout > 0 {
		opts.ReadTimeout = time.Duration(conf.ReadTimeout) * time.Millisecond
	}
	if conf.WriteTimeout > 0 {
		opts.WriteTimeout = time.Duration(conf.WriteTimeout) * time.Millisecond
	}
	if conf.PoolTimeout > 0 {
		opts.PoolTimeout = time.Duration(conf.PoolTimeout) * time.Millisecond
	}

	client := redis.NewClient(opts)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return client, nil
}
