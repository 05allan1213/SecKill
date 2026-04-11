#!/bin/bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! command -v wrk >/dev/null 2>&1; then
  echo "wrk is required for benchmark" >&2
  exit 1
fi

echo "========================================="
echo "秒杀系统全面性能压测"
echo "========================================="
echo ""

docker compose -f "${ROOT_DIR}/docker-compose.yml" up -d etcd mysql redis kafka

cleanup() {
  echo ""
  echo "清理服务进程..."
  for pid in ${GATEWAY_PID:-} ${SECKILL_PID:-} ${USER_PID:-}; do
    if [[ -n "${pid}" ]] && kill -0 "${pid}" 2>/dev/null; then
      kill "${pid}" 2>/dev/null || true
      for _ in $(seq 1 10); do
        if ! kill -0 "${pid}" 2>/dev/null; then
          break
        fi
        sleep 1
      done
      if kill -0 "${pid}" 2>/dev/null; then
        kill -9 "${pid}" 2>/dev/null || true
      fi
      wait "${pid}" 2>/dev/null || true
    fi
  done
}
trap cleanup EXIT

echo "启动服务..."
cd "${ROOT_DIR}/user-main"
GOCACHE=/tmp/go-build-cache-user go run ./cmd/user -f "${ROOT_DIR}/user-main/etc/user.yaml" >/tmp/user-bench.log 2>&1 &
USER_PID=$!

cd "${ROOT_DIR}/seckill-main"
GOCACHE=/tmp/go-build-cache-seckill go run ./cmd/sec_kill -f "${ROOT_DIR}/seckill-main/etc/seckill.yaml" >/tmp/seckill-bench.log 2>&1 &
SECKILL_PID=$!

cd "${ROOT_DIR}/gateway-main"
GOCACHE=/tmp/go-build-cache-gateway go run ./cmd/gateway -f "${ROOT_DIR}/gateway-main/etc/gateway.yaml" >/tmp/gateway-bench.log 2>&1 &
GATEWAY_PID=$!

echo "等待服务启动..."
for _ in $(seq 1 60); do
  if curl -sf "http://127.0.0.1:8998/metrics" >/dev/null 2>&1; then
    echo "服务启动成功！"
    break
  fi
  sleep 1
done

echo ""
echo "获取认证 Token..."
TOKEN="$(curl -sf -X POST 'http://127.0.0.1:8998/login' \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"123321"}' | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')"

test -n "${TOKEN}"
echo "Token 获取成功！"

export BENCH_TOKEN="${TOKEN}"

echo ""
echo "重置 Redis 库存..."
docker exec mks-redis redis-cli -a 123456 SET "SK:Stock:abc123" 10000 >/dev/null 2>&1
echo "库存重置完成！"

run_benchmark() {
  local version=$1
  local connections=$2
  local threads=$3
  local duration=$4
  
  echo ""
  echo "----------------------------------------"
  echo "压测配置: v${version} | ${connections} 连接 | ${threads} 线程 | ${duration}s"
  echo "----------------------------------------"
  
  wrk -t${threads} -c${connections} -d${duration}s \
    -s "${ROOT_DIR}/seckill-main/wrkbench/gateway_sec_kill_v${version}.lua" \
    --latency "http://127.0.0.1:8998/bitstorm/v${version}/sec_kill" 2>&1 | tee -a "${ROOT_DIR}/benchmark_results.log"
  
  sleep 2
}

echo ""
echo "========================================="
echo "开始全面性能压测"
echo "========================================="

echo "" > "${ROOT_DIR}/benchmark_results.log"
echo "秒杀系统全面性能压测报告" >> "${ROOT_DIR}/benchmark_results.log"
echo "测试时间: $(date '+%Y-%m-%d %H:%M:%S')" >> "${ROOT_DIR}/benchmark_results.log"
echo "=========================================" >> "${ROOT_DIR}/benchmark_results.log"

for version in 1 2 3; do
  echo ""
  echo "=========================================" >> "${ROOT_DIR}/benchmark_results.log"
  echo "版本 v${version} 压测" >> "${ROOT_DIR}/benchmark_results.log"
  echo "=========================================" >> "${ROOT_DIR}/benchmark_results.log"
  
  run_benchmark ${version} 50 4 10
  run_benchmark ${version} 100 4 10
  run_benchmark ${version} 200 8 10
  run_benchmark ${version} 500 12 10
  run_benchmark ${version} 1000 16 10
  
  echo ""
  echo "重置 Redis 库存..."
  docker exec mks-redis redis-cli -a 123456 SET "SK:Stock:abc123" 10000 >/dev/null 2>&1
done

echo ""
echo "========================================="
echo "压测完成！结果已保存到 benchmark_results.log"
echo "========================================="
