# seckill-main

`seckill-main` 提供秒杀 RPC 服务，负责商品查询、同步秒杀、异步秒杀和秒杀状态查询。

## 启动

依赖：
- `docker compose up -d etcd mysql redis kafka`

本地配置启动：
```bash
cd seckill-main
GOCACHE=/tmp/go-build-cache-seckill go run ./cmd/sec_kill -f etc/seckill.yaml
```

Etcd 运行时配置启动：
```bash
./scripts/sync_runtime_configs.sh
cd seckill-main
GOCACHE=/tmp/go-build-cache-seckill go run ./cmd/sec_kill -f ../configs/etcd/seckill.yaml
```

## 配置来源优先级

1. 本地 YAML
2. Etcd 运行时配置：`/bitstorm/seckill/runtime`
3. 环境变量覆盖：`SECKILL_MYSQL_PASSWORD`、`MYSQL_PASSWORD`、`SECKILL_REDIS_PASSWORD`、`REDIS_PASSWORD`、`SECKILL_KAFKA_*`

`ConfigCenter.Enabled=false` 时只使用本地 YAML + Env。`ConfigCenter.Enabled=true` 时会从 Etcd 加载 `Data`，并在 `Watch=true` 时热切换 `Store + Model bundle`。

## 验证与测试

```bash
make test-unit
make test-integration
make smoke
```

单服务基线：
```bash
cd seckill-main
GOCACHE=/tmp/go-build-cache-seckill go test ./...
GOCACHE=/tmp/go-build-cache-seckill go test -tags=integration ./...
```

## 观测

- 访问日志：`seckill-main/logs/access.log`
- 结构化运行日志：`seckill-main/logs/stat.log`
- Prometheus：`http://127.0.0.1:9101/metrics`
- Trace 文件：`seckill-main/logs/trace.json`

## 压测

统一通过网关压测：
```bash
make bench-smoke
```

Lua 脚本位于 [gateway_sec_kill.lua](/home/monody/project/Microsecond%20killing%20service/seckill-main/wrkbench/gateway_sec_kill.lua)。

## RPC 参考

- [rpc-reference.md](/home/monody/project/Microsecond%20killing%20service/docs/rpc-reference.md)
