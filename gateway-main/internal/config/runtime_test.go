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
	t.Setenv("GATEWAY_TRACE_ENABLED", "true")
	t.Setenv("GATEWAY_TRACE_SAMPLER", "0.25")
	t.Setenv("GATEWAY_LOG_MAX_SIZE_MB", "32")
	t.Setenv("GATEWAY_LOG_MAX_BACKUPS", "5")
	t.Setenv("GATEWAY_LOG_COMPRESS", "false")
	t.Setenv("GATEWAY_ACCESS_SUMMARY_MAX_BYTES", "96")

	cfg := Config{
		Auth:  AuthConf{Secret: "secret-from-config"},
		Redis: RedisConf{PassWord: "redis-from-config"},
	}
	cfg.Telemetry.Sampler = 0.01

	ApplyEnvOverrides(&cfg)
	applyObservability(&cfg)

	if cfg.Auth.Secret != "secret-from-env" {
		t.Fatalf("expected auth secret override, got %q", cfg.Auth.Secret)
	}
	if cfg.Redis.PassWord != "redis-from-env" {
		t.Fatalf("expected redis password override, got %q", cfg.Redis.PassWord)
	}
	if !cfg.Observability.Trace.Enabled {
		t.Fatal("expected trace to be enabled from env")
	}
	if cfg.Telemetry.Sampler != 0.25 {
		t.Fatalf("expected sampler override, got %v", cfg.Telemetry.Sampler)
	}
	if cfg.Observability.LogRotation.MaxSizeMB != 32 {
		t.Fatalf("expected max size override, got %d", cfg.Observability.LogRotation.MaxSizeMB)
	}
	if cfg.Observability.LogRotation.MaxBackups != 5 {
		t.Fatalf("expected max backups override, got %d", cfg.Observability.LogRotation.MaxBackups)
	}
	if cfg.Observability.LogRotation.Compress {
		t.Fatal("expected compress override to be false")
	}
	if cfg.Observability.AccessLog.SummaryMaxBytes != 96 {
		t.Fatalf("expected access summary bytes override, got %d", cfg.Observability.AccessLog.SummaryMaxBytes)
	}
}

func TestApplyObservability_DisablesTrace(t *testing.T) {
	cfg := Config{}
	cfg.Telemetry.Name = "gateway"
	cfg.Telemetry.Endpoint = "logs/trace.json"
	cfg.Telemetry.Batcher = "file"
	cfg.Telemetry.Sampler = 1
	cfg.Middlewares.Trace = true

	applyObservability(&cfg)

	if cfg.Middlewares.Trace {
		t.Fatal("expected trace middleware to be disabled")
	}
	if cfg.Telemetry.Endpoint != "" || cfg.Telemetry.Name != "" || cfg.Telemetry.Batcher != "" || cfg.Telemetry.Sampler != 0 {
		t.Fatalf("expected telemetry to be disabled, got %#v", cfg.Telemetry)
	}
}
