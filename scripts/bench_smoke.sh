#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! command -v wrk >/dev/null 2>&1; then
  echo "wrk is required for bench-smoke" >&2
  exit 1
fi

docker compose -f "${ROOT_DIR}/docker-compose.yml" up -d etcd mysql redis kafka

cleanup() {
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

cd "${ROOT_DIR}/user-main"
GOCACHE=/tmp/go-build-cache-user go run ./cmd/user -f "${ROOT_DIR}/user-main/etc/user.yaml" >/tmp/user-bench.log 2>&1 &
USER_PID=$!

cd "${ROOT_DIR}/seckill-main"
GOCACHE=/tmp/go-build-cache-seckill go run ./cmd/sec_kill -f "${ROOT_DIR}/seckill-main/etc/seckill.yaml" >/tmp/seckill-bench.log 2>&1 &
SECKILL_PID=$!

cd "${ROOT_DIR}/gateway-main"
GOCACHE=/tmp/go-build-cache-gateway go run ./cmd/gateway -f "${ROOT_DIR}/gateway-main/etc/gateway.yaml" >/tmp/gateway-bench.log 2>&1 &
GATEWAY_PID=$!

for _ in $(seq 1 60); do
  if curl -sf "http://127.0.0.1:8998/metrics" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

TOKEN="$(curl -sf -X POST 'http://127.0.0.1:8998/login' \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"123321"}' | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')"

test -n "${TOKEN}"

export BENCH_TOKEN="${TOKEN}"
wrk -t4 -c50 -d10s -s "${ROOT_DIR}/seckill-main/wrkbench/gateway_sec_kill.lua" --latency \
  "http://127.0.0.1:8998/bitstorm/v1/sec_kill"
