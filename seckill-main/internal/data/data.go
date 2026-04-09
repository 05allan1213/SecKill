package data

import (
	"github.com/BitofferHub/pkg/middlewares/cache"
	"github.com/BitofferHub/pkg/middlewares/gormcli"
	"github.com/BitofferHub/pkg/middlewares/mq"
	cfg "github.com/BitofferHub/seckill/internal/config"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"sync"
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

func NewDataFromConfig(dt cfg.DataConf) (*Data, error) {
	gormcli.Init(
		gormcli.WithAddr(dt.Database.Addr),
		gormcli.WithUser(dt.Database.User),
		gormcli.WithPassword(dt.Database.Password),
		gormcli.WithDataBase(dt.Database.DataBase),
		gormcli.WithMaxIdleConn(int(dt.Database.MaxIdleConn)),
		gormcli.WithMaxOpenConn(int(dt.Database.MaxOpenConn)),
		gormcli.WithMaxIdleTime(int64(dt.Database.MaxIdleTime)),
		gormcli.WithSlowThresholdMillisecond(10),
	)
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
	consumer := newManagedKafkaConsumer(dt.Kafka.Consumer)
	dta := &Data{db: gormcli.GetDB(), rdb: cache.GetRedisCli(), mqProducer: producer, mqConsumer: consumer}
	data = dta
	return dta, nil
}
