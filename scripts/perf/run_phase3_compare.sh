#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT_DIR="${ROOT_DIR}/tmp/perf"
mkdir -p "${OUT_DIR}"

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
USERNAME="${USERNAME:-admin}"
USER_PASSWORD="${USER_PASSWORD:-123321}"
GOODS_ID="${GOODS_ID:-1}"
GOODS_NUM="${GOODS_NUM:-abc123}"
STOCK="${STOCK:-200}"
USER_LIMIT="${USER_LIMIT:-1}"
REQUESTS="${REQUESTS:-200}"
CONNECTIONS="${CONNECTIONS:-50}"
POLL_TIMES="${POLL_TIMES:-20}"
POLL_INTERVAL="${POLL_INTERVAL:-0.2}"
LIMITER_PROFILE="${LIMITER_PROFILE:-compare}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing command: $1" >&2
    exit 1
  fi
}

require_cmd curl
require_cmd hey

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

login() {
  local response token
  response="$(curl -sS -X POST "${GATEWAY_URL}/login" \
    -H 'Content-Type: application/json' \
    -d "{\"username\":\"${USERNAME}\",\"password\":\"${USER_PASSWORD}\"}")"
  token="$(printf '%s' "${response}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)"
  if [[ -z "${token}" ]]; then
    echo "login failed: ${response}" >&2
    exit 1
  fi
  TOKEN="${token}"
}

load_goods_snapshot() {
  local row
  row="$(mysql_exec "SELECT id, goods_num, goods_name, price, pic_url, seller FROM t_goods WHERE goods_num='${GOODS_NUM}' LIMIT 1;")"
  if [[ -z "${row}" ]]; then
    echo "goods not found: ${GOODS_NUM}" >&2
    exit 1
  fi
  IFS=$'\t' read -r DB_GOODS_ID DB_GOODS_NUM DB_GOODS_NAME DB_GOODS_PRICE DB_GOODS_PIC_URL DB_GOODS_SELLER <<< "${row}"
}

reset_mysql_state() {
  mysql_exec "
    DELETE FROM t_order;
    DELETE FROM t_seckill_record;
    DELETE FROM t_user_quota;
    INSERT INTO t_quota(goods_id, num) VALUES (${GOODS_ID}, ${USER_LIMIT})
      ON DUPLICATE KEY UPDATE num = VALUES(num);
    UPDATE t_seckill_stock SET stock = ${STOCK} WHERE goods_id = ${GOODS_ID};
  "
}

reset_redis_state() {
  redis_exec "${SECKILL_REDIS_DB}" FLUSHDB
  redis_exec "${GATEWAY_REDIS_DB}" FLUSHDB
}

preheat_hot_keys() {
  local goods_json
  goods_json="$(printf '{"ID":%s,"GoodsNum":"%s","GoodsName":"%s","Price":%s,"PicUrl":"%s","Seller":%s}' \
    "${DB_GOODS_ID}" "${DB_GOODS_NUM}" "${DB_GOODS_NAME}" "${DB_GOODS_PRICE}" "${DB_GOODS_PIC_URL}" "${DB_GOODS_SELLER}")"
  redis_exec "${SECKILL_REDIS_DB}" SET "goodsInfo:${GOODS_NUM}" "${goods_json}" EX 600
  redis_exec "${SECKILL_REDIS_DB}" SET "SK:Stock:${GOODS_ID}" "${STOCK}" EX 600
  redis_exec "${SECKILL_REDIS_DB}" SET "SK:Limit${GOODS_ID}" "${USER_LIMIT}" EX 600
}

prepare_case() {
  reset_mysql_state
  reset_redis_state
  preheat_hot_keys
}

parse_hey_metric() {
  local file="$1"
  local key="$2"
  case "${key}" in
    qps)
      awk '/Requests\/sec:/ {print $2}' "${file}"
      ;;
    p95)
      awk '/  95% in / {print $3}' "${file}"
      ;;
  esac
}

