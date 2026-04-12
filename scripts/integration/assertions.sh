#!/usr/bin/env bash

set -euo pipefail

ASSERTIONS_PASSED=0
ASSERTIONS_FAILED=0
ASSERTIONS_LOG_DIR="${ASSERTIONS_LOG_DIR:-/tmp/e2e_test_logs}"

MYSQL_CONTAINER="${MYSQL_CONTAINER:-mks-mysql}"
REDIS_CONTAINER="${REDIS_CONTAINER:-mks-redis}"
USE_DOCKER="${USE_DOCKER:-true}"

mkdir -p "${ASSERTIONS_LOG_DIR}"

_mysql_exec() {
    local query="$1"
    if [[ "${USE_DOCKER}" == "true" ]]; then
        docker exec "${MYSQL_CONTAINER}" mysql -uroot -p123456 \
            --batch --raw --skip-column-names \
            bitstorm -e "${query}" 2>/dev/null
    else
        MYSQL_PWD="${MYSQL_PASSWORD:-123456}" mysql \
            -h "${MYSQL_HOST:-127.0.0.1}" \
            -P "${MYSQL_PORT:-3307}" \
            -u "${MYSQL_USER:-root}" \
            --batch --raw --skip-column-names \
            "${MYSQL_DB:-bitstorm}" \
            -e "${query}" 2>/dev/null
    fi
}

_redis_exec() {
    local db="$1"
    shift
    if [[ "${USE_DOCKER}" == "true" ]]; then
        docker exec "${REDIS_CONTAINER}" redis-cli -a 123456 -n "${db}" "$@" 2>/dev/null
    else
        redis-cli -h "${REDIS_HOST:-127.0.0.1}" -p "${REDIS_PORT:-6379}" -a "${REDIS_PASSWORD:-123456}" -n "${db}" "$@" 2>/dev/null
    fi
}

assert_equals() {
    local expected="$1"
    local actual="$2"
    local message="${3:-值相等检查}"

    if [[ "${expected}" == "${actual}" ]]; then
        echo "[PASS] ${message}: expected='${expected}', actual='${actual}'"
        ((ASSERTIONS_PASSED++))
        return 0
    else
        echo "[FAIL] ${message}: expected='${expected}', actual='${actual}'"
        ((ASSERTIONS_FAILED++))
        return 1
    fi
}

assert_not_equals() {
    local expected="$1"
    local actual="$2"
    local message="${3:-值不相等检查}"

    if [[ "${expected}" != "${actual}" ]]; then
        echo "[PASS] ${message}: not equals '${expected}'"
        ((ASSERTIONS_PASSED++))
        return 0
    else
        echo "[FAIL] ${message}: should not equal '${expected}'"
        ((ASSERTIONS_FAILED++))
        return 1
    fi
}

assert_greater_than() {
    local expected="$1"
    local actual="$2"
    local message="${3:-大于检查}"

    if [[ "${actual}" -gt "${expected}" ]]; then
        echo "[PASS] ${message}: ${actual} > ${expected}"
        ((ASSERTIONS_PASSED++))
        return 0
    else
        echo "[FAIL] ${message}: ${actual} <= ${expected}"
        ((ASSERTIONS_FAILED++))
        return 1
    fi
}

assert_less_or_equals() {
    local expected="$1"
    local actual="$2"
    local message="${3:-小于等于检查}"

    if [[ "${actual}" -le "${expected}" ]]; then
        echo "[PASS] ${message}: ${actual} <= ${expected}"
        ((ASSERTIONS_PASSED++))
        return 0
    else
        echo "[FAIL] ${message}: ${actual} > ${expected}"
        ((ASSERTIONS_FAILED++))
        return 1
    fi
}

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local message="${3:-包含检查}"

    if [[ "${haystack}" == *"${needle}"* ]]; then
        echo "[PASS] ${message}: contains '${needle}'"
        ((ASSERTIONS_PASSED++))
        return 0
    else
        echo "[FAIL] ${message}: does not contain '${needle}'"
        ((ASSERTIONS_FAILED++))
        return 1
    fi
}

assert_not_empty() {
    local value="$1"
    local message="${2:-非空检查}"

    if [[ -n "${value}" ]]; then
        echo "[PASS] ${message}: value is not empty"
        ((ASSERTIONS_PASSED++))
        return 0
    else
        echo "[FAIL] ${message}: value is empty"
        ((ASSERTIONS_FAILED++))
        return 1
    fi
}

assert_response_code() {
    local response="$1"
    local expected_code="$2"
    local message="${3:-响应码检查}"

    local actual_code
    actual_code=$(echo "${response}" | grep -o '"code":[0-9]*' | cut -d':' -f2 | head -1)

    if [[ -z "${actual_code}" ]]; then
        actual_code=$(echo "${response}" | grep -o '"Code":[0-9]*' | cut -d':' -f2 | head -1)
    fi

    if [[ "${actual_code}" == "${expected_code}" ]]; then
        echo "[PASS] ${message}: code=${actual_code}"
        ((ASSERTIONS_PASSED++))
        return 0
    else
        echo "[FAIL] ${message}: expected code=${expected_code}, actual code=${actual_code}"
        echo "       Response: ${response}"
        ((ASSERTIONS_FAILED++))
        return 1
    fi
}

