#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-local}"
TOKEN=""

case "${MODE}" in
  local)
    USER_CFG="${ROOT_DIR}/user-main/etc/user.yaml"
    SECKILL_CFG="${ROOT_DIR}/seckill-main/etc/seckill.yaml"
    GATEWAY_CFG="${ROOT_DIR}/gateway-main/etc/gateway.yaml"
    ;;
  etcd)
    USER_CFG="${ROOT_DIR}/configs/etcd/user.yaml"
    SECKILL_CFG="${ROOT_DIR}/configs/etcd/seckill.yaml"
    GATEWAY_CFG="${ROOT_DIR}/configs/etcd/gateway.yaml"
    ;;
  *)
    echo "unsupported mode: ${MODE}" >&2
    exit 1
    ;;
esac

docker compose -f "${ROOT_DIR}/docker-compose.yml" up -d etcd mysql redis kafka

if [[ "${MODE}" == "etcd" ]]; then
  "${ROOT_DIR}/scripts/sync_runtime_configs.sh"
fi

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
GOCACHE=/tmp/go-build-cache-user go run ./cmd/user -f "${USER_CFG}" >/tmp/user-smoke.log 2>&1 &
USER_PID=$!

cd "${ROOT_DIR}/seckill-main"
GOCACHE=/tmp/go-build-cache-seckill go run ./cmd/sec_kill -f "${SECKILL_CFG}" >/tmp/seckill-smoke.log 2>&1 &
SECKILL_PID=$!

cd "${ROOT_DIR}/gateway-main"
GOCACHE=/tmp/go-build-cache-gateway go run ./cmd/gateway -f "${GATEWAY_CFG}" >/tmp/gateway-smoke.log 2>&1 &
GATEWAY_PID=$!

for _ in $(seq 1 60); do
  if curl -sf "http://127.0.0.1:8998/metrics" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

LOGIN_RESPONSE="$(curl -sf -X POST 'http://127.0.0.1:8998/login' \
  -H 'Content-Type: application/json' \
  -H 'Trace-ID: smoke-login-001' \
  -d '{"username":"admin","password":"123321"}')"
TOKEN="$(printf '%s' "${LOGIN_RESPONSE}" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')"

test -n "${TOKEN}"

curl -sf 'http://127.0.0.1:8998/bitstorm/get_user_info_by_name?user_name=admin' \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Trace-ID: smoke-user-001' >/tmp/gateway-user.json

curl -sf -X POST 'http://127.0.0.1:8998/bitstorm/v1/sec_kill' \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -H 'Trace-ID: smoke-seckill-001' \
  -d '{"goodsNum":"abc123","num":1}' >/tmp/gateway-seckill.json

curl -sf 'http://127.0.0.1:9101/metrics' | rg 'rpc_server_requests' >/tmp/seckill-metrics.txt
curl -sf 'http://127.0.0.1:9102/metrics' | rg 'rpc_server_requests' >/tmp/user-metrics.txt
curl -sf 'http://127.0.0.1:9103/metrics' | rg 'http_server_requests' >/tmp/gateway-metrics.txt

test -s "${ROOT_DIR}/seckill-main/logs/trace.json"
test -s "${ROOT_DIR}/user-main/logs/trace.json"
test -s "${ROOT_DIR}/gateway-main/logs/trace.json"

sleep 1

rg 'Trace-ID' "${ROOT_DIR}/seckill-main/logs/access.log" >/tmp/seckill-access.txt
rg 'Trace-ID' "${ROOT_DIR}/user-main/logs/access.log" >/tmp/user-access.txt
rg 'Trace-ID' "${ROOT_DIR}/gateway-main/logs/access.log" >/tmp/gateway-access.txt

echo "smoke ${MODE} passed"
