#!/usr/bin/env bash
set -euo pipefail

# Reproduction script for DOCMGR-BUG-001
# Demonstrates that docmgr doc list --ticket fails in multi-repo setup
# when running from workspace root, but works when running from project1/ttmp

ROOT_DIR="${1:-/tmp/docmgr-bug-repro}"
WORKSPACE="${ROOT_DIR}/workspace"
PROJECT1="${WORKSPACE}/project1"
PROJECT2="${WORKSPACE}/project2"

# Try to find docmgr in various locations
if [ -n "${DOCMGR_PATH:-}" ]; then
    DOCMGR="${DOCMGR_PATH}"
elif command -v docmgr >/dev/null 2>&1; then
    DOCMGR="docmgr"
elif [ -f "./docmgr" ]; then
    DOCMGR="./docmgr"
elif [ -f "../docmgr" ]; then
    DOCMGR="../docmgr"
else
    echo "Error: docmgr not found. Please set DOCMGR_PATH or ensure docmgr is in PATH"
    exit 1
fi

echo "=========================================="
echo "Reproducing DOCMGR-BUG-001: Multi-repo doc list bug"
echo "=========================================="
echo ""

# Clean up previous run
rm -rf "${ROOT_DIR}"
mkdir -p "${WORKSPACE}" "${PROJECT1}" "${PROJECT2}"

echo "[1/7] Setting up multi-repo workspace structure..."

# Create workspace-level .ttmp.yaml pointing to project1/ttmp
cat > "${WORKSPACE}/.ttmp.yaml" <<'EOF'
root: project1/ttmp
EOF

echo "  Created ${WORKSPACE}/.ttmp.yaml with root: project1/ttmp"
echo ""

# Initialize project1/ttmp
echo "[2/7] Initializing project1/ttmp..."
cd "${PROJECT1}"
"${DOCMGR}" init --seed-vocabulary || true
echo "  Initialized ${PROJECT1}/ttmp"
echo ""

# Create a ticket and document in project1/ttmp
echo "[3/7] Creating ticket SOME-TICKET in project1/ttmp..."
cd "${PROJECT1}"
"${DOCMGR}" ticket create-ticket \
  --ticket SOME-TICKET \
  --title "Test Ticket for Bug Reproduction" \
  --topics test,bug

echo "  Created ticket SOME-TICKET"
echo ""

echo "[4/7] Adding a test document to SOME-TICKET..."
cd "${PROJECT1}"
"${DOCMGR}" doc add \
  --ticket SOME-TICKET \
  --doc-type reference \
  --title "Test Document"

echo "  Created test document"
echo ""

# Verify the document exists when running from project1/ttmp
echo "[5/7] Testing: doc list --ticket SOME-TICKET from project1/ttmp (should work)..."
cd "${PROJECT1}/ttmp"
echo "  Current directory: $(pwd)"
echo "  Running: ${DOCMGR} doc list --ticket SOME-TICKET"
echo ""

RESULT_FROM_PROJECT1=$("${DOCMGR}" doc list --ticket SOME-TICKET --with-glaze-output --output json 2>&1 || echo "FAILED")
if echo "${RESULT_FROM_PROJECT1}" | grep -q "SOME-TICKET" || echo "${RESULT_FROM_PROJECT1}" | grep -q "Test Document"; then
    echo "  ✅ SUCCESS: Found documents when running from project1/ttmp"
    echo "${RESULT_FROM_PROJECT1}" | head -20
else
    echo "  ❌ FAILED: No documents found (unexpected!)"
    echo "${RESULT_FROM_PROJECT1}"
fi
echo ""

# Try to list documents from workspace root (should fail due to bug)
echo "[6/7] Testing: doc list --ticket SOME-TICKET from workspace root (should fail due to bug)..."
cd "${WORKSPACE}"
echo "  Current directory: $(pwd)"
echo "  Config file: ${WORKSPACE}/.ttmp.yaml"
echo "  Config contents:"
cat "${WORKSPACE}/.ttmp.yaml" | sed 's/^/    /'
echo ""
echo "  Running: ${DOCMGR} doc list --ticket SOME-TICKET"
echo ""

# Check what root is being resolved
echo "  Debug: Checking resolved root..."
RESOLVED_ROOT=$("${DOCMGR}" config show --with-glaze-output --output json 2>/dev/null | grep -o '"root":"[^"]*"' | cut -d'"' -f4 || echo "unknown")
echo "  Resolved root: ${RESOLVED_ROOT}"
echo ""

# Test both JSON and human output
RESULT_FROM_WORKSPACE_JSON=$("${DOCMGR}" doc list --ticket SOME-TICKET --with-glaze-output --output json 2>&1 || echo "FAILED")
RESULT_FROM_WORKSPACE_HUMAN=$("${DOCMGR}" doc list --ticket SOME-TICKET 2>&1 || echo "FAILED")

