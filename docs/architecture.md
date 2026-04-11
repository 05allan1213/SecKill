# 架构文档

## 整体结构

当前仓库已经收敛到 3 层主结构：

- `gateway-main`：`handler -> logic -> svc`
- `user-main`：`server -> logic -> model`
- `seckill-main`：`server -> logic -> model`

## 调用链

外部请求主链路：

1. 客户端请求 `gateway-main`
2. `gateway-main` 写入/透传 `Trace-ID`
3. `gateway-main` 调用 `user-main` 或 `seckill-main` gRPC
4. 下游服务执行 `logic -> model`
5. 日志、Prometheus、trace 在各服务侧分别产出

## 关键数据流

### 用户查询

`gateway /bitstorm/get_user_info_by_name -> user RPC GetUserByName -> user model -> MySQL/Redis`

### 秒杀同步链路

`gateway /bitstorm/v1/sec_kill -> seckill RPC SecKillV1 -> quota/stock/order/record model -> MySQL`

### 秒杀异步链路

`gateway /bitstorm/v3/sec_kill -> seckill RPC SecKillV3 -> Redis 预扣库存 -> Kafka -> MQ consume -> MySQL 落单 -> Redis 回写状态`
