package config

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

type RuntimeWatcher struct {
	client *clientv3.Client
	cancel context.CancelFunc
	done   chan struct{}
}

func Load(path string) (Config, error) {
	var c Config
	if err := conf.Load(path, &c); err != nil {
		return c, err
	}

	if c.ConfigCenter.Enabled {
		runtimeCfg, err := FetchRuntimeConfig(context.Background(), c.ConfigCenter)
		if err != nil {
			return c, err
		}
		ApplyRuntimeConfig(&c, runtimeCfg)
	}

	ApplyEnvOverrides(&c)
	applyObservability(&c)
	return c, nil
}

func ApplyRuntimeConfig(c *Config, runtimeCfg RuntimeConfig) {
	if c == nil {
		return
	}
	if runtimeCfg.Auth != nil {
		c.Auth = *runtimeCfg.Auth
	}
	if runtimeCfg.Redis != nil {
		c.Redis = *runtimeCfg.Redis
	}
	if runtimeCfg.UserRpc != nil {
		c.UserRpc = *runtimeCfg.UserRpc
	}
	if runtimeCfg.SeckillRpc != nil {
		c.SeckillRpc = *runtimeCfg.SeckillRpc
	}
	if runtimeCfg.RoutePolicies != nil {
		c.RoutePolicies = runtimeCfg.RoutePolicies
	}
}

func FetchRuntimeConfig(ctx context.Context, center ConfigCenterConf) (RuntimeConfig, error) {
	var runtimeCfg RuntimeConfig
	if !center.Enabled {
		return runtimeCfg, nil
	}
	if len(center.Endpoints) == 0 || center.Key == "" {
		return runtimeCfg, fmt.Errorf("config center endpoints/key are required when enabled")
	}

	client, err := newConfigClient(center)
	if err != nil {
		return runtimeCfg, err
	}
	defer client.Close()

	value, err := getRuntimeValue(ctx, client, center.Key)
	if err != nil || value == "" {
		return runtimeCfg, err
	}
	if err := conf.LoadFromYamlBytes([]byte(value), &runtimeCfg); err != nil {
		return runtimeCfg, fmt.Errorf("load runtime config from etcd key %q: %w", center.Key, err)
	}

	return runtimeCfg, nil
}

func WatchRuntimeConfig(parent context.Context, center ConfigCenterConf, onChange func(RuntimeConfig)) (*RuntimeWatcher, error) {
	if !center.Enabled || !center.Watch {
		return nil, nil
	}
	if len(center.Endpoints) == 0 || center.Key == "" {
		return nil, fmt.Errorf("config center endpoints/key are required when watch is enabled")
	}

	client, err := newConfigClient(center)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(parent)
	watcher := &RuntimeWatcher{
		client: client,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	go func() {
		defer close(watcher.done)
		defer client.Close()

		watchCh := client.Watch(ctx, center.Key)
		for {
			select {
			case <-ctx.Done():
				return
			case resp, ok := <-watchCh:
				if !ok {
					return
				}
				if err := resp.Err(); err != nil {
					logx.Errorf("watch runtime config failed for %s: %v", center.Key, err)
					continue
				}
				for _, event := range resp.Events {
					if event.Type != clientv3.EventTypePut {
						continue
					}

					var runtimeCfg RuntimeConfig
					if err := conf.LoadFromYamlBytes(event.Kv.Value, &runtimeCfg); err != nil {
						logx.Errorf("decode runtime config failed for %s: %v", center.Key, err)
						continue
					}
					onChange(runtimeCfg)
				}
			}
		}
	}()

	return watcher, nil
}

func (w *RuntimeWatcher) Close() {
	if w == nil {
		return
	}
	w.cancel()
	select {
	case <-w.done:
	case <-time.After(5 * time.Second):
	}
}

func newConfigClient(center ConfigCenterConf) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:   center.Endpoints,
		DialTimeout: 3 * time.Second,
	})
}

func getRuntimeValue(ctx context.Context, client *clientv3.Client, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := client.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return string(resp.Kvs[0].Value), nil
}
