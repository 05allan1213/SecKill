package tb

import (
	"context"
	"testing"
)

func TestTBLimiterAllowNNilReceiverReturnsError(t *testing.T) {
	var limiter *TBLimiter

	result, err := limiter.AllowN(context.Background(), "route", &TBLimit{Rate: 1, Burst: 1, Expire: 10}, 1)
	if err == nil {
		t.Fatal("expected error for nil limiter receiver")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}

func TestTBLimiterAllowNNilLimitReturnsError(t *testing.T) {
	limiter := &TBLimiter{}

	result, err := limiter.AllowN(context.Background(), "route", nil, 1)
	if err == nil {
		t.Fatal("expected error for nil limit config")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %#v", result)
	}
}
