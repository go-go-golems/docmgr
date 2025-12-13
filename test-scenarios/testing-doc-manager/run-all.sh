#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
export DOCMGR_PATH="${DOCMGR_PATH:-}"

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

cd "$(dirname "$0")"

bash ./00-reset.sh "${ROOT_DIR}"
bash ./01-create-mock-codebase.sh "${ROOT_DIR}"
bash ./02-init-ticket.sh "${ROOT_DIR}"
bash ./03-create-docs-and-meta.sh "${ROOT_DIR}"
bash ./04-relate-and-doctor.sh "${ROOT_DIR}"
bash ./05-search-scenarios.sh "${ROOT_DIR}"
bash ./06-doctor-advanced.sh "${ROOT_DIR}"
bash ./07-status.sh "${ROOT_DIR}"
bash ./08-configure.sh "${ROOT_DIR}"
bash ./09-relate-from-git.sh "${ROOT_DIR}"
bash ./10-status-warnings.sh "${ROOT_DIR}"
bash ./11-changelog-file-notes.sh "${ROOT_DIR}"
bash ./12-vocab-add-output.sh "${ROOT_DIR}"
bash ./13-template-schema-output.sh "${ROOT_DIR}"
bash ./14-path-normalization.sh "${ROOT_DIR}"
bash ./19-export-sqlite.sh "${ROOT_DIR}"

echo "[ok] Scenario completed at ${ROOT_DIR}/acme-chat-app"
