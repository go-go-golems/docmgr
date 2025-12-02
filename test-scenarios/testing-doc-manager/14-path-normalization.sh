#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

TICKET_DIR="$(find ttmp -maxdepth 5 -type d -name 'MEN-4242--*' | head -n 1)"
if [[ -z "${TICKET_DIR}" ]]; then
  echo "Ticket directory for MEN-4242 not found under ttmp/" >&2
  exit 1
fi

DOC_PATH="${TICKET_DIR}/index.md"
DOC_DIR="$(cd "$(dirname "${DOC_PATH}")" && pwd)"
BACKEND_FILE="$(pwd)/backend/chat/api/register.go"
WS_FILE="$(pwd)/backend/chat/ws/manager.go"
TTMP_ROOT="$(pwd)/ttmp"

rel_from_doc="$(python3 - "${BACKEND_FILE}" "${DOC_DIR}" <<'PY'
import os, sys
target = os.path.abspath(sys.argv[1])
base = os.path.abspath(sys.argv[2])
print(os.path.relpath(target, base))
PY
)"
rel_from_doc="${rel_from_doc//\\//}"

rel_from_ttmp="$(python3 - "${BACKEND_FILE}" "${TTMP_ROOT}" <<'PY'
import os, sys
target = os.path.abspath(sys.argv[1])
base = os.path.abspath(sys.argv[2])
print(os.path.relpath(target, base))
PY
)"
rel_from_ttmp="${rel_from_ttmp//\\//}"

file_name="$(basename "${BACKEND_FILE}")"

# Relate the same files using different path forms (doc-relative, ttmp-relative, absolute)
"${DOCMGR}" doc relate --doc "${DOC_PATH}" \
  --file-note "${rel_from_doc}:Doc-relative path reference" \
  --file-note "${rel_from_ttmp}:ttmp-relative path reference" \
  --file-note "${WS_FILE}:Absolute path reference"

# Verify search matches regardless of how the --file flag is specified
"${DOCMGR}" doc search --ticket MEN-4242 --file "${rel_from_doc}"
"${DOCMGR}" doc search --ticket MEN-4242 --file "${rel_from_ttmp}"
"${DOCMGR}" doc search --ticket MEN-4242 --file "${WS_FILE}"
"${DOCMGR}" doc search --ticket MEN-4242 --file "${file_name}"

