#!/usr/bin/env bash
set -euo pipefail

# Helper script to compare scenariolog results from two docmgr runs
#
# Usage:
#   ./compare-results.sh <system-db> <local-db>
#
# Example:
#   ./compare-results.sh /tmp/docmgr-system/.scenario-run.db /tmp/docmgr-local/.scenario-run.db

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <system-db> <local-db>" >&2
  echo "" >&2
  echo "Example:" >&2
  echo "  $0 /tmp/docmgr-system/.scenario-run.db /tmp/docmgr-local/.scenario-run.db" >&2
  exit 1
fi

SYSTEM_DB="$1"
LOCAL_DB="$2"

if [[ ! -f "${SYSTEM_DB}" ]]; then
  echo "[fail] System database not found: ${SYSTEM_DB}" >&2
  exit 1
fi

if [[ ! -f "${LOCAL_DB}" ]]; then
  echo "[fail] Local database not found: ${LOCAL_DB}" >&2
  exit 1
fi

SYSTEM_RUN_ID=$(sqlite3 "${SYSTEM_DB}" "SELECT run_id FROM scenario_runs ORDER BY started_at DESC LIMIT 1;" 2>/dev/null || echo "")
LOCAL_RUN_ID=$(sqlite3 "${LOCAL_DB}" "SELECT run_id FROM scenario_runs ORDER BY started_at DESC LIMIT 1;" 2>/dev/null || echo "")

if [[ -z "${SYSTEM_RUN_ID}" ]]; then
  echo "[fail] No runs found in system database" >&2
  exit 1
fi

if [[ -z "${LOCAL_RUN_ID}" ]]; then
  echo "[fail] No runs found in local database" >&2
  exit 1
fi

echo "System run ID: ${SYSTEM_RUN_ID}"
echo "Local run ID: ${LOCAL_RUN_ID}"
echo ""

echo "========================================"
echo "Step Exit Code Comparison"
echo "========================================"
sqlite3 -header -column "${SYSTEM_DB}" <<SQL
ATTACH DATABASE '${LOCAL_DB}' AS local_db;
SELECT 
  s1.step_num,
  s1.step_name,
  s1.exit_code as system_exit,
  s2.exit_code as local_exit,
  CASE 
    WHEN s1.exit_code = s2.exit_code THEN 'match'
    ELSE 'DIFFERENT'
  END as status
FROM steps s1
JOIN local_db.steps s2 ON s1.step_num = s2.step_num AND s1.step_name = s2.step_name
WHERE s1.run_id = '${SYSTEM_RUN_ID}' AND s2.run_id = '${LOCAL_RUN_ID}'
ORDER BY s1.step_num;
DETACH DATABASE local_db;
SQL

echo ""
echo "========================================"
echo "Step Duration Comparison (sorted by difference)"
echo "========================================"
sqlite3 -header -column "${SYSTEM_DB}" <<SQL
ATTACH DATABASE '${LOCAL_DB}' AS local_db;
SELECT 
  s1.step_num,
  s1.step_name,
  ROUND((julianday(s1.completed_at) - julianday(s1.started_at)) * 86400, 2) as system_sec,
  ROUND((julianday(s2.completed_at) - julianday(s2.started_at)) * 86400, 2) as local_sec,
  ROUND(
    ((julianday(s2.completed_at) - julianday(s2.started_at)) - 
     (julianday(s1.completed_at) - julianday(s1.started_at))) * 86400, 2
  ) as diff_sec
FROM steps s1
JOIN local_db.steps s2 ON s1.step_num = s2.step_num AND s1.step_name = s2.step_name
WHERE s1.run_id = '${SYSTEM_RUN_ID}' AND s2.run_id = '${LOCAL_RUN_ID}'
ORDER BY ABS(
  (julianday(s2.completed_at) - julianday(s2.started_at)) - 
  (julianday(s1.completed_at) - julianday(s1.started_at))
) DESC;
DETACH DATABASE local_db;
SQL

echo ""
echo "========================================"
echo "Steps with Different Exit Codes"
echo "========================================"
sqlite3 -header -column "${SYSTEM_DB}" <<SQL
ATTACH DATABASE '${LOCAL_DB}' AS local_db;
SELECT 
  s1.step_num,
  s1.step_name,
  s1.exit_code as system_exit,
  s2.exit_code as local_exit
FROM steps s1
JOIN local_db.steps s2 ON s1.step_num = s2.step_num AND s1.step_name = s2.step_name
WHERE s1.run_id = '${SYSTEM_RUN_ID}' 
  AND s2.run_id = '${LOCAL_RUN_ID}'
  AND s1.exit_code != s2.exit_code
ORDER BY s1.step_num;
DETACH DATABASE local_db;
SQL

echo ""
echo "========================================"
echo "Total Run Duration"
echo "========================================"
sqlite3 -header -column "${SYSTEM_DB}" <<SQL
ATTACH DATABASE '${LOCAL_DB}' AS local_db;
SELECT 
  'system' as version,
  ROUND((julianday(completed_at) - julianday(started_at)) * 86400, 2) as duration_sec
FROM scenario_runs
WHERE run_id = '${SYSTEM_RUN_ID}'
UNION ALL
SELECT 
  'local' as version,
  ROUND((julianday(completed_at) - julianday(started_at)) * 86400, 2) as duration_sec
FROM local_db.scenario_runs
WHERE run_id = '${LOCAL_RUN_ID}';
DETACH DATABASE local_db;
SQL

