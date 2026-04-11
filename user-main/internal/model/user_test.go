package model

import "testing"

func TestUserCacheKey(t *testing.T) {
	if got := userCacheKey(42); got != "userinfo:42" {
		t.Fatalf("unexpected cache key: %q", got)
	}
}
