package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/BitofferHub/gateway/limiter"
	"github.com/redis/go-redis/v9"
)

func TestRouteLimitMiddlewareReturnsStructured429(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "123456",
		DB:       8,
	})
	if err := redisClient.Ping(t.Context()).Err(); err != nil {
		t.Skipf("redis not available: %v", err)
	}

	routeKey := "/test/rate-limit/" + time.Now().Format("150405.000000000")
	rl, err := limiter.NewRateLimiter(limiter.RateLimiterConfig{
		Routes: map[string]limiter.RoutePolicy{
			routeKey: {
				LimitTimeout: 2000,
				LimitRate:    1,
				RetryTime:    50,
				Remarks:      "test route",
			},
		},
		DefaultRetryTime:    50,
		DefaultLimitTimeout: 2000,
		DefaultLimitRate:    1,
	}, redisClient)
	if err != nil {
		t.Skipf("rate limiter unavailable: %v", err)
	}

	middleware := NewRouteLimitMiddleware(&svc.ServiceContext{Limiter: rl}, routeKey)
	handler := middleware.Handle(func(w http.ResponseWriter, r *http.Request) {
		WriteCodeMessage(w, http.StatusOK, 0, "")
	})

	first := httptest.NewRecorder()
	handler(first, httptest.NewRequest(http.MethodPost, routeKey, nil))
	if first.Code != http.StatusOK {
		t.Fatalf("unexpected first response status: got %d want %d", first.Code, http.StatusOK)
	}

	second := httptest.NewRecorder()
	handler(second, httptest.NewRequest(http.MethodPost, routeKey, nil))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("unexpected second response status: got %d want %d", second.Code, http.StatusTooManyRequests)
	}

	var payload map[string]any
	if err := json.Unmarshal(second.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if got := int(payload["code"].(float64)); got != rateLimitCode {
		t.Fatalf("unexpected rate limit code: got %d want %d", got, rateLimitCode)
	}
	if got := payload["message"].(string); got != "rate limited" {
		t.Fatalf("unexpected rate limit message: got %q", got)
	}
}
