#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/.." >/dev/null 2>&1 && pwd)"

canonical="${ROOT_DIR}/install.sh"
legacy_web_copy="${ROOT_DIR}/apps/web/src/install.sh"
wrapper_script="${ROOT_DIR}/scripts/install.sh"

if [[ ! -f "$canonical" ]]; then
  echo "Canonical installer missing: $canonical" >&2
  exit 1
fi

if [[ -f "$legacy_web_copy" ]]; then
  echo "Legacy installer copy still exists: $legacy_web_copy" >&2
  echo "Web endpoint should import canonical installer from repository root." >&2
  exit 1
fi

if ! grep -q 'exec "${ROOT_DIR}/install.sh"' "$wrapper_script"; then
  echo "scripts/install.sh must delegate to canonical install.sh" >&2
  exit 1
fi

echo "Installer source-of-truth checks passed"
