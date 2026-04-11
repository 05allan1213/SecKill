# seckill-main

`seckill-main` 是整个项目的主角，秒杀核心流程都在这里。

## 模块职责

- `internal/server`
  - gRPC 接入层
- `internal/logic/seckill`
  - 秒杀业务流程
- `internal/data`
  - DB、Redis、Kafka 访问
- `internal/svc`
  - 依赖注入

## 最值得先看的文件

- [internal/logic/seckill/seckillv1logic.go](/home/monody/project/SecKill/seckill-main/internal/logic/seckill/seckillv1logic.go)
- [internal/logic/seckill/seckillv2logic.go](/home/monody/project/SecKill/seckill-main/internal/logic/seckill/seckillv2logic.go)
- [internal/logic/seckill/seckillv3logic.go](/home/monody/project/SecKill/seckill-main/internal/logic/seckill/seckillv3logic.go)
- [internal/logic/seckill/flow.go](/home/monody/project/SecKill/seckill-main/internal/logic/seckill/flow.go)
- [internal/data/seckill_stock_redis.go](/home/monody/project/SecKill/seckill-main/internal/data/seckill_stock_redis.go)

## 三版秒杀

- `V1`
  - 数据库事务基线版本
- `V2`
  - Redis 预扣库存 + 同步落库
- `V3`
  - Redis 预扣库存 + MQ 异步下单 + 查询闭环

## 启动

```bash
cd seckill-main
go run ./cmd/sec_kill -f etc/seckill.yaml
```

生成 proto：

```bash
make api
```

完整学习路径见根目录 [README.md](/home/monody/project/SecKill/README.md) 和 [学习.md](/home/monody/project/SecKill/学习.md)。
