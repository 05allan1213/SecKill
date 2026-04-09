package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Data DataConf
}

type DataConf struct {
	Database DatabaseConf
	Redis    RedisConf
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
