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
- 请求摘要（默认最多 128B）
- 响应摘要（默认最多 128B）

默认策略：

- 应用内日志按大小滚动：单文件 `100MB`、保留 `7` 份、压缩归档
- 日志目录按 `KeepDays: 7` 清理旧文件
- `gateway` 关闭 go-zero REST `Log` 请求日志
- `user-main` / `seckill-main` 关闭 go-zero RPC `Stat` 请求日志
- trace 默认关闭；需要显式开启 `*_TRACE_ENABLED=true`

## Prometheus

- `gateway-main`: `http://127.0.0.1:9103/metrics`
- `user-main`: `http://127.0.0.1:9102/metrics`
- `seckill-main`: `http://127.0.0.1:9101/metrics`

## Trace

当前默认使用 file exporter：

- `gateway-main/logs/trace.json`
- `user-main/logs/trace.json`
- `seckill-main/logs/trace.json`

默认配置不会生成上述文件；只有显式开启 trace 后才会落盘。

## 巡检

- 运行 `make check-logs` 检查每个服务日志目录是否超过 `500MB`
- 脚本同时检查单文件是否超过 `100MB`，并输出每个目录最大的 5 个文件
- 宿主机 `logrotate` 示例见 `deploy/logrotate/microsecond-killing-service.conf`
