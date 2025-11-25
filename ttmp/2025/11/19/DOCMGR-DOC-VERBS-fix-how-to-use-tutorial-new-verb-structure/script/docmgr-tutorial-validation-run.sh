#!/usr/bin/env bash

set -euo pipefail

TARGET="${1:-/tmp/test-git-repo}"
DOCMGR_BIN="${DOCMGR_BIN:-/home/manuel/.local/bin/docmgr}"
LOG_PATH="${LOG_PATH:-$TARGET/docmgr-run.log}"

if [[ ! -x "$DOCMGR_BIN" ]]; then
  echo "error: docmgr binary not executable at $DOCMGR_BIN" >&2
  exit 1
fi

echo ">>> Cleaning $TARGET"
rm -rf "$TARGET"
mkdir -p "$TARGET"
cd "$TARGET"

exec > >(tee "$LOG_PATH") 2>&1

echo ">>> Initialized logging at $LOG_PATH"
echo ">>> Using docmgr binary: $DOCMGR_BIN"
echo ">>> Working directory: $(pwd)"

echo ">>> Initializing git repo and seed files"
git init -q
mkdir -p backend/api web/src/store/api
cat > backend/api/register.go <<'EOF'
package api

// Registers API routes (placeholder)

func Register() {}
EOF
cat > web/src/store/api/chatApi.ts <<'EOF'
export const chatApi = {};
EOF
git add backend/api/register.go web/src/store/api/chatApi.ts
git commit -q -m "Initial placeholder files"

echo ">>> Running docmgr init"
"$DOCMGR_BIN" init --seed-vocabulary --root ttmp

echo ">>> Creating ticket MEN-3083"
"$DOCMGR_BIN" ticket create-ticket --ticket MEN-3083 --title "Tutorial validation ticket" --topics test,backend

echo ">>> Adding design doc"
"$DOCMGR_BIN" doc add --ticket MEN-3083 --doc-type design-doc --title "Placeholder design context"

echo ">>> Relating implementation files"
"$DOCMGR_BIN" doc relate --ticket MEN-3083 \
  --file-note "backend/api/register.go:Registers API routes (normalization logic)" \
  --file-note "web/src/store/api/chatApi.ts:Frontend integration"

echo ">>> Adding representative task"
"$DOCMGR_BIN" task add --ticket MEN-3083 --text "Update API docs for /chat/v2"

echo ">>> Recording changelog entry"
"$DOCMGR_BIN" changelog update --ticket MEN-3083 \
  --entry "Initial tutorial validation pass" \
  --file-note "backend/api/register.go:Source implementation for normalization"

echo ">>> Running doctor validation"
"$DOCMGR_BIN" doctor --root ttmp --ticket MEN-3083 --stale-after 30 --fail-on error

echo ">>> Listing ticket files"
find ttmp -maxdepth 5 -type f -print

echo ">>> Script complete. Log captured at $LOG_PATH"

