# SecKill 秒杀系统学习项目

这是一个以秒杀主流程为中心的微服务学习项目，通过 `V1 / V2 / V3` 三版实现演示典型的高并发秒杀优化路径。

## 项目目标

- 理解请求如何进入系统
- 掌握库存和资格校验机制
- 了解 `V1 / V2 / V3` 的演进原因
- 理解 Redis 预扣库存和 MQ 异步下单的作用
- 掌握异步秒杀的结果查询闭环设计

## 技术栈

| 类别 | 技术 |
|------|------|
| 核心框架 | go-zero + gRPC + protobuf + GORM |
| 数据库 | MySQL |
| 缓存 | Redis |
| 消息队列 | Kafka |
| 服务注册发现 | Etcd |
| 认证 | JWT |
| 基础设施 | Docker Compose |

## 系统架构

```
┌─────────┐     HTTP      ┌──────────────┐     gRPC      ┌────────────┐
│  用户   │ ────────────> │ gateway-main │ ────────────> │ user-main  │
└─────────┘               │  (HTTP 网关)  │               │ (用户服务)  │
                          │  - JWT 鉴权   │               └────────────┘
                          │  - 限流       │
                          │  - 路由转发   │
                          └──────────────┘
                                    │
                                    │ gRPC
                                    ▼
                          ┌──────────────────┐
                          │  seckill-main    │
                          │   (秒杀服务)      │
                          │  - V1/V2/V3 实现  │
                          │  - MQ Consumer   │
                          └──────────────────┘
                                    │
           ┌────────────────────────┼────────────────────────┐
           │                        │                        │
           ▼                        ▼                        ▼
    ┌─────────────┐         ┌─────────────┐         ┌─────────────┐
    │   MySQL     │         │    Redis    │         │    Kafka    │
    │ (订单/库存)  │         │ (库存预扣)   │         │ (异步消息)   │
    └─────────────┘         └─────────────┘         └─────────────┘
```

## 秒杀版本演进

### V1: 数据库事务秒杀

最基础的实现，直接在数据库层面完成库存扣减和订单创建。

- 优点：简单直接
- 缺点：高并发下数据库压力巨大

### V2: Redis 预扣 + 同步落库

在 V1 基础上加入 Redis 预扣库存、防重复秒杀、前置限购校验。

- 使用 Lua 脚本保证 Redis 操作的原子性
- 数据库操作作为最终落库
- 大幅降低数据库压力

### V3: Redis 预扣 + MQ 异步

在 V2 基础上引入 Kafka 异步下单，实现削峰填谷。

- 秒杀请求快速响应（返回 secNum）
- 异步 Consumer 处理实际下单
- 提供 `GetSecKillInfo` 查询接口闭环

## 目录结构

```
SecKill
├── docker/                      # Docker 初始化资源
│   └── mysql/initdb/           # MySQL 建库建表和初始化数据
├── gateway-main/               # HTTP 入口层
│   ├── cmd/                    # 启动入口
│   ├── etc/                    # 网关配置
│   ├── internal/               # handler/logic/middleware/svc
│   └── limiter/                # 限流实现
├── scripts/
│   ├── integration/            # 端到端集成测试
│   │   ├── test_cases.sh      # 测试用例
│   │   ├── assertions.sh       # 断言工具
│   │   └── run_e2e_test.sh     # E2E 测试运行器
│   └── perf/                   # 性能压测脚本
│       ├── run_perf_optimized.sh  # 详细压测脚本
│       └── extreme_perf.sh     # 极简压测脚本
├── seckill-main/               # 秒杀核心服务
│   ├── api/                    # protobuf 定义
│   ├── cmd/                    # RPC 服务和 MQ Consumer 入口
│   ├── etc/                    # 服务配置
│   ├── internal/
│   │   ├── logic/seckill/      # V1/V2/V3 和结果查询主流程
│   │   ├── data/               # MySQL/Redis/Kafka 访问层
│   │   ├── server/             # gRPC 接入层
│   │   └── svc/                # 依赖注入
│   └── sql/                    # 表结构参考
├── user-main/                  # 用户服务
│   ├── api/                    # protobuf 定义
│   ├── cmd/                    # RPC 启动入口
│   ├── etc/                    # 服务配置
│   └── internal/               # 内部实现
├── docker-compose.yml          # 基础设施编排
├── start_and_test.sh           # 一键启动和测试脚本
└── 测试报告.md                 # 测试报告
```

## 快速启动

### 环境要求

- Docker / Docker Compose
- Go
- curl、nc、lsof

### 一键启动和测试

```bash
bash ./start_and_test.sh
```

### 分步操作

