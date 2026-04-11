# gateway-main

`gateway-main` 提供对外 HTTP API、JWT 认证、链路透传和路由级限流，并转发到 `user-main` / `seckill-main`。

## 启动

依赖：
- `docker compose up -d etcd mysql redis kafka`

本地配置启动：
```bash
cd gateway-main
GOCACHE=/tmp/go-build-cache-gateway go run ./cmd/gateway -f etc/gateway.yaml
```

Etcd 运行时配置启动：
```bash
./scripts/sync_runtime_configs.sh
cd gateway-main
GOCACHE=/tmp/go-build-cache-gateway go run ./cmd/gateway -f ../configs/etcd/gateway.yaml
```

## 配置来源优先级

1. 本地 YAML
2. Etcd 运行时配置：`/bitstorm/gateway/runtime`
3. 环境变量覆盖：`GATEWAY_AUTH_SECRET`、`AUTH_SECRET`、`GATEWAY_REDIS_PASSWORD`、`REDIS_PASSWORD`

Etcd 模式热更新：
- `Auth`
- `Redis`
- `UserRpc`
- `SeckillRpc`
- `RoutePolicies`

不热更新：
- `RestConf`
- `Log`
- `Prometheus`
- `Telemetry`
- REST 中间件开关

## 接口文档

- API 源：`gateway-main/gateway.api`
- OpenAPI YAML：[gateway.openapi.yaml](/home/monody/project/Microsecond%20killing%20service/docs/openapi/gateway.openapi.yaml)
- OpenAPI JSON：[gateway.openapi.json](/home/monody/project/Microsecond%20killing%20service/docs/openapi/gateway.openapi.json)

## 验证与测试

```bash
make test-unit
make test-integration
make smoke
```

单服务基线：
```bash
cd gateway-main
GOCACHE=/tmp/go-build-cache-gateway go test ./...
GOCACHE=/tmp/go-build-cache-gateway go test -tags=integration ./...
```

## 观测

- 访问日志：`gateway-main/logs/access.log`
- 结构化运行日志：`gateway-main/logs/stat.log`
- Prometheus：`http://127.0.0.1:9103/metrics`
- Trace 文件：`gateway-main/logs/trace.json`
