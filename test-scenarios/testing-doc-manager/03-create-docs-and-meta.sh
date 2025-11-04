#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

# Add documents
${DOCMGR} add --ticket MEN-4242 --doc-type design-doc --title "Path Normalization Strategy"
${DOCMGR} add --ticket MEN-4242 --doc-type reference --title "Chat WebSocket Lifecycle"
${DOCMGR} add --ticket MEN-4242 --doc-type playbook --title "Smoke Tests for Chat"

# Show guidelines for design-doc
${DOCMGR} guidelines --doc-type design-doc --output markdown || true

# Enrich metadata on index.md
INDEX_MD="ttmp/MEN-4242-normalize-chat-api-paths-and-websocket-lifecycle/index.md"
${DOCMGR} meta update --doc "${INDEX_MD}" --field Owners --value "manuel,alex"
${DOCMGR} meta update --doc "${INDEX_MD}" --field Summary --value "Unify chat HTTP paths and stabilize WebSocket flows."
${DOCMGR} meta update --doc "${INDEX_MD}" --field ExternalSources --value "https://example.com/rfc/chat-api,https://example.com/ws-lifecycle"

# Add documents for second ticket
${DOCMGR} add --ticket MEN-5678 --doc-type design-doc --title "WebSocket Reconnection Strategy"
${DOCMGR} add --ticket MEN-5678 --doc-type reference  --title "Reconnect Lifecycle"
${DOCMGR} add --ticket MEN-5678 --doc-type playbook   --title "Reconnect Smoke Tests"

# Enrich metadata on second ticket index.md
INDEX2_MD="ttmp/MEN-5678-secondary-ticket-websocket-reconnection-plan/index.md"
${DOCMGR} meta update --doc "${INDEX2_MD}" --field Owners --value "manuel"
${DOCMGR} meta update --doc "${INDEX2_MD}" --field Summary --value "Plan WebSocket reconnection strategy."
${DOCMGR} meta update --doc "${INDEX2_MD}" --field ExternalSources --value "https://example.com/ws-reconnect"

# List docs and tickets for both tickets
${DOCMGR} list tickets --ticket MEN-4242
${DOCMGR} list docs --ticket MEN-4242
${DOCMGR} list tickets --ticket MEN-5678
${DOCMGR} list docs --ticket MEN-5678
