package svc

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/BitofferHub/gateway/internal/config"
	"github.com/BitofferHub/gateway/limiter"
	"github.com/BitofferHub/pkg/constant"
	secproto "github.com/BitofferHub/seckill/api/sec_kill/proto"
	userv1 "github.com/BitofferHub/user/api/user/v1"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ServiceContext struct {
	Config       config.Config
	current      atomic.Pointer[serviceBundle]
	watchContext context.CancelFunc
	watcher      *config.RuntimeWatcher
	gracePeriod  time.Duration
}

func NewServiceContext(c config.Config) *ServiceContext {
	bundle, err := newServiceBundle(c)
	if err != nil {
		panic(err)
	}

	svc := &ServiceContext{
		Config:      c,
		gracePeriod: c.ConfigCenter.GracePeriod,
	}
	svc.current.Store(bundle)

	if c.ConfigCenter.Enabled && c.ConfigCenter.Watch {
		ctx, cancel := context.WithCancel(context.Background())
		svc.watchContext = cancel
		watcher, err := config.WatchRuntimeConfig(ctx, c.ConfigCenter, func(runtimeCfg config.RuntimeConfig) {
			next := c
			config.ApplyRuntimeConfig(&next, runtimeCfg)
			config.ApplyEnvOverrides(&next)
			svc.reload(next)
		})
		if err != nil {
			panic(err)
		}
		svc.watcher = watcher
	}

	return svc
}

func (s *ServiceContext) Close() {
	if s.watchContext != nil {
		s.watchContext()
	}
	if s.watcher != nil {
		s.watcher.Close()
	}
	if bundle := s.current.Load(); bundle != nil {
		bundle.Close()
	}
}

func (s *ServiceContext) UserClient() userv1.UserClient {
	return s.currentBundle().UserClient
}

func (s *ServiceContext) SeckillClient() secproto.SecKillClient {
	return s.currentBundle().SeckillClient
}

func (s *ServiceContext) Limiter() *limiter.RateLimiter {
	return s.currentBundle().Limiter
}

func (s *ServiceContext) AuthConfig() config.AuthConf {
	return s.currentBundle().Auth
}

func (s *ServiceContext) currentBundle() *serviceBundle {
	bundle := s.current.Load()
	if bundle == nil {
		panic("service bundle is not initialized")
	}
	return bundle
}

func (s *ServiceContext) reload(next config.Config) {
	bundle, err := newServiceBundle(next)
	if err != nil {
		logx.Errorf("reload gateway runtime config failed: %v", err)
		return
	}

	prev := s.current.Swap(bundle)
	if prev == nil {
		return
	}

	go func(old *serviceBundle) {
		time.Sleep(s.gracePeriod)
		old.Close()
	}(prev)
}

type serviceBundle struct {
	Auth          config.AuthConf
	Redis         *redis.Client
	Limiter       *limiter.RateLimiter
	UserClient    userv1.UserClient
	SeckillClient secproto.SecKillClient
	userRPC       zrpc.Client
	seckillRPC    zrpc.Client
}

func newServiceBundle(c config.Config) (*serviceBundle, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.PassWord,
		DB:           c.Redis.DB,
		ReadTimeout:  c.Redis.ReadTimeout,
		WriteTimeout: c.Redis.WriteTimeout,
	})

	routePolicies := make(map[string]limiter.RoutePolicy, len(c.RoutePolicies))
	for path, policy := range c.RoutePolicies {
		routePolicies[path] = limiter.RoutePolicy{
			LimitTimeout: policy.LimitTimeout,
			LimitRate:    policy.LimitRate,
			RetryTime:    policy.RetryTime,
			Remarks:      policy.Remarks,
		}
	}

	rl, err := limiter.NewRateLimiter(limiter.RateLimiterConfig{
		Routes:              routePolicies,
		DefaultRetryTime:    50,
		DefaultLimitTimeout: 2000,
		DefaultLimitRate:    1000,
	}, rdb)
	if err != nil {
		_ = rdb.Close()
		return nil, err
	}

	userRPC := zrpc.MustNewClient(c.UserRpc,
		zrpc.WithUnaryClientInterceptor(newTraceForwardInterceptor()),
	)
	seckillRPC := zrpc.MustNewClient(c.SeckillRpc,
		zrpc.WithUnaryClientInterceptor(newTraceForwardInterceptor()),
	)

	return &serviceBundle{
		Auth:          c.Auth,
		Redis:         rdb,
		Limiter:       rl,
		UserClient:    userv1.NewUserClient(userRPC.Conn()),
		SeckillClient: secproto.NewSecKillClient(seckillRPC.Conn()),
		userRPC:       userRPC,
		seckillRPC:    seckillRPC,
	}, nil
}

func (b *serviceBundle) Close() {
	if b == nil {
		return
	}
	if b.userRPC != nil && b.userRPC.Conn() != nil {
		_ = b.userRPC.Conn().Close()
	}
	if b.seckillRPC != nil && b.seckillRPC.Conn() != nil {
		_ = b.seckillRPC.Conn().Close()
	}
	if b.Redis != nil {
		_ = b.Redis.Close()
	}
}

func newTraceForwardInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if traceID, ok := ctx.Value(constant.TraceID).(string); ok && traceID != "" {
			lowerKey := strings.ToLower(constant.TraceID)
			ctx = metadata.AppendToOutgoingContext(ctx,
				constant.TraceID, traceID,
				lowerKey, traceID,
				"traceid", traceID,
				"x-md-global-"+lowerKey, traceID,
				"x-md-global-traceid", traceID,
			)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
