package svc

import (
	"github.com/BitofferHub/gateway/internal/config"
	"github.com/BitofferHub/gateway/limiter"
	"github.com/redis/go-redis/v9"
)

type ServiceContext struct {
	Config  config.Config
	Redis   *redis.Client
	Limiter *limiter.RateLimiter
	*RPCClients
}

func NewServiceContext(c config.Config) *ServiceContext {
	rdb := newRedisClient(c.Redis)
	rl, err := newRateLimiter(c, rdb)
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:     c,
		Redis:      rdb,
		Limiter:    rl,
		RPCClients: newRPCClients(c),
	}
}
