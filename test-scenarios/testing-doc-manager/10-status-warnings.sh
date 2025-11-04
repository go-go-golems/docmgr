#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

# Create a second ttmp at /tmp to trigger multiple-root warning during status
TMP_TTMP="/tmp/ttmp"
mkdir -p "$TMP_TTMP"

trap 'rm -rf "$TMP_TTMP"' EXIT

# Run in glaze mode to surface warning rows
${DOCMGR} status --with-glaze-output --summary-only

echo "[ok] status warnings (multiple roots/fallback) surfaced"