assert_no_oversell() {
    local goods_id="$1"
    local initial_stock="$2"

    local current_stock
    current_stock=$(_mysql_exec "SELECT stock FROM t_seckill_stock WHERE goods_id=${goods_id};")

    if [[ -z "${current_stock}" ]]; then
        echo "[FAIL] 库存检查: 无法获取商品 ${goods_id} 的库存"
        ((ASSERTIONS_FAILED++))
        return 1
    fi

    if [[ "${current_stock}" -lt 0 ]]; then
        echo "[FAIL] 库存超卖检查: 商品 ${goods_id} 库存为负数 (${current_stock})"
        ((ASSERTIONS_FAILED++))
        return 1
    fi

    local sold_count=$((initial_stock - current_stock))
    if [[ "${sold_count}" -gt "${initial_stock}" ]]; then
        echo "[FAIL] 库存超卖检查: 商品 ${goods_id} 售出 ${sold_count} 超过初始库存 ${initial_stock}"
        ((ASSERTIONS_FAILED++))
        return 1
    fi

    echo "[PASS] 库存超卖检查: 商品 ${goods_id} 初始库存=${initial_stock}, 剩余库存=${current_stock}, 售出=${sold_count}"
    ((ASSERTIONS_PASSED++))
    return 0
}

assert_quota_not_exceeded() {
    local user_id="$1"
    local goods_id="$2"
    local max_quota="$3"

    local user_quota
    user_quota=$(_mysql_exec "SELECT num FROM t_user_quota WHERE user_id=${user_id} AND goods_id=${goods_id};")

    if [[ -z "${user_quota}" ]]; then
        user_quota=0
    fi

    if [[ "${user_quota}" -gt "${max_quota}" ]]; then
        echo "[FAIL] 限购检查: 用户 ${user_id} 购买商品 ${goods_id} 数量 ${user_quota} 超过限购 ${max_quota}"
        ((ASSERTIONS_FAILED++))
        return 1
    fi

    echo "[PASS] 限购检查: 用户 ${user_id} 购买商品 ${goods_id} 数量 ${user_quota} <= 限购 ${max_quota}"
    ((ASSERTIONS_PASSED++))
    return 0
}

assert_order_count() {
    local expected_min="$1"
    local expected_max="$2"
    local extra_where="${3:-}"

    local where_clause="1=1"
    [[ -n "${extra_where}" ]] && where_clause="${extra_where}"

    local order_count
    order_count=$(_mysql_exec "SELECT COUNT(*) FROM t_order WHERE ${where_clause};")

    if [[ "${order_count}" -lt "${expected_min}" ]]; then
        echo "[FAIL] 订单数量检查: 订单数 ${order_count} < 最小期望 ${expected_min}"
        ((ASSERTIONS_FAILED++))
        return 1
    fi

    if [[ "${order_count}" -gt "${expected_max}" ]]; then
        echo "[FAIL] 订单数量检查: 订单数 ${order_count} > 最大期望 ${expected_max}"
        ((ASSERTIONS_FAILED++))
        return 1
    fi

    echo "[PASS] 订单数量检查: 订单数 ${order_count} 在范围 [${expected_min}, ${expected_max}] 内"
    ((ASSERTIONS_PASSED++))
    return 0
}

assert_seckill_record_count() {
    local expected_min="$1"
    local expected_max="$2"
    local status="${3:-}"

    local where_clause="1=1"
    [[ -n "${status}" ]] && where_clause="status=${status}"

    local record_count
    record_count=$(_mysql_exec "SELECT COUNT(*) FROM t_seckill_record WHERE ${where_clause};")

    if [[ "${record_count}" -lt "${expected_min}" ]]; then
        echo "[FAIL] 秒杀记录数量检查: 记录数 ${record_count} < 最小期望 ${expected_min}"
        ((ASSERTIONS_FAILED++))
        return 1
    fi

    if [[ "${record_count}" -gt "${expected_max}" ]]; then
        echo "[FAIL] 秒杀记录数量检查: 记录数 ${record_count} > 最大期望 ${expected_max}"
        ((ASSERTIONS_FAILED++))
        return 1
    fi

    echo "[PASS] 秒杀记录数量检查: 记录数 ${record_count} 在范围 [${expected_min}, ${expected_max}] 内"
    ((ASSERTIONS_PASSED++))
    return 0
}

assert_redis_stock_consistent() {
    local goods_id="$1"
    local redis_db="${2:-0}"

    local redis_stock
    redis_stock=$(_redis_exec "${redis_db}" GET "SK:Stock:${goods_id}")

    local mysql_stock
    mysql_stock=$(_mysql_exec "SELECT stock FROM t_seckill_stock WHERE goods_id=${goods_id};")

    if [[ -z "${redis_stock}" ]]; then
        echo "[WARN] Redis库存检查: 商品 ${goods_id} Redis库存不存在"
        return 0
    fi

    if [[ "${redis_stock}" == "${mysql_stock}" ]]; then
        echo "[PASS] Redis库存一致性检查: 商品 ${goods_id} Redis=${redis_stock}, MySQL=${mysql_stock}"
        ((ASSERTIONS_PASSED++))
        return 0
    else
        echo "[FAIL] Redis库存一致性检查: 商品 ${goods_id} Redis=${redis_stock}, MySQL=${mysql_stock} 不一致"
        ((ASSERTIONS_FAILED++))
        return 1
    fi
}

print_assertion_summary() {
    echo ""
    echo "========================================"
    echo "断言统计汇总"
    echo "========================================"
    echo "通过: ${ASSERTIONS_PASSED}"
    echo "失败: ${ASSERTIONS_FAILED}"
    echo "总计: $((ASSERTIONS_PASSED + ASSERTIONS_FAILED))"
    echo ""

    if [[ "${ASSERTIONS_FAILED}" -gt 0 ]]; then
        echo "❌ 存在失败的断言"
        return 1
    else
        echo "✅ 所有断言通过"
        return 0
    fi
}

reset_assertion_counters() {
    ASSERTIONS_PASSED=0
    ASSERTIONS_FAILED=0
}
