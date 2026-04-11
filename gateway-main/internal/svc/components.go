package svc

import (
	"context"
	"strings"

	"github.com/BitofferHub/gateway/internal/config"
	"github.com/BitofferHub/gateway/limiter"
	"github.com/BitofferHub/pkg/constant"
	secproto "github.com/BitofferHub/seckill/api/sec_kill/proto"
	userv1 "github.com/BitofferHub/user/api/user/v1"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type RPCClients struct {
	UserClient    userv1.UserClient
	SeckillClient secproto.SecKillClient
}

func newRedisClient(c config.RedisConf) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         c.Addr,
		Password:     c.PassWord,
		DB:           c.DB,
		ReadTimeout:  c.ReadTimeout,
		WriteTimeout: c.WriteTimeout,
	})
}

func newRateLimiter(c config.Config, rdb *redis.Client) (*limiter.RateLimiter, error) {
	selectedPolicies := c.RoutePolicies
	if profilePolicies, ok := c.RoutePolicyProfiles[c.LimiterProfile]; ok && len(profilePolicies) > 0 {
		selectedPolicies = profilePolicies
	}

	routePolicies := make(map[string]limiter.RoutePolicy, len(selectedPolicies))
	for path, policy := range selectedPolicies {
		routePolicies[path] = limiter.RoutePolicy{
			LimitTimeout: policy.LimitTimeout,
			LimitRate:    policy.LimitRate,
			RetryTime:    policy.RetryTime,
			Remarks:      policy.Remarks,
		}
	}

	return limiter.NewRateLimiter(limiter.RateLimiterConfig{
		Routes:              routePolicies,
		DefaultRetryTime:    50,
		DefaultLimitTimeout: 2000,
		DefaultLimitRate:    1000,
	}, rdb)
}

func newRPCClients(c config.Config) *RPCClients {
	userConn := zrpc.MustNewClient(c.UserRpc,
		zrpc.WithUnaryClientInterceptor(newTraceForwardInterceptor()),
	).Conn()
	seckillConn := zrpc.MustNewClient(c.SeckillRpc,
		zrpc.WithUnaryClientInterceptor(newTraceForwardInterceptor()),
	).Conn()

	return &RPCClients{
		UserClient:    userv1.NewUserClient(userConn),
		SeckillClient: secproto.NewSecKillClient(seckillConn),
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
