package data

import (
	"fmt"
	"github.com/BitofferHub/pkg/middlewares/cache"
	cfg "github.com/BitofferHub/user/internal/config"
	bitlog "github.com/BitofferHub/user/internal/log"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"time"
)

// Data .
type Data struct {
	db  *gorm.DB
	rdb *cache.Client
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

func NewDataFromConfig(dt cfg.DataConf) (*Data, error) {
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
	dta := &Data{db: db, rdb: cache.GetRedisCli()}
	data = dta
	return dta, nil
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
	return db, nil
}
