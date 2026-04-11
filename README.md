# 秒杀系统微服务项目

一个基于 go-zero 框架的高性能秒杀系统，支持高并发场景下的商品秒杀功能。

## 📝 项目简介

本项目是一个完整的微服务秒杀系统，采用 DDD（领域驱动设计）架构，实现了从用户认证、商品管理到秒杀下单的完整业务流程。系统经过多轮迭代优化，具备生产级别的可观测性、配置管理和性能优化能力。

### 🎯 项目目标

- ✅ **高并发处理**：支持万级 QPS 的秒杀场景
- ✅ **架构清晰**：采用 go-zero 标准三层架构（Server -> Logic -> Model）
- ✅ **可观测性**：完整的日志、指标、链路追踪体系
- ✅ **配置管理**：支持 Etcd 配置中心和热更新
- ✅ **性能优化**：Redis 预扣库存、DB 连接池优化、缓存策略
- ✅ **测试完善**：单元测试、集成测试、冒烟测试、压测脚本

### ✨ 核心特性

#### 🚀 高性能设计
- **Redis 预扣库存**：使用 Lua 脚本保证原子性，避免超卖
- **Kafka 异步处理**：削峰填谷，提升系统吞吐量
- **多级缓存**：Redis 缓存 + 本地缓存，减少数据库压力
- **连接池优化**：DB 连接池配置，提升数据库性能

#### 🏗️ 微服务架构
- **服务拆分**：网关服务、用户服务、秒杀服务独立部署
- **服务发现**：基于 Etcd 的服务注册与发现
- **配置中心**：Etcd 配置中心，支持热更新
- **统一网关**：API 网关统一入口，JWT 认证、限流、链路追踪

#### 📊 可观测性
- **日志规范**：统一使用 go-zero logx，结构化日志输出
- **Prometheus 指标**：请求计数、延迟分布、错误率监控
- **链路追踪**：OpenTelemetry 标准，跨服务调用链追踪
- **访问日志**：包含 traceID、方法、耗时、状态码等关键字段

#### 🔧 配置管理
- **多环境支持**：本地 YAML、Etcd 运行时配置、环境变量
- **配置优先级**：本地 YAML -> Etcd 运行时配置 -> Env
- **热更新**：支持运行时配置热更新，无需重启服务
- **配置同步工具**：一键同步配置到 Etcd

#### 🧪 测试完善
- **单元测试**：覆盖核心业务逻辑
- **集成测试**：端到端测试流程
- **冒烟测试**：快速验证系统功能
- **压测脚本**：wrk 压测，性能基线验证

### 📈 性能指标

基于 wrk 压测结果（本地环境，三个版本对比）：

#### v1 版本（同步处理）

| 并发数 | QPS | P50 延迟 | P90 延迟 | P99 延迟 | 错误率 |
|--------|-----|----------|----------|----------|--------|
| **50** | 17,228 req/s | 2.28ms | 6.52ms | 11.57ms | 0% |
| **100** | 21,314 req/s | 3.95ms | 9.25ms | 18.58ms | 0% |
| **200** | 21,644 req/s | 7.94ms | 17.38ms | 32.62ms | 0% |

#### v2 版本（Redis 预扣）

| 并发数 | QPS | P50 延迟 | P90 延迟 | P99 延迟 | 错误率 |
|--------|-----|----------|----------|----------|--------|
| **50** | 48,355 req/s | 0.82ms | 1.75ms | 3.74ms | 0% |
| **100** | 55,101 req/s | 1.42ms | 3.94ms | 6.86ms | 0% |
| **200** | 41,498 req/s | 4.89ms | 9.28ms | 13.93ms | 0% |

#### v3 版本（Kafka 异步）⭐ 推荐

| 并发数 | QPS | P50 延迟 | P90 延迟 | P99 延迟 | 错误率 |
|--------|-----|----------|----------|----------|--------|
| **50** | 50,589 req/s | 0.79ms | 1.65ms | 3.47ms | 0% |
| **100** | 56,669 req/s | 1.38ms | 3.83ms | 6.67ms | 0% |
| **200** | 43,285 req/s | 4.53ms | 9.06ms | 13.68ms | 0% |

