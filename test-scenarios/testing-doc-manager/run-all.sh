#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
export DOCMGR_PATH="${DOCMGR_PATH:-docmgr}"

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

echo "[ok] Scenario completed at ${ROOT_DIR}/acme-chat-app"
