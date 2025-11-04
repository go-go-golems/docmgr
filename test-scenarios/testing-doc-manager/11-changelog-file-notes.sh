#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

# Append a changelog entry with files and file-notes; verify notes are present
${DOCMGR} changelog update --ticket MEN-4242 \
  --entry "Test changelog file notes rendering" \
  --files backend/chat/api/register.go,web/src/store/api/chatApi.ts \
  --file-note "backend/chat/api/register.go:Source of path normalization" \
  --file-note "web/src/store/api/chatApi.ts=Frontend integration"

CHG=$(ls ttmp/MEN-4242-*/changelog.md | head -n1)
echo "--- BEGIN changelog excerpt: ${CHG} ---"
tail -n 30 "$CHG"
echo "--- END changelog excerpt ---"

echo "[ok] changelog file-notes appended"