```bash
# 启动所有服务（基础设施 + 微服务）
bash ./start_and_test.sh start

# 运行所有测试用例
bash ./scripts/integration/run_e2e_test.sh test all

# 停止服务
bash ./start_and_test.sh stop
```

## 集成测试

### 测试用例列表

| 测试用例 | 说明 |
|---------|------|
| `test_login` | 用户登录，验证 JWT Token 获取 |
| `test_v1_flow` | V1 完整流程：数据库事务秒杀 |
| `test_v2_flow` | V2 完整流程：Redis 预扣 + 同步落库 |
| `test_v3_flow` | V3 完整流程：Redis 预扣 + MQ 异步 |
| `test_concurrent_seckill` | 并发秒杀测试，验证超卖和限购 |
| `test_quota_limit` | 限购功能测试，验证超限拦截 |
| `test_stock_exhausted` | 库存耗尽测试，验证库存不足处理 |

### 运行特定测试

```bash
# 运行所有测试
bash ./scripts/integration/run_e2e_test.sh test all

# 仅测试 V1 流程
bash ./scripts/integration/run_e2e_test.sh test v1

# 仅测试并发
bash ./scripts/integration/run_e2e_test.sh test concurrent

# 仅测试限购
bash ./scripts/integration/run_e2e_test.sh test quota
```

### 测试断言关键点

- **库存不超卖**：成功订单数 ≤ 初始库存
- **限购不超限**：同一用户成功秒杀次数 ≤ 限购数量
- **响应码正确**：HTTP 200 且业务 code = 0
- **订单数量正确**：数据库订单数与成功秒杀数一致

## 性能测试

### 压测脚本说明

#### 详细压测脚本 (run_perf_optimized.sh)

多用户并发压测脚本，具备以下特性：
- 多用户模拟真实秒杀场景
- 预热阶段确保连接池就绪
- 详细的统计报告和性能分析
- 可配置库存、用户数、并发数等参数

```bash
# 默认参数压测
STOCK=10000 USER_LIMIT=10 NUM_USERS=200 REQUESTS=20000 CONNECTIONS=200 \
  bash ./scripts/perf/run_perf_optimized.sh
```

#### 极简压测脚本 (extreme_perf.sh)

快速压测脚本，适合简单验证：
- 快速执行，无额外配置
- 自动获取用户 Token
- 依次压测 V1/V2/V3 版本

```bash
# 默认参数压测
bash ./scripts/perf/extreme_perf.sh

# 自定义参数
STOCK=10000 USER_LIMIT=10 NUM_USERS=100 REQUESTS=10000 CONNECTIONS=200 \
  bash ./scripts/perf/extreme_perf.sh
```

### 压测参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| STOCK | 库存数量 | 10000 |
| USER_LIMIT | 每用户限购数量 | 10 |
| NUM_USERS | 压测用户数 | 100 |
| REQUESTS | 总请求数 | 10000 |
| CONNECTIONS | 并发连接数 | 200 |

### 性能测试结果参考

| 版本 | 平均 QPS | 平均 P50 | 平均 P99 | 成功率 |
|------|----------|----------|----------|--------|
| v1 | ~5,000 | 44ms | 374ms | ~99.9% |
| v2 | ~8,300 | 34ms | 68ms | ~99.7% |
| v3 | ~8,300 | 34ms | 64ms | ~98.3% |

> 注：实际结果取决于硬件配置和并发参数

## API 接口

### 登录

```bash
curl -X POST http://127.0.0.1:8998/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"123321"}'
```

### V1 秒杀

```bash
curl -X POST http://127.0.0.1:8998/bitstorm/v1/sec_kill \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"goodsNum":"abc123","num":1}'
```

### V2 秒杀

```bash
curl -X POST http://127.0.0.1:8998/bitstorm/v2/sec_kill \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"goodsNum":"abc123","num":1}'
```

### V3 秒杀

```bash
curl -X POST http://127.0.0.1:8998/bitstorm/v3/sec_kill \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"goodsNum":"abc123","num":1}'
```

### V3 查询秒杀结果

```bash
curl "http://127.0.0.1:8998/bitstorm/v3/get_sec_kill_info?sec_num=<secNum>" \
  -H "Authorization: Bearer <token>"
```

## 端口配置

| 服务 | 端口 |
|------|------|
| Gateway HTTP | 8998 |
| User RPC | 8669 |
| User Health | 8670 |
| Seckill RPC | 8002 |
| Seckill Health | 8003 |
| MySQL | 3307 |
| Redis | 6379 |
| Kafka | 9092 |
| Etcd | 20001 |

## 测试账号

- 用户名：`admin`
- 密码：`123321`
- 商品编号：`abc123`
- 商品 ID：`1`

## 测试报告

详细测试报告请查看 [测试报告.md](测试报告.md)。
