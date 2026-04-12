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
	Addr            string
	User            string
	Password        string
	DataBase        string
	MaxIdleConn     int32 `json:",default=50"`
	MaxOpenConn     int32 `json:",default=100"`
	MaxIdleTime     int32 `json:",default=300"`        // 空闲连接存活时间(秒)
	ConnMaxLifetime int32 `json:",default=3600"`      // 连接最大存活时间(秒)
}

type RedisConf struct {
	Addr         string `json:",default=127.0.0.1:6379"`
	PassWord     string `json:",default="`
	Db           int32  `json:",default=0"`
	PoolSize     int32  `json:",default=100"`        // 连接池大小
	MinIdleConns int32  `json:",default=10"`        // 最小空闲连接数
	MaxRetries   int32  `json:",default=3"`         // 最大重试次数
	DialTimeout  int32  `json:",default=1000"`      // 连接超时(毫秒)
	ReadTimeout  int32  `json:",default=1000"`      // 读超时(毫秒)
	WriteTimeout int32  `json:",default=1000"`      // 写超时(毫秒)
	PoolTimeout  int32  `json:",default=2000"`      // 连接池超时(毫秒)
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
	Brokers      []string
	Topic        string
	Offset       int64
	GroupID      string `json:",default=seckill-consumer"`
	NumConsumers int32  `json:",default=1"`      // 消费者数量，建议与分区数匹配
}

type KafkaRetryConf struct {
	Topic      string `json:",default=seckill-retry"`
	MaxAttempts int   `json:",default=3"`
	BackoffMs  int    `json:",default=1000"`
}

type KafkaDLQConf struct {
	Topic string `json:",default=seckill-dlq"`
}
