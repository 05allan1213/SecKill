#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

docker compose -f "${ROOT_DIR}/docker-compose.yml" up -d etcd mysql redis kafka

cd "${ROOT_DIR}/seckill-main"
GOCACHE=/tmp/go-build-cache-seckill go test -tags=integration ./...

cd "${ROOT_DIR}/user-main"
GOCACHE=/tmp/go-build-cache-user go test -tags=integration ./...

cd "${ROOT_DIR}/gateway-main"
GOCACHE=/tmp/go-build-cache-gateway go test -tags=integration ./...
