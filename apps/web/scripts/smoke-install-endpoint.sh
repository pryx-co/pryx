#!/usr/bin/env bash

set -euo pipefail

URL="${1:-https://pryx.dev/install}"

headers_file="$(mktemp)"
body_file="$(mktemp)"
trap 'rm -f "$headers_file" "$body_file"' EXIT

curl -fsSL -D "$headers_file" "$URL" -o "$body_file"

if ! grep -qi '^content-type:.*text/x-shellscript' "$headers_file"; then
  echo "Expected text/x-shellscript content-type from $URL" >&2
  exit 1
fi

first_line="$(sed -n '1p' "$body_file")"
if [[ "$first_line" != '#!/usr/bin/env bash' ]]; then
  echo "Expected shell script response from $URL" >&2
  exit 1
fi

if grep -qi '<html' "$body_file"; then
  echo "Expected raw shell script body, got HTML-like content from $URL" >&2
  exit 1
fi

echo "Install endpoint smoke test passed: $URL"
