#!/usr/bin/env bash
set -euo pipefail

# Compare system docmgr vs locally built docmgr using scenariolog
#
# This script runs the docmgr scenario suite twice:
# 1. With system docmgr (from PATH)
# 2. With locally built docmgr (from repo)
#
# Both runs are recorded in scenariolog SQLite databases, allowing you to:
# - Compare exit codes and timing
# - Search logs for differences
# - Query failures and warnings
# - Diff outputs between versions
#
# Usage:
#   ./compare-docmgr-versions.sh [system-root] [local-root]
#
# Example:
#   ./compare-docmgr-versions.sh /tmp/docmgr-system /tmp/docmgr-local
#
# Then query:
#   DB=/tmp/docmgr-system/.scenario-run.db
#   /tmp/scenariolog-local summary --db "$DB" --output table
#   /tmp/scenariolog-local failures --db "$DB" --output table

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../../../.." && pwd)"
SCENARIO_DIR="${REPO_ROOT}/test-scenarios/testing-doc-manager"

SYSTEM_ROOT="${1:-/tmp/docmgr-system}"
LOCAL_ROOT="${2:-/tmp/docmgr-local}"

# Build scenariolog if needed
SCENARIOLOG_PATH="${SCENARIOLOG_PATH:-/tmp/scenariolog-local}"
if [[ ! -f "${SCENARIOLOG_PATH}" ]]; then
  echo "[info] Building scenariolog: ${REPO_ROOT}/scenariolog -> ${SCENARIOLOG_PATH}" >&2
  go -C "${REPO_ROOT}/scenariolog" build -tags sqlite_fts5 -o "${SCENARIOLOG_PATH}" ./cmd/scenariolog || {
    echo "[warn] Failed to build with FTS5; building without FTS5 (search will be disabled)" >&2
    go -C "${REPO_ROOT}/scenariolog" build -o "${SCENARIOLOG_PATH}" ./cmd/scenariolog
  }
fi

if [[ ! -f "${SCENARIOLOG_PATH}" ]]; then
  echo "[fail] Failed to build scenariolog" >&2
  exit 1
fi

echo "[info] Using SCENARIOLOG_PATH=${SCENARIOLOG_PATH}" >&2

# Build local docmgr binary
LOCAL_DOCMGR="${LOCAL_ROOT}/docmgr-local"
echo "[info] Building local docmgr: ${REPO_ROOT}/cmd/docmgr -> ${LOCAL_DOCMGR}" >&2
mkdir -p "$(dirname "${LOCAL_DOCMGR}")"
go build -o "${LOCAL_DOCMGR}" "${REPO_ROOT}/cmd/docmgr" || {
  echo "[fail] Failed to build local docmgr" >&2
  exit 1
}

# Find system docmgr
SYSTEM_DOCMGR="$(command -v docmgr)" || {
  echo "[fail] System docmgr not found in PATH" >&2
  exit 1
}
echo "[info] Using SYSTEM_DOCMGR=${SYSTEM_DOCMGR}" >&2
echo "[info] Using LOCAL_DOCMGR=${LOCAL_DOCMGR}" >&2

# Verify scenario directory exists
if [[ ! -d "${SCENARIO_DIR}" ]]; then
  echo "[fail] Scenario directory not found: ${SCENARIO_DIR}" >&2
  exit 1
fi

# Run system docmgr scenario suite
echo "" >&2
echo "========================================" >&2
echo "Running scenario suite with SYSTEM docmgr" >&2
echo "========================================" >&2
echo "" >&2

SYSTEM_DB="${SYSTEM_ROOT}/.scenario-run.db"
SYSTEM_LOG_DIR="${SYSTEM_ROOT}/.logs"
mkdir -p "${SYSTEM_LOG_DIR}"

cd "${SCENARIO_DIR}"
DOCMGR_PATH="${SYSTEM_DOCMGR}" SCENARIOLOG_PATH="${SCENARIOLOG_PATH}" bash ./run-all.sh "${SYSTEM_ROOT}" || {
  SYSTEM_EXIT_CODE=$?
  echo "[warn] System docmgr scenario suite exited with code ${SYSTEM_EXIT_CODE}" >&2
}

SYSTEM_RUN_ID=""
if [[ -f "${SYSTEM_DB}" ]]; then
  SYSTEM_RUN_ID=$(sqlite3 "${SYSTEM_DB}" "SELECT run_id FROM scenario_runs ORDER BY started_at DESC LIMIT 1;" 2>/dev/null || echo "")
fi

# Run local docmgr scenario suite
echo "" >&2
echo "========================================" >&2
echo "Running scenario suite with LOCAL docmgr" >&2
echo "========================================" >&2
echo "" >&2

LOCAL_DB="${LOCAL_ROOT}/.scenario-run.db"
LOCAL_LOG_DIR="${LOCAL_ROOT}/.logs"
mkdir -p "${LOCAL_LOG_DIR}"

cd "${SCENARIO_DIR}"
DOCMGR_PATH="${LOCAL_DOCMGR}" SCENARIOLOG_PATH="${SCENARIOLOG_PATH}" bash ./run-all.sh "${LOCAL_ROOT}" || {
  LOCAL_EXIT_CODE=$?
  echo "[warn] Local docmgr scenario suite exited with code ${LOCAL_EXIT_CODE}" >&2
}

