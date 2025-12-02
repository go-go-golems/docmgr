#!/usr/bin/env bash
set -euo pipefail

# Diagnostics smoke: exercises docmgr binary against a mock repo to ensure
# taxonomy/rule wiring fires for vocabulary, related files, listing skips,
# workspace missing_index/stale, and frontmatter/template parse errors.
#
# Usage: ./15-diagnostics-smoke.sh [/tmp/docmgr-scenario]

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
DOCMGR="${DOCMGR_PATH:-docmgr}"
DIAG_JSON="${ROOT_DIR}/diagnostics-output.json"

if [[ ! -d "${REPO}" ]]; then
  echo "Repository not found at ${REPO}. Run 01-create-mock-codebase.sh and 02-init-ticket.sh first." >&2
  exit 1
fi

cd "${REPO}"

# Find the MEN-4242 ticket index (created by earlier scripts)
INDEX_MD="$(find ttmp -maxdepth 5 -type f -name index.md -path '*MEN-4242--*' | head -n1 || true)"
if [[ -z "${INDEX_MD}" ]]; then
  echo "Could not locate MEN-4242 index.md under ttmp/. Ensure earlier scenario steps ran." >&2
  exit 1
fi

echo "==> Using doc: ${INDEX_MD}"

# 1) Introduce an unknown topic to trigger vocabulary diagnostics (warning-level)
${DOCMGR} meta update --doc "${INDEX_MD}" --field Topics --value "chat,unknown-taxo"

# 2) Add a missing related file entry to exercise related-file handling
${DOCMGR} doc relate --doc "${INDEX_MD}" \
  --file-note "missing/path.go:expected missing file for diagnostics smoke"

# 3) Create a broken frontmatter doc to trigger YAML parse error (listing skip + doctor invalid frontmatter)
BROKEN_DOC="ttmp/2025/12/01/MEN-4242--normalize-chat-api-paths-and-websocket-lifecycle/reference/zz-broken.md"
cat > "${BROKEN_DOC}" <<'EOF'
---
Title: Broken FM
DocType: reference
Ticket: MEN-4242
Topics: [chat
---
This body will be ignored.
EOF

# 4) Create a stale doc by setting LastUpdated far in the past
sed -i 's/^LastUpdated:.*/LastUpdated: 2020-01-01T00:00:00Z/' "${INDEX_MD}"

# 5) Create a template with a parse error (under ttmp/templates)
BAD_TEMPLATE_REL="templates/bad-template.templ"
BAD_TEMPLATE_ABS="ttmp/${BAD_TEMPLATE_REL}"
mkdir -p "$(dirname "${BAD_TEMPLATE_ABS}")"
cat > "${BAD_TEMPLATE_ABS}" <<'EOF'
{{ define "bad" }}
Value: {{ .Missing
{{ end }}
EOF

# Run doctor to surface warnings (do not fail the run)
${DOCMGR} doctor \
  --ignore-dir _templates --ignore-dir _guidelines \
  --stale-after 30 \
  --fail-on none \
  --diagnostics-json "${DIAG_JSON}"

if [[ ! -s "${DIAG_JSON}" ]]; then
  echo "Diagnostics JSON not generated at ${DIAG_JSON}" >&2
  exit 1
fi
echo "Diagnostics JSON written to ${DIAG_JSON}"

# Run list docs to surface listing skip taxonomy (broken frontmatter)
${DOCMGR} list docs --ticket MEN-4242 || true

# Run template validate to trigger template parse taxonomy
${DOCMGR} template validate --path "${BAD_TEMPLATE_REL}" || true
