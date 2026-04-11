package tb

import (
	"context"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestTBLimiter_Allow(t *testing.T) {
	miniRedis, err := miniredis.Run()
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("miniredis listen not permitted in current environment: %v", err)
		}
		t.Fatalf("start miniredis failed: %v", err)
	}
	defer miniRedis.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: miniRedis.Addr(),
	})
	tbLimiter, err := NewTBLimiter(context.Background(), redisClient)
	if err != nil {
		t.Fatalf("new limiter failed: %v", err)
	}

	tbLimit := &TBLimit{Rate: 100, Burst: 100, Expire: 10}
	result, err := tbLimiter.Allow(context.Background(), "/foo", tbLimit)

	if err != nil {
		t.Fatalf("allow failed: %v", err)
	}
	if result.Allowed != 1 {
		t.Fatalf("expected allowed=1, got %d", result.Allowed)
	}
	if result.Remaining < 0 {
		t.Fatalf("expected non-negative remaining tokens, got %d", result.Remaining)
	}
}
