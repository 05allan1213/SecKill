# SecKill 学习项目

这是一个以秒杀主流程为中心的微服务学习项目，不是生产级系统。

项目的核心目标不是做一个大而全的电商平台，而是用尽量短的链路，把一次秒杀请求里最关键的几个问题讲清楚：

- 请求怎么进入系统
- 库存和资格怎么校验
- 为什么 `V1 / V2 / V3` 要逐步演进
- 为什么高并发场景里会引入 Redis 和 MQ
- 为什么异步秒杀需要结果查询接口

如果把这个项目看成一句话，就是：

`一个用 go-zero + gRPC + MySQL + Redis + Kafka 搭出来的秒杀演进学习样例`

## 这个项目是干什么的

这个项目用三版秒杀实现来演示典型的高并发秒杀优化路径：

- `V1`
  - 最基础的数据库事务秒杀
- `V2`
  - 在 `V1` 上加入 Redis 预扣库存、防重复秒杀、前置限购校验
- `V3`
  - 在 `V2` 上加入 Kafka 异步下单和 `GetSecKillInfo` 查询闭环

所以它不是为了展示“功能很多”，而是为了展示“为什么要这样演进”。

适合的使用方式：

- 想快速理解秒杀系统的主流程
- 想看数据库事务、Redis 预扣、MQ 异步三种方案怎么落到代码里
- 想把一个微服务项目从 HTTP 入口一路追到库存、订单和结果查询

## 技术栈

### 核心框架

- `go-zero`
  - 网关 HTTP 服务、RPC 服务、配置加载
- `gRPC + protobuf`
  - `gateway-main` 到 `user-main / seckill-main` 的服务间调用
- `GORM`
  - MySQL 数据访问

### 中间件与基础组件

- `MySQL`
  - 商品、库存、订单、秒杀记录、额度数据
- `Redis`
  - 热点商品缓存、库存预扣、重复秒杀和限购前置校验、限流依赖
- `Kafka`
  - `V3` 的异步削峰和下单处理
- `Etcd`
  - RPC 服务注册与发现
- `JWT`
  - 登录后网关鉴权
- `Docker Compose`
  - 本地依赖启动

### 代码层面还用到了

- `segmentio/kafka-go`
- `redis/go-redis`
- `golang-jwt/jwt`
- `automaxprocs`

## 仓库目录结构

```text
SecKill
├── docker/                      # Docker 初始化资源
│   └── mysql/initdb/            # MySQL 建库建表和初始化数据
├── gateway-main/                # HTTP 入口层
│   ├── cmd/                     # 启动入口
│   ├── etc/                     # 网关配置
│   ├── internal/                # handler/logic/middleware/svc 等实现
│   └── limiter/                 # 路由限流实现
├── scripts/
│   └── perf/                    # 压测和对比脚本
├── seckill-main/                # 秒杀核心服务
│   ├── api/                     # protobuf 定义
│   ├── cmd/                     # RPC 服务和 MQ consumer 启动入口
│   ├── etc/                     # 秒杀服务配置
│   ├── internal/
│   │   ├── logic/seckill/       # V1/V2/V3 和结果查询主流程
│   │   ├── data/                # MySQL/Redis/Kafka 访问
│   │   ├── server/              # gRPC 接入层
│   │   └── svc/                 # 依赖注入
│   └── sql/                     # 秒杀相关表结构参考
├── user-main/                   # 用户与登录支撑服务
│   ├── api/                     # protobuf 定义
│   ├── cmd/                     # RPC 启动入口
│   ├── etc/                     # 用户服务配置
│   ├── internal/                # user logic/data/server/svc
│   └── sql/                     # 用户表结构参考
├── docker-compose.yml           # 本地依赖编排
├── start_and_test.sh            # 一键启动和功能演示脚本
├── 学习.md                       # 主流程、代码阅读顺序、V1/V2/V3 说明
├── 第三阶段性能优化说明.md       # 性能对比和压测说明
└── 迭代优化计划.md               # 本轮迭代目标与阶段计划
```

## 各目录作用

- `gateway-main`
  - 项目的 HTTP 门面，负责登录、鉴权、限流、转 RPC
- `seckill-main`
  - 项目主角，核心秒杀逻辑都在这里
- `user-main`
  - 为登录和用户查询提供最小支撑能力
- `docker`
  - 依赖服务初始化脚本，尤其是 MySQL 初始表结构和测试数据
- `scripts/perf`
  - 第三阶段压测和结果对比脚本
- `start_and_test.sh`
  - 最快的演示入口，会拉起依赖、启动服务、重置测试数据并跑接口

## 学习主线

推荐把项目理解成一条主链路：

