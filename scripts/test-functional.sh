#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${NIMBUS_BASE_URL:-http://127.0.0.1:58080}"
TENANT_NAME="${NIMBUS_BOOTSTRAP_TENANT:-Nimbus Framework}"
USERNAME="${NIMBUS_BOOTSTRAP_USERNAME:-admin}"
PASSWORD="${NIMBUS_BOOTSTRAP_PASSWORD:-admin123}"

for command in curl jq; do
  command -v "${command}" >/dev/null 2>&1 || {
    echo "Missing command: ${command}" >&2
    exit 1
  }
done

expect_success() {
  local label="$1"
  local response="$2"
  local code
  code="$(jq -r '.code // -1' <<<"${response}")"
  if [[ "${code}" != "0" && "${code}" != "200" ]]; then
    echo "[FAIL] ${label}: ${response}" >&2
    exit 1
  fi
  echo "[PASS] ${label}"
}

health="$(curl -fsS "${BASE_URL}/health")"
[[ "$(jq -r '.data.status // .status // empty' <<<"${health}")" == "UP" ]]
echo "[PASS] Health"

tenant_response="$(curl -fsS --get "${BASE_URL}/admin-api/system/tenant/get-id-by-name" --data-urlencode "name=${TENANT_NAME}")"
expect_success "Tenant lookup" "${tenant_response}"
tenant_id="$(jq -r '.data' <<<"${tenant_response}")"

login_response="$(curl -fsS -X POST "${BASE_URL}/admin-api/system/auth/login" \
  -H "tenant-id: ${tenant_id}" \
  -H "Content-Type: application/json" \
  --data "{\"username\":\"${USERNAME}\",\"password\":\"${PASSWORD}\"}")"
expect_success "Admin login" "${login_response}"
access_token="$(jq -r '.data.accessToken' <<<"${login_response}")"
refresh_token="$(jq -r '.data.refreshToken' <<<"${login_response}")"
[[ -n "${access_token}" && -n "${refresh_token}" && "${access_token}" != "${refresh_token}" ]]
echo "[PASS] Separate access and refresh tokens"

refresh_response="$(curl -fsS -X POST --get "${BASE_URL}/admin-api/system/auth/refresh-token" \
  --data-urlencode "refreshToken=${refresh_token}")"
expect_success "Token refresh" "${refresh_response}"
rotated_access="$(jq -r '.data.accessToken' <<<"${refresh_response}")"
rotated_refresh="$(jq -r '.data.refreshToken' <<<"${refresh_response}")"
[[ "${rotated_access}" != "${access_token}" && "${rotated_refresh}" != "${refresh_token}" ]]
echo "[PASS] Token rotation"

auth_headers=(-H "tenant-id: ${tenant_id}" -H "Authorization: Bearer ${rotated_access}")
for endpoint in \
  "/admin-api/system/auth/get-permission-info" \
  "/admin-api/system/user/page?pageNo=1&pageSize=10" \
  "/admin-api/system/role/page?pageNo=1&pageSize=10" \
  "/admin-api/system/menu/list" \
  "/admin-api/system/dept/list" \
  "/admin-api/system/post/page?pageNo=1&pageSize=10" \
  "/admin-api/system/dict-type/page?pageNo=1&pageSize=10" \
  "/admin-api/system/tenant/page?pageNo=1&pageSize=10" \
  "/admin-api/system/login-log/page?pageNo=1&pageSize=10" \
  "/admin-api/system/operate-log/page?pageNo=1&pageSize=10" \
  "/admin-api/system/oauth2-client/page?pageNo=1&pageSize=10" \
  "/admin-api/system/oauth2-token/page?pageNo=1&pageSize=10" \
  "/admin-api/system/notice/page?pageNo=1&pageSize=10" \
  "/admin-api/system/notify-template/page?pageNo=1&pageSize=10" \
  "/admin-api/system/notify-message/page?pageNo=1&pageSize=10" \
  "/admin-api/system/mail-account/page?pageNo=1&pageSize=10" \
  "/admin-api/system/mail-template/page?pageNo=1&pageSize=10" \
  "/admin-api/system/mail-log/page?pageNo=1&pageSize=10" \
  "/admin-api/system/sms-channel/page?pageNo=1&pageSize=10" \
  "/admin-api/system/sms-template/page?pageNo=1&pageSize=10" \
  "/admin-api/system/sms-log/page?pageNo=1&pageSize=10" \
  "/admin-api/infra/config/page?pageNo=1&pageSize=10" \
  "/admin-api/infra/file-config/page?pageNo=1&pageSize=10" \
  "/admin-api/infra/api-access-log/page?pageNo=1&pageSize=10" \
  "/admin-api/infra/file/page?pageNo=1&pageSize=10" \
  "/admin-api/infra/api-error-log/page?pageNo=1&pageSize=10" \
  "/admin-api/infra/data-source-config/list" \
  "/admin-api/infra/job/page?pageNo=1&pageSize=10" \
  "/admin-api/infra/job-log/page?pageNo=1&pageSize=10" \
  "/admin-api/infra/redis/get-monitor-info" \
  "/admin-api/member/user/page?pageNo=1&pageSize=10" \
  "/admin-api/member/level/list" \
  "/admin-api/member/group/page?pageNo=1&pageSize=10" \
  "/admin-api/member/tag/page?pageNo=1&pageSize=10" \
  "/admin-api/pay/app/page?pageNo=1&pageSize=10" \
  "/admin-api/pay/order/page?pageNo=1&pageSize=10" \
  "/admin-api/pay/refund/page?pageNo=1&pageSize=10"; do
  response="$(curl -fsS "${auth_headers[@]}" "${BASE_URL}${endpoint}")"
  expect_success "${endpoint}" "${response}"
done

echo "Nimbus Framework Go functional smoke test passed."
