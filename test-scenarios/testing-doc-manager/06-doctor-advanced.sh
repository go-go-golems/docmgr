#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"
TICKET_DIR="ttmp/MEN-4242-normalize-chat-api-paths-and-websocket-lifecycle"
INDEX_MD="${TICKET_DIR}/index.md"

# 1) Introduce issues: unknown topic, missing related file, duplicate index
${DOCMGR} meta update --doc "${INDEX_MD}" --field Topics --value "chat,backend,websocket,nonexistent-topic"
${DOCMGR} meta update --doc "${INDEX_MD}" --field RelatedFiles --value \
"backend/chat/api/register.go,backend/chat/ws/manager.go,web/src/store/api/chatApi.ts,backend/chat/api/does-not-exist.go"

# Create duplicate index inside design-doc/
mkdir -p "${TICKET_DIR}/design-doc"
cat > "${TICKET_DIR}/design-doc/index.md" <<'EOF'
---
Title: Design Section Index
Ticket: MEN-4242
Status: active
Topics: [chat]
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-03T00:00:00Z
---

# Design section index (for testing multiple_index)
EOF

# 2) Run doctor expecting warnings and nonzero exit with --fail-on warning
set +e
${DOCMGR} doctor --ignore-dir _templates --ignore-dir _guidelines --fail-on warning
DOCTOR_RC=$?
set -e
if [ ${DOCTOR_RC} -eq 0 ]; then
  echo "[warn] Expected doctor to fail with warnings, but it passed"
fi

# 3) Use ignore-glob to suppress duplicate index, keeping other warnings
set +e
${DOCMGR} doctor --ignore-dir _templates --ignore-dir _guidelines \
  --ignore-glob "ttmp/*/design-doc/index.md" --fail-on warning
DOCTOR_RC2=$?
set -e
if [ ${DOCTOR_RC2} -eq 0 ]; then
  echo "[warn] Still expected warnings due to unknown topic / missing file"
fi

# 4) Fix metadata and re-run doctor to pass
${DOCMGR} meta update --doc "${INDEX_MD}" --field Topics --value "chat,backend,websocket"
${DOCMGR} meta update --doc "${INDEX_MD}" --field RelatedFiles --value \
"backend/chat/api/register.go,backend/chat/ws/manager.go,web/src/store/api/chatApi.ts"

# Remove test duplicate index
rm -f "${TICKET_DIR}/design-doc/index.md"

${DOCMGR} doctor --ignore-dir _templates --ignore-dir _guidelines --stale-after 30 --fail-on error

echo "[ok] Advanced doctor scenario completed"
