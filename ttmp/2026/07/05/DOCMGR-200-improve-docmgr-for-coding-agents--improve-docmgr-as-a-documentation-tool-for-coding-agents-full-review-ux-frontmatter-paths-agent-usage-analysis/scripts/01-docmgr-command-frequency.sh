#!/usr/bin/env bash
# Full-corpus frequency analysis of docmgr invocations across coding-agent
# transcript stores (codex, pi, claude-code). Greps raw JSONL directly —
# breadth pass; the minitrace/DuckDB scripts do the structured depth pass.
#
# Usage: ./01-docmgr-command-frequency.sh <output-dir>
set -euo pipefail
OUT=${1:?output dir}
mkdir -p "$OUT"

declare -A STORES=(
  [codex]="$HOME/.codex/sessions"
  [pi]="$HOME/.pi/agent/sessions"
  [claude]="$HOME/.claude/projects"
)

for name in codex pi claude; do
  dir=${STORES[$name]}
  # Extract "docmgr <verb> <subverb>" tokens from tool-call text.
  rg -o --no-filename 'docmgr [a-z][a-z-]+( [a-z][a-z-]+)?' "$dir" --glob '*.jsonl' 2>/dev/null \
    | sort | uniq -c | sort -rn > "$OUT/${name}_command_freq.txt"
  # Flags used with docmgr on the same escaped-JSON command line.
  rg -o --no-filename 'docmgr [^"\\]{0,200}' "$dir" --glob '*.jsonl' 2>/dev/null \
    | rg -o -- '--[a-z][a-z-]+' | sort | uniq -c | sort -rn > "$OUT/${name}_flag_freq.txt"
  echo "$name done"
done
