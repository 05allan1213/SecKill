# 秒杀系统测试文档

## 项目概述

这是一个基于 go-zero 框架的微服务秒杀系统，包含以下核心服务：

- **用户服务**: 用户管理和认证
- **秒杀服务**: 秒杀核心业务逻辑
- **网关服务**: HTTP API 网关

## 环境要求

### 基础设施
- Docker & Docker Compose
- Go 1.26+

### 服务端口
- MySQL: 3307
- Redis: 6379
- Etcd: 20001
- Kafka: 9092
- Kafka UI: 8080
- 用户服务: 8669
- 秒杀服务: 8002
- 网关服务: 8998

## 启动步骤

### 1. 启动基础设施

```bash
cd "/home/monody/project/Microsecond killing service"
docker compose up -d
```

验证服务状态：
```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

### 2. 构建微服务

```bash
# 构建用户服务
cd user-main
go build -o /tmp/user-server ./cmd/user

# 构建秒杀服务
cd ../seckill-main
go build -o /tmp/seckill-server ./cmd/sec_kill

# 构建网关服务
cd ../gateway-main
go build -o /tmp/gateway-server ./cmd/gateway
```

### 3. 启动微服务

在三个独立的终端中运行：

**终端 1 - 用户服务：**
```bash
cd "/home/monody/project/Microsecond killing service/user-main"
/tmp/user-server -f etc/user.yaml
```

**终端 2 - 秒杀服务：**
```bash
cd "/home/monody/project/Microsecond killing service/seckill-main"
/tmp/seckill-server -f etc/seckill.yaml
```

**终端 3 - 网关服务：**
```bash
cd "/home/monody/project/Microsecond killing service/gateway-main"
/tmp/gateway-server -f etc/gateway.yaml
```

### 4. 初始化 Redis 库存数据

```bash
# 设置商品ID为1的库存为3
docker exec mks-redis redis-cli -a 123456 SET "SK:Stock:1" 3
```

## API 测试

### 1. 用户认证测试

#### 1.1 登录获取 JWT Token

**请求：**
```bash
curl -X POST http://localhost:8998/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123321"}'
```

**预期响应：**
```json
{
  "code": 200,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expire": "2026-04-10T22:16:18+08:00"
}
```

**测试说明：**
- 使用预置的测试账号：admin/123321
- 返回的 token 用于后续所有需要认证的接口
- token 有效期为 1 小时

#### 1.2 未认证访问测试

**请求：**
```bash
curl -X POST http://localhost:8998/bitstorm/sec_kill \
  -H "Content-Type: application/json" \
  -d '{"goodsNum":"abc123","num":1}'
```

**预期响应：**
```json
{
  "code": 401,
  "message": "missing or malformed jwt"
}
```

**测试说明：**
- 未携带 token 的请求应被拦截
- 返回 401 未授权错误

### 2. 用户信息测试

#### 2.1 获取用户信息

**请求：**
```bash
curl -X GET "http://localhost:8998/get_user_info" \
  -H "Authorization: Bearer <your_token>"
```

**预期响应：**
```json
{
  "welcome": "admin"
}
```

### 3. 秒杀功能测试

#### 3.1 秒杀接口（默认 v3 版本，推荐）

**请求：**
```bash
curl -X POST http://localhost:8998/bitstorm/sec_kill \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your_token>" \
  -d '{"goodsNum":"abc123","num":1}'
```

**成功响应：**
```json
{
  "code": 0,
  "message": "",
  "data": {
    "secNum": "071971d5-25f9-4fdc-93f5-49008fb00b81"
  }
}
```

**库存不足响应：**
```json
{
  "code": 500,
  "message": "rpc error: code = Unknown desc = stock not enough"
}
```

**测试说明：**
- 默认接口自动使用 v3 版本（Kafka 异步处理）
- 返回秒杀单号（secNum）用于查询状态
- 库存在 Redis 中实时扣减
- 性能最优，适合高并发场景

#### 3.2 指定版本秒杀接口

**v1 版本（同步处理）：**
```bash
curl -X POST http://localhost:8998/bitstorm/v1/sec_kill \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your_token>" \
  -d '{"goodsNum":"abc123","num":1}'
