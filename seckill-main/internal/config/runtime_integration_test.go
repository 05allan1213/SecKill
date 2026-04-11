//go:build integration

package config

import (
	"context"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestFetchAndWatchRuntimeConfig(t *testing.T) {
	center := ConfigCenterConf{
		Enabled:   true,
		Endpoints: []string{"127.0.0.1:20001"},
		Key:       "/bitstorm/seckill/runtime",
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

	initial := `Data:
  Database:
    Addr: 127.0.0.1:3307
    User: root
    Password: "from-etcd"
    DataBase: bitstorm
    MaxIdleConn: 10
    MaxOpenConn: 20
    MaxIdleTime: 30
  Redis:
    Addr: 127.0.0.1:6379
    Db: 0
    PassWord: "123456"
    PoolSize: 10
  Kafka:
    Producer:
      Brokers: [127.0.0.1:9092]
      Topic: seckill
      Ack: 0
    Consumer:
      Brokers: [127.0.0.1:9092]
      Topic: seckill
      Offset: 0
`
	if _, err := client.Put(context.Background(), center.Key, initial); err != nil {
		t.Fatalf("put initial runtime config failed: %v", err)
	}

	runtimeCfg, err := FetchRuntimeConfig(context.Background(), center)
	if err != nil {
		t.Fatalf("fetch runtime config failed: %v", err)
	}
	if runtimeCfg.Data == nil || runtimeCfg.Data.Database.Password != "from-etcd" {
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

	updated := `Data:
  Database:
    Addr: 127.0.0.1:3308
    User: root
    Password: "updated"
    DataBase: bitstorm
    MaxIdleConn: 10
    MaxOpenConn: 20
    MaxIdleTime: 30
  Redis:
    Addr: 127.0.0.1:6379
    Db: 0
    PassWord: "123456"
    PoolSize: 10
  Kafka:
    Producer:
      Brokers: [127.0.0.1:9092]
      Topic: seckill
      Ack: 0
    Consumer:
      Brokers: [127.0.0.1:9092]
      Topic: seckill
      Offset: 0
`
	if _, err := client.Put(context.Background(), center.Key, updated); err != nil {
		t.Fatalf("put updated runtime config failed: %v", err)
	}

	select {
	case cfg := <-done:
		if cfg.Data == nil || cfg.Data.Database.Addr != "127.0.0.1:3308" {
			t.Fatalf("unexpected watched config: %#v", cfg)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for runtime config watch")
	}
}
