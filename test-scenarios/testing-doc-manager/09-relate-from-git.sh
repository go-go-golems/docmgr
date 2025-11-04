#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

# Create an untracked file to ensure git-status suggestions fire
mkdir -p backend/chat/api
cat > backend/chat/api/relate_from_git_test.go <<'EOF'
package api

// Temporary test file to validate docmgr relate --from-git suggestions
func tempRelateTest() {}
EOF

# Suggest and apply changed files from git working tree
${DOCMGR} relate --ticket MEN-4242 --suggest --from-git --apply-suggestions

echo "[ok] relate --from-git suggestions applied"


