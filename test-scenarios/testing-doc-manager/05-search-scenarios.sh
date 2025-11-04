#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

# Content search with snippet
${DOCMGR} search --query "WebSocket" --ticket MEN-4242

# Content search for second ticket
${DOCMGR} search --query "WebSocket" --ticket MEN-5678

# Metadata-only search
${DOCMGR} search --ticket MEN-4242 --topics websocket,backend --doc-type design-doc

# Metadata-only search for second ticket
${DOCMGR} search --ticket MEN-5678 --topics chat,backend --doc-type design-doc

# Reverse lookup
${DOCMGR} search --file backend/chat/api/register.go

# Directory lookup
${DOCMGR} search --dir web/src/store/api/

# External sources
${DOCMGR} search --external-source "https://example.com/ws-lifecycle"

# Date filtering
${DOCMGR} search --updated-since "1 day ago" --ticket MEN-4242
${DOCMGR} search --since "last month" --ticket MEN-4242

# File suggestions via heuristics (git + ripgrep/grep)
${DOCMGR} search --ticket MEN-4242 --topics chat --files

# Date filtering for second ticket
${DOCMGR} search --updated-since "1 day ago" --ticket MEN-5678
${DOCMGR} search --since "last month" --ticket MEN-5678

# File suggestions for second ticket
${DOCMGR} search --ticket MEN-5678 --topics chat --files
