#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

# Re-write .ttmp.yaml using the CLI (idempotent with --force)
${DOCMGR} configure --root ttmp \
  --owners manuel \
  --intent long-term \
  --vocabulary ttmp/vocabulary.yaml \
  --force

echo "[ok] configure wrote .ttmp.yaml via CLI"


