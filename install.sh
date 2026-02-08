#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -x "${SCRIPT_DIR}/xray/install.sh" ]]; then
  exec "${SCRIPT_DIR}/xray/install.sh" "$@"
fi

echo "error: ${SCRIPT_DIR}/xray/install.sh not found or not executable."
echo "run: chmod +x ${SCRIPT_DIR}/xray/install.sh"
exit 1
