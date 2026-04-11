package config

import (
	"testing"

	"github.com/zeromicro/go-zero/zrpc"
)

func TestApplyRuntimeConfig(t *testing.T) {
	cfg := Config{
		Auth:       AuthConf{Secret: "old-secret"},
		Redis:      RedisConf{Addr: "redis-old"},
		UserRpc:    zrpc.RpcClientConf{Timeout: 1111},
		SeckillRpc: zrpc.RpcClientConf{Timeout: 2222},
		RoutePolicies: map[string]RoutePolicy{
			"/old": {LimitRate: 1},
		},
	}

	userRPC := zrpc.RpcClientConf{Timeout: 3333}
	seckillRPC := zrpc.RpcClientConf{Timeout: 4444}
	ApplyRuntimeConfig(&cfg, RuntimeConfig{
		Auth:       &AuthConf{Secret: "new-secret"},
		Redis:      &RedisConf{Addr: "redis-new"},
		UserRpc:    &userRPC,
		SeckillRpc: &seckillRPC,
		RoutePolicies: map[string]RoutePolicy{
			"/new": {LimitRate: 2},
		},
	})

	if cfg.Auth.Secret != "new-secret" {
		t.Fatalf("expected auth override, got %q", cfg.Auth.Secret)
	}
	if cfg.Redis.Addr != "redis-new" {
		t.Fatalf("expected redis override, got %q", cfg.Redis.Addr)
	}
	if cfg.UserRpc.Timeout != 3333 {
		t.Fatalf("expected user rpc timeout override, got %d", cfg.UserRpc.Timeout)
	}
	if cfg.SeckillRpc.Timeout != 4444 {
		t.Fatalf("expected seckill rpc timeout override, got %d", cfg.SeckillRpc.Timeout)
	}
	if _, ok := cfg.RoutePolicies["/new"]; !ok {
		t.Fatalf("expected route policies override, got %#v", cfg.RoutePolicies)
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	t.Setenv("GATEWAY_AUTH_SECRET", "secret-from-env")
	t.Setenv("GATEWAY_REDIS_PASSWORD", "redis-from-env")

	cfg := Config{
		Auth:  AuthConf{Secret: "secret-from-config"},
		Redis: RedisConf{PassWord: "redis-from-config"},
	}

	ApplyEnvOverrides(&cfg)

	if cfg.Auth.Secret != "secret-from-env" {
		t.Fatalf("expected auth secret override, got %q", cfg.Auth.Secret)
	}
	if cfg.Redis.PassWord != "redis-from-env" {
		t.Fatalf("expected redis password override, got %q", cfg.Redis.PassWord)
	}
}
