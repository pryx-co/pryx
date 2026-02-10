#!/usr/bin/env bash

set -euo pipefail

base_url="${1:-https://pryx.dev}"
base_url="${base_url%/}"

headers_file="$(mktemp)"
body_file="$(mktemp)"
trap 'rm -f "$headers_file" "$body_file"' EXIT

home_status="$(curl -sS -o /dev/null -w '%{http_code}' "$base_url/")"
if [[ "$home_status" -lt 200 || "$home_status" -ge 400 ]]; then
  echo "Expected / to return 2xx/3xx, got $home_status" >&2
  exit 1
fi

curl -fsSL -D "$headers_file" "$base_url/install" -o "$body_file"

if ! grep -qi '^content-type:.*text/x-shellscript' "$headers_file"; then
  echo "Expected /install to return text/x-shellscript" >&2
  exit 1
fi

if [[ "$(sed -n '1p' "$body_file")" != '#!/usr/bin/env bash' ]]; then
  echo "Expected /install response body to be installer shell script" >&2
  exit 1
fi

api_body="$(curl -fsSL "$base_url/api")"
if ! printf '%s' "$api_body" | grep -q '"status":"operational"'; then
  echo "Expected /api health payload with operational status" >&2
  exit 1
fi

admin_status="$(curl -sS -o /dev/null -w '%{http_code}' "$base_url/api/admin/health")"
case "$admin_status" in
  200|401|403)
    ;;
  *)
    echo "Expected /api/admin/health reachable (200/401/403), got $admin_status" >&2
    exit 1
    ;;
esac

echo "Post-deploy checks passed for $base_url"
