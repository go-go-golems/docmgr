#!/usr/bin/env bash
set -euo pipefail

# Frontmatter validation smoke: exercise the `docmgr validate frontmatter` verb
# against a known-bad YAML block, then confirm it passes once fixed.
#
# Usage: ./18-validate-frontmatter-smoke.sh [/tmp/docmgr-scenario]

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
DOCMGR="${DOCMGR_PATH:-docmgr}"

if [[ ! -d "${REPO}" ]]; then
  echo "Repository not found at ${REPO}. Run 01-create-mock-codebase.sh and 02-init-ticket.sh first." >&2
  exit 1
fi

cd "${REPO}"

INDEX_MD="$(find ttmp -maxdepth 5 -type f -name index.md -path '*MEN-4242-*' | head -n1 || true)"
if [[ -z "${INDEX_MD}" ]]; then
  echo "Could not locate MEN-4242 index.md under ttmp/. Ensure earlier scenario steps ran." >&2
  exit 1
fi

BROKEN_DOC="ttmp/2025/12/01/MEN-4242-normalize-chat-api-paths-and-websocket-lifecycle/reference/validate-broken.md"
mkdir -p "$(dirname "${BROKEN_DOC}")"
cat > "${BROKEN_DOC}" <<'EOF'
---
Title: Broken validation doc
Ticket: MEN-4242
DocType: reference
Summary: unquoted: colon causes parse error
---
Body is irrelevant for this smoke test.
EOF

echo "==> Expecting validation failure (YAML syntax) ..."
if ${DOCMGR} validate frontmatter --doc "${BROKEN_DOC}"; then
  echo "Expected validate frontmatter to fail on bad YAML, but it succeeded" >&2
  exit 1
else
  echo "Validation failed as expected for broken frontmatter"
fi

cat > "${BROKEN_DOC}" <<'EOF'
---
Title: Fixed validation doc
Ticket: MEN-4242
DocType: reference
Summary: "quoted colon now parses"
---
Body is irrelevant for this smoke test.
EOF

echo "==> Expecting validation success after fix ..."
${DOCMGR} validate frontmatter --doc "${BROKEN_DOC}"
echo "Frontmatter validation smoke completed."
