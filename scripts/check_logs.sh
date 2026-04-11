#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIR_LIMIT_MB="${DIR_LIMIT_MB:-500}"
FILE_LIMIT_MB="${FILE_LIMIT_MB:-100}"

dirs=(
  "${ROOT_DIR}/gateway-main/logs"
  "${ROOT_DIR}/user-main/logs"
  "${ROOT_DIR}/seckill-main/logs"
)

status=0

for dir in "${dirs[@]}"; do
  mkdir -p "${dir}"

  dir_bytes="$(du -sb "${dir}" | awk '{print $1}')"
  dir_limit_bytes=$((DIR_LIMIT_MB * 1024 * 1024))
  if (( dir_bytes > dir_limit_bytes )); then
    echo "directory exceeds limit: ${dir} (${dir_bytes} bytes > ${dir_limit_bytes} bytes)" >&2
    status=1
  else
    echo "directory within limit: ${dir} (${dir_bytes} bytes)"
  fi

  max_file_bytes=0
  max_file_path=""
  while IFS= read -r entry; do
    size="${entry%% *}"
    path="${entry#* }"
    if (( size > max_file_bytes )); then
      max_file_bytes="${size}"
      max_file_path="${path}"
    fi
  done < <(find "${dir}" -maxdepth 1 -type f -printf '%s %p\n' | sort -nr)

  file_limit_bytes=$((FILE_LIMIT_MB * 1024 * 1024))
  if (( max_file_bytes > file_limit_bytes )); then
    echo "largest file exceeds limit: ${max_file_path} (${max_file_bytes} bytes > ${file_limit_bytes} bytes)" >&2
    status=1
  elif [[ -n "${max_file_path}" ]]; then
    echo "largest file within limit: ${max_file_path} (${max_file_bytes} bytes)"
  else
    echo "no log files found in ${dir}"
  fi

  echo "top files for ${dir}:"
  find "${dir}" -maxdepth 1 -type f -printf '%s %p\n' | sort -nr | head -n 5
  echo "---"
done

exit "${status}"
