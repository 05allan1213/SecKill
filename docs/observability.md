# 观测说明

## 日志

### gateway-main

- `gateway-main/logs/access.log`
- `gateway-main/logs/stat.log`

### user-main

- `user-main/logs/access.log`
- `user-main/logs/stat.log`

### seckill-main

- `seckill-main/logs/access.log`
- `seckill-main/logs/stat.log`

访问日志统一包含：

- `Trace-ID`
- 方法或路径
- `duration_ms`
- gRPC / HTTP 状态
- 请求摘要
- 响应摘要

## Prometheus

- `gateway-main`: `http://127.0.0.1:9103/metrics`
- `user-main`: `http://127.0.0.1:9102/metrics`
- `seckill-main`: `http://127.0.0.1:9101/metrics`

## Trace

当前默认使用 file exporter：

- `gateway-main/logs/trace.json`
- `user-main/logs/trace.json`
- `seckill-main/logs/trace.json`

这套配置适合本地验证；如需接 Jaeger/OTLP，只改 `Telemetry` 配置即可。
