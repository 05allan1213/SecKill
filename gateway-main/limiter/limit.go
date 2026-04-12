package limiter

import (
	"context"
	"crypto/tls"
	"errors"

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
var newTokenBucketLimiter = tb.NewTBLimiter

func NewRateLimiter(config RateLimiterConfig, redisClient *redis.Client) (*RateLimiter, error) {
	rateLimiter := &RateLimiter{
		RateLimiterConfig: config,
	}
	var err error
	ctx := context.Background()
	if redisClient == nil {
		return nil, errors.New("redis client is nil")
	}
	tbLimiter, err = newTokenBucketLimiter(ctx, redisClient)
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
	result := &Result{}
	if r == nil {
		gwlog.WarnEvery(ctx, "gateway.rate_limit.nil", 5*time.Second, "rate limiter is nil, allow request",
			gwlog.Field(gwlog.FieldAction, "rate_limit.nil"),
			gwlog.Field(gwlog.FieldRouteKey, url),
		)
		result.IsAllowed = true
		return result, nil
	}

	limit := routeLimits[url]
	if limit == nil {
		gwlog.WarnEvery(ctx, "gateway.rate_limit.route_missing", 5*time.Second, "route rate limiter config missing, allow request",
			gwlog.Field(gwlog.FieldAction, "rate_limit.route_missing"),
			gwlog.Field(gwlog.FieldRouteKey, url),
		)
		result.IsAllowed = true
		return result, nil
	}

	if tbLimiter == nil {
		gwlog.WarnEvery(ctx, "gateway.rate_limit.redis_unavailable", 5*time.Second, "token bucket limiter unavailable, fallback to local limiter",
			gwlog.Field(gwlog.FieldAction, "rate_limit.redis_unavailable"),
			gwlog.Field(gwlog.FieldRouteKey, url),
		)
		return allowWithLocalFallback(ctx, url), nil
	}

	res, err := tbLimiter.Allow(ctx, url, limit)

	if err != nil {
		gwlog.WarnEvery(ctx, "gateway.rate_limit.redis_fallback", 5*time.Second, "rate limiter redis fallback",
			gwlog.Field(gwlog.FieldAction, "rate_limit.redis_fallback"),
			gwlog.Field(gwlog.FieldRouteKey, url),
			gwlog.Field(gwlog.FieldError, err.Error()),
		)
		return allowWithLocalFallback(ctx, url), nil
	}

	if res.Allowed > 0 {
		result.IsAllowed = true
		return result, nil
	} else {
		result.IsAllowed = false
		return result, nil
	}
}

func allowWithLocalFallback(ctx context.Context, url string) *Result {
	result := &Result{}
	localLimit := localRouteLimits[url]
	if localLimit == nil {
		gwlog.WarnEvery(ctx, "gateway.rate_limit.local_missing", 5*time.Second, "local rate limiter missing, allow request",
			gwlog.Field(gwlog.FieldAction, "rate_limit.local_missing"),
			gwlog.Field(gwlog.FieldRouteKey, url),
		)
		result.IsAllowed = true
		return result
	}

	result.IsAllowed = localLimit.Allow()
	return result
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
