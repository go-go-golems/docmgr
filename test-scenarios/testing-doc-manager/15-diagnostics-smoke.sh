#!/usr/bin/env bash
set -euo pipefail

# Diagnostics smoke: exercises docmgr binary against a mock repo to ensure
# vocabulary and related-file diagnostics remain wired after recent changes.
#
# Usage: ./15-diagnostics-smoke.sh [/tmp/docmgr-scenario]

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
DOCMGR="${DOCMGR_PATH:-docmgr}"

if [[ ! -d "${REPO}" ]]; then
  echo "Repository not found at ${REPO}. Run 01-create-mock-codebase.sh and 02-init-ticket.sh first." >&2
  exit 1
fi

cd "${REPO}"

# Find the MEN-4242 ticket index (created by earlier scripts)
INDEX_MD="$(find ttmp -maxdepth 5 -type f -name index.md -path '*MEN-4242-*' | head -n1 || true)"
if [[ -z "${INDEX_MD}" ]]; then
  echo "Could not locate MEN-4242 index.md under ttmp/. Ensure earlier scenario steps ran." >&2
  exit 1
fi

echo "==> Using doc: ${INDEX_MD}"

# Introduce an unknown topic to trigger vocabulary diagnostics (warning-level)
${DOCMGR} meta update --doc "${INDEX_MD}" --field Topics --value "chat,unknown-taxo"

# Add a missing related file entry to exercise related-file handling
${DOCMGR} doc relate --doc "${INDEX_MD}" \
  --file-note "missing/path.go:expected missing file for diagnostics smoke"

# Run doctor to surface warnings (do not fail the run)
${DOCMGR} doctor \
  --ignore-dir _templates --ignore-dir _guidelines \
  --stale-after 30 \
  --fail-on none
