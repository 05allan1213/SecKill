# user-main

`user-main` 提供用户 RPC 服务，负责用户创建、按 ID 查询和按用户名查询。

## 启动

依赖：
- `docker compose up -d etcd mysql redis`

本地配置启动：
```bash
cd user-main
GOCACHE=/tmp/go-build-cache-user go run ./cmd/user -f etc/user.yaml
```

Etcd 运行时配置启动：
```bash
./scripts/sync_runtime_configs.sh
cd user-main
GOCACHE=/tmp/go-build-cache-user go run ./cmd/user -f ../configs/etcd/user.yaml
```

## 配置来源优先级

1. 本地 YAML
2. Etcd 运行时配置：`/bitstorm/user/runtime`
3. 环境变量覆盖：`USER_MYSQL_PASSWORD`、`MYSQL_PASSWORD`、`USER_REDIS_PASSWORD`、`REDIS_PASSWORD`

Etcd 模式只热更新 `Data`，不会热更新监听地址、日志、Prometheus、Telemetry。

## 验证与测试

```bash
make test-unit
make test-integration
make smoke
```

单服务基线：
```bash
cd user-main
GOCACHE=/tmp/go-build-cache-user go test ./...
GOCACHE=/tmp/go-build-cache-user go test -tags=integration ./...
```

## 观测

- 访问日志：`user-main/logs/access.log`
- 结构化运行日志：`user-main/logs/stat.log`
- Prometheus：`http://127.0.0.1:9102/metrics`
- Trace 文件：`user-main/logs/trace.json`

默认策略：
- 应用内日志按 `100MB` 滚动、保留 `7` 份并压缩
- 请求/响应摘要默认限制为 `128B`
- go-zero RPC `Stat` 默认关闭，避免重复记录原始请求
- trace 默认关闭；需要 `USER_TRACE_ENABLED=true` 才会生成 `trace.json`

## RPC 参考

- [rpc-reference.md](/home/monody/project/Microsecond%20killing%20service/docs/rpc-reference.md)