```

**v2 版本（Redis 预扣）：**
```bash
curl -X POST http://localhost:8998/bitstorm/v2/sec_kill \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your_token>" \
  -d '{"goodsNum":"abc123","num":1}'
```

**v3 版本（Kafka 异步）：**
```bash
curl -X POST http://localhost:8998/bitstorm/v3/sec_kill \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your_token>" \
  -d '{"goodsNum":"abc123","num":1}'
```

#### 3.3 查询秒杀状态

**请求：**
```bash
curl -X GET "http://localhost:8998/bitstorm/v3/get_sec_kill_info?sec_num=<secNum>" \
  -H "Authorization: Bearer <your_token>"
```

**预期响应：**
```json
{
  "code": 0,
  "message": "",
  "data": {
    "status": 2,
    "orderNum": "3341461e-058f-40e1-a04e-1b1fc96fbb1b",
    "secNum": "071971d5-25f9-4fdc-93f5-49008fb00b81",
    "goodsNum": ""
  }
}
```

**状态说明：**
- status: 1 - 秒杀中
- status: 2 - 已生成订单，待支付

### 4. 库存管理测试

#### 4.1 查看当前库存

**请求：**
```bash
docker exec mks-redis redis-cli -a 123456 GET "SK:Stock:1"
```

**预期响应：**
```
"2"
```

#### 4.2 查看数据库库存

**请求：**
```bash
docker exec mks-mysql mysql -uroot -p123456 \
  -e "USE bitstorm; SELECT * FROM t_seckill_stock;"
```

**预期响应：**
```
id      goods_id        stock   create_time             modify_time
1       1               3       2026-04-10 12:53:25     2026-04-10 12:53:25
```

## 功能验证清单

### ✅ 认证功能
- [ ] 登录接口返回有效 JWT Token
- [ ] 未认证请求被正确拦截
- [ ] Token 过期后自动失效

### ✅ 秒杀功能
- [ ] 秒杀成功返回秒杀单号
- [ ] Redis 库存正确扣减
- [ ] 库存不足时返回错误
- [ ] 重复秒杀被正确拦截

### ✅ 订单功能
- [ ] 秒杀状态查询正常
- [ ] 订单号正确生成
- [ ] 订单状态正确更新

### ✅ 数据一致性
- [ ] Redis 库存与数据库一致
- [ ] 秒杀记录正确保存
- [ ] 订单记录正确保存

## 性能测试

### 并发测试（可选）

使用 Apache Bench 进行简单的并发测试：

```bash
# 安装 ab
sudo apt-get install apache2-utils

# 并发 100 个请求
ab -n 100 -c 10 \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -p payload.json \
  http://localhost:8998/bitstorm/v3/sec_kill
```

payload.json 内容：
```json
{"goodsNum":"abc123","num":1}
```

## 监控和调试

### 查看服务日志

```bash
# 查看用户服务日志
# 在运行用户服务的终端查看输出

# 查看秒杀服务日志
# 在运行秒杀服务的终端查看输出

# 查看网关服务日志
# 在运行网关服务的终端查看输出
```

### 查看 Kafka 消息

访问 Kafka UI：http://localhost:8080

### 查看 Etcd 服务注册

```bash
docker exec mks-etcd etcdctl --endpoints=http://localhost:2379 get "" --prefix
```

**预期输出：**
```
user.rpc/4773144316427125317
127.0.0.1:8669
seckill.rpc/4773144316427125320
127.0.0.1:8002
```

## 常见问题

### 1. 服务无法启动

**问题：** 端口被占用

**解决方案：**
```bash
# 查看端口占用
ss -tlnp | grep -E ':(8669|8002|8998)'

# 杀死占用进程
kill -9 <PID>
```

### 2. RPC 连接失败

**问题：** "connection error" 或 "bad resolver state"

**解决方案：**
- 检查 etcd 服务是否运行
- 确认服务已正确注册到 etcd
- 验证配置文件中的 etcd 地址

### 3. 库存不足错误

**问题：** "stock not enough"

**解决方案：**
```bash
# 重新初始化库存
docker exec mks-redis redis-cli -a 123456 SET "SK:Stock:1" 10
```

### 4. Token 失效

**问题：** "token is invalid"

**解决方案：**
- 重新登录获取新 token
- 检查 token 是否正确复制（包含完整字符串）

## 测试数据

### 预置用户数据
- 用户名: admin
- 密码: 123321
- 用户ID: 1

### 预置商品数据
- 商品编号: abc123
- 商品名称: redhat
- 价格: 18
- 初始库存: 3

## 清理和重置

### 重置库存数据

```bash
# 重置 Redis 库存
docker exec mks-redis redis-cli -a 123456 SET "SK:Stock:1" 3

