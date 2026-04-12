#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/assertions.sh"

GATEWAY_URL="${GATEWAY_URL:-http://127.0.0.1:8998}"
SECKILL_REDIS_DB="${SECKILL_REDIS_DB:-0}"

TEST_GOODS_ID="${TEST_GOODS_ID:-1}"
TEST_GOODS_NUM="${TEST_GOODS_NUM:-abc123}"
TEST_USERNAME="${TEST_USERNAME:-admin}"
TEST_PASSWORD="${TEST_PASSWORD:-123321}"
TEST_USER_ID="${TEST_USER_ID:-1}"
TEST_STOCK="${TEST_STOCK:-10}"
TEST_QUOTA="${TEST_QUOTA:-5}"

TOKEN=""

mysql_exec() {
    _mysql_exec "$1"
}

redis_exec() {
    local db="$1"
    shift
    _redis_exec "${db}" "$@"
}

login_and_get_token() {
    echo ">>> 登录获取 Token..."

    local response
    response=$(curl -sS -X POST "${GATEWAY_URL}/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"${TEST_USERNAME}\",\"password\":\"${TEST_PASSWORD}\"}")

    TOKEN=$(echo "${response}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

    if [[ -z "${TOKEN}" ]]; then
        echo "[ERROR] 登录失败: ${response}"
        return 1
    fi

    echo "[INFO] 登录成功, Token: ${TOKEN:0:30}..."
    return 0
}

reset_test_data() {
    echo ">>> 重置测试数据..."

    mysql_exec "
        DELETE FROM t_order WHERE goods_id = ${TEST_GOODS_ID};
        DELETE FROM t_seckill_record WHERE goods_id = ${TEST_GOODS_ID};
        DELETE FROM t_user_quota WHERE goods_id = ${TEST_GOODS_ID};
        DELETE FROM t_seckill_async_result WHERE goods_id = ${TEST_GOODS_ID};
        INSERT INTO t_quota(goods_id, num) VALUES (${TEST_GOODS_ID}, ${TEST_QUOTA})
          ON DUPLICATE KEY UPDATE num = VALUES(num);
        UPDATE t_seckill_stock SET stock = ${TEST_STOCK} WHERE goods_id = ${TEST_GOODS_ID};
    "

    redis_exec "${SECKILL_REDIS_DB}" DEL "SK:Stock:${TEST_GOODS_ID}"
    redis_exec "${SECKILL_REDIS_DB}" DEL "SK:Limit${TEST_GOODS_ID}"
    redis_exec "${SECKILL_REDIS_DB}" DEL "SK:UserGoodsSecNum:${TEST_USER_ID}:${TEST_GOODS_ID}"
    redis_exec "${SECKILL_REDIS_DB}" DEL "SK:UserSecKilledNum:${TEST_USER_ID}:${TEST_GOODS_ID}"

    redis_exec "${SECKILL_REDIS_DB}" SET "SK:Stock:${TEST_GOODS_ID}" "${TEST_STOCK}"
    redis_exec "${SECKILL_REDIS_DB}" SET "SK:Limit${TEST_GOODS_ID}" "${TEST_QUOTA}"

    echo "[INFO] 测试数据已重置 (stock=${TEST_STOCK}, quota=${TEST_QUOTA})"
}

test_login() {
    echo ""
    echo "========================================"
    echo "测试用例: 用户登录"
    echo "========================================"

    reset_assertion_counters

    local response
    response=$(curl -sS -X POST "${GATEWAY_URL}/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"${TEST_USERNAME}\",\"password\":\"${TEST_PASSWORD}\"}")

    assert_response_code "${response}" "200" "登录接口响应码"

    local token
    token=$(echo "${response}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    assert_not_empty "${token}" "Token 不为空"

    print_assertion_summary
}

test_v1_flow() {
    echo ""
    echo "========================================"
    echo "测试用例: V1 完整流程 (数据库扣减)"
    echo "========================================"

    reset_assertion_counters
    reset_test_data

    login_and_get_token || return 1

    echo ">>> 执行 V1 秒杀..."
    local response
    response=$(curl -sS -X POST "${GATEWAY_URL}/bitstorm/v1/sec_kill" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{\"goodsNum\":\"${TEST_GOODS_NUM}\",\"num\":1}")

    assert_response_code "${response}" "0" "V1 秒杀接口业务响应码"

    local order_num
    order_num=$(echo "${response}" | grep -o '"orderNum":"[^"]*"' | cut -d'"' -f4)
    assert_not_empty "${order_num}" "V1 秒杀返回订单号"

    echo ">>> 验证数据库状态..."
    assert_no_oversell "${TEST_GOODS_ID}" "${TEST_STOCK}"

    assert_quota_not_exceeded "${TEST_USER_ID}" "${TEST_GOODS_ID}" "${TEST_QUOTA}"

    local order_count
    order_count=$(mysql_exec "SELECT COUNT(*) FROM t_order WHERE goods_id = ${TEST_GOODS_ID};")
    assert_equals "1" "${order_count}" "V1 订单数量为 1"

    print_assertion_summary
}

test_v2_flow() {
    echo ""
    echo "========================================"
    echo "测试用例: V2 完整流程 (Redis 扣减)"
    echo "========================================"

    reset_assertion_counters
    reset_test_data

    login_and_get_token || return 1

    echo ">>> 执行 V2 秒杀..."
    local response
    response=$(curl -sS -X POST "${GATEWAY_URL}/bitstorm/v2/sec_kill" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{\"goodsNum\":\"${TEST_GOODS_NUM}\",\"num\":1}")

    assert_response_code "${response}" "0" "V2 秒杀接口业务响应码"

    local order_num
    order_num=$(echo "${response}" | grep -o '"orderNum":"[^"]*"' | cut -d'"' -f4)
    assert_not_empty "${order_num}" "V2 秒杀返回订单号"

    echo ">>> 验证数据库状态..."
    assert_no_oversell "${TEST_GOODS_ID}" "${TEST_STOCK}"

    assert_quota_not_exceeded "${TEST_USER_ID}" "${TEST_GOODS_ID}" "${TEST_QUOTA}"

    print_assertion_summary
}

test_v3_flow() {
    echo ""
    echo "========================================"
    echo "测试用例: V3 完整流程 (异步消息队列)"
    echo "========================================"

    reset_assertion_counters
    reset_test_data

    login_and_get_token || return 1

    echo ">>> 执行 V3 秒杀..."
    local response
    response=$(curl -sS -X POST "${GATEWAY_URL}/bitstorm/v3/sec_kill" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{\"goodsNum\":\"${TEST_GOODS_NUM}\",\"num\":1}")

    assert_response_code "${response}" "0" "V3 秒杀接口业务响应码"

    local sec_num
    sec_num=$(echo "${response}" | grep -o '"secNum":"[^"]*"' | cut -d'"' -f4)
    assert_not_empty "${sec_num}" "V3 秒杀返回秒杀号"

    echo ">>> 等待异步处理完成..."
    local max_wait=30
    local wait_count=0
    local status=""

    while [[ "${status}" != "2" && "${status}" != "6" && "${wait_count}" -lt "${max_wait}" ]]; do
        sleep 1
        ((wait_count++))
        local status_response
        status_response=$(curl -sS "${GATEWAY_URL}/bitstorm/v3/get_sec_kill_info?sec_num=${sec_num}" \
            -H "Authorization: Bearer ${TOKEN}")
        status=$(echo "${status_response}" | grep -o '"status":[0-9]*' | cut -d':' -f2)
        echo "    [${wait_count}/${max_wait}] 状态: ${status}"
    done

    if [[ "${status}" == "2" ]]; then
        echo "[PASS] V3 异步处理成功 (status=2)"
        ((ASSERTIONS_PASSED++))
    elif [[ "${status}" == "6" ]]; then
        echo "[FAIL] V3 异步处理失败 (status=6)"
        ((ASSERTIONS_FAILED++))
    else
        echo "[FAIL] V3 异步处理超时"
        ((ASSERTIONS_FAILED++))
    fi

    echo ">>> 验证数据库状态..."
    assert_no_oversell "${TEST_GOODS_ID}" "${TEST_STOCK}"

    assert_quota_not_exceeded "${TEST_USER_ID}" "${TEST_GOODS_ID}" "${TEST_QUOTA}"

    print_assertion_summary
}

test_concurrent_seckill() {
    echo ""
    echo "========================================"
    echo "测试用例: 并发秒杀测试"
    echo "========================================"

    reset_assertion_counters

    local concurrent_stock="${CONCURRENT_STOCK:-20}"
    local concurrent_requests="${CONCURRENT_REQUESTS:-50}"
    local concurrent_users="${CONCURRENT_USERS:-10}"
    local concurrent_quota="${CONCURRENT_QUOTA:-3}"

    echo ">>> 配置: 库存=${concurrent_stock}, 请求数=${concurrent_requests}, 用户数=${concurrent_users}, 限购=${concurrent_quota}"

    echo ">>> 重置测试数据..."
    mysql_exec "
        DELETE FROM t_order WHERE goods_id = ${TEST_GOODS_ID};
        DELETE FROM t_seckill_record WHERE goods_id = ${TEST_GOODS_ID};
        DELETE FROM t_user_quota WHERE goods_id = ${TEST_GOODS_ID};
        INSERT INTO t_quota(goods_id, num) VALUES (${TEST_GOODS_ID}, ${concurrent_quota})
          ON DUPLICATE KEY UPDATE num = VALUES(num);
        UPDATE t_seckill_stock SET stock = ${concurrent_stock} WHERE goods_id = ${TEST_GOODS_ID};
    "

    redis_exec "${SECKILL_REDIS_DB}" DEL "SK:Stock:${TEST_GOODS_ID}"
    redis_exec "${SECKILL_REDIS_DB}" DEL "SK:Limit${TEST_GOODS_ID}"
    redis_exec "${SECKILL_REDIS_DB}" SET "SK:Stock:${TEST_GOODS_ID}" "${concurrent_stock}"
    redis_exec "${SECKILL_REDIS_DB}" SET "SK:Limit${TEST_GOODS_ID}" "${concurrent_quota}"

    for uid in $(seq 1 "${concurrent_users}"); do
        redis_exec "${SECKILL_REDIS_DB}" DEL "SK:UserGoodsSecNum:${uid}:${TEST_GOODS_ID}"
        redis_exec "${SECKILL_REDIS_DB}" DEL "SK:UserSecKilledNum:${uid}:${TEST_GOODS_ID}"
    done

    echo ">>> 执行并发秒杀 (V2 版本)..."
    local success_count=0
    local fail_count=0
    local temp_dir
    temp_dir=$(mktemp -d)
    local pids=()

    for i in $(seq 1 "${concurrent_requests}"); do
        local uid=$(( (i - 1) % concurrent_users + 1 ))
        local username="user${uid}"

        (
            local user_token
            user_token=$(curl -sS -X POST "${GATEWAY_URL}/login" \
                -H "Content-Type: application/json" \
                -d "{\"username\":\"${username}\",\"password\":\"123321\"}" | \
                grep -o '"token":"[^"]*"' | cut -d'"' -f4)

            if [[ -z "${user_token}" ]]; then
                echo "fail" > "${temp_dir}/${i}.result"
                exit 0
            fi

            local response
            response=$(curl -sS -X POST "${GATEWAY_URL}/bitstorm/v2/sec_kill" \
                -H "Authorization: Bearer ${user_token}" \
                -H "Content-Type: application/json" \
                -d "{\"goodsNum\":\"${TEST_GOODS_NUM}\",\"num\":1}")

            local code
            code=$(echo "${response}" | grep -o '"code":[0-9]*' | cut -d':' -f2 | head -1)

            if [[ "${code}" == "0" ]]; then
                echo "success" > "${temp_dir}/${i}.result"
            else
                echo "fail" > "${temp_dir}/${i}.result"
            fi
        ) &
        pids+=($!)

        if [[ $(( i % 10 )) -eq 0 ]]; then
            wait "${pids[@]}" 2>/dev/null || true
            pids=()
        fi
    done

    [[ ${#pids[@]} -gt 0 ]] && wait "${pids[@]}" 2>/dev/null || true

    success_count=$(grep -c "success" "${temp_dir}"/*.result 2>/dev/null || echo "0")
    fail_count=$(grep -c "fail" "${temp_dir}"/*.result 2>/dev/null || echo "0")
    rm -rf "${temp_dir}"

    echo "[INFO] 并发请求完成: 成功=${success_count}, 失败=${fail_count}"

    echo ">>> 验证库存不超卖..."
    assert_no_oversell "${TEST_GOODS_ID}" "${concurrent_stock}"

    echo ">>> 验证限购不超限..."
    for uid in $(seq 1 "${concurrent_users}"); do
        local user_purchases
        user_purchases=$(mysql_exec "SELECT COALESCE(SUM(num), 0) FROM t_order WHERE user_id = ${uid} AND goods_id = ${TEST_GOODS_ID};")
        if [[ "${user_purchases}" -gt "${concurrent_quota}" ]]; then
            echo "[FAIL] 用户 ${uid} 购买数量 ${user_purchases} 超过限购 ${concurrent_quota}"
            ((ASSERTIONS_FAILED++))
        else
            echo "[PASS] 用户 ${uid} 购买数量 ${user_purchases} <= 限购 ${concurrent_quota}"
            ((ASSERTIONS_PASSED++))
        fi
    done

    echo ">>> 验证订单数量..."
    local order_count
    order_count=$(mysql_exec "SELECT COUNT(*) FROM t_order WHERE goods_id = ${TEST_GOODS_ID};")

    if [[ "${order_count}" -le "${concurrent_stock}" ]]; then
        echo "[PASS] 订单数量 ${order_count} <= 库存 ${concurrent_stock}"
        ((ASSERTIONS_PASSED++))
    else
        echo "[FAIL] 订单数量 ${order_count} > 库存 ${concurrent_stock}"
        ((ASSERTIONS_FAILED++))
    fi

    print_assertion_summary
}

test_quota_limit() {
    echo ""
    echo "========================================"
    echo "测试用例: 限购功能测试"
    echo "========================================"

    reset_assertion_counters

    local quota="${TEST_QUOTA}"
    reset_test_data

    login_and_get_token || return 1

    echo ">>> 执行 ${quota} 次秒杀 (应该全部成功)..."
    local success_count=0
    for i in $(seq 1 "${quota}"); do
        local response
        response=$(curl -sS -X POST "${GATEWAY_URL}/bitstorm/v2/sec_kill" \
            -H "Authorization: Bearer ${TOKEN}" \
            -H "Content-Type: application/json" \
            -d "{\"goodsNum\":\"${TEST_GOODS_NUM}\",\"num\":1}")

        local code
        code=$(echo "${response}" | grep -o '"code":[0-9]*' | cut -d':' -f2 | head -1)

        if [[ "${code}" == "0" ]]; then
            ((success_count++))
            echo "    第 ${i} 次秒杀: 成功"
        else
            echo "    第 ${i} 次秒杀: 失败 (code=${code})"
        fi
    done

    assert_equals "${quota}" "${success_count}" "限购内秒杀成功次数"

    echo ">>> 执行第 ${quota}+1 次秒杀 (应该失败)..."
    local response
    response=$(curl -sS -X POST "${GATEWAY_URL}/bitstorm/v2/sec_kill" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{\"goodsNum\":\"${TEST_GOODS_NUM}\",\"num\":1}")

    local code
    code=$(echo "${response}" | grep -o '"code":[0-9]*' | cut -d':' -f2 | head -1)

    if [[ "${code}" != "0" ]]; then
        echo "[PASS] 超限购秒杀被拒绝 (code=${code})"
        ((ASSERTIONS_PASSED++))
    else
        echo "[FAIL] 超限购秒杀未被拒绝"
        ((ASSERTIONS_FAILED++))
    fi

    assert_quota_not_exceeded "${TEST_USER_ID}" "${TEST_GOODS_ID}" "${quota}"

    print_assertion_summary
}

test_stock_exhausted() {
    echo ""
    echo "========================================"
    echo "测试用例: 库存耗尽测试"
    echo "========================================"

    reset_assertion_counters

    local small_stock="${SMALL_STOCK:-2}"
    local test_requests="${TEST_REQUESTS:-5}"

    echo ">>> 设置小库存: ${small_stock}"
    mysql_exec "UPDATE t_seckill_stock SET stock = ${small_stock} WHERE goods_id = ${TEST_GOODS_ID};"
    redis_exec "${SECKILL_REDIS_DB}" SET "SK:Stock:${TEST_GOODS_ID}" "${small_stock}"

    mysql_exec "
        DELETE FROM t_order WHERE goods_id = ${TEST_GOODS_ID};
        DELETE FROM t_seckill_record WHERE goods_id = ${TEST_GOODS_ID};
        DELETE FROM t_user_quota WHERE goods_id = ${TEST_GOODS_ID};
    "

    redis_exec "${SECKILL_REDIS_DB}" DEL "SK:UserGoodsSecNum:${TEST_USER_ID}:${TEST_GOODS_ID}"
    redis_exec "${SECKILL_REDIS_DB}" DEL "SK:UserSecKilledNum:${TEST_USER_ID}:${TEST_GOODS_ID}"

    login_and_get_token || return 1

    echo ">>> 执行 ${test_requests} 次秒杀 (库存仅 ${small_stock})..."
    local success_count=0
    local fail_count=0

    for i in $(seq 1 "${test_requests}"); do
        local response
        response=$(curl -sS -X POST "${GATEWAY_URL}/bitstorm/v2/sec_kill" \
            -H "Authorization: Bearer ${TOKEN}" \
            -H "Content-Type: application/json" \
            -d "{\"goodsNum\":\"${TEST_GOODS_NUM}\",\"num\":1}")

        local code
        code=$(echo "${response}" | grep -o '"code":[0-9]*' | cut -d':' -f2 | head -1)

        if [[ "${code}" == "0" ]]; then
            ((success_count++))
            echo "    第 ${i} 次秒杀: 成功"
        else
            ((fail_count++))
            echo "    第 ${i} 次秒杀: 失败 (code=${code})"
        fi
    done

    echo ">>> 验证成功次数 <= 库存..."
    if [[ "${success_count}" -le "${small_stock}" ]]; then
        echo "[PASS] 成功秒杀次数 ${success_count} <= 库存 ${small_stock}"
        ((ASSERTIONS_PASSED++))
    else
        echo "[FAIL] 成功秒杀次数 ${success_count} > 库存 ${small_stock}"
        ((ASSERTIONS_FAILED++))
    fi

    assert_no_oversell "${TEST_GOODS_ID}" "${small_stock}"

    print_assertion_summary
}

run_all_tests() {
    echo "========================================"
    echo "运行所有测试用例"
    echo "========================================"

    local total_passed=0
    local total_failed=0

    if test_login; then
        total_passed=$((total_passed + 1))
    else
        total_failed=$((total_failed + 1))
    fi

    if test_v1_flow; then
        total_passed=$((total_passed + 1))
    else
        total_failed=$((total_failed + 1))
    fi

    if test_v2_flow; then
        total_passed=$((total_passed + 1))
    else
        total_failed=$((total_failed + 1))
    fi

    if test_v3_flow; then
        total_passed=$((total_passed + 1))
    else
        total_failed=$((total_failed + 1))
    fi

    if test_quota_limit; then
        total_passed=$((total_passed + 1))
    else
        total_failed=$((total_failed + 1))
    fi

    if test_stock_exhausted; then
        total_passed=$((total_passed + 1))
    else
        total_failed=$((total_failed + 1))
    fi

    echo ""
    echo "========================================"
    echo "测试用例统计汇总"
    echo "========================================"
    echo "通过的测试用例: ${total_passed}"
    echo "失败的测试用例: ${total_failed}"
    echo ""

    if [[ "${total_failed}" -gt 0 ]]; then
        echo "❌ 存在失败的测试用例"
        return 1
    else
        echo "✅ 所有测试用例通过"
        return 0
    fi
}
