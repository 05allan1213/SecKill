package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	HealthProbe HealthProbeConf
	Data        DataConf
	Log         LogConf
}

type HealthProbeConf struct {
	Host string `json:",default=127.0.0.1"`
	Port int    `json:",default=8670"`
}

type LogConf struct {
	AccessDetail       string `json:",default=request"`
	SQLMode            string `json:",default=slow"`
	SQLSlowThresholdMs int64  `json:",default=100"`
}

type DataConf struct {
	Database DatabaseConf
	Redis    RedisConf
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
