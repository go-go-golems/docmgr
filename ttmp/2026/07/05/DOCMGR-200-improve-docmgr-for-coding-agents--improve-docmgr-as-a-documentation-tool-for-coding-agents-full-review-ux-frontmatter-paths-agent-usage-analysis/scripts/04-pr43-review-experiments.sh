#!/usr/bin/env bash
set -euo pipefail

# PR #43 review validation harness.
# Run from the docmgr repository root. It records the commands used for the
# project/code review and reproduces one API/CLI parity edge case around stable
# task IDs.

repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

out_dir="$repo_root/ttmp/2026/07/05/DOCMGR-200-improve-docmgr-for-coding-agents--improve-docmgr-as-a-documentation-tool-for-coding-agents-full-review-ux-frontmatter-paths-agent-usage-analysis/sources"
mkdir -p "$out_dir"
out="$out_dir/pr43-review-experiments.txt"
: >"$out"

log() {
  printf '\n### %s\n' "$*" | tee -a "$out"
}
run() {
  log "$*"
  "$@" 2>&1 | tee -a "$out"
}

log "repository"
git rev-parse HEAD | tee -a "$out"
git status --short | tee -a "$out"

run go test ./... -count=1
run go test -tags sqlite_fts5 ./... -count=1

log "ui build"
(
  cd ui
  pnpm build
) 2>&1 | tee -a "$out"

log "ui lint"
(
  cd ui
  pnpm lint
) 2>&1 | tee -a "$out"

log "github checks"
if command -v gh >/dev/null 2>&1; then
  gh pr checks 43 2>&1 | tee -a "$out" || true
  gh api "repos/go-go-golems/docmgr/commits/$(git rev-parse HEAD)/check-runs" \
    --jq '.check_runs[] | {name:.name, status:.status, conclusion:.conclusion, html_url:.html_url, app:.app.slug}' \
    2>&1 | tee -a "$out" || true
  gh api repos/go-go-golems/docmgr/check-runs/85406559070/annotations \
    --jq '.[] | {path,start_line,end_line,annotation_level,message,title}' \
    2>&1 | tee -a "$out" || true
else
  echo "gh not installed; skipped" | tee -a "$out"
fi

log "stable task id API parity experiment"
work=$(mktemp -d)
bin="$work/docmgr"
trap 'if [[ -n "${server_pid:-}" ]]; then kill "$server_pid" 2>/dev/null || true; fi; rm -rf "$work"' EXIT

go build -tags sqlite_fts5 -o "$bin" ./cmd/docmgr
(
  cd "$work"
  git init -q
  "$bin" init --root ttmp >/dev/null
  "$bin" ticket create --root ttmp --ticket API-1 --title "API stable id review" >/dev/null
  "$bin" task add --root ttmp --ticket API-1 --text "task created with a stable marker" >/dev/null
  task_file=$(find ttmp -path '*/API-1*/tasks.md' -print -quit)
  echo "tasks file: $task_file" | tee -a "$out"
  grep -n '<!-- t:' "$task_file" | tee -a "$out"
  stable_id=$(grep -o '<!-- t:[a-z0-9]* -->' "$task_file" | head -n1 | sed -E 's/.*t:([a-z0-9]+).*/\1/')
  echo "stable id: $stable_id" | tee -a "$out"

  "$bin" api serve --root ttmp --addr 127.0.0.1:18787 >server.log 2>&1 &
  server_pid=$!
  for _ in {1..50}; do
    if curl -fsS 'http://127.0.0.1:18787/api/v1/healthz' >/dev/null 2>&1; then
      break
    fi
    sleep 0.1
  done

  echo "GET /tickets/tasks exposes stableId:" | tee -a "$out"
  curl -fsS 'http://127.0.0.1:18787/api/v1/tickets/tasks?ticket=API-1' | tee -a "$out"
  printf '\n' | tee -a "$out"

  echo "POST /tickets/tasks/check with stable id string (expected to fail today):" | tee -a "$out"
  curl -sS -i \
    -H 'Content-Type: application/json' \
    -d "{\"ticket\":\"API-1\",\"ids\":[\"$stable_id\"],\"checked\":true}" \
    'http://127.0.0.1:18787/api/v1/tickets/tasks/check' | tee -a "$out" || true
  printf '\n' | tee -a "$out"

  echo "POST /tickets/tasks/check with positional id 1 (succeeds):" | tee -a "$out"
  curl -sS -i \
    -H 'Content-Type: application/json' \
    -d '{"ticket":"API-1","ids":[1],"checked":true}' \
    'http://127.0.0.1:18787/api/v1/tickets/tasks/check' | tee -a "$out" || true
  printf '\n' | tee -a "$out"
)

log "done"
printf 'wrote %s\n' "$out" | tee -a "$out"