run_hey_case() {
  local version="$1"
  local endpoint="$2"
  local output="${OUT_DIR}/${version}.hey.txt"
  hey -n "${REQUESTS}" -c "${CONNECTIONS}" \
    -m POST \
    -H "Authorization: Bearer ${TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{\"goodsNum\":\"${GOODS_NUM}\",\"num\":1}" \
    "${GATEWAY_URL}${endpoint}" > "${output}"
  printf '%s\t%s\t%s\n' "${version}" "$(parse_hey_metric "${output}" qps)" "$(parse_hey_metric "${output}" p95)"
}

run_outcome_case() {
  local version="$1"
  local endpoint="$2"
  local outfile="${OUT_DIR}/${version}.responses.txt"
  : > "${outfile}"

  seq "${REQUESTS}" | xargs -P "${CONNECTIONS}" -I{} bash -lc '
    curl -sS -o - -w "\t%{http_code}\n" \
      -X POST "'"${GATEWAY_URL}${endpoint}"'" \
      -H "Authorization: Bearer '"${TOKEN}"'" \
      -H "Content-Type: application/json" \
      -d "{\"goodsNum\":\"'"${GOODS_NUM}"'\",\"num\":1}"
  ' >> "${outfile}"
}

count_success() {
  awk -F'\t' '$2=="200" && $1 ~ /"code":0/ {c++} END {print c+0}' "$1"
}

count_limited() {
  awk -F'\t' '$2=="429" || $1 ~ /"code":42901/ {c++} END {print c+0}' "$1"
}

count_business_failed() {
  awk -F'\t' '$2=="200" && $1 !~ /"code":0/ {c++} END {print c+0}' "$1"
}

poll_v3_results() {
  local outfile="$1"
  local secnums_file="${OUT_DIR}/v3.secnums.txt"
  local status_file="${OUT_DIR}/v3.status.txt"
  : > "${status_file}"

  grep -o '"secNum":"[^"]*"' "${outfile}" | cut -d'"' -f4 | sort -u > "${secnums_file}" || true
  while IFS= read -r secnum; do
    [[ -z "${secnum}" ]] && continue
    local response status tries
    tries=0
    while (( tries < POLL_TIMES )); do
      response="$(curl -sS "${GATEWAY_URL}/bitstorm/v3/get_sec_kill_info?sec_num=${secnum}" \
        -H "Authorization: Bearer ${TOKEN}")"
      status="$(printf '%s' "${response}" | grep -o '"status":[0-9]*' | cut -d: -f2)"
      if [[ "${status}" != "1" && -n "${status}" ]]; then
        printf '%s\t%s\n' "${secnum}" "${response}" >> "${status_file}"
        break
      fi
      tries=$((tries + 1))
      sleep "${POLL_INTERVAL}"
    done
    if (( tries == POLL_TIMES )); then
      printf '%s\t{"status":1}\n' "${secnum}" >> "${status_file}"
    fi
  done < "${secnums_file}"

  local completed failed pending
  completed="$(grep -c '"status":2' "${status_file}" || true)"
  failed="$(grep -c '"status":6' "${status_file}" || true)"
  pending="$(grep -c '"status":1' "${status_file}" || true)"
  printf '%s\t%s\t%s\n' "${completed:-0}" "${failed:-0}" "${pending:-0}"
}

print_summary() {
  local version="$1"
  local qps="$2"
  local p95="$3"
  local success="$4"
  local limited="$5"
  local failed="$6"
  local v3_completed="${7:-0}"
  local v3_failed="${8:-0}"
  local v3_pending="${9:-0}"

  printf '%-4s | qps=%-10s | p95=%-10s | success=%-4s | limited=%-4s | business_failed=%-4s' \
    "${version}" "${qps}" "${p95}" "${success}" "${limited}" "${failed}"
  if [[ "${version}" == "v3" ]]; then
    printf ' | v3_completed=%-4s | v3_failed=%-4s | v3_pending=%-4s' \
      "${v3_completed}" "${v3_failed}" "${v3_pending}"
  fi
  printf '\n'
}

