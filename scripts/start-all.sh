#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="${ROOT_DIR}/.run"
LOG_DIR="${RUN_DIR}/logs"
mkdir -p "${LOG_DIR}"

"${ROOT_DIR}/scripts/init-local.sh"

BACKEND_BINARY="${ROOT_DIR}/backend/bin/nimbus-server"
if [[ ! -x "${BACKEND_BINARY}" ]]; then
  echo "Missing ${BACKEND_BINARY}; run 'cd backend && make build' first" >&2
  exit 1
fi

start_process() {
  local name="$1"
  local port="$2"
  shift 2
  local pid_file="${RUN_DIR}/${name}.pid"

  if [[ -f "${pid_file}" ]] && kill -0 "$(<"${pid_file}")" 2>/dev/null; then
    echo "[running] ${name} pid=$(<"${pid_file}") port=${port}"
    return
  fi
  rm -f "${pid_file}"
  if nc -z 127.0.0.1 "${port}" 2>/dev/null; then
    echo "Port ${port} is already occupied; stop the existing process first" >&2
    exit 1
  fi

  nohup "$@" </dev/null >"${LOG_DIR}/${name}.log" 2>&1 &
  echo "$!" >"${pid_file}"
  echo "[starting] ${name} pid=$! port=${port}"
}

start_process backend 58080 env \
  GIN_MODE="${GIN_MODE:-release}" \
  NIMBUS_DB_DSN="${NIMBUS_DB_DSN:-nimbus:nimbus_dev@tcp(127.0.0.1:23316)/nimbus_platform_go?charset=utf8mb4&parseTime=True&loc=Local}" \
  NIMBUS_REDIS_ADDR="${NIMBUS_REDIS_ADDR:-127.0.0.1:27316}" \
  "${BACKEND_BINARY}"

start_process frontend 3000 \
  bash -lc \
  "cd '${ROOT_DIR}/frontend' && exec ./node_modules/.bin/vite --mode env.local --host 0.0.0.0 --port 3000"

for target in "backend:58080/health" "frontend:3000/"; do
  name="${target%%:*}"
  address="${target#*:}"
  for _ in {1..30}; do
    curl -fsS "http://127.0.0.1:${address}" >/dev/null 2>&1 && break
    sleep 1
  done
  curl -fsS "http://127.0.0.1:${address}" >/dev/null || {
    echo "${name} failed to become healthy; see ${LOG_DIR}/${name}.log" >&2
    exit 1
  }
done

"${ROOT_DIR}/scripts/status-all.sh"