`login -> gateway -> seckill rpc -> 库存/资格校验 -> 下单或排队 -> 查询秒杀结果`

其中最关键的是 `seckill-main/internal/logic/seckill`。

三版秒杀的定位：

- `V1`
  - 纯数据库事务秒杀，作为基线版本
- `V2`
  - Redis 预扣库存 + 防重复秒杀 + 同步落库
- `V3`
  - Redis 预扣库存 + MQ 异步下单 + `GetSecKillInfo` 查询闭环

## 快速启动

### 依赖

- Docker / Docker Compose
- Go
- curl
- nc
- lsof

### 一键演示

在仓库根目录运行：

```bash
bash ./start_and_test.sh
```

这个脚本会做这些事：

- 启动 `etcd / mysql / redis / kafka`
- 启动 `user-main / seckill-main / gateway-main`
- 重置 MySQL 秒杀测试数据
- 设置 Redis 热点库存和限购 key
- 依次调用登录、`V1`、`V2`、`V3`、`GetSecKillInfo`

如果只是启动服务：

```bash
bash ./start_and_test.sh start
```

如果只跑接口测试：

```bash
bash ./start_and_test.sh test
```

停止服务：

```bash
bash ./start_and_test.sh stop
```

## 演示账号和测试数据

- 用户名：`admin`
- 密码：`123321`
- 商品编号：`abc123`
- 商品 ID：`1`

默认演示环境：

- Gateway：`127.0.0.1:8998`
- User RPC：`127.0.0.1:8669`
- Seckill RPC：`127.0.0.1:8002`
- MySQL：`127.0.0.1:3307`
- Redis：`127.0.0.1:6379`
- Kafka：`127.0.0.1:9092`
- Etcd：`127.0.0.1:20001`

## 常用请求

登录：

```bash
curl -X POST http://127.0.0.1:8998/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"123321"}'
```

发起 `V1` 秒杀：

```bash
curl -X POST http://127.0.0.1:8998/bitstorm/v1/sec_kill \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"goodsNum":"abc123","num":1}'
```

发起 `V2` 秒杀：

```bash
curl -X POST http://127.0.0.1:8998/bitstorm/v2/sec_kill \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"goodsNum":"abc123","num":1}'
```

发起 `V3` 秒杀：

```bash
curl -X POST http://127.0.0.1:8998/bitstorm/v3/sec_kill \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"goodsNum":"abc123","num":1}'
```

查询 `V3` 秒杀结果：

```bash
curl "http://127.0.0.1:8998/bitstorm/v3/get_sec_kill_info?sec_num=<secNum>" \
  -H "Authorization: Bearer <token>"
```

## V1 / V2 / V3 怎么看

- `V1`
  - 先看最基础的数据库事务闭环
- `V2`
  - 再看为什么要把重复秒杀、限购和库存预扣前置到 Redis
- `V3`
  - 最后看为什么要异步化，以及为什么需要 `GetSecKillInfo`

更详细的对比和代码阅读顺序见：

- [学习.md](/home/monody/project/SecKill/学习.md)
- [第三阶段性能优化说明.md](/home/monody/project/SecKill/第三阶段性能优化说明.md)
- [迭代优化计划.md](/home/monody/project/SecKill/迭代优化计划.md)

## 建议阅读顺序

1. 先读 [学习.md](/home/monody/project/SecKill/学习.md)
2. 再跑一次 `bash ./start_and_test.sh`
3. 然后按下面顺序看代码：
   - `gateway-main/internal/handler/routes.go`
   - `gateway-main/internal/logic/loginlogic.go`
   - `gateway-main/internal/logic/bitstormseckillv1logic.go`
   - `gateway-main/internal/logic/bitstormseckillv2logic.go`
   - `gateway-main/internal/logic/bitstormseckillv3logic.go`
   - `gateway-main/internal/logic/bitstormgetseckillinfologic.go`
   - `seckill-main/internal/logic/seckill/seckillv1logic.go`
   - `seckill-main/internal/logic/seckill/seckillv2logic.go`
   - `seckill-main/internal/logic/seckill/seckillv3logic.go`
   - `seckill-main/internal/logic/seckill/flow.go`

## 限流和压测

网关提供两套限流档位：

- `compare`
  - 用于比较 `V1 / V2 / V3` 后端链路差异
- `protect`
  - 用于展示入口限流的拦截效果

切换位置在 [gateway.yaml](/home/monody/project/SecKill/gateway-main/etc/gateway.yaml) 的 `LimiterProfile`。

压测脚本：

```bash
bash ./scripts/perf/run_phase3_compare.sh
```

这一步依赖本地已经安装 `hey`。