declare -A CASE_QPS
declare -A CASE_P95
declare -A CASE_SUCCESS
declare -A CASE_LIMITED
declare -A CASE_FAILED
declare -A CASE_V3_COMPLETED
declare -A CASE_V3_FAILED
declare -A CASE_V3_PENDING

run_case() {
  local version="$1"
  local endpoint="$2"
  local metrics qps p95 responses success limited failed v3_stats v3_completed v3_failed v3_pending

  prepare_case
  metrics="$(run_hey_case "${version}" "${endpoint}")"
  qps="$(printf '%s' "${metrics}" | cut -f2)"
  p95="$(printf '%s' "${metrics}" | cut -f3)"

  prepare_case
  run_outcome_case "${version}" "${endpoint}"
  responses="${OUT_DIR}/${version}.responses.txt"
  success="$(count_success "${responses}")"
  limited="$(count_limited "${responses}")"
  failed="$(count_business_failed "${responses}")"

  CASE_QPS["${version}"]="${qps}"
  CASE_P95["${version}"]="${p95}"
  CASE_SUCCESS["${version}"]="${success}"
  CASE_LIMITED["${version}"]="${limited}"
  CASE_FAILED["${version}"]="${failed}"

  if [[ "${version}" == "v3" ]]; then
    v3_stats="$(poll_v3_results "${responses}")"
    v3_completed="$(printf '%s' "${v3_stats}" | cut -f1)"
    v3_failed="$(printf '%s' "${v3_stats}" | cut -f2)"
    v3_pending="$(printf '%s' "${v3_stats}" | cut -f3)"
    CASE_V3_COMPLETED["${version}"]="${v3_completed}"
    CASE_V3_FAILED["${version}"]="${v3_failed}"
    CASE_V3_PENDING["${version}"]="${v3_pending}"
    print_summary "${version}" "${qps}" "${p95}" "${success}" "${limited}" "${failed}" "${v3_completed}" "${v3_failed}" "${v3_pending}"
    return
  fi

  print_summary "${version}" "${qps}" "${p95}" "${success}" "${limited}" "${failed}"
}

generate_json_report() {
  local report_file="${OUT_DIR}/report.json"
  local timestamp
  timestamp="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

  cat > "${report_file}" << EOF
{
  "timestamp": "${timestamp}",
  "config": {
    "gateway_url": "${GATEWAY_URL}",
    "requests": ${REQUESTS},
    "connections": ${CONNECTIONS},
    "stock": ${STOCK},
    "user_limit": ${USER_LIMIT}
  },
  "results": {
    "v1": {
      "qps": ${CASE_QPS["v1"]:-0},
      "p95": "${CASE_P95["v1"]:-0}",
      "success": ${CASE_SUCCESS["v1"]:-0},
      "limited": ${CASE_LIMITED["v1"]:-0},
      "business_failed": ${CASE_FAILED["v1"]:-0}
    },
    "v2": {
      "qps": ${CASE_QPS["v2"]:-0},
      "p95": "${CASE_P95["v2"]:-0}",
      "success": ${CASE_SUCCESS["v2"]:-0},
      "limited": ${CASE_LIMITED["v2"]:-0},
      "business_failed": ${CASE_FAILED["v2"]:-0}
    },
    "v3": {
      "qps": ${CASE_QPS["v3"]:-0},
      "p95": "${CASE_P95["v3"]:-0}",
      "success": ${CASE_SUCCESS["v3"]:-0},
      "limited": ${CASE_LIMITED["v3"]:-0},
      "business_failed": ${CASE_FAILED["v3"]:-0},
      "v3_completed": ${CASE_V3_COMPLETED["v3"]:-0},
      "v3_failed": ${CASE_V3_FAILED["v3"]:-0},
      "v3_pending": ${CASE_V3_PENDING["v3"]:-0}
    }
  }
}
EOF
  echo "JSON report saved to ${report_file}"
}

