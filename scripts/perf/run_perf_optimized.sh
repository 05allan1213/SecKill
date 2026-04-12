#!/usr/bin/env bash

# 优化版压测脚本
# 主要改进：
# 1. 多用户并发压测，真实模拟秒杀场景
# 2. 使用 hey 工具直接压测，避免 xargs+curl 开销
# 3. 添加预热阶段，确保连接池预热
# 4. 详细的统计报告和性能分析
# 5. 支持分布式压测准备

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT_DIR="${ROOT_DIR}/tmp/perf_optimized"
mkdir -p "${OUT_DIR}"

# ==================== 可配置参数 ====================
GATEWAY_URL="${GATEWAY_URL:-http://127.0.0.1:8998}"
MYSQL_HOST="${MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${MYSQL_PORT:-3307}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-123456}"
MYSQL_DB="${MYSQL_DB:-bitstorm}"
MYSQL_CONTAINER="${MYSQL_CONTAINER:-mks-mysql}"
REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-123456}"
REDIS_CONTAINER="${REDIS_CONTAINER:-mks-redis}"
SECKILL_REDIS_DB="${SECKILL_REDIS_DB:-0}"
GATEWAY_REDIS_DB="${GATEWAY_REDIS_DB:-8}"
USE_DOCKER="${USE_DOCKER:-true}"

# 压测参数
GOODS_ID="${GOODS_ID:-1}"
GOODS_NUM="${GOODS_NUM:-abc123}"
STOCK="${STOCK:-1000}"              # 库存数量
USER_LIMIT="${USER_LIMIT:-5}"       # 每用户限购
NUM_USERS="${NUM_USERS:-100}"       # 压测用户数
REQUESTS="${REQUESTS:-5000}"        # 总请求数
CONNECTIONS="${CONNECTIONS:-200}"   # 并发连接数
DURATION="${DURATION:-30s}"         # 压测时长（与REQUESTS二选一）
WARMUP_REQUESTS="${WARMUP_REQUESTS:-100}"  # 预热请求数

# 输出颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# ==================== 工具函数 ====================

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo -e "${RED}缺少命令: $1${NC}" >&2
    exit 1
  fi
}

print_header() {
  echo -e "\n${BLUE}========================================${NC}"
  echo -e "${BLUE}  $1${NC}"
  echo -e "${BLUE}========================================${NC}\n"
}

