#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

# Add documents
${DOCMGR} doc add --ticket MEN-4242 --doc-type design-doc --title "Path Normalization Strategy"
${DOCMGR} doc add --ticket MEN-4242 --doc-type reference --title "Chat WebSocket Lifecycle"
${DOCMGR} doc add --ticket MEN-4242 --doc-type playbook --title "Smoke Tests for Chat"

# Show guidelines for design-doc
${DOCMGR} doc guidelines --doc-type design-doc --output markdown || true

# Enrich metadata on index.md
INDEX_MD=$(find ttmp -type f -path "*/MEN-4242-*/index.md" -print -quit)
if [[ -z "${INDEX_MD}" ]]; then
	echo "Could not locate MEN-4242 index.md" >&2
	exit 1
fi
${DOCMGR} meta update --doc "${INDEX_MD}" --field Owners --value "manuel,alex"
${DOCMGR} meta update --doc "${INDEX_MD}" --field Summary --value "Unify chat HTTP paths and stabilize WebSocket flows."
${DOCMGR} meta update --doc "${INDEX_MD}" --field ExternalSources --value "https://example.com/rfc/chat-api,https://example.com/ws-lifecycle"

# Add documents for second ticket
${DOCMGR} doc add --ticket MEN-5678 --doc-type design-doc --title "WebSocket Reconnection Strategy"
${DOCMGR} doc add --ticket MEN-5678 --doc-type reference  --title "Reconnect Lifecycle"
${DOCMGR} doc add --ticket MEN-5678 --doc-type playbook   --title "Reconnect Smoke Tests"

# Enrich metadata on second ticket index.md
INDEX2_MD=$(find ttmp -type f -path "*/MEN-5678-*/index.md" -print -quit)
if [[ -z "${INDEX2_MD}" ]]; then
	echo "Could not locate MEN-5678 index.md" >&2
	exit 1
fi
${DOCMGR} meta update --doc "${INDEX2_MD}" --field Owners --value "manuel"
${DOCMGR} meta update --doc "${INDEX2_MD}" --field Summary --value "Plan WebSocket reconnection strategy."
${DOCMGR} meta update --doc "${INDEX2_MD}" --field ExternalSources --value "https://example.com/ws-reconnect"

# Add a reference doc with intentionally denormalized paths for MEN-4242
TICKET_DIR=$(dirname "${INDEX_MD}")
WONKY_DOC="${TICKET_DIR}/reference/99-wonky-paths-fixture.md"
ABS_REGISTER="${REPO}/backend/chat/api/register.go"
ABS_WS="${REPO}/backend/chat/ws/manager.go"
cat > "${WONKY_DOC}" <<EOF
---
Title: Wonky Path Fixture
Ticket: MEN-4242
Status: active
Topics:
  - chat
  - backend
DocType: reference
Intent: long-term
Owners:
  - manuel
RelatedFiles:
  - Path: ../../../../../backend/chat/api/register.go
    Note: Doc-relative path reference (deep traversal)
  - Path: ../backend/chat/api/register.go
    Note: Ttmp-relative path reference (shallower traversal)
  - Path: ${ABS_WS}
    Note: Absolute path reference (host-specific)
Summary: >
  Fixture document that records RelatedFiles using doc-relative, ttmp-relative,
  and absolute paths so search regression scripts can exercise fuzzy matching.
LastUpdated: $(date -Iseconds)
---

# Wonky Path Fixture

This document is created directly by the scenario scripts to ensure we have
frontmatter entries with denormalized paths stored as-is.
EOF

# List docs and tickets for both tickets
${DOCMGR} list tickets --ticket MEN-4242
${DOCMGR} list docs --ticket MEN-4242
${DOCMGR} list tickets --ticket MEN-5678
${DOCMGR} list docs --ticket MEN-5678