print_comparison_table() {
  echo ""
  echo "========================================"
  echo "Performance Comparison Table"
  echo "========================================"
  printf "%-8s | %-12s | %-12s | %-8s | %-8s | %-8s\n" "Version" "QPS" "P95" "Success" "Limited" "Failed"
  echo "-------------------------------------------------------------------------"
  
  for v in v1 v2 v3; do
    local qps="${CASE_QPS["${v}"]:-0}"
    local p95="${CASE_P95["${v}"]:-0}"
    local success="${CASE_SUCCESS["${v}"]:-0}"
    local limited="${CASE_LIMITED["${v}"]:-0}"
    local failed="${CASE_FAILED["${v}"]:-0}"
    printf "%-8s | %-12s | %-12s | %-8s | %-8s | %-8s\n" "${v}" "${qps}" "${p95}" "${success}" "${limited}" "${failed}"
  done
  
  echo ""
  echo "V3 Async Results:"
  printf "  Completed: %s, Failed: %s, Pending: %s\n" \
    "${CASE_V3_COMPLETED["v3"]:-0}" "${CASE_V3_FAILED["v3"]:-0}" "${CASE_V3_PENDING["v3"]:-0}"
}

run_assertions() {
  local assertions_passed=0
  local assertions_failed=0

  echo ""
  echo "========================================"
  echo "Assertions"
  echo "========================================"

  local v1_qps v2_qps
  v1_qps="${CASE_QPS["v1"]:-0}"
  v2_qps="${CASE_QPS["v2"]:-0}"

  if awk "BEGIN {exit !(${v2_qps} >= ${v1_qps} * 0.9)}"; then
    echo "✅ PASS: V2 QPS >= V1 QPS (10% tolerance)"
    assertions_passed=$((assertions_passed + 1))
  else
    echo "❌ FAIL: V2 QPS < V1 QPS (V1=${v1_qps}, V2=${v2_qps})"
    assertions_failed=$((assertions_failed + 1))
  fi

  for v in v1 v2 v3; do
    local success="${CASE_SUCCESS["${v}"]:-0}"
    if (( success <= STOCK )); then
      echo "✅ PASS: ${v} orders not exceed stock (success=${success}, stock=${STOCK})"
      assertions_passed=$((assertions_passed + 1))
    else
      echo "❌ FAIL: ${v} orders exceed stock (success=${success}, stock=${STOCK})"
      assertions_failed=$((assertions_failed + 1))
    fi
  done

  for v in v1 v2 v3; do
    local limited="${CASE_LIMITED["${v}"]:-0}"
    if (( limited > 0 )); then
      echo "✅ PASS: ${v} quota enforced (limited=${limited})"
      assertions_passed=$((assertions_passed + 1))
    else
      echo "⚠️  WARN: ${v} no quota limit triggered"
    fi
  done

  echo ""
  echo "Assertions Summary: ${assertions_passed} passed, ${assertions_failed} failed"

  if (( assertions_failed > 0 )); then
    return 1
  fi
  return 0
}

echo "Phase3 performance compare"
echo "Gateway: ${GATEWAY_URL}"
echo "LimiterProfile (expected in gateway.yaml): ${LIMITER_PROFILE}"
echo "Requests=${REQUESTS}, Connections=${CONNECTIONS}, Stock=${STOCK}, UserLimit=${USER_LIMIT}"

login
load_goods_snapshot

echo "version | metrics"
run_case "v1" "/bitstorm/v1/sec_kill"
run_case "v2" "/bitstorm/v2/sec_kill"
run_case "v3" "/bitstorm/v3/sec_kill"

echo "raw outputs saved in ${OUT_DIR}"

print_comparison_table

generate_json_report

run_assertions