**性能特点**：
- ✅ **v3 性能最优**：QPS 达到 5.6万+，延迟 P50 < 1ms
- ✅ **v2 性能优秀**：QPS 达到 5.5万+，延迟 P50 < 2ms
- ✅ **v1 性能稳定**：QPS 达到 2.1万+，错误率为 0
- ✅ **架构优化效果显著**：v3 相比 v1 性能提升 162%，延迟降低 83%

**压测配置**：
- 工具：wrk
- 线程：4-16
- 连接数：50-1000
- 持续时间：10s
- 接口：`/bitstorm/v1/sec_kill`、`/bitstorm/v2/sec_kill`、`/bitstorm/v3/sec_kill`

**详细报告**：[docs/benchmark_report.md](./docs/benchmark_report.md)

### 🔐 安全特性

- **JWT 认证**：基于 JWT 的用户认证机制
- **API 限流**：令牌桶限流，防止恶意请求
- **用户限额**：单个用户购买数量限制
- **库存保护**：Redis 预扣 + 数据库校验，防止超卖

### 🎨 技术亮点

1. **三层架构简化**：从传统的五层架构简化为 Server -> Logic -> Model 三层
2. **配置中心集成**：支持 Etcd 配置中心和热更新
3. **可观测性完善**：日志、指标、链路追踪三位一体
4. **性能优化落地**：DB 连接池、缓存策略、压测验证
5. **测试体系完善**：单元测试、集成测试、冒烟测试、压测脚本
6. **文档体系完整**：架构文档、API 文档、性能文档、测试文档

## 🚀 快速开始

### 方式 1：一键启动（推荐）

使用冒烟测试脚本自动启动所有服务并验证：

```bash
# 本地配置模式（自动启动 + 测试 + 清理）
./scripts/smoke.sh local

# 或 Etcd 配置模式
./scripts/smoke.sh etcd
```

**这个命令会自动完成：**
- ✅ 启动 Docker 基础设施（MySQL、Redis、Kafka、Etcd）
- ✅ 启动三个微服务（后台进程）
- ✅ 运行完整的冒烟测试
- ✅ 验证指标和日志
- ✅ 退出时自动清理进程

### 方式 2：手动启动（开发调试）

如果需要手动启动服务进行调试：

```bash
# 1. 启动基础设施
docker compose up -d

# 2. 在三个终端中分别启动服务
# 终端 1 - 用户服务
cd user-main && go run ./cmd/user -f etc/user.yaml

# 终端 2 - 秒杀服务
cd seckill-main && go run ./cmd/sec_kill -f etc/seckill.yaml

# 终端 3 - 网关服务
cd gateway-main && go run ./cmd/gateway -f etc/gateway.yaml
```

### 其他测试命令

```bash
# 单元测试
make test-unit

# 集成测试
make test-integration

# 压力测试
make bench-smoke
```

## 📖 文档

- [docs/testing.md](./docs/testing.md) - 完整的测试指南和 API 文档
- [docs/architecture.md](./docs/architecture.md) - 架构设计文档
- [docs/config-center.md](./docs/config-center.md) - 配置中心文档
- [docs/performance.md](./docs/performance.md) - 性能优化文档
- [docs/observability.md](./docs/observability.md) - 可观测性文档
- [迭代优化计划.md](./迭代优化计划.md) - 迭代优化计划

## 🏗️ 系统架构

### 项目结构

