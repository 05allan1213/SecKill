package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	HealthProbe HealthProbeConf
	Data        DataConf
	Log         LogConf
	Fallback    FallbackConf
}

type HealthProbeConf struct {
	Host string `json:",default=127.0.0.1"`
	Port int    `json:",default=8003"`
}

type LogConf struct {
	AccessDetail       string `json:",default=request"`
	SQLMode            string `json:",default=slow"`
	SQLSlowThresholdMs int64  `json:",default=100"`
}

type FallbackConf struct {
	Enabled   bool `json:",default=true"`
	TimeoutMs int  `json:",default=500"`
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
	Retry    KafkaRetryConf
	DLQ      KafkaDLQConf
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
	GroupID string `json:",default=seckill-consumer"`
}

type KafkaRetryConf struct {
	Topic      string `json:",default=seckill-retry"`
	MaxAttempts int   `json:",default=3"`
	BackoffMs  int    `json:",default=1000"`
}

type KafkaDLQConf struct {
	Topic string `json:",default=seckill-dlq"`
}
