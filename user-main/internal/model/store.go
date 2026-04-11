package model

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/BitofferHub/pkg/middlewares/cache"
	cfg "github.com/BitofferHub/user/internal/config"
	bitlog "github.com/BitofferHub/user/internal/log"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Store struct {
	db        *gorm.DB
	rdb       *cache.Client
	closeOnce sync.Once
}

func NewStoreFromConfig(dt cfg.DataConf) (*Store, error) {
	db, err := openDB(dt.Database, 0)
	if err != nil {
		return nil, err
	}

	cache.Init(
		cache.WithAddr(dt.Redis.Addr),
		cache.WithPassWord(dt.Redis.PassWord),
		cache.WithDB(int(dt.Redis.Db)),
		cache.WithPoolSize(int(dt.Redis.PoolSize)),
	)

	return &Store{
		db:  db,
		rdb: cache.GetRedisCli(),
	}, nil
}

func (s *Store) DB() *gorm.DB {
	return s.db
}

func (s *Store) Cache() *cache.Client {
	return s.rdb
}

func (s *Store) CloneWithDB(db *gorm.DB) *Store {
	if db == nil {
		db = s.db
	}
	return &Store{
		db:  db,
		rdb: s.rdb,
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
		if s.db == nil {
			return
		}
		sqlDB, err := s.db.DB()
		if err != nil {
			return
		}
		_ = sqlDB.Close()
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
