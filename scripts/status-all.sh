#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

for entry in "backend:58080:/health" "frontend:3000:/"; do
  name="${entry%%:*}"
  rest="${entry#*:}"
  port="${rest%%:*}"
  path="${rest#*:}"
  state="DOWN"
  pid="-"
  pid_file="${ROOT_DIR}/.run/${name}.pid"
  if [[ -f "${pid_file}" ]]; then
    pid="$(<"${pid_file}")"
  fi
  if curl -fsS "http://127.0.0.1:${port}${path}" >/dev/null 2>&1; then
    state="UP"
  fi
  printf '%-10s port=%-5s pid=%-7s %s\n' "${name}" "${port}" "${pid}" "${state}"
done

mysql_state="DOWN"
if docker compose -f "${ROOT_DIR}/compose.yaml" exec -T mysql mysqladmin ping -h 127.0.0.1 -u nimbus -pnimbus_dev --silent >/dev/null 2>&1; then
  mysql_state="UP"
fi
printf '%-10s port=%-5s pid=%-7s %s\n' "mysql" "23316" "docker" "${mysql_state}"

redis_state="DOWN"
if docker compose -f "${ROOT_DIR}/compose.yaml" exec -T redis redis-cli ping 2>/dev/null | grep -q PONG; then
  redis_state="UP"
fi
printf '%-10s port=%-5s pid=%-7s %s\n' "redis" "27316" "docker" "${redis_state}"