```
Microsecond killing service/
├── README.md                      # 项目主文档
├── 迭代优化计划.md                 # 迭代优化计划
├── Makefile                       # 统一构建入口
├── docker-compose.yml             # Docker 编排配置
│
├── configs/                       # 配置文件目录
│   ├── etcd/                      # Etcd 配置模式启动文件
│   │   ├── gateway.yaml
│   │   ├── user.yaml
│   │   └── seckill.yaml
│   └── runtime/                   # 运行时配置源（同步到 Etcd）
│       ├── gateway.yaml
│       ├── user.yaml
│       └── seckill.yaml
│
├── docker/                        # Docker 相关文件
│   └── mysql/
│       └── initdb/                # MySQL 初始化脚本
│           └── 01-init.sql
│
├── docs/                          # 项目文档
│   ├── architecture.md            # 架构设计文档
│   ├── config-center.md           # 配置中心文档
│   ├── testing.md                 # 测试文档
│   ├── observability.md           # 可观测性文档
│   ├── performance.md             # 性能优化文档
│   ├── rpc-reference.md           # RPC 参考文档
│   └── openapi/                   # OpenAPI 规范
│       ├── gateway.openapi.yaml
│       └── gateway.openapi.json
│
├── scripts/                       # 脚本目录
│   ├── test_unit.sh               # 单元测试
│   ├── test_integration.sh        # 集成测试
│   ├── smoke.sh                   # 冒烟测试（支持 local/etcd 模式）
│   ├── bench_smoke.sh             # 压力测试
│   └── sync_runtime_configs.sh    # 同步配置到 Etcd
│
├── gateway-main/                  # 网关服务
│   ├── cmd/                       # 命令入口
│   │   ├── gateway/               # 网关主程序
│   │   └── configsync/            # 配置同步工具
│   ├── etc/                       # 本地配置文件
│   ├── internal/                  # 内部代码
│   │   ├── config/                # 配置定义
│   │   ├── handler/               # HTTP 处理器
│   │   ├── logic/                 # 业务逻辑
│   │   ├── middleware/            # 中间件
│   │   ├── svc/                   # 服务上下文
│   │   └── types/                 # 类型定义
│   ├── limiter/                   # 限流器
│   ├── logs/                      # 日志目录（gitignore）
│   └── README.md                  # 服务文档
│
├── user-main/                     # 用户服务
│   ├── cmd/                       # 命令入口
│   │   └── user/                  # 用户服务主程序
│   ├── etc/                       # 本地配置文件
│   ├── internal/                  # 内部代码
│   │   ├── config/                # 配置定义
│   │   ├── logic/                 # 业务逻辑
│   │   ├── model/                 # 数据模型
│   │   ├── server/                # gRPC 服务器
│   │   └── svc/                   # 服务上下文
│   ├── api/                       # Proto 定义
│   ├── sql/                       # SQL 脚本
│   ├── logs/                      # 日志目录（gitignore）
│   └── README.md                  # 服务文档
│
└── seckill-main/                  # 秒杀服务
    ├── cmd/                       # 命令入口
    │   └── sec_kill/              # 秒杀服务主程序
    ├── etc/                       # 本地配置文件
    ├── internal/                  # 内部代码
    │   ├── config/                # 配置定义
    │   ├── logic/                 # 业务逻辑
    │   ├── model/                 # 数据模型
    │   ├── mq/                    # 消息队列
    │   ├── server/                # gRPC 服务器
    │   └── svc/                   # 服务上下文
    ├── api/                       # Proto 定义
    ├── sql/                       # SQL 脚本
    ├── wrkbench/                  # 压测脚本
    ├── logs/                      # 日志目录（gitignore）
    └── README.md                  # 服务文档
```

### 架构图

```
┌─────────────┐
│   客户端     │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  网关服务    │ :8998
│  (Gateway)  │
└──────┬──────┘
       │
       ├─────────────┬─────────────┐
       ▼             ▼             ▼
┌──────────┐  ┌──────────┐  ┌──────────┐
│ 用户服务  │  │ 秒杀服务  │  │  Etcd    │
│ :8669    │  │ :8002    │  │ :20001   │
└─────┬────┘  └─────┬────┘  └──────────┘
      │             │
      │             ├─────────────┐
      │             ▼             ▼
      │        ┌────────┐   ┌─────────┐
      │        │ Redis  │   │  Kafka  │
      │        │ :6379  │   │ :9092   │
      │        └────────┘   └─────────┘
      │
      └──────────────┐
                     ▼
              ┌──────────┐
              │  MySQL   │
              │ :3307    │
              └──────────┘
```

## 🔧 技术栈

- **框架**: go-zero v1.10.1
- **语言**: Go 1.26+
- **数据库**: MySQL 8.0
- **缓存**: Redis
- **消息队列**: Kafka
- **服务发现**: Etcd
- **认证**: JWT

## 📦 服务说明

### 用户服务
- 端口: 8669
- 功能: 用户管理、认证
- 数据库: lottery_system

### 秒杀服务
- 端口: 8002
- 功能: 秒杀核心逻辑、库存管理
- 数据库: bitstorm
- 消息队列: Kafka

### 网关服务
- 端口: 8998
- 功能: HTTP API 网关、认证、限流
- 中间件: JWT 认证、限流、链路追踪

## 🎯 核心功能

### ✅ 已实现功能

- [x] 用户登录认证 (JWT)
- [x] 商品秒杀 (v1/v2/v3 三个版本)
- [x] Redis 库存扣减 (Lua 脚本原子操作)
- [x] Kafka 异步订单处理
- [x] 用户购买限额控制
- [x] 库存不足保护
- [x] 秒杀状态查询
- [x] API 限流保护