# Check JSON output (count occurrences, handle empty result)
JSON_MATCHES=$(echo "${RESULT_FROM_WORKSPACE_JSON}" | grep -o '"ticket":"SOME-TICKET"' 2>/dev/null || true)
if [ -n "${JSON_MATCHES}" ]; then
    JSON_HAS_DOCS=$(echo "${JSON_MATCHES}" | wc -l | xargs)
else
    JSON_HAS_DOCS=0
fi

# Check human output (look for ticket name or "No documents found")
if echo "${RESULT_FROM_WORKSPACE_HUMAN}" | grep -qE "(SOME-TICKET|Test Document)"; then
    HUMAN_HAS_DOCS=1
else
    HUMAN_HAS_DOCS=0
fi
if echo "${RESULT_FROM_WORKSPACE_HUMAN}" | grep -q "No documents found"; then
    HUMAN_NO_DOCS=1
else
    HUMAN_NO_DOCS=0
fi

echo "  JSON output check: ${JSON_HAS_DOCS} document(s) found"
echo "  Human output check: $([ "${HUMAN_HAS_DOCS}" = "1" ] && echo "documents found" || ([ "${HUMAN_NO_DOCS}" = "1" ] && echo "no documents found" || echo "unclear"))"
echo ""

if [ "${JSON_HAS_DOCS}" -gt 0 ] && [ "${HUMAN_HAS_DOCS}" = "1" ]; then
    echo "  ✅ SUCCESS: Found documents in both JSON and human output"
    echo "  (Bug may be fixed, or not reproducing in this scenario)"
    echo ""
    echo "  JSON output:"
    echo "${RESULT_FROM_WORKSPACE_JSON}" | head -15
    echo ""
    echo "  Human output (first 10 lines):"
    echo "${RESULT_FROM_WORKSPACE_HUMAN}" | head -10
elif [ "${JSON_HAS_DOCS}" -gt 0 ] && [ "${HUMAN_NO_DOCS}" = "1" ]; then
    echo "  ⚠️  PARTIAL BUG: JSON output works, but human output shows 'No documents found'"
    echo "  This suggests a bug in the Run method (human output) but not RunIntoGlazeProcessor"
    echo ""
    echo "  JSON output:"
    echo "${RESULT_FROM_WORKSPACE_JSON}" | head -15
    echo ""
    echo "  Human output:"
    echo "${RESULT_FROM_WORKSPACE_HUMAN}"
elif [ "${JSON_HAS_DOCS}" = "0" ] && [ "${HUMAN_NO_DOCS}" = "1" ]; then
    echo "  ❌ BUG CONFIRMED: No documents found in either output format"
    echo "  This is the expected bug behavior."
    echo ""
    echo "  JSON output:"
    echo "${RESULT_FROM_WORKSPACE_JSON}"
    echo ""
    echo "  Human output:"
    echo "${RESULT_FROM_WORKSPACE_HUMAN}"
    echo ""
    echo "  Debug info:"
    echo "    - Resolved root: ${RESOLVED_ROOT}"
    echo "    - Expected: ${PROJECT1}/ttmp"
    if [ "${RESOLVED_ROOT}" != "unknown" ]; then
        echo "    - Root exists: $([ -d "${RESOLVED_ROOT}" ] && echo "yes" || echo "no")"
        echo "    - Root is absolute: $([ "${RESOLVED_ROOT}" = /* ] && echo "yes" || echo "no")"
    fi
else
    echo "  ⚠️  UNCLEAR: Mixed results - need manual inspection"
    echo ""
    echo "  JSON output:"
    echo "${RESULT_FROM_WORKSPACE_JSON}"
    echo ""
    echo "  Human output:"
    echo "${RESULT_FROM_WORKSPACE_HUMAN}"
fi
echo ""

# Show the directory structure for reference
echo "[7/7] Directory structure:"
echo ""
echo "  ${ROOT_DIR}/"
echo "  └── workspace/"
echo "      ├── .ttmp.yaml          (root: project1/ttmp)"
echo "      ├── project1/"
echo "      │   └── ttmp/"
echo "      │       └── YYYY/MM/DD/"
echo "      │           └── SOME-TICKET-.../"
echo "      │               ├── index.md"
echo "      │               └── reference/"
echo "      │                   └── 01-test-document.md"
echo "      └── project2/"
echo "          └── ttmp/"
echo ""

# Summary
echo "=========================================="
echo "Summary"
echo "=========================================="
echo ""
echo "Expected behavior:"
echo "  - From project1/ttmp: ✅ Should find documents"
echo "  - From workspace root: ❌ Should find documents (currently fails in some cases)"
echo ""
echo "Note:"
echo "  If the script shows documents are found from workspace root, the bug"
echo "  may be fixed or may only occur under specific conditions. The script"
echo "  can be used to verify the fix after implementing the solution."
echo ""
echo "Root cause:"
echo "  The doc list command's RunIntoGlazeProcessor method does not"
echo "  ensure the resolved root path is absolute before calling filepath.Walk."
echo ""
echo "Fix:"
echo "  Add absolute path check in RunIntoGlazeProcessor (similar to Run method)"
echo "  See bug report for details."
echo ""

