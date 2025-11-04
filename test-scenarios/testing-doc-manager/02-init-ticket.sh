#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

# Initialize docs root (creates ttmp/, vocabulary.yaml, templates, guidelines) and seed default vocabulary
${DOCMGR} init --seed-vocabulary || true

# Create ticket workspace (RFC-aligned under ttmp/)
${DOCMGR} create-ticket --ticket MEN-4242 --title "Normalize chat API paths and WebSocket lifecycle" --topics chat,backend,websocket

# Create a second ticket using explicit --root ttmp (tests flag-based root)
${DOCMGR} create-ticket --root ttmp --ticket MEN-5678 --title "Secondary ticket — WebSocket reconnection plan" --topics chat,backend

# Verify scaffolding (both tickets should be listed)
${DOCMGR} list tickets
${DOCMGR} list tickets --ticket MEN-5678
