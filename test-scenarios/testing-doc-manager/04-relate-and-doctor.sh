#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

# Relate files to the ticket index (sets RelatedFiles) with notes
${DOCMGR} doc relate --ticket MEN-4242 \
  --file-note "backend/chat/api/register.go:Registers API routes (scenario)" \
  --file-note "backend/chat/ws/manager.go:WebSocket lifecycle (scenario)" \
  --file-note "web/src/store/api/chatApi.ts:Frontend API integration (scenario)"

# Relate files to the second ticket (focus on WS lifecycle) with notes
${DOCMGR} doc relate --ticket MEN-5678 \
  --file-note "backend/chat/ws/manager.go:WebSocket lifecycle (scenario)" \
  --file-note "web/src/ui/chat/ChatPanel.tsx:Frontend chat panel (scenario)"

# Optional: see suggestions with reasons (no changes applied)
# ${DOCMGR} doc relate --ticket MEN-4242 --suggest --query WebSocket --topics chat

# Doctor checks with ignore and thresholds
${DOCMGR} doctor \
  --ignore-dir _templates --ignore-dir _guidelines \
  --stale-after 30 \
  --fail-on error

# Optional: simulate staleness (uncomment)
# sed -i 's/LastUpdated:.*/LastUpdated: 2025-09-01T00:00:00Z/' "${INDEX_MD}"
# ${DOCMGR} doctor --ignore-dir _templates --ignore-dir _guidelines --stale-after 14 --fail-on warning