LOCAL_RUN_ID=""
if [[ -f "${LOCAL_DB}" ]]; then
  LOCAL_RUN_ID=$(sqlite3 "${LOCAL_DB}" "SELECT run_id FROM scenario_runs ORDER BY started_at DESC LIMIT 1;" 2>/dev/null || echo "")
fi

# Print summary and comparison instructions
echo "" >&2
echo "========================================" >&2
echo "Comparison Summary" >&2
echo "========================================" >&2
echo "" >&2

echo "System docmgr run:" >&2
echo "  Root: ${SYSTEM_ROOT}" >&2
echo "  Database: ${SYSTEM_DB}" >&2
echo "  Run ID: ${SYSTEM_RUN_ID:-<not found>}" >&2
echo "" >&2

echo "Local docmgr run:" >&2
echo "  Root: ${LOCAL_ROOT}" >&2
echo "  Database: ${LOCAL_DB}" >&2
echo "  Run ID: ${LOCAL_RUN_ID:-<not found>}" >&2
echo "" >&2

echo "========================================" >&2
echo "Query Commands" >&2
echo "========================================" >&2
echo "" >&2

cat <<EOF
# System docmgr summary
${SCENARIOLOG_PATH} summary --db "${SYSTEM_DB}" --output table

# Local docmgr summary
${SCENARIOLOG_PATH} summary --db "${LOCAL_DB}" --output table

# System docmgr failures
${SCENARIOLOG_PATH} failures --db "${SYSTEM_DB}" --output table

# Local docmgr failures
${SCENARIOLOG_PATH} failures --db "${LOCAL_DB}" --output table

# System docmgr timings (top 10 slowest steps)
${SCENARIOLOG_PATH} timings --db "${SYSTEM_DB}" --top 10 --output table

# Local docmgr timings (top 10 slowest steps)
${SCENARIOLOG_PATH} timings --db "${LOCAL_DB}" --top 10 --output table

# Search system docmgr logs for errors/warnings
${SCENARIOLOG_PATH} search --db "${SYSTEM_DB}" --run-id "${SYSTEM_RUN_ID}" --query "error OR warning OR fail" --limit 20 --output table

# Search local docmgr logs for errors/warnings
${SCENARIOLOG_PATH} search --db "${LOCAL_DB}" --run-id "${LOCAL_RUN_ID}" --query "error OR warning OR fail" --limit 20 --output table

# Compare step exit codes (if both runs completed)
sqlite3 <<SQL
.mode column
.headers on
SELECT 
  s1.step_num,
  s1.name,
  s1.exit_code as system_exit,
  s2.exit_code as local_exit,
  CASE 
    WHEN s1.exit_code = s2.exit_code THEN 'match'
    ELSE 'DIFFERENT'
  END as status
FROM scenario_steps s1
JOIN scenario_steps s2 ON s1.step_num = s2.step_num AND s1.name = s2.name
WHERE s1.run_id = '${SYSTEM_RUN_ID}' AND s2.run_id = '${LOCAL_RUN_ID}'
ORDER BY s1.step_num;
SQL

# Compare step durations
sqlite3 <<SQL
.mode column
.headers on
SELECT 
  s1.step_num,
  s1.name,
  ROUND((julianday(s1.finished_at) - julianday(s1.started_at)) * 86400, 2) as system_sec,
  ROUND((julianday(s2.finished_at) - julianday(s2.started_at)) * 86400, 2) as local_sec,
  ROUND(
    ((julianday(s2.finished_at) - julianday(s2.started_at)) - 
     (julianday(s1.finished_at) - julianday(s1.started_at))) * 86400, 2
  ) as diff_sec
FROM scenario_steps s1
JOIN scenario_steps s2 ON s1.step_num = s2.step_num AND s1.name = s2.name
WHERE s1.run_id = '${SYSTEM_RUN_ID}' AND s2.run_id = '${LOCAL_RUN_ID}'
ORDER BY ABS(diff_sec) DESC;
SQL

# View artifacts (captured stdout/stderr) for a specific step
# Replace STEP_NUM with the step number you want to inspect
${SCENARIOLOG_PATH} artifacts --db "${SYSTEM_DB}" --run-id "${SYSTEM_RUN_ID}" --step-num STEP_NUM --output table
${SCENARIOLOG_PATH} artifacts --db "${LOCAL_DB}" --run-id "${LOCAL_RUN_ID}" --step-num STEP_NUM --output table

EOF

echo "" >&2
echo "========================================" >&2
echo "Manual Comparison" >&2
echo "========================================" >&2
echo "" >&2

cat <<EOF
# Compare scenario outputs side-by-side
diff -r "${SYSTEM_ROOT}/acme-chat-app" "${LOCAL_ROOT}/acme-chat-app" || true

# Compare specific command outputs (example: search results)
# System docmgr search output:
DOCMGR_PATH="${SYSTEM_DOCMGR}" docmgr doc search --root "${SYSTEM_ROOT}/acme-chat-app/ttmp" --query "chat" --output json > /tmp/system-search.json

# Local docmgr search output:
DOCMGR_PATH="${LOCAL_DOCMGR}" docmgr doc search --root "${LOCAL_ROOT}/acme-chat-app/ttmp" --query "chat" --output json > /tmp/local-search.json

# Diff the JSON outputs:
diff /tmp/system-search.json /tmp/local-search.json || true

EOF

echo "[ok] Comparison complete. Use the query commands above to analyze differences." >&2

