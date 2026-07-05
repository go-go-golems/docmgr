#!/usr/bin/env bash
# Stage a sample of agent sessions that actually invoke docmgr into
# store-shaped trees, then convert them into minitrace archives.
#
# Inputs: <hits-dir> containing {codex,pi,claude}_hits.txt (rg -c output:
#         "<path>:<match-count>"), produced by grepping the native stores
#         for real docmgr subcommand invocations.
# Sample: top TOP_N sessions by docmgr-hit count per store, plus every
#         STRIDE-th session of the remainder for spread.
#
# Usage: ./03-stage-and-convert.sh <hits-dir> <work-dir> [TOP_N] [STRIDE]
set -euo pipefail
HITS=${1:?hits dir}
WORK=${2:?work dir}
TOP_N=${3:-40}
STRIDE=${4:-25}

select_sessions() { # <hits-file>
  sort -t: -k2 -rn "$1" | cut -d: -f1 | awk -v top="$TOP_N" -v stride="$STRIDE" \
    'NR<=top {print; next} (NR-top)%stride==0 {print}'
}

stage() { # <hits-file> <store-root> <staged-root>
  local hits=$1 root=$2 staged=$3
  mkdir -p "$staged"
  while IFS= read -r f; do
    rel=${f#"$root"/}
    mkdir -p "$staged/$(dirname -- "$rel")"
    ln -sf "$f" "$staged/$rel"
  done < <(select_sessions "$hits")
}

echo "--- staging codex ---"
stage "$HITS/codex_hits.txt" "$HOME/.codex/sessions" "$WORK/staged/codex-home/sessions"
echo "--- staging pi ---"
stage "$HITS/pi_hits.txt" "$HOME/.pi/agent/sessions" "$WORK/staged/pi-sessions"
echo "--- staging claude ---"
stage "$HITS/claude_hits.txt" "$HOME/.claude/projects" "$WORK/staged/claude-projects"

for n in codex pi claude; do
  find "$WORK/staged" -name '*.jsonl' | grep -c "$n" || true
done

echo "--- converting ---"
go-minitrace convert codex --source-dir "$WORK/staged/codex-home" --output-dir "$WORK/archive/codex"
go-minitrace convert pi --source-dir "$WORK/staged/pi-sessions" --output-dir "$WORK/archive/pi"
go-minitrace convert claude-code --source-dir "$WORK/staged/claude-projects" --output-dir "$WORK/archive/claude"

echo "--- archive counts ---"
for n in codex pi claude; do
  echo "$n: $(find "$WORK/archive/$n" -name '*.minitrace.json' 2>/dev/null | wc -l)"
done
