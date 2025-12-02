#!/usr/bin/env bash
set -euo pipefail

# Scenario 16: Move a document between tickets and rewrite the Ticket field.
#
# Prereqs: run 01-create-mock-codebase.sh, 02-init-ticket.sh, 03-create-docs-and-meta.sh.
#
# This script:
#   - Locates a MEN-4242 reference doc
#   - Moves it to MEN-5678 under the same relative path
#   - Rewrites Ticket in frontmatter
#   - Verifies source removal and destination rewrite

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
DOCMGR="${DOCMGR_PATH:-docmgr}"

if [[ ! -d "${REPO}" ]]; then
  echo "Repository not found at ${REPO}. Run setup scripts first." >&2
  exit 1
fi

cd "${REPO}"

SRC_DOC="$(find ttmp -maxdepth 6 -type f -path '*MEN-4242*reference/01-chat-websocket-lifecycle.md' | head -n1 || true)"
if [[ -z "${SRC_DOC}" ]]; then
  echo "Source doc not found (expected MEN-4242 reference). Run earlier scenarios." >&2
  exit 1
fi

DEST_TICKET="MEN-5678"
echo "==> Moving ${SRC_DOC} to ticket ${DEST_TICKET}"

${DOCMGR} doc move \
  --doc "${SRC_DOC}" \
  --dest-ticket "${DEST_TICKET}" \
  --overwrite

DEST_DOC="$(find ttmp -maxdepth 6 -type f -path "*${DEST_TICKET}*/01-chat-websocket-lifecycle.md" | head -n1 || true)"
if [[ -z "${DEST_DOC}" ]]; then
  echo "Destination doc not found after move" >&2
  exit 1
fi

if [[ -e "${SRC_DOC}" ]]; then
  echo "Source doc still exists after move: ${SRC_DOC}" >&2
  exit 1
fi

if ! head -n 10 "${DEST_DOC}" | grep -q "Ticket: ${DEST_TICKET}"; then
  echo "Ticket field not rewritten in destination doc: ${DEST_DOC}" >&2
  exit 1
fi

echo "Doc move succeeded: ${DEST_DOC}"
