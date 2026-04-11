package config

import (
	"time"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Auth                AuthConf                          `json:",optional"`
	Log                 LogConf                           `json:",optional"`
	Redis               RedisConf                         `json:",optional"`
	LimiterProfile      string                            `json:",default=compare"`
	UserRpc             zrpc.RpcClientConf                `json:",optional"`
	SeckillRpc          zrpc.RpcClientConf                `json:",optional"`
	RoutePolicies       map[string]RoutePolicy            `json:",optional"`
	RoutePolicyProfiles map[string]map[string]RoutePolicy `json:",optional"`
}

type LogConf struct {
	AccessDetail string `json:",default=request"`
}

type AuthConf struct {
	Secret  string        `json:",default=secret key"`
	Timeout time.Duration `json:",default=1h"`
}

type RedisConf struct {
	Addr         string        `json:",optional"`
	PassWord     string        `json:"passWord,optional"`
	DB           int           `json:",default=0"`
	ReadTimeout  time.Duration `json:"read_timeout,default=2s"`
	WriteTimeout time.Duration `json:"write_timeout,default=2s"`
}

type RoutePolicy struct {
	LimitTimeout int    `json:"limit_timeout,default=2000"`
	LimitRate    int    `json:"limit_rate,default=1000"`
	RetryTime    int    `json:"retry_time,default=50"`
	Remarks      string `json:"remarks,optional"`
}