# 清空用户购买记录
docker exec mks-redis redis-cli -a 123456 DEL "SK:UserSecKilledNum:1:1"
docker exec mks-redis redis-cli -a 123456 DEL "SK:UserGoodsSecNum:1:1"
```

### 重置数据库

```bash
# 重新初始化数据库
docker exec mks-mysql mysql -uroot -p123456 \
  -e "USE bitstorm; UPDATE t_seckill_stock SET stock=3 WHERE id=1;"
```

### 停止所有服务

```bash
# 停止微服务
pkill -f "(user-server|seckill-server|gateway-server)"

# 停止基础设施
cd "/home/monody/project/Microsecond killing service"
docker compose down
```

## 架构说明

### 技术栈
- **框架**: go-zero
- **数据库**: MySQL 8.0
- **缓存**: Redis
- **消息队列**: Kafka
- **服务发现**: Etcd
- **认证**: JWT

### 秒杀流程（v3 版本）

1. 用户发起秒杀请求
2. 网关验证 JWT Token
3. 秒杀服务执行 Redis Lua 脚本扣减库存
4. 发送消息到 Kafka
5. 消费者异步处理订单
6. 用户查询秒杀状态

### 库存扣减机制

使用 Redis Lua 脚本保证原子性：
- 检查用户是否已在秒杀中
- 检查用户购买限额
- 检查库存是否充足
- 扣减库存并记录用户购买数量

## 测试报告模板

### 测试环境
- 操作系统: 
- Go 版本: 
- Docker 版本: 
- 测试时间: 

### 测试结果
| 测试项 | 预期结果 | 实际结果 | 状态 |
|--------|----------|----------|------|
| 用户登录 | 返回 JWT Token | | ✅/❌ |
| 未认证访问 | 返回 401 错误 | | ✅/❌ |
| 秒杀成功 | 返回秒杀单号 | | ✅/❌ |
| 库存扣减 | Redis 库存减少 | | ✅/❌ |
| 库存不足 | 返回错误信息 | | ✅/❌ |
| 查询状态 | 返回订单信息 | | ✅/❌ |

### 性能数据
- 平均响应时间: 
- 并发处理能力: 
- 错误率: 

### 问题记录
| 问题描述 | 严重程度 | 解决方案 | 状态 |
|----------|----------|----------|------|
| | | | |

## 附录

### 完整测试脚本

创建文件 `test_api.sh`：

```bash
#!/bin/bash

BASE_URL="http://localhost:8998"

echo "=== 1. 登录测试 ==="
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123321"}')

echo "登录响应: $LOGIN_RESPONSE"

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
echo "Token: $TOKEN"

echo -e "\n=== 2. 获取用户信息测试 ==="
curl -s -X GET "$BASE_URL/get_user_info" \
  -H "Authorization: Bearer $TOKEN"
echo

echo -e "\n=== 3. 秒杀测试 ==="
SECKILL_RESPONSE=$(curl -s -X POST $BASE_URL/bitstorm/v3/sec_kill \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"goodsNum":"abc123","num":1}')

echo "秒杀响应: $SECKILL_RESPONSE"

SEC_NUM=$(echo $SECKILL_RESPONSE | jq -r '.data.secNum')
echo "秒杀单号: $SEC_NUM"

echo -e "\n=== 4. 查询秒杀状态测试 ==="
curl -s -X GET "$BASE_URL/bitstorm/v3/get_sec_kill_info?sec_num=$SEC_NUM" \
  -H "Authorization: Bearer $TOKEN"
echo

echo -e "\n=== 5. 查看库存 ==="
docker exec mks-redis redis-cli -a 123456 GET "SK:Stock:1"

echo -e "\n测试完成！"
```

运行测试：
```bash
chmod +x test_api.sh
./test_api.sh
```
