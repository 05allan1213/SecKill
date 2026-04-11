//go:build integration

package config

import (
	"context"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestGatewayFetchAndWatchRuntimeConfig(t *testing.T) {
	center := ConfigCenterConf{
		Enabled:   true,
		Endpoints: []string{"127.0.0.1:20001"},
		Key:       "/bitstorm/gateway/runtime",
		Watch:     true,
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   center.Endpoints,
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		t.Fatalf("create etcd client failed: %v", err)
	}
	defer client.Close()

	initial := `Auth:
  Secret: "secret key"
  Timeout: 1h
Redis:
  Addr: 127.0.0.1:6379
  DB: 8
  passWord: "123456"
  read_timeout: 2s
  write_timeout: 2s
RoutePolicies:
  /bitstorm/v1/sec_kill:
    limit_timeout: 2000
    limit_rate: 1000
    retry_time: 50
    remarks: 秒杀v1
`
	if _, err := client.Put(context.Background(), center.Key, initial); err != nil {
		t.Fatalf("put initial runtime config failed: %v", err)
	}

	runtimeCfg, err := FetchRuntimeConfig(context.Background(), center)
	if err != nil {
		t.Fatalf("fetch runtime config failed: %v", err)
	}
	if runtimeCfg.Auth == nil || runtimeCfg.Auth.Secret != "secret key" {
		t.Fatalf("unexpected runtime config: %#v", runtimeCfg)
	}

	done := make(chan RuntimeConfig, 1)
	watcher, err := WatchRuntimeConfig(context.Background(), center, func(cfg RuntimeConfig) {
		done <- cfg
	})
	if err != nil {
		t.Fatalf("watch runtime config failed: %v", err)
	}
	defer watcher.Close()
	time.Sleep(200 * time.Millisecond)

	updated := `Auth:
  Secret: "updated-secret"
  Timeout: 2h
Redis:
  Addr: 127.0.0.1:6379
  DB: 8
  passWord: "123456"
  read_timeout: 2s
  write_timeout: 2s
RoutePolicies:
  /bitstorm/v1/sec_kill:
    limit_timeout: 2000
    limit_rate: 10
    retry_time: 50
    remarks: 秒杀v1
`
	if _, err := client.Put(context.Background(), center.Key, updated); err != nil {
		t.Fatalf("put updated runtime config failed: %v", err)
	}

	select {
	case cfg := <-done:
		if cfg.Auth == nil || cfg.Auth.Secret != "updated-secret" {
			t.Fatalf("unexpected watched config: %#v", cfg)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for runtime config watch")
	}
}
