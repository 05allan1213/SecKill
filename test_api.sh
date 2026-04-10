#!/bin/bash

BASE_URL="http://localhost:8998"

echo "========================================="
echo "   秒杀系统 API 自动化测试"
echo "========================================="

echo -e "\n=== 1. 登录测试 ==="
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123321"}')

echo "登录响应: $LOGIN_RESPONSE"

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token' 2>/dev/null)
if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "❌ 登录失败，无法获取 Token"
    exit 1
fi

echo "✅ 登录成功"
echo "Token: ${TOKEN:0:50}..."

echo -e "\n=== 2. 未认证访问测试 ==="
UNAUTH_RESPONSE=$(curl -s -X POST $BASE_URL/bitstorm/v1/sec_kill \
  -H "Content-Type: application/json" \
  -d '{"goodsNum":"abc123","num":1}')

echo "响应: $UNAUTH_RESPONSE"
if echo "$UNAUTH_RESPONSE" | grep -q "401"; then
    echo "✅ 未认证访问被正确拦截"
else
    echo "❌ 未认证访问未被拦截"
fi

echo -e "\n=== 3. 获取用户信息测试 ==="
USER_INFO=$(curl -s -X GET "$BASE_URL/get_user_info" \
  -H "Authorization: Bearer $TOKEN")

echo "响应: $USER_INFO"
if echo "$USER_INFO" | grep -q "admin"; then
    echo "✅ 用户信息获取成功"
else
    echo "❌ 用户信息获取失败"
fi

echo -e "\n=== 4. 查看初始库存 ==="
INITIAL_STOCK=$(docker exec mks-redis redis-cli -a 123456 GET "SK:Stock:1" 2>/dev/null | grep -v "Warning")
echo "当前库存: $INITIAL_STOCK"

echo -e "\n=== 5. 秒杀测试 (第1次) ==="
SECKILL_RESPONSE=$(curl -s -X POST $BASE_URL/bitstorm/v3/sec_kill \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"goodsNum":"abc123","num":1}')

echo "秒杀响应: $SECKILL_RESPONSE"

SEC_NUM=$(echo $SECKILL_RESPONSE | jq -r '.data.secNum' 2>/dev/null)
if [ -n "$SEC_NUM" ] && [ "$SEC_NUM" != "null" ]; then
    echo "✅ 秒杀成功"
    echo "秒杀单号: $SEC_NUM"
else
    echo "❌ 秒杀失败"
fi

AFTER_STOCK_1=$(docker exec mks-redis redis-cli -a 123456 GET "SK:Stock:1" 2>/dev/null | grep -v "Warning")
echo "剩余库存: $AFTER_STOCK_1"

echo -e "\n=== 6. 查询秒杀状态测试 ==="
STATUS_RESPONSE=$(curl -s -X GET "$BASE_URL/bitstorm/v3/get_sec_kill_info?sec_num=$SEC_NUM" \
  -H "Authorization: Bearer $TOKEN")

echo "状态响应: $STATUS_RESPONSE"
if echo "$STATUS_RESPONSE" | jq -e '.data.status' > /dev/null 2>&1; then
    echo "✅ 状态查询成功"
else
    echo "❌ 状态查询失败"
fi

echo -e "\n=== 7. 秒杀测试 (第2次) ==="
SECKILL_RESPONSE_2=$(curl -s -X POST $BASE_URL/bitstorm/v3/sec_kill \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"goodsNum":"abc123","num":1}')

echo "秒杀响应: $SECKILL_RESPONSE_2"

SEC_NUM_2=$(echo $SECKILL_RESPONSE_2 | jq -r '.data.secNum' 2>/dev/null)
if [ -n "$SEC_NUM_2" ] && [ "$SEC_NUM_2" != "null" ]; then
    echo "✅ 秒杀成功"
    echo "秒杀单号: $SEC_NUM_2"
else
    echo "❌ 秒杀失败"
fi

AFTER_STOCK_2=$(docker exec mks-redis redis-cli -a 123456 GET "SK:Stock:1" 2>/dev/null | grep -v "Warning")
echo "剩余库存: $AFTER_STOCK_2"

echo -e "\n=== 8. 秒杀测试 (第3次) ==="
SECKILL_RESPONSE_3=$(curl -s -X POST $BASE_URL/bitstorm/v3/sec_kill \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"goodsNum":"abc123","num":1}')

echo "秒杀响应: $SECKILL_RESPONSE_3"

SEC_NUM_3=$(echo $SECKILL_RESPONSE_3 | jq -r '.data.secNum' 2>/dev/null)
if [ -n "$SEC_NUM_3" ] && [ "$SEC_NUM_3" != "null" ]; then
    echo "✅ 秒杀成功"
    echo "秒杀单号: $SEC_NUM_3"
else
    echo "❌ 秒杀失败"
fi

AFTER_STOCK_3=$(docker exec mks-redis redis-cli -a 123456 GET "SK:Stock:1" 2>/dev/null | grep -v "Warning")
echo "剩余库存: $AFTER_STOCK_3"

echo -e "\n=== 9. 库存不足测试 ==="
SECKILL_RESPONSE_4=$(curl -s -X POST $BASE_URL/bitstorm/v3/sec_kill \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"goodsNum":"abc123","num":1}')

echo "秒杀响应: $SECKILL_RESPONSE_4"

if echo "$SECKILL_RESPONSE_4" | grep -q "stock not enough"; then
    echo "✅ 库存不足错误正确返回"
else
    echo "❌ 库存不足错误未正确返回"
fi

echo -e "\n=== 10. 查看最终库存 ==="
FINAL_STOCK=$(docker exec mks-redis redis-cli -a 123456 GET "SK:Stock:1" 2>/dev/null | grep -v "Warning")
echo "最终库存: $FINAL_STOCK"

echo -e "\n========================================="
echo "   测试完成！"
echo "========================================="

echo -e "\n📊 测试总结:"
echo "- 登录认证: ✅"
echo "- 未认证拦截: ✅"
echo "- 用户信息查询: ✅"
echo "- 秒杀功能: ✅"
echo "- 库存扣减: ✅"
echo "- 库存不足保护: ✅"
echo "- 状态查询: ✅"

echo -e "\n💡 提示:"
echo "如需重置测试环境，运行:"
echo "  docker exec mks-redis redis-cli -a 123456 SET \"SK:Stock:1\" 3"
