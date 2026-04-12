#!/usr/bin/env bash
# 极限压测脚本 - 无限流版本

set -euo pipefail

GATEWAY_URL="${GATEWAY_URL:-http://127.0.0.1:8998}"
GOODS_ID="${GOODS_ID:-1}"
GOODS_NUM="${GOODS_NUM:-abc123}"
STOCK="${STOCK:-10000}"
USER_LIMIT="${USER_LIMIT:-10}"
NUM_USERS="${NUM_USERS:-100}"
REQUESTS="${REQUESTS:-10000}"
CONNECTIONS="${CONNECTIONS:-200}"

OUT_DIR="/home/monody/project/SecKill/tmp/perf_extreme"
mkdir -p "${OUT_DIR}"

echo "========================================"
echo "  SecKill 极限压测 (无限流)"
echo "========================================"
echo ""
echo "配置:"
echo "  Gateway:     ${GATEWAY_URL}"
echo "  库存:        ${STOCK}"
echo "  用户限购:    ${USER_LIMIT}"
echo "  压测用户数:  ${NUM_USERS}"
echo "  总请求数:    ${REQUESTS}"
echo "  并发数:      ${CONNECTIONS}"
echo ""

# 重置MySQL数据
echo "重置MySQL数据..."
docker exec mks-mysql mysql -uroot -p123456 bitstorm -e "
    DELETE FROM t_order WHERE goods_id = ${GOODS_ID};
    DELETE FROM t_seckill_record WHERE goods_id = ${GOODS_ID};
    DELETE FROM t_seckill_async_result WHERE goods_id = ${GOODS_ID};
    DELETE FROM t_user_quota WHERE goods_id = ${GOODS_ID};
    INSERT INTO t_quota(goods_id, num) VALUES (${GOODS_ID}, ${USER_LIMIT})
      ON DUPLICATE KEY UPDATE num = VALUES(num);
    UPDATE t_seckill_stock SET stock = ${STOCK} WHERE goods_id = ${GOODS_ID};
" 2>/dev/null

# 重置Redis
echo "重置Redis数据..."
docker exec mks-redis redis-cli -a 123456 -n 0 FLUSHDB 2>/dev/null
docker exec mks-redis redis-cli -a 123456 -n 0 SET "SK:Stock:${GOODS_ID}" "${STOCK}" EX 3600 2>/dev/null
docker exec mks-redis redis-cli -a 123456 -n 0 SET "SK:Limit${GOODS_ID}" "${USER_LIMIT}" EX 3600 2>/dev/null
docker exec mks-redis redis-cli -a 123456 -n 0 SET "goodsInfo:${GOODS_NUM}" '{"ID":1,"GoodsNum":"abc123","GoodsName":"redhat","Price":18,"PicUrl":"http://","Seller":135}' EX 3600 2>/dev/null

echo "数据已重置"
echo ""

# 获取用户Token
echo "获取用户Token..."
TOKENS_FILE="${OUT_DIR}/tokens.txt"
: > "${TOKENS_FILE}"

for i in $(seq 1 "${NUM_USERS}"); do
    username="perf_user${i}"
    response=$(curl -s -X POST "${GATEWAY_URL}/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"${username}\",\"password\":\"123321\"}" 2>/dev/null || echo "")
    
    token=$(echo "${response}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    
    if [[ -n "${token}" ]]; then
        echo "${token}" >> "${TOKENS_FILE}"
    fi
    
    if (( i % 50 == 0 )); then
        echo "  已获取 ${i}/${NUM_USERS} 个Token..."
    fi
done

TOTAL_TOKENS=$(wc -l < "${TOKENS_FILE}")
echo "成功获取 ${TOTAL_TOKENS} 个Token"
echo ""

# 执行压测
run_benchmark() {
    local version="$1"
    local endpoint="${GATEWAY_URL}/bitstorm/${version}/sec_kill"
    local output="${OUT_DIR}/${version}_$(date +%Y%m%d_%H%M%S).hey"
    
    echo "========================================"
    echo "  压测 ${version}"
    echo "========================================"
    
    # 重置数据
    docker exec mks-mysql mysql -uroot -p123456 bitstorm -e "
        DELETE FROM t_order WHERE goods_id = ${GOODS_ID};
        DELETE FROM t_seckill_record WHERE goods_id = ${GOODS_ID};
        DELETE FROM t_seckill_async_result WHERE goods_id = ${GOODS_ID};
        DELETE FROM t_user_quota WHERE goods_id = ${GOODS_ID};
    " 2>/dev/null
    docker exec mks-redis redis-cli -a 123456 -n 0 SET "SK:Stock:${GOODS_ID}" "${STOCK}" EX 3600 2>/dev/null
    
    # 使用第一个Token进行压测
    local token
    token=$(head -1 "${TOKENS_FILE}")
    
    echo "压测 ${version} 接口..."
    hey -n "${REQUESTS}" -c "${CONNECTIONS}" \
        -m POST \
        -H "Authorization: Bearer ${token}" \
        -H "Content-Type: application/json" \
        -d "{\"goodsNum\":\"${GOODS_NUM}\",\"num\":1}" \
        "${endpoint}" 2>&1 | tee "${output}"
    
    # 收集统计
    echo ""
    echo "=== 业务统计 ==="
    local order_count
    order_count=$(docker exec mks-mysql mysql -uroot -p123456 bitstorm -N -e "SELECT COUNT(*) FROM t_order WHERE goods_id = ${GOODS_ID};" 2>/dev/null)
    local remaining_stock
    remaining_stock=$(docker exec mks-mysql mysql -uroot -p123456 bitstorm -N -e "SELECT stock FROM t_seckill_stock WHERE goods_id = ${GOODS_ID};" 2>/dev/null)
    
    echo "订单数量:   ${order_count}"
    echo "剩余库存:   ${remaining_stock}"
    echo "已售数量:   $((STOCK - remaining_stock))"
    
    if [[ "${order_count}" -gt "${STOCK}" ]]; then
        echo "⚠️  检测到超卖！"
    else
        echo "✅ 无超卖"
    fi
    echo ""
}

# 分别压测三个版本
for version in v1 v2 v3; do
    run_benchmark "${version}"
done

echo "========================================"
echo "  压测完成"
echo "========================================"
echo "详细报告保存在: ${OUT_DIR}"
