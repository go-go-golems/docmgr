#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
mkdir -p "${REPO}"
cd "${REPO}"

git init -q
git config user.email "dev@example.com"
git config user.name "Dev"

# Add .docmgrignore to exclude common directories
cat > .docmgrignore <<'EOF'
# Ignore typical VCS and build artifacts
.git/
node_modules/
dist/
EOF

# Add .ttmp.yaml configuration
cat > .ttmp.yaml <<'EOF'
root: ttmp
defaults:
  owners: [manuel]
  intent: long-term
filenamePrefixPolicy: off
# Optional: override default vocabulary path (defaults to "ttmp/vocabulary.yaml")
# vocabulary: ttmp/vocabulary.yaml
EOF

# Backend (Go)
mkdir -p backend/chat/api backend/chat/ws
cat > backend/chat/api/register.go <<'EOF'
package api

// Register routes for chat REST API
// TODO: Normalize chat API paths across services
func Register() {}
EOF

cat > backend/chat/ws/manager.go <<'EOF'
package ws

// WebSocket manager for chat lifecycle
// TODO: Stabilize reconnect logic and support multi-tenant room mapping
func Start() {}
EOF

# Frontend (TypeScript)
mkdir -p web/src/store/api web/src/ui/chat
cat > web/src/store/api/chatApi.ts <<'EOF'
/**
 * Chat API integration
 * TODO: Align path normalization with backend Register()
 */
export const fetchMessages = async () => {};
EOF

cat > web/src/ui/chat/ChatPanel.tsx <<'EOF'
/**
 * Chat UI panel
 * Handles WebSocket lifecycle and transcript rendering
 */
export const ChatPanel = () => null;
EOF

# Add README and initial commits
echo "# acme-chat-app" > README.md

git add .
git commit -m "feat: bootstrap mock repo with backend/frontend chat files and .ttmp.yaml" -q

# Add more commits for git-based heuristics
echo "// Normalize HTTP paths for chat across services" >> backend/chat/api/register.go
git add backend/chat/api/register.go
git commit -m "chore(api): note path normalization TODO" -q

echo "// websocket reconnect stabilization plan" >> backend/chat/ws/manager.go
git add backend/chat/ws/manager.go
git commit -m "chore(ws): note reconnect stabilization" -q

echo "[ok] Mock codebase created at ${REPO}"
