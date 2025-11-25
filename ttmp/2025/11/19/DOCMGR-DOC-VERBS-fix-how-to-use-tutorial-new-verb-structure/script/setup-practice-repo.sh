#!/usr/bin/env bash

set -euo pipefail

TARGET="${1:-/tmp/docmgr-practice}"
DOCMGR_BIN="${DOCMGR_BIN:-/home/manuel/.local/bin/docmgr}"
TICKET="${TICKET:-MEN-PRAC}"
TITLE="${TITLE:-\"Practice workspace for docmgr tutorial\"}"
TOPICS="${TOPICS:-tutorial,practice}"

if [[ ! -x "$DOCMGR_BIN" ]]; then
  echo "error: docmgr binary not executable at $DOCMGR_BIN" >&2
  exit 1
fi

echo ">>> Preparing practice workspace at $TARGET"
rm -rf "$TARGET"
mkdir -p "$TARGET"
cd "$TARGET"

echo ">>> Initializing empty git repository (optional but keeps commands consistent)"
git init -q

echo ">>> Running docmgr init --seed-vocabulary"
"$DOCMGR_BIN" init --seed-vocabulary --root ttmp

echo ">>> Creating starter ticket ($TICKET)"
"$DOCMGR_BIN" ticket create-ticket \
  --ticket "$TICKET" \
  --title "$TITLE" \
  --topics "$TOPICS"

cat <<EOF

Setup complete!
- Workspace path: $TARGET
- Docs root: $TARGET/ttmp
- Ticket created: $TICKET

Next steps:
1. Open $TARGET in your editor.
2. Follow the tutorial manually (doc add, relate, tasks, changelog).
3. When you're done practicing, rerun this script to reset.
EOF