print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }
print_info() { echo -e "${CYAN}ℹ️  $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }

mysql_exec() {
  local query="$1"
  if [[ "${USE_DOCKER}" == "true" ]]; then
    docker exec "${MYSQL_CONTAINER}" mysql -uroot -p"${MYSQL_PASSWORD}" \
      --batch --raw --skip-column-names \
      "${MYSQL_DB}" -e "${query}" 2>/dev/null
  else
    MYSQL_PWD="${MYSQL_PASSWORD}" mysql \
      -h "${MYSQL_HOST}" \
      -P "${MYSQL_PORT}" \
      -u "${MYSQL_USER}" \
      --batch --raw --skip-column-names \
      "${MYSQL_DB}" \
      -e "${query}"
  fi
}

redis_exec() {
  local db="$1"
  shift
  if [[ "${USE_DOCKER}" == "true" ]]; then
    docker exec "${REDIS_CONTAINER}" redis-cli -a "${REDIS_PASSWORD}" -n "${db}" "$@" >/dev/null 2>&1
  else
    redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" -a "${REDIS_PASSWORD}" -n "${db}" "$@" >/dev/null 2>&1
  fi
}

# ==================== 用户管理 ====================

# 创建压测用户（如果不存在）
create_test_users() {
  print_header "准备压测用户"
  
  local existing_users
  existing_users=$(mysql_exec "SELECT COUNT(*) FROM t_user_info WHERE user_name LIKE 'perf_user%';")
  
  if [[ "${existing_users}" -ge "${NUM_USERS}" ]]; then
    print_info "已存在 ${existing_users} 个压测用户，跳过创建"
    return 0
  fi
  
  print_info "创建 ${NUM_USERS} 个压测用户..."
  
  local values=""
  for i in $(seq 1 "${NUM_USERS}"); do
    local uid=$((i + 10000))
    local mobile="188${uid}"
    local id_card="510681${uid}0000"
    if [[ -n "${values}" ]]; then
      values="${values},"
    fi
    values="${values}(${uid}, 'perf_user${i}', '123321', 25, 1, '${mobile}', '${id_card}')"
  done
  
  mysql_exec "INSERT IGNORE INTO t_user_info (id, user_name, pwd, age, sex, mobile, id_card) VALUES ${values};"
  print_success "已创建 ${NUM_USERS} 个压测用户"
}

# 获取用户Token列表
get_user_tokens() {
  print_header "获取用户Token"
  
  local tokens_file="${OUT_DIR}/tokens.txt"
  : > "${tokens_file}"
  
  local success=0
  local failed=0
  
  for i in $(seq 1 "${NUM_USERS}"); do
    local username="perf_user${i}"
    local response
    response=$(curl -sS -X POST "${GATEWAY_URL}/login" \
      -H 'Content-Type: application/json' \
      -d "{\"username\":\"${username}\",\"password\":\"123321\"}" 2>/dev/null || echo "")
    
    local token
    token=$(printf '%s' "${response}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    
    if [[ -n "${token}" ]]; then
      echo "${token}" >> "${tokens_file}"
      ((success++))
    else
      ((failed++))
    fi
    
    # 显示进度
    if (( i % 20 == 0 )); then
      print_info "已处理 ${i}/${NUM_USERS} 个用户..."
    fi
  done
  
  print_success "成功获取 ${success} 个Token"
  if [[ "${failed}" -gt 0 ]]; then
    print_warning "获取失败 ${failed} 个"
  fi
  
  TOKENS_FILE="${tokens_file}"
}

# ==================== 数据准备 ====================

# 重置测试数据
reset_test_data() {
  print_header "重置测试数据"
  
  print_info "清理历史数据..."
  mysql_exec "
    DELETE FROM t_order WHERE goods_id = ${GOODS_ID};
    DELETE FROM t_seckill_record WHERE goods_id = ${GOODS_ID};
    DELETE FROM t_seckill_async_result WHERE goods_id = ${GOODS_ID};
    DELETE FROM t_user_quota WHERE goods_id = ${GOODS_ID};
  "
  
  print_info "设置库存=${STOCK}, 限购=${USER_LIMIT}..."
  mysql_exec "
    INSERT INTO t_quota(goods_id, num) VALUES (${GOODS_ID}, ${USER_LIMIT})
      ON DUPLICATE KEY UPDATE num = VALUES(num);
    UPDATE t_seckill_stock SET stock = ${STOCK} WHERE goods_id = ${GOODS_ID};
  "
  
  print_info "重置Redis热点数据..."
  redis_exec "${SECKILL_REDIS_DB}" FLUSHDB
  redis_exec "${GATEWAY_REDIS_DB}" FLUSHDB
  
  # 预热Redis热点Key
  local goods_json
  goods_json='{"ID":1,"GoodsNum":"abc123","GoodsName":"redhat","Price":18,"PicUrl":"http://","Seller":135}'
  
  redis_exec "${SECKILL_REDIS_DB}" SET "goodsInfo:${GOODS_NUM}" "${goods_json}" EX 3600
  redis_exec "${SECKILL_REDIS_DB}" SET "SK:Stock:${GOODS_ID}" "${STOCK}" EX 3600
  redis_exec "${SECKILL_REDIS_DB}" SET "SK:Limit${GOODS_ID}" "${USER_LIMIT}" EX 3600
  
  # 清理用户购买记录
  for i in $(seq 1 "${NUM_USERS}"); do
    local uid=$((i + 10000))
    redis_exec "${SECKILL_REDIS_DB}" DEL "SK:UserGoodsSecNum:${uid}:${GOODS_ID}"
    redis_exec "${SECKILL_REDIS_DB}" DEL "SK:UserSecKilledNum:${uid}:${GOODS_ID}"
  done
  
  print_success "测试数据已重置"
}

# ==================== 压测执行 ====================

# 预热阶段
run_warmup() {
  print_header "预热阶段"
  
  print_info "发送 ${WARMUP_REQUESTS} 个预热请求..."
  
  local token
  token=$(head -1 "${TOKENS_FILE}")
  
  hey -n "${WARMUP_REQUESTS}" -c 10 \
    -m POST \
    -H "Authorization: Bearer ${token}" \
    -H "Content-Type: application/json" \
    -d "{\"goodsNum\":\"${GOODS_NUM}\",\"num\":1}" \
    "${GATEWAY_URL}/bitstorm/v2/sec_kill" > /dev/null 2>&1
  
  print_success "预热完成"
  sleep 2
}

# 使用多Token压测（轮询方式）
run_multi_user_benchmark() {
  local version="$1"
  local endpoint="${GATEWAY_URL}/bitstorm/${version}/sec_kill"
  local output="${OUT_DIR}/${version}_$(date +%Y%m%d_%H%M%S).hey"
  
  print_info "压测 ${version} 接口: ${endpoint}"
  print_info "参数: 请求=${REQUESTS}, 并发=${CONNECTIONS}, 用户数=${NUM_USERS}"
  
  # 生成请求体文件（hey支持从文件读取body，但Token需要特殊处理）
  # 这里使用自定义方式：生成多个请求脚本并发执行
  
  if [[ "${NUM_USERS}" -le 1 ]]; then
    # 单用户模式：使用hey直接压测
    local token
    token=$(head -1 "${TOKENS_FILE}")
    
    hey -n "${REQUESTS}" -c "${CONNECTIONS}" \
      -m POST \
      -H "Authorization: Bearer ${token}" \
      -H "Content-Type: application/json" \
      -d "{\"goodsNum\":\"${GOODS_NUM}\",\"num\":1}" \
      "${endpoint}" > "${output}" 2>&1
  else
    # 多用户模式：使用hey的-workers参数配合生成的headers文件
    # 生成headers文件，每个请求轮询选择Token
    local headers_file="${OUT_DIR}/headers.txt"
    : > "${headers_file}"
    
    local total_tokens
    total_tokens=$(wc -l < "${TOKENS_FILE}")
    
    for i in $(seq 1 "${REQUESTS}"); do
      local line_num=$(((i - 1) % total_tokens + 1))
      local token
      token=$(sed -n "${line_num}p" "${TOKENS_FILE}")
      echo "Authorization: Bearer ${token}" >> "${headers_file}"
    done
    
    # 使用 hey 的 -H 文件模式（如果支持）
    # hey 不支持动态header，改用替代方案
    print_warning "hey不支持动态Header，使用替代方案..."
    
    # 替代方案：使用多个hey进程分担请求
    local num_workers="${CONNECTIONS}"
    local req_per_worker=$((REQUESTS / num_workers))
    local remainder=$((REQUESTS % num_workers))
    
    local temp_dir="${OUT_DIR}/workers"
    mkdir -p "${temp_dir}"
    
    for w in $(seq 1 "${num_workers}"); do
      local worker_output="${temp_dir}/worker_${w}.log" &
      local token_idx=$(((w - 1) % total_tokens + 1))
      local token
      token=$(sed -n "${token_idx}p" "${TOKENS_FILE}")
      local req_count="${req_per_worker}"
      if [[ "${w}" -le "${remainder}" ]]; then
        ((req_count++))
      fi
      
      if [[ "${req_count}" -gt 0 ]]; then
        hey -n "${req_count}" -c 1 \
          -m POST \
          -H "Authorization: Bearer ${token}" \
          -H "Content-Type: application/json" \
          -d "{\"goodsNum\":\"${GOODS_NUM}\",\"num\":1}" \
          "${endpoint}" > "${worker_output}" 2>&1 &
      fi
    done
    
    # 等待所有worker完成
    wait
    
    # 汇总结果
    local total_requests=0
    local total_duration=0
    local success_count=0
    
    for w in $(seq 1 "${num_workers}"); do
      local worker_output="${temp_dir}/worker_${w}.log"
      if [[ -f "${worker_output}" ]]; then
        local worker_req
        worker_req=$(grep "Total requests:" "${worker_output}" 2>/dev/null | awk '{print $3}' || echo "0")
        total_requests=$((total_requests + worker_req))
      fi
    done
    
    # 生成汇总报告
    {
      echo "=== Multi-User Benchmark Summary ==="
      echo "Version: ${version}"
      echo "Total Users: ${NUM_USERS}"
      echo "Total Requests: ${REQUESTS}"
      echo "Concurrency: ${CONNECTIONS}"
      echo "Stock: ${STOCK}"
      echo "User Limit: ${USER_LIMIT}"
      echo ""
      echo "=== Worker Outputs ==="
      cat "${temp_dir}"/*.log 2>/dev/null || true
    } > "${output}"
  fi
  
  echo "${output}"
}

# 简化版压测：使用单用户但高并发（用于压测服务端性能）
run_single_user_benchmark() {
  local version="$1"
  local endpoint="${GATEWAY_URL}/bitstorm/${version}/sec_kill"
  local output="${OUT_DIR}/${version}_$(date +%Y%m%d_%H%M%S).hey"
  
  print_info "压测 ${version} 接口（单用户高并发模式）"
  
  local token
  token=$(head -1 "${TOKENS_FILE}")
  
  # 使用 -z 参数进行时长压测
  if [[ "${DURATION}" != "0" ]]; then
    hey -z "${DURATION}" -c "${CONNECTIONS}" \
      -m POST \
      -H "Authorization: Bearer ${token}" \
      -H "Content-Type: application/json" \
      -d "{\"goodsNum\":\"${GOODS_NUM}\",\"num\":1}" \
      "${endpoint}" > "${output}" 2>&1
  else
    hey -n "${REQUESTS}" -c "${CONNECTIONS}" \
      -m POST \
      -H "Authorization: Bearer ${token}" \
      -H "Content-Type: application/json" \
      -d "{\"goodsNum\":\"${GOODS_NUM}\",\"num\":1}" \
      "${endpoint}" > "${output}" 2>&1
  fi
  
  echo "${output}"
}

# 解析hey输出
parse_hey_output() {
  local file="$1"
  
  if [[ ! -f "${file}" ]]; then
    echo "QPS=0 P50=0 P95=0 P99=0 SUCCESS=0 FAILED=0"
    return
  fi
  
  local qps p50 p95 p99
  qps=$(grep "Requests/sec:" "${file}" | awk '{print $2}' || echo "0")
  p50=$(grep "  50% in " "${file}" | awk '{print $3}' || echo "0")
  p95=$(grep "  95% in " "${file}" | awk '{print $3}' || echo "0")
  p99=$(grep "  99% in " "${file}" | awk '{print $3}' || echo "0")
  
  echo "QPS=${qps} P50=${p50} P95=${p95} P99=${p99}"
}

# 统计业务结果
collect_business_stats() {
  print_header "业务统计"
  
  local order_count
  order_count=$(mysql_exec "SELECT COUNT(*) FROM t_order WHERE goods_id = ${GOODS_ID};")
  
  local remaining_stock
  remaining_stock=$(mysql_exec "SELECT stock FROM t_seckill_stock WHERE goods_id = ${GOODS_ID};")
  
  local sold_count=$((STOCK - remaining_stock))
  
  local redis_stock
  redis_stock=$(redis_exec "${SECKILL_REDIS_DB}" GET "SK:Stock:${GOODS_ID}" 2>/dev/null || echo "N/A")
  
  echo -e "${CYAN}订单数量:${NC}     ${order_count}"
  echo -e "${CYAN}剩余库存:${NC}     ${remaining_stock} (MySQL)"
  echo -e "${CYAN}已售数量:${NC}     ${sold_count}"
  echo -e "${CYAN}Redis库存:${NC}    ${redis_stock}"
  
  # 检查超卖
  if [[ "${order_count}" -gt "${STOCK}" ]]; then
    print_error "检测到超卖！订单数 ${order_count} > 库存 ${STOCK}"
  else
    print_success "无超卖：订单数 ${order_count} <= 库存 ${STOCK}"
  fi
  
  # 检查Redis和MySQL一致性
  if [[ "${redis_stock}" != "N/A" && "${redis_stock}" != "${remaining_stock}" ]]; then
    print_warning "Redis与MySQL库存不一致: Redis=${redis_stock}, MySQL=${remaining_stock}"
  fi
}

# ==================== 主流程 ====================

main() {
  print_header "SecKill 优化版压测脚本"
  
  # 检查依赖
  require_cmd curl
  require_cmd hey
  require_cmd docker
  
  echo -e "${CYAN}压测配置:${NC}"
  echo "  Gateway:     ${GATEWAY_URL}"
  echo "  库存:        ${STOCK}"
  echo "  用户限购:    ${USER_LIMIT}"
  echo "  压测用户数:  ${NUM_USERS}"
  echo "  总请求数:    ${REQUESTS}"
  echo "  并发数:      ${CONNECTIONS}"
  echo "  压测时长:    ${DURATION}"
  echo ""
  
  # 1. 准备用户
  create_test_users
  
  # 2. 获取Token
  get_user_tokens
  
  # 3. 重置测试数据
  reset_test_data
  
  # 4. 预热
  run_warmup
  
  # 5. 执行压测
  print_header "执行压测"
  
  local results=()
  
  for version in v1 v2 v3; do
    print_info "压测 ${version}..."
    reset_test_data
    local output
    output=$(run_single_user_benchmark "${version}")
    results+=("${version}:${output}")
    
    # 收集业务统计
    collect_business_stats
    echo ""
  done
  
  # 6. 汇总报告
  print_header "压测报告汇总"
  
  printf "%-8s | %-12s | %-10s | %-10s | %-10s\n" "Version" "QPS" "P50" "P95" "P99"
  echo "-------------------------------------------------------------------------"
  
  for result in "${results[@]}"; do
    local version="${result%%:*}"
    local file="${result#*:}"
    local stats
    stats=$(parse_hey_output "${file}")
    
    local qps p50 p95 p99
    eval "${stats}"
    
    printf "%-8s | %-12s | %-10s | %-10s | %-10s\n" "${version}" "${QPS}" "${P50}" "${P95}" "${P99}"
  done
  
  print_info "详细报告保存在: ${OUT_DIR}"
}

# 显示帮助
show_usage() {
  echo ""
  echo "Usage: $0 [options]"
  echo ""
  echo "Environment Variables:"
  echo "  GATEWAY_URL    Gateway地址 (default: http://127.0.0.1:8998)"
  echo "  STOCK          库存数量 (default: 1000)"
  echo "  USER_LIMIT     每用户限购 (default: 5)"
  echo "  NUM_USERS      压测用户数 (default: 100)"
  echo "  REQUESTS       总请求数 (default: 5000)"
  echo "  CONNECTIONS    并发数 (default: 200)"
  echo "  DURATION       压测时长 (default: 30s)"
  echo ""
  echo "Examples:"
  echo "  # 默认压测"
  echo "  $0"
  echo ""
  echo "  # 高并发压测"
  echo "  CONNECTIONS=500 REQUESTS=10000 $0"
  echo ""
  echo "  # 多用户压测"
  echo "  NUM_USERS=1000 STOCK=5000 $0"
  echo ""
}

# 解析参数
case "${1:-}" in
  help|--help|-h)
    show_usage
    exit 0
    ;;
  *)
    main
    ;;
esac
