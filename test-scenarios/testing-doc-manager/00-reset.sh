#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
rm -rf "${ROOT_DIR}"
mkdir -p "${ROOT_DIR}"
cd "${ROOT_DIR}"

echo "[ok] Reset to ${ROOT_DIR}"
