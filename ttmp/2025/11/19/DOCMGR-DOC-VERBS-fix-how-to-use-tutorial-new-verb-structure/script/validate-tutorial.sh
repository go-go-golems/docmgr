#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_SCRIPT="$SCRIPT_DIR/docmgr-tutorial-validation-run.sh"

TARGET="${1:-/tmp/test-git-repo}"
ITERATIONS="${ITERATIONS:-1}"

if [[ ! -x "$RUN_SCRIPT" ]]; then
  echo "error: expected runnable script at $RUN_SCRIPT" >&2
  exit 1
fi

echo ">>> Reset helper starting (target=$TARGET, iterations=$ITERATIONS)"

for ((i = 1; i <= ITERATIONS; i++)); do
  echo ""
  echo ">>> Iteration $i/$ITERATIONS"
LOG_PATH_BASE="${LOG_PATH_BASE:-/tmp/docmgr-validation-logs}"
LOG_PATH="${LOG_PATH_BASE%/}/docmgr-run-${i}.log"
mkdir -p "$(dirname "$LOG_PATH")"
LOG_PATH="$LOG_PATH" DOCMGR_BIN="${DOCMGR_BIN:-/home/manuel/.local/bin/docmgr}" \
    "$RUN_SCRIPT" "$TARGET"
done

echo ""
echo ">>> Reset helper finished. Latest repo at $TARGET"

