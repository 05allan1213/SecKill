package limiter

import (
	"encoding/json"
	"os"

	gwlog "github.com/BitofferHub/gateway/internal/log"
	"github.com/redis/go-redis/v9"
)

var Rl *RateLimiter

func InitLimiter(routeConfigPath string, redisClient *redis.Client,
	defaultRetryTime, defaultLimitTimeout, defaultLimitRate int) error {

	routes, err := os.ReadFile(routeConfigPath)
	if err != nil {
		gwlog.Error(nil, "read limiter config failed",
			gwlog.Field(gwlog.FieldAction, "limiter.init"),
			gwlog.Field(gwlog.FieldError, err.Error()),
		)
		return err
	}
	routePolicies := make(map[string]RoutePolicy, 0)
	err = json.Unmarshal(routes, &routePolicies)
	if err != nil {
		gwlog.Error(nil, "parse limiter config failed",
			gwlog.Field(gwlog.FieldAction, "limiter.init"),
			gwlog.Field(gwlog.FieldError, err.Error()),
		)
		return err
	}

	rateLimiterConfig := RateLimiterConfig{
		Routes:              routePolicies,
		DefaultRetryTime:    defaultRetryTime,
		DefaultLimitTimeout: defaultLimitTimeout,
		DefaultLimitRate:    defaultLimitRate,
	}
	rl, err := NewRateLimiter(rateLimiterConfig, redisClient)

	if err != nil {
		gwlog.Error(nil, "create rate limiter failed",
			gwlog.Field(gwlog.FieldAction, "limiter.init"),
			gwlog.Field(gwlog.FieldError, err.Error()),
		)
		return err
	}
	Rl = rl
	return nil
}
