# 配置中心

## 键位

- `/bitstorm/user/runtime`
- `/bitstorm/seckill/runtime`
- `/bitstorm/gateway/runtime`

## 来源优先级

每个服务的配置加载顺序固定为：

1. 本地 YAML
2. Etcd 运行时配置
3. 环境变量覆盖

敏感信息不写入 Etcd，仍由环境变量覆盖。

## 环境变量

### user-main

- `USER_MYSQL_PASSWORD`
- `MYSQL_PASSWORD`
- `USER_REDIS_PASSWORD`
- `REDIS_PASSWORD`

### seckill-main

- `SECKILL_MYSQL_PASSWORD`
- `MYSQL_PASSWORD`
- `SECKILL_REDIS_PASSWORD`
- `REDIS_PASSWORD`
- `SECKILL_KAFKA_PRODUCER_BROKERS`
- `SECKILL_KAFKA_CONSUMER_BROKERS`
- `SECKILL_KAFKA_BROKERS`
- `KAFKA_BROKERS`

### gateway-main

- `GATEWAY_AUTH_SECRET`
- `AUTH_SECRET`
- `GATEWAY_REDIS_PASSWORD`
- `REDIS_PASSWORD`

## 热更新边界

### 会热更新

- `user-main` / `seckill-main`：`Data`
- `gateway-main`：`Auth`、`Redis`、`UserRpc`、`SeckillRpc`、`RoutePolicies`

### 不会热更新

- 监听地址和端口
- `Log`
- `Prometheus`
- `Telemetry`
- 中间件开关

## 同步运行时配置

```bash
make sync-config
```

等价脚本：
```bash
./scripts/sync_runtime_configs.sh
```

运行时配置源文件位于：

- `configs/runtime/user.yaml`
- `configs/runtime/seckill.yaml`
- `configs/runtime/gateway.yaml`
