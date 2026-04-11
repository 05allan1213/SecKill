package limiter

import (
	"context"
	"crypto/tls"

	gwlog "github.com/BitofferHub/gateway/internal/log"
	"github.com/BitofferHub/gateway/limiter/tb"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
	"time"
)

type RateLimiterConfig struct {
	Routes              map[string]RoutePolicy
	DefaultLimitRate    int
	DefaultRetryTime    int
	DefaultLimitTimeout int
}

type RoutePolicy struct {
	LimitTimeout int
	LimitRate    int
	RetryTime    int
	Remarks      string
}

var tbLimiter *tb.TBLimiter

var routeLimits map[string]*tb.TBLimit
var localRouteLimits map[string]*rate.Limiter

func NewRateLimiter(config RateLimiterConfig, redisClient *redis.Client) (*RateLimiter, error) {
	rateLimiter := &RateLimiter{
		RateLimiterConfig: config,
	}
	var err error
	ctx := context.Background()
	tbLimiter, err = tb.NewTBLimiter(ctx, redisClient)
	if err != nil {
		gwlog.Error(ctx, "rate limiter init failed",
			gwlog.Field(gwlog.FieldAction, "rate_limit.init"),
			gwlog.Field(gwlog.FieldError, err.Error()),
		)
		return nil, err
	}
	// 为每个接口生成限流配置
	routeLimits = make(map[string]*tb.TBLimit)
	localRouteLimits = make(map[string]*rate.Limiter)
	for k, v := range config.Routes {
		if v.LimitRate == 0 {
			routeLimits[k] = &tb.TBLimit{Burst: config.DefaultLimitRate, Rate: config.DefaultLimitRate, Expire: 10}
			localLimit := rate.Limit(config.DefaultLimitRate)
			localRouteLimits[k] = rate.NewLimiter(localLimit, 1) // 把令牌桶的大小设置为 1，把令牌桶当做漏桶来使用
		} else {
			routeLimits[k] = &tb.TBLimit{Burst: v.LimitRate, Rate: v.LimitRate, Expire: 10}
			localLimit := rate.Limit(v.LimitRate)
			localRouteLimits[k] = rate.NewLimiter(localLimit, 1) // 把令牌桶的大小设置为 1，把令牌桶当做漏桶来使用
		}
	}
	return rateLimiter, nil
}

func (r *RateLimiter) Allow(ctx context.Context, url string) (*Result, error) {
	res, err := tbLimiter.Allow(ctx, url, routeLimits[url])
	result := &Result{}

	if err != nil {
		gwlog.Warn(ctx, "rate limiter redis fallback",
			gwlog.Field(gwlog.FieldAction, "rate_limit.redis_fallback"),
			gwlog.Field(gwlog.FieldRouteKey, url),
			gwlog.Field(gwlog.FieldError, err.Error()),
		)
		// 如果出错，使用本地限流器
		localLimit := localRouteLimits[url]
		if localLimit == nil {
			// 如果限流器不存在，直接通过
			gwlog.Warn(ctx, "local rate limiter missing",
				gwlog.Field(gwlog.FieldAction, "rate_limit.local_missing"),
				gwlog.Field(gwlog.FieldRouteKey, url),
			)
			result.IsAllowed = true
			return result, nil
		}
		if localLimit.Allow() {
			result.IsAllowed = true
			return result, nil
		} else {
			result.IsAllowed = false
			return result, nil
		}
	}

	if res.Allowed > 0 {
		result.IsAllowed = true
		return result, nil
	} else {
		result.IsAllowed = false
		return result, nil
	}
}

type RedisConfig struct {
	Addr               string
	Password           string
	DB                 int
	MaxRetries         int
	MinRetryBackoff    time.Duration
	MaxRetryBackoff    time.Duration
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	PoolSize           int
	MinIdleConns       int
	MaxConnAge         time.Duration
	PoolTimeout        time.Duration
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration
	readOnly           bool
	TLSConfig          *tls.Config
}

type RateLimiter struct {
	RateLimiterConfig RateLimiterConfig
}

type Result struct {
	IsAllowed bool
	IsTimeout bool
}

type Limiter interface {
	Allow(ctx context.Context, url string) (*Result, error)
}
