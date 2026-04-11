# gateway-main

`gateway-main` 是项目的 HTTP 入口层，不负责真正的秒杀业务。

它的职责只有四件事：

- 提供 `/login`
- 做 JWT 鉴权
- 做路由级限流
- 把 HTTP 请求转成 `user-main` 和 `seckill-main` 的 RPC 请求

## 关键入口

- 路由注册：
  - [internal/handler/routes.go](/home/monody/project/SecKill/gateway-main/internal/handler/routes.go)
- 登录逻辑：
  - [internal/logic/loginlogic.go](/home/monody/project/SecKill/gateway-main/internal/logic/loginlogic.go)
- 秒杀转发：
  - [internal/logic/bitstormseckillv1logic.go](/home/monody/project/SecKill/gateway-main/internal/logic/bitstormseckillv1logic.go)
  - [internal/logic/bitstormseckillv2logic.go](/home/monody/project/SecKill/gateway-main/internal/logic/bitstormseckillv2logic.go)
  - [internal/logic/bitstormseckillv3logic.go](/home/monody/project/SecKill/gateway-main/internal/logic/bitstormseckillv3logic.go)
- 秒杀结果查询：
  - [internal/logic/bitstormgetseckillinfologic.go](/home/monody/project/SecKill/gateway-main/internal/logic/bitstormgetseckillinfologic.go)

## 限流

限流配置在：

- [etc/gateway.yaml](/home/monody/project/SecKill/gateway-main/etc/gateway.yaml)

支持两套默认档位：

- `compare`
- `protect`

命中限流后返回 HTTP `429` 和结构化 JSON 错误响应。

## 启动

```bash
cd gateway-main
go run ./cmd/gateway -f etc/gateway.yaml
```

更完整的学习入口请回到根目录 [README.md](/home/monody/project/SecKill/README.md)。
