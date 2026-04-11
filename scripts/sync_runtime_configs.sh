#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENDPOINTS="${CONFIG_CENTER_ENDPOINTS:-127.0.0.1:20001}"

cd "${ROOT_DIR}/gateway-main"
go run ./cmd/configsync -endpoints "${ENDPOINTS}" -dir "${ROOT_DIR}/configs/runtime"
