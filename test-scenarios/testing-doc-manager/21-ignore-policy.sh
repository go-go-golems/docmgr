#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

TICKET_DIR="$(find ttmp -type d -name 'MEN-4242--*' -print -quit)"
if [[ -z "${TICKET_DIR}" ]]; then
  echo "[fail] MEN-4242 ticket dir not found" >&2
  exit 1
fi

mkdir -p "${TICKET_DIR}/scripts/node_modules/pkg"
cat > "${TICKET_DIR}/scripts/node_modules/pkg/README.md" <<'EOF'
# Package README without docmgr frontmatter

This file should be ignored by built-in node_modules handling.
EOF

mkdir -p "${TICKET_DIR}/scripts/local-cache"
cat > "${TICKET_DIR}/scripts/local-cache/.docmgrignore" <<'EOF'
*.md
EOF
cat > "${TICKET_DIR}/scripts/local-cache/bad.md" <<'EOF'
# Local generated markdown without frontmatter

This file should be ignored by the nested .docmgrignore.
EOF

NODE_DECISION="$(${DOCMGR} ignore explain --root ttmp "${TICKET_DIR}/scripts/node_modules/pkg/README.md" --with-glaze-output --output json)"
echo "${NODE_DECISION}"
if ! grep -q '"ignored": true' <<<"${NODE_DECISION}"; then
  echo "[fail] expected node_modules README to be ignored" >&2
  exit 1
fi

NESTED_DECISION="$(${DOCMGR} ignore explain --root ttmp "${TICKET_DIR}/scripts/local-cache/bad.md" --with-glaze-output --output json)"
echo "${NESTED_DECISION}"
if ! grep -q '"ignored": true' <<<"${NESTED_DECISION}"; then
  echo "[fail] expected nested .docmgrignore markdown to be ignored" >&2
  exit 1
fi

DOCTOR_OUTPUT="$(${DOCMGR} doctor --ticket MEN-4242 --stale-after 30 --fail-on error 2>&1)"
echo "${DOCTOR_OUTPUT}"
if grep -q 'node_modules/pkg/README.md\|scripts/local-cache/bad.md' <<<"${DOCTOR_OUTPUT}"; then
  echo "[fail] ignored markdown appeared in doctor output" >&2
  exit 1
fi

echo "[ok] Ignore policy scenario completed"
