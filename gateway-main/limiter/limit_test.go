package limiter

import (
	"context"
	"testing"

	"github.com/BitofferHub/gateway/limiter/tb"
	"golang.org/x/time/rate"
)

func TestRateLimiterAllowMissingRouteAllows(t *testing.T) {
	originalTBLimiter := tbLimiter
	originalRouteLimits := routeLimits
	originalLocalRouteLimits := localRouteLimits
	t.Cleanup(func() {
		tbLimiter = originalTBLimiter
		routeLimits = originalRouteLimits
		localRouteLimits = originalLocalRouteLimits
	})

	tbLimiter = nil
	routeLimits = map[string]*tb.TBLimit{}
	localRouteLimits = map[string]*rate.Limiter{}

	result, err := (&RateLimiter{}).Allow(context.Background(), "/missing")
	if err != nil {
		t.Fatalf("Allow returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsAllowed {
		t.Fatal("expected missing route config to allow request")
	}
}
