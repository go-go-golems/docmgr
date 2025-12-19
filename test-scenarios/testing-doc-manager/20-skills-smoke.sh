#!/usr/bin/env bash
set -euo pipefail

# Skills smoke: exercises docmgr skill commands against a mock repo to ensure
# skill list/show work correctly with filtering (--ticket, --topics, --file, --dir).
#
# Usage: ./20-skills-smoke.sh [/tmp/docmgr-scenario]

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
REPO="${ROOT_DIR}/acme-chat-app"
DOCMGR="${DOCMGR_PATH:-docmgr}"
DOCS_ROOT="${REPO}/ttmp"

if [[ ! -d "${REPO}" ]]; then
  echo "Repository not found at ${REPO}. Run 01-create-mock-codebase.sh and 02-init-ticket.sh first." >&2
  exit 1
fi

cd "${REPO}"

# Find the MEN-4242 ticket directory (created by earlier scripts)
TICKET_DIR="$(find "${DOCS_ROOT}" -maxdepth 4 -type d -name '*MEN-4242--*' | head -n1 || true)"
if [[ -z "${TICKET_DIR}" ]]; then
  echo "Could not locate MEN-4242 ticket directory under ${DOCS_ROOT}. Ensure earlier scenario steps ran." >&2
  exit 1
fi

echo "==> Using ticket directory: ${TICKET_DIR}"

# Create skills directory in ticket
SKILLS_DIR="${TICKET_DIR}/skills"
mkdir -p "${SKILLS_DIR}"

# Create skill 1: API Design skill
cat > "${SKILLS_DIR}/01-api-design.md" <<'EOF'
---
Title: "Skill: API Design"
Ticket: MEN-4242
DocType: skill
Status: active
Topics: [backend, api]
WhatFor: "Designing RESTful APIs with proper error handling and versioning"
WhenToUse: "Use this skill when designing new API endpoints or refactoring existing ones"
RelatedFiles:
  - Path: backend/chat/api/register.go
    Note: "Example API endpoint implementation"
  - Path: backend/chat/api/handlers.go
    Note: "API handler patterns"
Intent: long-term
Owners: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-19T12:00:00Z
---

# Skill: API Design

## Overview

This skill covers designing RESTful APIs with proper error handling, versioning, and documentation.

## What This Skill Is For

Designing RESTful APIs that are maintainable, well-documented, and follow best practices.

## When To Use This Skill

Use this skill when:
- Designing new API endpoints
- Refactoring existing APIs
- Establishing API standards for a project
EOF

# Create skill 2: WebSocket Management skill
cat > "${SKILLS_DIR}/02-websocket-management.md" <<'EOF'
---
Title: "Skill: WebSocket Management"
Ticket: MEN-4242
DocType: skill
Status: active
Topics: [websocket, backend]
WhatFor: "Managing WebSocket connections, lifecycle, and event handling"
WhenToUse: "Use this skill when implementing real-time features or WebSocket-based communication"
RelatedFiles:
  - Path: backend/chat/ws/manager.go
    Note: "WebSocket connection manager"
  - Path: backend/chat/ws/handlers.go
    Note: "WebSocket event handlers"
Intent: long-term
Owners: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-19T12:00:00Z
---

# Skill: WebSocket Management

## Overview

This skill covers managing WebSocket connections, handling lifecycle events, and implementing real-time features.

## What This Skill Is For

Building robust WebSocket-based real-time features with proper connection management.

## When To Use This Skill

Use this skill when:
- Implementing real-time features
- Building WebSocket-based communication
- Managing connection lifecycle
EOF

# Create workspace-level skill
WORKSPACE_SKILLS_DIR="${DOCS_ROOT}/skills"
mkdir -p "${WORKSPACE_SKILLS_DIR}"

cat > "${WORKSPACE_SKILLS_DIR}/workspace-testing.md" <<'EOF'
---
Title: "Skill: Workspace Testing"
DocType: skill
Status: active
Topics: [testing, tooling]
WhatFor: "Testing workspace configurations and scenarios"
WhenToUse: "Use this skill when setting up test scenarios or validating workspace configurations"
RelatedFiles:
  - Path: test-scenarios/testing-doc-manager/run-all.sh
    Note: "Test scenario runner"
Intent: long-term
Owners: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-19T12:00:00Z
---

# Skill: Workspace Testing

## Overview

This skill covers testing workspace configurations and running test scenarios.

## What This Skill Is For

Ensuring workspace configurations work correctly and test scenarios validate functionality.

## When To Use This Skill

Use this skill when:
- Setting up test scenarios
- Validating workspace configurations
- Debugging workspace issues
EOF

echo "==> Created skill documents"

# Assert: seeded vocabulary includes 'skill' docType
echo "==> Check: vocabulary includes docType 'skill'"
if ! ${DOCMGR} vocab list --category docTypes --root "${DOCS_ROOT}" | grep -q "docTypes: skill"; then
  echo "[fail] vocabulary is missing docType 'skill' (expected init --seed-vocabulary to include it)" >&2
  exit 1
fi

