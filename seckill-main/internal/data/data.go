package data

import (
	"fmt"
	"github.com/BitofferHub/pkg/middlewares/cache"
	"github.com/BitofferHub/pkg/middlewares/gormcli"
	"github.com/BitofferHub/pkg/middlewares/mq"
	"github.com/BitofferHub/seckill/internal/conf"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// Data .
type Data struct {
	db         *gorm.DB
	rdb        *cache.Client
	mqProducer mq.Producer
	mqConsumer mq.Consumer
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

// NewData
//
//	@Author <a href="https://bitoffer.cn">狂飙训练营</a>
//	@Description:
//	@param dt
//	@return *Data
//	@return error
func NewData(dt *conf.Data) (*Data, error) {
	fmt.Printf("conf is %+v\n", dt)
	gormcli.Init(
		gormcli.WithAddr(dt.GetDatabase().GetAddr()),
		gormcli.WithUser(dt.GetDatabase().GetUser()),
		gormcli.WithPassword(dt.GetDatabase().GetPassword()),
		gormcli.WithDataBase(dt.GetDatabase().GetDataBase()),
		gormcli.WithMaxIdleConn(int(dt.GetDatabase().GetMaxIdleConn())),
		gormcli.WithMaxOpenConn(int(dt.GetDatabase().GetMaxOpenConn())),
		gormcli.WithMaxIdleTime(int64(dt.GetDatabase().GetMaxIdleTime())),
		gormcli.WithSlowThresholdMillisecond(10),
	)
	cache.Init(
		cache.WithAddr(dt.GetRedis().Addr),
		cache.WithPassWord(dt.GetRedis().PassWord),
		cache.WithDB(int(dt.GetRedis().Db)),
		cache.WithPoolSize(int(dt.GetRedis().PoolSize)))
	producer := mq.NewKafkaProducer(
		mq.WithBrokers(dt.GetKafka().GetProducer().Brokers),
		mq.WithTopic(dt.GetKafka().GetProducer().Topic),
		mq.WithAck(int8(dt.GetKafka().GetProducer().Ack)),
		mq.WithAsync())
	if producer == nil {
		panic("nil producer")
	}
	consumer := mq.NewKafkaConsumer(
		mq.WithBrokers(dt.GetKafka().GetConsumer().Brokers),
		mq.WithTopic(dt.GetKafka().GetConsumer().Topic),
		mq.WithOffset(dt.GetKafka().GetConsumer().Offset))
	if consumer == nil {
		panic("nil consumer")
	}
	fmt.Println("producer ", producer)
	dta := &Data{db: gormcli.GetDB(), rdb: cache.GetRedisCli(), mqProducer: producer, mqConsumer: consumer}
	data = dta
	fmt.Println("producer 2", data.GetMQProducer())
	fmt.Printf("data is %+v\n", data)
	fmt.Printf("data db is %+v\n", data.GetDB())
	return dta, nil
}