### 🔄 秒杀版本对比

| 版本 | 特点 | 适用场景 |
|------|------|----------|
| v1 | 同步处理，直接扣减数据库库存 | 低并发场景 |
| v2 | Redis 预扣库存，同步创建订单 | 中等并发场景 |
| v3 | Redis 预扣库存，Kafka 异步处理 | 高并发场景 |

## 🧪 测试账号

- 用户名: admin
- 密码: 123321
- 用户ID: 1

## 📝 API 接口

### 认证接口

```bash
# 登录
POST /login
Content-Type: application/json
{"username":"admin","password":"123321"}
```

### 用户接口

```bash
# 获取用户信息
GET /get_user_info
Authorization: Bearer <token>
```

### 秒杀接口

**推荐使用默认接口（v3 版本）**：

```bash
# 秒杀（默认 v3 版本，推荐）
POST /bitstorm/sec_kill
Authorization: Bearer <token>
Content-Type: application/json
{"goodsNum":"abc123","num":1}

# 查询秒杀状态
GET /bitstorm/v3/get_sec_kill_info?sec_num=<secNum>
Authorization: Bearer <token>
```

**指定版本接口**：

```bash
# v1 版本（同步处理，低并发场景）
POST /bitstorm/v1/sec_kill
Authorization: Bearer <token>
Content-Type: application/json
{"goodsNum":"abc123","num":1}

# v2 版本（Redis 预扣，中等并发场景）
POST /bitstorm/v2/sec_kill
Authorization: Bearer <token>
Content-Type: application/json
{"goodsNum":"abc123","num":1}

# v3 版本（Kafka 异步，高并发场景）⭐ 推荐
POST /bitstorm/v3/sec_kill
Authorization: Bearer <token>
Content-Type: application/json
{"goodsNum":"abc123","num":1}
```

**版本说明**：
- **默认接口** `/bitstorm/sec_kill` -> 自动使用 v3 版本（推荐）
- **v1 版本**：同步处理，直接扣减数据库库存，适合低并发场景
- **v2 版本**：Redis 预扣库存，同步创建订单，适合中等并发场景
- **v3 版本**：Redis 预扣库存，Kafka 异步处理，适合高并发场景 ⭐ 推荐

## 🔍 监控和管理

### Kafka UI
访问: http://localhost:8080

### 查看服务注册
```bash
docker exec mks-etcd etcdctl --endpoints=http://localhost:2379 get "" --prefix
```

### 查看 Redis 库存
```bash
docker exec mks-redis redis-cli -a 123456 GET "SK:Stock:1"
```

## 🛠️ 开发指南

### 构建服务

```bash
# 用户服务
cd user-main && go build -o /tmp/user-server ./cmd/user

# 秒杀服务
cd seckill-main && go build -o /tmp/seckill-server ./cmd/sec_kill

# 网关服务
cd gateway-main && go build -o /tmp/gateway-server ./cmd/gateway
```

### 配置文件

- 用户服务: `user-main/etc/user.yaml`
- 秒杀服务: `seckill-main/etc/seckill.yaml`
- 网关服务: `gateway-main/etc/gateway.yaml`

## 📊 性能优化

### Redis Lua 脚本优势
- 原子性操作，避免竞态条件
- 减少网络往返，提升性能
- 支持复杂的库存检查逻辑

### Kafka 异步处理优势
- 削峰填谷，应对高并发
- 解耦秒杀和订单服务
- 提升系统吞吐量

## 🐛 故障排查

### 服务无法启动
```bash
# 检查端口占用
ss -tlnp | grep -E ':(8669|8002|8998)'

# 杀死占用进程
pkill -f "(user-server|seckill-server|gateway-server)"
```

### RPC 连接失败
```bash
# 检查 etcd 服务
docker ps | grep etcd

# 检查服务注册
docker exec mks-etcd etcdctl --endpoints=http://localhost:2379 get "" --prefix
```

### 库存不足
```bash
# 重置库存
docker exec mks-redis redis-cli -a 123456 SET "SK:Stock:1" 10
```

## 📄 许可证

本项目仅供学习和研究使用。

## 🔗 相关链接

- [go-zero 官方文档](https://go-zero.dev/)
- [项目官网](https://bitoffer.cn)
