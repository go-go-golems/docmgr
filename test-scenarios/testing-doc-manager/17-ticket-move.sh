#!/usr/bin/env bash
set -euo pipefail

# Scenario 17: Move a ticket directory to the current path template.
#
# Prereqs: run 01-create-mock-codebase.sh, 02-init-ticket.sh, 03-create-docs-and-meta.sh.
# This simulates a legacy flat ticket path and uses `docmgr ticket move` to relocate
# it to the date-based template.

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
DOCMGR="${DOCMGR_PATH:-docmgr}"
TICKET="MEN-5678"

if [[ ! -d "${REPO}" ]]; then
  echo "Repository not found at ${REPO}. Run setup scripts first." >&2
  exit 1
fi

cd "${REPO}"

# Locate current ticket directory
CUR_DIR="$(find ttmp -maxdepth 4 -type d \( -name "${TICKET}" -o -path "*${TICKET}-*" \) | head -n1 || true)"
if [[ -z "${CUR_DIR}" ]]; then
  echo "Ticket directory for ${TICKET} not found. Ensure earlier scripts created it." >&2
  exit 1
fi

# Simulate legacy flat layout by moving the ticket to ttmp/<ticket>-legacy
LEGACY_DIR="ttmp/${TICKET}-legacy"
if [[ -d "${LEGACY_DIR}" ]]; then
  rm -rf "${LEGACY_DIR}"
fi
mv "${CUR_DIR}" "${LEGACY_DIR}"

echo "==> Legacy location: ${LEGACY_DIR}"

# Run ticket move to relocate into current path template
${DOCMGR} ticket move --ticket "${TICKET}" --overwrite

NEW_DIR="$(find ttmp -maxdepth 4 -type d -path "*${TICKET}-*" | head -n1 || true)"
if [[ -z "${NEW_DIR}" ]]; then
  echo "Failed to locate new ticket directory after move" >&2
  exit 1
fi

if [[ -d "${LEGACY_DIR}" ]]; then
  echo "Legacy directory still exists after move: ${LEGACY_DIR}" >&2
  exit 1
fi

if [[ ! -f "${NEW_DIR}/index.md" ]]; then
  echo "index.md missing in new location" >&2
  exit 1
fi

echo "Ticket move succeeded: ${NEW_DIR}"
