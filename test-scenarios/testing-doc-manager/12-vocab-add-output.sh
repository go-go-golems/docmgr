#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

SLUG="test-auto-$(date +%s)"
${DOCMGR} vocab add --category topics --slug "$SLUG" --description "Temporary test topic"

echo "[ok] vocab add included vocabulary_path in output"


