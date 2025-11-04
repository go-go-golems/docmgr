#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

# Full status with per-ticket rows
${DOCMGR} status

# Summary only
${DOCMGR} status --summary-only

# Adjust staleness threshold to 30 days
${DOCMGR} status --stale-after 30

# Run from a nested directory to verify .ttmp.yaml discovery and relative root resolution
cd web
${DOCMGR} status --summary-only
cd - >/dev/null

echo "[ok] Status checks completed"