# Test 1: List all skills
echo ""
echo "==> Test 1: List all skills"
OUT_1="$(${DOCMGR} skill list --root "${DOCS_ROOT}")"
printf '%s\n' "${OUT_1}"
printf '%s\n' "${OUT_1}" | grep -q "Skill: Skill: API Design"
printf '%s\n' "${OUT_1}" | grep -q "Skill: Skill: WebSocket Management"
printf '%s\n' "${OUT_1}" | grep -q "Skill: Skill: Workspace Testing"

# Test 2: List skills for ticket
echo ""
echo "==> Test 2: List skills for ticket MEN-4242"
OUT_2="$(${DOCMGR} skill list --ticket MEN-4242 --root "${DOCS_ROOT}")"
printf '%s\n' "${OUT_2}"
printf '%s\n' "${OUT_2}" | grep -q "Skill: Skill: API Design"
printf '%s\n' "${OUT_2}" | grep -q "Skill: Skill: WebSocket Management"

# Test 3: List skills by topic
echo ""
echo "==> Test 3: List skills by topic backend"
OUT_3="$(${DOCMGR} skill list --topics backend --root "${DOCS_ROOT}")"
printf '%s\n' "${OUT_3}"
printf '%s\n' "${OUT_3}" | grep -q "Skill: Skill: API Design"
printf '%s\n' "${OUT_3}" | grep -q "Skill: Skill: WebSocket Management"

# Test 4: List skills by multiple topics
echo ""
echo "==> Test 4: List skills by topics backend,websocket"
OUT_4="$(${DOCMGR} skill list --topics backend,websocket --root "${DOCS_ROOT}")"
printf '%s\n' "${OUT_4}"
printf '%s\n' "${OUT_4}" | grep -q "Skill: Skill: API Design"
printf '%s\n' "${OUT_4}" | grep -q "Skill: Skill: WebSocket Management"

# Test 5: List skills by file (reverse lookup)
echo ""
echo "==> Test 5: List skills related to file backend/chat/api/register.go"
OUT_5="$(${DOCMGR} skill list --file backend/chat/api/register.go --root "${DOCS_ROOT}")"
printf '%s\n' "${OUT_5}"
printf '%s\n' "${OUT_5}" | grep -q "Skill: Skill: API Design"

# Test 6: List skills by directory
echo ""
echo "==> Test 6: List skills related to directory backend/chat/api/"
OUT_6="$(${DOCMGR} skill list --dir backend/chat/api/ --root "${DOCS_ROOT}")"
printf '%s\n' "${OUT_6}"
printf '%s\n' "${OUT_6}" | grep -q "Skill: Skill: API Design"

# Test 7: List skills with structured output
echo ""
echo "==> Test 7: List skills with structured output (JSON)"
OUT_7="$(${DOCMGR} skill list --with-glaze-output --output json --root "${DOCS_ROOT}" | head -n 50)"
printf '%s\n' "${OUT_7}"
printf '%s\n' "${OUT_7}" | grep -q "\"skill\": \"Skill: API Design\""

# Test 8: Show skill by exact title
echo ""
echo "==> Test 8: Show skill by exact title"
OUT_8="$(${DOCMGR} skill show --skill "API Design" --root "${DOCS_ROOT}")"
printf '%s\n' "${OUT_8}"
printf '%s\n' "${OUT_8}" | grep -q "Title: Skill: API Design"
printf '%s\n' "${OUT_8}" | grep -q "What this skill is for:"
printf '%s\n' "${OUT_8}" | grep -q "# Skill: API Design"

# Test 9: Show skill by partial match
echo ""
echo "==> Test 9: Show skill by partial match (websocket)"
OUT_9="$(${DOCMGR} skill show --skill websocket --root "${DOCS_ROOT}")"
printf '%s\n' "${OUT_9}"
printf '%s\n' "${OUT_9}" | grep -q "Title: Skill: WebSocket Management"

# Test 10: Show workspace-level skill
echo ""
echo "==> Test 10: Show workspace-level skill"
OUT_10="$(${DOCMGR} skill show --skill "Workspace Testing" --root "${DOCS_ROOT}")"
printf '%s\n' "${OUT_10}"
printf '%s\n' "${OUT_10}" | grep -q "Title: Skill: Workspace Testing"

# Test 11: Verify skill list filters work together
echo ""
echo "==> Test 11: Combined filters (ticket + topic)"
OUT_11="$(${DOCMGR} skill list --ticket MEN-4242 --topics backend --root "${DOCS_ROOT}")"
printf '%s\n' "${OUT_11}"
printf '%s\n' "${OUT_11}" | grep -q "Skill: Skill: API Design"

# Test 12: Verify file filter works with ticket filter
echo ""
echo "==> Test 12: Combined filters (ticket + file)"
OUT_12="$(${DOCMGR} skill list --ticket MEN-4242 --file backend/chat/api/register.go --root "${DOCS_ROOT}")"
printf '%s\n' "${OUT_12}"
printf '%s\n' "${OUT_12}" | grep -q "Skill: Skill: API Design"

echo ""
echo "==> Skills smoke tests completed successfully!"

