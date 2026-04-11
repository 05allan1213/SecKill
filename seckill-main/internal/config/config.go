package config

import (
	"time"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Data         DataConf
	ConfigCenter ConfigCenterConf `json:",optional"`
}

type ConfigCenterConf struct {
	Enabled     bool          `json:",default=false"`
	Endpoints   []string      `json:",optional"`
	Key         string        `json:",optional"`
	Watch       bool          `json:",default=false"`
	GracePeriod time.Duration `json:",default=30s"`
}

type DataConf struct {
	Database DatabaseConf
	Redis    RedisConf
	Kafka    KafkaConf
}

type DatabaseConf struct {
	Addr        string
	User        string
	Password    string
	DataBase    string
	MaxIdleConn int32
	MaxOpenConn int32
	MaxIdleTime int32
}

type RedisConf struct {
	Addr     string
	PassWord string
	Db       int32
	PoolSize int32
}

type KafkaConf struct {
	Producer KafkaProducerConf
	Consumer KafkaConsumerConf
}

type KafkaProducerConf struct {
	Brokers []string
	Topic   string
	Ack     int8
}

type KafkaConsumerConf struct {
	Brokers []string
	Topic   string
	Offset  int64
}

type RuntimeConfig struct {
	Data *DataConf `json:",optional"`
}
