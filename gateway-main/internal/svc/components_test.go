package svc

import (
	"testing"

	"github.com/BitofferHub/gateway/internal/config"
	"github.com/BitofferHub/gateway/limiter"
	"github.com/redis/go-redis/v9"
)

func TestNewRateLimiterCompareProfile(t *testing.T) {
	original := newGatewayLimiter
	t.Cleanup(func() {
		newGatewayLimiter = original
	})

	var captured limiter.RateLimiterConfig
	newGatewayLimiter = func(cfg limiter.RateLimiterConfig, _ *redis.Client) (*limiter.RateLimiter, error) {
		captured = cfg
		return &limiter.RateLimiter{}, nil
	}

	rl, err := newRateLimiter(config.Config{
		LimiterProfile: "compare",
		RoutePolicies: map[string]config.RoutePolicy{
			"/fallback": {LimitRate: 1},
		},
		RoutePolicyProfiles: map[string]map[string]config.RoutePolicy{
			"compare": {
				"/compare": {LimitRate: 100, RetryTime: 50, LimitTimeout: 2000, Remarks: "compare"},
			},
			"protect": {
				"/protect": {LimitRate: 10, RetryTime: 50, LimitTimeout: 2000, Remarks: "protect"},
			},
		},
	}, nil)
	if err != nil {
		t.Fatalf("newRateLimiter returned error: %v", err)
	}
	if rl == nil {
		t.Fatal("expected limiter for compare profile")
	}
	if len(captured.Routes) != 1 {
		t.Fatalf("expected 1 selected route, got %d", len(captured.Routes))
	}
	policy, ok := captured.Routes["/compare"]
	if !ok {
		t.Fatalf("expected compare profile route to be selected, got %#v", captured.Routes)
	}
	if policy.LimitRate != 100 {
		t.Fatalf("expected compare limit rate 100, got %d", policy.LimitRate)
	}
}

func TestNewRateLimiterProtectProfile(t *testing.T) {
	original := newGatewayLimiter
	t.Cleanup(func() {
		newGatewayLimiter = original
	})

	var captured limiter.RateLimiterConfig
	newGatewayLimiter = func(cfg limiter.RateLimiterConfig, _ *redis.Client) (*limiter.RateLimiter, error) {
		captured = cfg
		return &limiter.RateLimiter{}, nil
	}

	rl, err := newRateLimiter(config.Config{
		LimiterProfile: "protect",
		RoutePolicyProfiles: map[string]map[string]config.RoutePolicy{
			"compare": {
				"/compare": {LimitRate: 100},
			},
			"protect": {
				"/protect": {LimitRate: 10},
			},
		},
	}, nil)
	if err != nil {
		t.Fatalf("newRateLimiter returned error: %v", err)
	}
	if rl == nil {
		t.Fatal("expected limiter for protect profile")
	}
	if len(captured.Routes) != 1 {
		t.Fatalf("expected 1 selected route, got %d", len(captured.Routes))
	}
	policy, ok := captured.Routes["/protect"]
	if !ok {
		t.Fatalf("expected protect profile route to be selected, got %#v", captured.Routes)
	}
	if policy.LimitRate != 10 {
		t.Fatalf("expected protect limit rate 10, got %d", policy.LimitRate)
	}
}

func TestNewRateLimiterNoneProfile(t *testing.T) {
	original := newGatewayLimiter
	t.Cleanup(func() {
		newGatewayLimiter = original
	})

	called := false
	newGatewayLimiter = func(cfg limiter.RateLimiterConfig, _ *redis.Client) (*limiter.RateLimiter, error) {
		called = true
		return &limiter.RateLimiter{}, nil
	}

	rl, err := newRateLimiter(config.Config{LimiterProfile: "none"}, nil)
	if err != nil {
		t.Fatalf("newRateLimiter returned error: %v", err)
	}
	if rl != nil {
		t.Fatalf("expected nil limiter for none profile, got %#v", rl)
	}
	if called {
		t.Fatal("expected none profile to skip limiter construction")
	}
}

func TestNewRateLimiterUnknownProfile(t *testing.T) {
	original := newGatewayLimiter
	t.Cleanup(func() {
		newGatewayLimiter = original
	})

	called := false
	newGatewayLimiter = func(cfg limiter.RateLimiterConfig, _ *redis.Client) (*limiter.RateLimiter, error) {
		called = true
		return &limiter.RateLimiter{}, nil
	}

	rl, err := newRateLimiter(config.Config{LimiterProfile: "unexpected"}, nil)
	if err == nil {
		t.Fatal("expected error for unknown profile")
	}
	if rl != nil {
		t.Fatalf("expected nil limiter when profile is invalid, got %#v", rl)
	}
	if called {
		t.Fatal("expected invalid profile to fail before limiter construction")
	}
}
