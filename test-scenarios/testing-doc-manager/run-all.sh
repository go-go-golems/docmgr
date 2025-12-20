#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
export DOCMGR_PATH="${DOCMGR_PATH:-}"
export SCENARIOLOG_PATH="${SCENARIOLOG_PATH:-}"

# NOTE: This scenario suite must run against an explicitly pinned docmgr binary.
# If DOCMGR_PATH is unset, we'd silently fall back to `docmgr` from PATH, which can be an
# older system install and can cause false failures (e.g. missing flags).
#
# To run against system docmgr intentionally, set:
#   DOCMGR_PATH="$(command -v docmgr)"
#
# To run against the repo code (recommended), build a binary and set DOCMGR_PATH:
#   go build -o /tmp/docmgr-local ./cmd/docmgr
#   DOCMGR_PATH=/tmp/docmgr-local bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario
if [[ -z "${DOCMGR_PATH}" ]]; then
  echo "[fail] DOCMGR_PATH is not set. Refusing to run with ambiguous system docmgr from PATH." >&2
  echo "       Set DOCMGR_PATH to a pinned binary (recommended: repo build)." >&2
  exit 2
fi

echo "[info] Using DOCMGR_PATH=${DOCMGR_PATH}" >&2

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}"

if [[ -z "${SCENARIOLOG_PATH}" ]]; then
  # Default to building scenariolog from this repo (self-contained module).
  REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
  SCENARIOLOG_PATH="/tmp/scenariolog-local"
  echo "[info] SCENARIOLOG_PATH is not set; building scenariolog: ${REPO_ROOT}/scenariolog -> ${SCENARIOLOG_PATH}" >&2
  # NOTE: This repo may be used with a top-level go.work (workspace mode), but scenariolog is
  # a nested module with its own go.mod. Force module mode for this build so it doesn't get
  # resolved against the go.work module set.
  GOWORK=off go -C "${REPO_ROOT}/scenariolog" build -tags sqlite_fts5 -o "${SCENARIOLOG_PATH}" ./cmd/scenariolog
fi

echo "[info] Using SCENARIOLOG_PATH=${SCENARIOLOG_PATH}" >&2

# Step 00 is special: it deletes ROOT_DIR, so we can't keep the sqlite DB inside ROOT_DIR yet.
bash ./00-reset.sh "${ROOT_DIR}"

DB="${ROOT_DIR}/.scenario-run.db"
LOG_DIR="${ROOT_DIR}/.logs"
mkdir -p "${LOG_DIR}"

RUN_ID="$("${SCENARIOLOG_PATH}" run start --db "${DB}" --root-dir "${ROOT_DIR}" --suite testing-doc-manager)"
echo "[info] scenariolog run_id=${RUN_ID}" >&2

finalize_run() {
  local exit_code="$1"
  if [[ -n "${RUN_ID:-}" ]]; then
    "${SCENARIOLOG_PATH}" run end --db "${DB}" --run-id "${RUN_ID}" --exit-code "${exit_code}" || true
  fi
  exit "${exit_code}"
}

trap 'finalize_run $?' EXIT

step() {
  local num="$1"
  local name="$2"
  local script="$3"
  shift 3
  "${SCENARIOLOG_PATH}" exec \
    --db "${DB}" \
    --run-id "${RUN_ID}" \
    --root-dir "${ROOT_DIR}" \
    --log-dir ".logs" \
    --step-num "${num}" \
    --name "${name}" \
    --script-path "${script}" \
    -- bash "${script}" "$@"
}

step 1 "01-create-mock-codebase" "./01-create-mock-codebase.sh" "${ROOT_DIR}"
step 2 "02-init-ticket" "./02-init-ticket.sh" "${ROOT_DIR}"
step 3 "03-create-docs-and-meta" "./03-create-docs-and-meta.sh" "${ROOT_DIR}"
step 4 "04-relate-and-doctor" "./04-relate-and-doctor.sh" "${ROOT_DIR}"
step 5 "05-search-scenarios" "./05-search-scenarios.sh" "${ROOT_DIR}"
step 6 "06-doctor-advanced" "./06-doctor-advanced.sh" "${ROOT_DIR}"
step 7 "07-status" "./07-status.sh" "${ROOT_DIR}"
step 8 "08-configure" "./08-configure.sh" "${ROOT_DIR}"
step 9 "09-relate-from-git" "./09-relate-from-git.sh" "${ROOT_DIR}"
step 10 "10-status-warnings" "./10-status-warnings.sh" "${ROOT_DIR}"
step 11 "11-changelog-file-notes" "./11-changelog-file-notes.sh" "${ROOT_DIR}"
step 12 "12-vocab-add-output" "./12-vocab-add-output.sh" "${ROOT_DIR}"
step 13 "13-template-schema-output" "./13-template-schema-output.sh" "${ROOT_DIR}"
step 14 "14-path-normalization" "./14-path-normalization.sh" "${ROOT_DIR}"
step 19 "19-export-sqlite" "./19-export-sqlite.sh" "${ROOT_DIR}"
step 20 "20-skills-smoke" "./20-skills-smoke.sh" "${ROOT_DIR}"

echo "[ok] Scenario completed at ${ROOT_DIR}/acme-chat-app"
