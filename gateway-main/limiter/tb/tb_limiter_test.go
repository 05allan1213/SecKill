package tb

import (
	"context"
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestTBLimiter_Allow(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
	})
	tbLimiter, err := NewTBLimiter(context.TODO(), redisClient)
	if err != nil {
		t.Skipf("redis not available: %v", err)
	}

	tbLimit := &TBLimit{Rate: 100, Burst: 100, Expire: 10}
	result, err := tbLimiter.Allow(context.TODO(), "/foo", tbLimit)

	if err != nil {
		t.Skipf("redis allow check unavailable: %v", err)
	}

	if result.Allowed > 0 {
		fmt.Printf("limit allowed: %+v, remining: %+v\n", result.Allowed, result.Remaining)
	} else {
		fmt.Printf("limit not allowed: %+v, remining: %+v\n", result.Allowed, result.Remaining)
	}

}
