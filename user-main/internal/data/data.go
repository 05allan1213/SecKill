package data

import (
	"github.com/BitofferHub/pkg/middlewares/cache"
	"github.com/BitofferHub/pkg/middlewares/gormcli"
	cfg "github.com/BitofferHub/user/internal/config"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
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
	gormcli.Init(
		gormcli.WithAddr(dt.Database.Addr),
		gormcli.WithUser(dt.Database.User),
		gormcli.WithPassword(dt.Database.Password),
		gormcli.WithDataBase(dt.Database.DataBase),
		gormcli.WithMaxIdleConn(int(dt.Database.MaxIdleConn)),
		gormcli.WithMaxOpenConn(int(dt.Database.MaxOpenConn)),
		gormcli.WithMaxIdleTime(int64(dt.Database.MaxIdleTime)),
	)
	cache.Init(
		cache.WithAddr(dt.Redis.Addr),
		cache.WithPassWord(dt.Redis.PassWord),
		cache.WithDB(int(dt.Redis.Db)),
		cache.WithPoolSize(int(dt.Redis.PoolSize)),
	)
	dta := &Data{db: gormcli.GetDB(), rdb: cache.GetRedisCli()}
	data = dta
	return dta, nil
}
