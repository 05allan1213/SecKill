package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BitofferHub/pkg/middlewares/cache"
	cfg "github.com/BitofferHub/user/internal/config"
	bitlog "github.com/BitofferHub/user/internal/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Data .
type Data struct {
	db          *gorm.DB
	rdb         *cache.Client    // cache 包封装的 Redis 客户端
	redisClient *redis.Client    // 原生 Redis 客户端，用于高性能场景
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

func (p *Data) CloneWithDB(db *gorm.DB) *Data {
	if db == nil {
		db = p.db
	}
	return &Data{
		db:  db,
		rdb: p.rdb,
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

	dta := &Data{db: db, rdb: cache.GetRedisCli(), redisClient: redisClient}
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
