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
BROKEN_DOC_ABS="${REPO}/${BROKEN_DOC}"
mkdir -p "$(dirname "${BROKEN_DOC_ABS}")"
cat > "${BROKEN_DOC_ABS}" <<'EOF'
---
Title: Broken validation doc
Ticket: MEN-4242
DocType: reference
Summary: broken frontmatter delimiters
----
Body is irrelevant for this smoke test (no closing ---).
EOF

echo "==> Expecting validation failure (YAML syntax) ..."
if ${DOCMGR} validate frontmatter --doc "${BROKEN_DOC_ABS}"; then
  echo "Expected validate frontmatter to fail on bad YAML, but it succeeded" >&2
  exit 1
else
  echo "Validation failed as expected for broken frontmatter"
fi

echo "==> Suggest fixes (should show quoting guidance) ..."
if ! ${DOCMGR} validate frontmatter --doc "${BROKEN_DOC_ABS}" --suggest-fixes; then
  echo "Suggest-fixes returned non-zero (expected due to parse error), continuing" >&2
fi

echo "==> Auto-fix (creates .bak and attempts repair) ..."
${DOCMGR} validate frontmatter --doc "${BROKEN_DOC_ABS}" --auto-fix || true

if [[ ! -f "${BROKEN_DOC_ABS}.bak" ]]; then
  echo "Expected backup file ${BROKEN_DOC_ABS}.bak" >&2
  exit 1
fi

cat > "${BROKEN_DOC_ABS}" <<'EOF'
---
Title: Fixed validation doc
Ticket: MEN-4242
DocType: reference
Summary: "quoted colon now parses"
---
Body is irrelevant for this smoke test.
EOF

echo "==> Expecting validation success after fix ..."
${DOCMGR} validate frontmatter --doc "${BROKEN_DOC_ABS}"
echo "Frontmatter validation smoke completed."
