#!/usr/bin/env bash
# Count docmgr failure signatures across agent transcript stores.
# Patterns were seeded from error strings in the docmgr source plus common
# Cobra/CLI failures, then refined by inspecting hits.
# Usage: ./02-docmgr-error-patterns.sh <output-dir>
set -euo pipefail
OUT=${1:?output dir}
mkdir -p "$OUT"

declare -A STORES=(
  [codex]="$HOME/.codex/sessions"
  [pi]="$HOME/.pi/agent/sessions"
  [claude]="$HOME/.claude/projects"
)

PATTERNS=(
  'unknown flag: --[a-z-]+'
  'unknown command \\\\"[a-z-]+\\\\" for \\\\"docmgr'
  'required flag\(s\) [^ ]+ not set'
  'ticket not found'
  'ticket ambiguous'
  'no ticket workspace found'
  'not part of the suggested vocabulary'
  'vocabulary warning'
  'file does not exist'
  'RelatedFiles.*does not exist'
  'could not find ticket'
  'no docs root'
  'failed to parse frontmatter'
  'invalid frontmatter'
  'missing frontmatter'
  'accepts at most'
  'Error: open .*ttmp'
  'docmgr: command not found'
)

for name in codex pi claude; do
  dir=${STORES[$name]}
  : > "$OUT/${name}_errors.txt"
  for p in "${PATTERNS[@]}"; do
    # rg exits 1 on zero matches; don't let that kill the script under set -e
    c=$( (rg -c --no-filename "$p" "$dir" --glob '*.jsonl' 2>/dev/null || true) | awk -F: '{s+=$NF} END {print s+0}')
    printf "%8d  %s\n" "$c" "$p" >> "$OUT/${name}_errors.txt"
  done
  sort -rn -o "$OUT/${name}_errors.txt" "$OUT/${name}_errors.txt"
  echo "== $name =="; cat "$OUT/${name}_errors.txt"
done
