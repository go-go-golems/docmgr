#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

# Content search with snippet
${DOCMGR} doc search --query "WebSocket" --ticket MEN-4242
${DOCMGR} doc search --query "WebSocket" --ticket MEN-4242 --order-by rank

# Content search for second ticket
${DOCMGR} doc search --query "WebSocket" --ticket MEN-5678

# Metadata-only search
${DOCMGR} doc search --ticket MEN-4242 --topics websocket,backend --doc-type design-doc

# Metadata-only search for second ticket
${DOCMGR} doc search --ticket MEN-5678 --topics chat,backend --doc-type design-doc

# Reverse lookup
${DOCMGR} doc search --file backend/chat/api/register.go

# Directory lookup
${DOCMGR} doc search --dir web/src/store/api/

# External sources
${DOCMGR} doc search --external-source "https://example.com/ws-lifecycle"

# Date filtering
${DOCMGR} doc search --updated-since "1 day ago" --ticket MEN-4242
${DOCMGR} doc search --since "last month" --ticket MEN-4242

# File suggestions via heuristics (git + ripgrep/grep)
${DOCMGR} doc search --ticket MEN-4242 --topics chat --files

# Wonky path regression: ensure doc search matches denormalized entries
${DOCMGR} doc search --ticket MEN-4242 --file "../../../../../backend/chat/api/register.go"
${DOCMGR} doc search --ticket MEN-4242 --file "../backend/chat/api/register.go"
${DOCMGR} doc search --ticket MEN-4242 --file "${ROOT_DIR}/acme-chat-app/backend/chat/ws/manager.go"
${DOCMGR} doc search --ticket MEN-4242 --file "register.go"

# Date filtering for second ticket
${DOCMGR} doc search --updated-since "1 day ago" --ticket MEN-5678
${DOCMGR} doc search --since "last month" --ticket MEN-5678

# File suggestions for second ticket
${DOCMGR} doc search --ticket MEN-5678 --topics chat --files
