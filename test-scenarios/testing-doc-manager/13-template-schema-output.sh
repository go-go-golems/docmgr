#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
cd "${REPO}"

DOCMGR="${DOCMGR_PATH:-docmgr}"

# Test that --print-template-schema outputs ONLY schema (no human-readable output)
# This verifies the early return pattern works correctly

echo "[test] Testing --print-template-schema for list docs"
OUTPUT=$(${DOCMGR} list docs --print-template-schema --schema-format yaml 2>&1)
if echo "$OUTPUT" | grep -q "Docs root:"; then
    echo "ERROR: list docs --print-template-schema should not output 'Docs root:'"
    exit 1
fi
if echo "$OUTPUT" | grep -q "## Documents"; then
    echo "ERROR: list docs --print-template-schema should not output '## Documents'"
    exit 1
fi
if ! echo "$OUTPUT" | grep -q "properties:"; then
    echo "ERROR: list docs --print-template-schema should output schema with 'properties:'"
    exit 1
fi
echo "[ok] list docs --print-template-schema outputs only schema"

echo "[test] Testing --print-template-schema for list tickets"
OUTPUT=$(${DOCMGR} list tickets --print-template-schema --schema-format yaml 2>&1)
if echo "$OUTPUT" | grep -q "Docs root:"; then
    echo "ERROR: list tickets --print-template-schema should not output 'Docs root:'"
    exit 1
fi
if echo "$OUTPUT" | grep -q "## Tickets"; then
    echo "ERROR: list tickets --print-template-schema should not output '## Tickets'"
    exit 1
fi
if ! echo "$OUTPUT" | grep -q "properties:"; then
    echo "ERROR: list tickets --print-template-schema should output schema with 'properties:'"
    exit 1
fi
echo "[ok] list tickets --print-template-schema outputs only schema"

echo "[test] Testing --print-template-schema for doctor"
OUTPUT=$(${DOCMGR} doctor --print-template-schema --schema-format yaml 2>&1)
if echo "$OUTPUT" | grep -q "findings_total"; then
    # This is OK - it's part of the schema
    :
fi
if echo "$OUTPUT" | grep -q "All checks passed"; then
    echo "ERROR: doctor --print-template-schema should not output 'All checks passed'"
    exit 1
fi
if ! echo "$OUTPUT" | grep -q "properties:"; then
    echo "ERROR: doctor --print-template-schema should output schema with 'properties:'"
    exit 1
fi
echo "[ok] doctor --print-template-schema outputs only schema"

echo "[test] Testing --print-template-schema for status"
OUTPUT=$(${DOCMGR} status --print-template-schema --schema-format yaml 2>&1)
if echo "$OUTPUT" | grep -q "Docs root:"; then
    echo "ERROR: status --print-template-schema should not output 'Docs root:'"
    exit 1
fi
if echo "$OUTPUT" | grep -q "Total tickets:"; then
    echo "ERROR: status --print-template-schema should not output 'Total tickets:'"
    exit 1
fi
if ! echo "$OUTPUT" | grep -q "properties:"; then
    echo "ERROR: status --print-template-schema should output schema with 'properties:'"
    exit 1
fi
echo "[ok] status --print-template-schema outputs only schema"

echo "[test] Testing --print-template-schema for tasks list"
OUTPUT=$(${DOCMGR} tasks list --print-template-schema --schema-format yaml 2>&1)
if echo "$OUTPUT" | grep -q "Total tasks:"; then
    echo "ERROR: tasks list --print-template-schema should not output 'Total tasks:'"
    exit 1
fi
if ! echo "$OUTPUT" | grep -q "properties:"; then
    echo "ERROR: tasks list --print-template-schema should output schema with 'properties:'"
    exit 1
fi
echo "[ok] tasks list --print-template-schema outputs only schema"

echo "[test] Testing --print-template-schema for search"
OUTPUT=$(${DOCMGR} doc search --query "test" --print-template-schema --schema-format yaml 2>&1)
if echo "$OUTPUT" | grep -q "::"; then
    echo "ERROR: doc search --print-template-schema should not output search results with '::'"
    exit 1
fi
if ! echo "$OUTPUT" | grep -q "properties:"; then
    echo "ERROR: doc search --print-template-schema should output schema with 'properties:'"
    exit 1
fi
echo "[ok] doc search --print-template-schema outputs only schema"

echo "[test] Testing --print-template-schema for vocab list"
OUTPUT=$(${DOCMGR} vocab list --print-template-schema --schema-format yaml 2>&1)
if echo "$OUTPUT" | grep -q "category:"; then
    # Check if it's schema format (properties.category) vs human output (category: slug)
    if ! echo "$OUTPUT" | grep -q "properties:"; then
        echo "ERROR: vocab list --print-template-schema should output schema format"
        exit 1
    fi
fi
if ! echo "$OUTPUT" | grep -q "properties:"; then
    echo "ERROR: vocab list --print-template-schema should output schema with 'properties:'"
    exit 1
fi
echo "[ok] vocab list --print-template-schema outputs only schema"

echo "[test] Testing --print-template-schema for guidelines"
OUTPUT=$(${DOCMGR} doc guidelines --doc-type design-doc --print-template-schema --schema-format yaml 2>&1)
if echo "$OUTPUT" | grep -q "## Required Elements"; then
    echo "ERROR: doc guidelines --print-template-schema should not output guideline content"
    exit 1
fi
if ! echo "$OUTPUT" | grep -q "properties:"; then
    echo "ERROR: doc guidelines --print-template-schema should output schema with 'properties:'"
    exit 1
fi
echo "[ok] doc guidelines --print-template-schema outputs only schema"

echo "[ok] All --print-template-schema tests passed"

