#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

for name in frontend backend; do
  pid_file="${ROOT_DIR}/.run/${name}.pid"
  [[ -f "${pid_file}" ]] || continue
  pid="$(<"${pid_file}")"
  kill "${pid}" 2>/dev/null || true
  for _ in {1..20}; do
    kill -0 "${pid}" 2>/dev/null || break
    sleep 0.25
  done
  kill -9 "${pid}" 2>/dev/null || true
  rm -f "${pid_file}"
done
