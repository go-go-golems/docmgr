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

# Keep output short: only print headings + success markers.
# If an assertion fails, dump the captured output for debugging.
dump_output() {
  local label="$1"
  local out="$2"
  echo "" >&2
  echo "[fail] ${label}" >&2
  echo "----- output (first 120 lines) -----" >&2
  printf '%s\n' "${out}" | head -n 120 >&2
  echo "----- output (last 80 lines) -----" >&2
  printf '%s\n' "${out}" | tail -n 80 >&2
}

assert_contains() {
  local label="$1"
  local out="$2"
  local needle="$3"
  if ! printf '%s\n' "${out}" | grep -Fq -- "${needle}"; then
    dump_output "${label} (missing: ${needle})" "${out}"
    exit 1
  fi
}

assert_not_contains() {
  local label="$1"
  local out="$2"
  local needle="$3"
  if printf '%s\n' "${out}" | grep -Fq -- "${needle}"; then
    dump_output "${label} (unexpected: ${needle})" "${out}"
    exit 1
  fi
}

assert_rc_ne_zero() {
  local label="$1"
  local rc="$2"
  local out="$3"
  if [[ "${rc}" -eq 0 ]]; then
    dump_output "${label} (expected non-zero exit code)" "${out}"
    exit 1
  fi
}

# Find the MEN-4242 ticket directory (created by earlier scripts)
TICKET_DIR="$(find "${DOCS_ROOT}" -maxdepth 4 -type d -name '*MEN-4242--*' | head -n1 || true)"
if [[ -z "${TICKET_DIR}" ]]; then
  echo "Could not locate MEN-4242 ticket directory under ${DOCS_ROOT}. Ensure earlier scenario steps ran." >&2
  exit 1
fi

echo "==> Using ticket directory: ${TICKET_DIR}"

# Find the MEN-5678 ticket directory (created by earlier scripts) - we'll close it to test filtering.
TICKET_DIR_5678="$(find "${DOCS_ROOT}" -maxdepth 4 -type d -name '*MEN-5678--*' | head -n1 || true)"
if [[ -z "${TICKET_DIR_5678}" ]]; then
  echo "Could not locate MEN-5678 ticket directory under ${DOCS_ROOT}. Ensure earlier scenario steps ran." >&2
  exit 1
fi

# Create skills directory in ticket
SKILLS_DIR="${TICKET_DIR}/skills"
mkdir -p "${SKILLS_DIR}/api-design" "${SKILLS_DIR}/websocket-management"

# Create skill 1: API Design plan
cat > "${SKILLS_DIR}/api-design/skill.yaml" <<'EOF'
skill:
  name: api-design
  title: API Design
  description: Plan for API design references in the ticket scope.
  what_for: Design RESTful APIs with consistent error handling and versioning.
  when_to_use: Use when designing new API endpoints or refactoring existing ones.
  topics: [backend, api]
  compatibility: Works without external binaries.

sources:
  - type: file
    path: backend/chat/api/register.go
    output: references/backend-chat-register.go

  - type: file
    path: web/src/store/api/chatApi.ts
    output: references/web-chat-api.ts

output:
  skill_dir_name: api-design
EOF

# Create skill 2: WebSocket Management plan
cat > "${SKILLS_DIR}/websocket-management/skill.yaml" <<'EOF'
skill:
  name: websocket-management
  title: WebSocket Management
  description: Plan for WebSocket lifecycle references in the ticket scope.
  what_for: Manage WebSocket connections, lifecycle, and event handling.
  when_to_use: Use when implementing real-time features or WebSocket communication.
  topics: [websocket, backend]
  compatibility: Works without external binaries.

sources:
  - type: file
    path: backend/chat/ws/manager.go
    output: references/backend-chat-ws-manager.go

  - type: file
    path: web/src/ui/chat/ChatPanel.tsx
    output: references/web-chat-panel.tsx

output:
  skill_dir_name: websocket-management
EOF

# Create workspace-level skills
WORKSPACE_SKILLS_DIR="${DOCS_ROOT}/skills"
mkdir -p "${WORKSPACE_SKILLS_DIR}/workspace-testing" "${WORKSPACE_SKILLS_DIR}/api-design"

cat > "${WORKSPACE_SKILLS_DIR}/workspace-testing/skill.yaml" <<'EOF'
skill:
  name: workspace-testing
  title: Workspace Testing
  description: Plan for testing workspace configurations and scenarios.
  what_for: Validate workspace configuration and scenario setup.
  when_to_use: Use when setting up test scenarios or validating workspace configs.
  topics: [testing, tooling]
  compatibility: Works without external binaries.

sources:
  - type: file
    path: README.md
    output: references/readme.md

output:
  skill_dir_name: workspace-testing
EOF

# Create a workspace-level skill that CLASHES by title/slug with the ticket-level one.
# This is used to verify ambiguity handling in `skill show`.
cat > "${WORKSPACE_SKILLS_DIR}/api-design/skill.yaml" <<'EOF'
skill:
  name: api-design
  title: API Design
  description: Workspace-level copy of API design guidance.
  what_for: Provide workspace-level API design references.
  when_to_use: Use when designing APIs (workspace-level copy).
  topics: [backend, api]
  compatibility: Works without external binaries.

sources:
  - type: file
    path: backend/chat/api/register.go
    output: references/backend-chat-register.go

output:
  skill_dir_name: api-design
EOF

echo "==> Created skill plans"

# Create a skill under MEN-5678, then close MEN-5678 (complete) to verify `skill show`
# excludes non-active tickets unless --ticket is provided.
SKILLS_DIR_5678="${TICKET_DIR_5678}/skills"
mkdir -p "${SKILLS_DIR_5678}/closed-ticket-only-skill"
cat > "${SKILLS_DIR_5678}/closed-ticket-only-skill/skill.yaml" <<'EOF'
skill:
  name: closed-ticket-only-skill
  title: Closed Ticket Only Skill
  description: Should be hidden unless --ticket MEN-5678 is passed.
  what_for: Test ticket status filtering with closed tickets.
  when_to_use: Only for testing ticket-status filtering.
  topics: [backend]
  compatibility: Works without external binaries.

sources: []

output:
  skill_dir_name: closed-ticket-only-skill
EOF

# Close MEN-5678.
${DOCMGR} ticket close --ticket MEN-5678 --root "${DOCS_ROOT}" --status complete --changelog-entry "Close ticket for skills filtering smoke test" >/dev/null

# Test 1: List all skills
echo ""
echo "==> Test 1: List all skills"
OUT_1="$(${DOCMGR} skill list --root "${DOCS_ROOT}")"
assert_contains "Test 1" "${OUT_1}" "Skill: API Design"
assert_contains "Test 1" "${OUT_1}" "Skill: WebSocket Management"
assert_contains "Test 1" "${OUT_1}" "Skill: Workspace Testing"
assert_contains "Test 1" "${OUT_1}" "Load: docmgr skill show"
assert_not_contains "Test 1" "${OUT_1}" "MEN-5678"
assert_not_contains "Test 1" "${OUT_1}" "Closed Ticket Only Skill"
echo "[ok] Test 1"

# Test 2: List skills for ticket
echo ""
echo "==> Test 2: List skills for ticket MEN-4242"
OUT_2="$(${DOCMGR} skill list --ticket MEN-4242 --root "${DOCS_ROOT}")"
assert_contains "Test 2" "${OUT_2}" "Skill: API Design"
assert_contains "Test 2" "${OUT_2}" "Skill: WebSocket Management"
assert_contains "Test 2" "${OUT_2}" "Skill: Workspace Testing"
assert_contains "Test 2" "${OUT_2}" "Load: docmgr skill show"
echo "[ok] Test 2"

# Test 3: List skills by topic
echo ""
echo "==> Test 3: List skills by topic backend"
OUT_3="$(${DOCMGR} skill list --topics backend --root "${DOCS_ROOT}")"
assert_contains "Test 3" "${OUT_3}" "Skill: API Design"
assert_contains "Test 3" "${OUT_3}" "Skill: WebSocket Management"
echo "[ok] Test 3"

# Test 4: List skills by multiple topics
echo ""
echo "==> Test 4: List skills by topics backend,websocket"
OUT_4="$(${DOCMGR} skill list --topics backend,websocket --root "${DOCS_ROOT}")"
assert_contains "Test 4" "${OUT_4}" "Skill: API Design"
assert_contains "Test 4" "${OUT_4}" "Skill: WebSocket Management"
echo "[ok] Test 4"

# Test 5: List skills by file (reverse lookup)
echo ""
echo "==> Test 5: List skills related to file backend/chat/api/register.go"
OUT_5="$(${DOCMGR} skill list --file backend/chat/api/register.go --root "${DOCS_ROOT}")"
assert_contains "Test 5" "${OUT_5}" "Skill: API Design"
echo "[ok] Test 5"

# Test 6: List skills by directory
echo ""
echo "==> Test 6: List skills related to directory backend/chat/api/"
OUT_6="$(${DOCMGR} skill list --dir backend/chat/api/ --root "${DOCS_ROOT}")"
assert_contains "Test 6" "${OUT_6}" "Skill: API Design"
echo "[ok] Test 6"

# Test 7: List skills with structured output
echo ""
echo "==> Test 7: List skills with structured output (JSON)"
OUT_7="$(${DOCMGR} skill list --with-glaze-output --output json --root "${DOCS_ROOT}")"
assert_contains "Test 7" "${OUT_7}" "\"skill\": \"API Design\""
assert_contains "Test 7" "${OUT_7}" "\"load_command\": \"docmgr skill show"
echo "[ok] Test 7"

# Test 8: Show skill by exact title
echo ""
echo "==> Test 8: Show skill by exact title"
set +e
OUT_8="$(${DOCMGR} skill show --skill "API Design" --root "${DOCS_ROOT}" 2>&1)"
RC_8=$?
set -e
assert_rc_ne_zero "Test 8" "${RC_8}" "${OUT_8}"
assert_contains "Test 8" "${OUT_8}" "Multiple skills match"
assert_contains "Test 8" "${OUT_8}" "Load: docmgr skill show"
echo "[ok] Test 8"

# Test 8b: Show skill by exact title with --ticket disambiguation (flag-based)
echo ""
echo "==> Test 8b: Show skill by exact title with --ticket (disambiguation)"
OUT_8B="$(${DOCMGR} skill show --skill "WebSocket Management" --ticket MEN-4242 --root "${DOCS_ROOT}")"
assert_contains "Test 8b" "${OUT_8B}" "Title: WebSocket Management"
assert_contains "Test 8b" "${OUT_8B}" "Name: websocket-management"
assert_contains "Test 8b" "${OUT_8B}" "Ticket: MEN-4242"
assert_contains "Test 8b" "${OUT_8B}" "Sources:"
echo "[ok] Test 8b"

# Test 8c: Show skill by positional argument with --ticket (disambiguation)
echo ""
echo "==> Test 8c: Show skill by positional argument with --ticket"
OUT_8C="$(${DOCMGR} skill show "WebSocket Management" --ticket MEN-4242 --root "${DOCS_ROOT}")"
assert_contains "Test 8c" "${OUT_8C}" "Title: WebSocket Management"
assert_contains "Test 8c" "${OUT_8C}" "Name: websocket-management"
assert_contains "Test 8c" "${OUT_8C}" "Ticket: MEN-4242"
assert_contains "Test 8c" "${OUT_8C}" "Sources:"
echo "[ok] Test 8c"

# Test 9: Show skill by partial match
echo ""
echo "==> Test 9: Show skill by partial match (websocket)"
OUT_9="$(${DOCMGR} skill show --skill websocket --root "${DOCS_ROOT}")"
assert_contains "Test 9" "${OUT_9}" "Title: WebSocket Management"
echo "[ok] Test 9"

# Test 10: Show workspace-level skill
echo ""
echo "==> Test 10: Show workspace-level skill"
OUT_10="$(${DOCMGR} skill show --skill "Workspace Testing" --root "${DOCS_ROOT}")"
assert_contains "Test 10" "${OUT_10}" "Title: Workspace Testing"
echo "[ok] Test 10"

# Test 10b: Show by filename/slug (should work, but will be ambiguous for api-design)
echo ""
echo "==> Test 10b: Show by filename/slug (ambiguity case: api-design)"
set +e
OUT_10B="$(${DOCMGR} skill show api-design --root "${DOCS_ROOT}" 2>&1)"
RC_10B=$?
set -e
assert_rc_ne_zero "Test 10b" "${RC_10B}" "${OUT_10B}"
assert_contains "Test 10b" "${OUT_10B}" "Multiple skills match"
echo "[ok] Test 10b"

# Test 10c: Show by explicit path (unambiguous)
echo ""
echo "==> Test 10c: Show by explicit path"
OUT_10C="$(${DOCMGR} skill show "${SKILLS_DIR}/api-design/skill.yaml" --root "${DOCS_ROOT}")"
assert_contains "Test 10c" "${OUT_10C}" "Title: API Design"
assert_contains "Test 10c" "${OUT_10C}" "Ticket: MEN-4242"
echo "[ok] Test 10c"

# Test 10d: Skills from non-active tickets are excluded unless --ticket is provided.
echo ""
echo "==> Test 10d: Exclude non-active ticket skills unless --ticket is provided"
set +e
OUT_10D="$(${DOCMGR} skill show closed-ticket-only-skill --root "${DOCS_ROOT}" 2>&1)"
RC_10D=$?
set -e
assert_rc_ne_zero "Test 10d (no --ticket)" "${RC_10D}" "${OUT_10D}"
assert_contains "Test 10d (no --ticket)" "${OUT_10D}" "no skills found matching"

OUT_10D2="$(${DOCMGR} skill show --ticket MEN-5678 closed-ticket-only-skill --root "${DOCS_ROOT}")"
assert_contains "Test 10d (with --ticket)" "${OUT_10D2}" "Title: Closed Ticket Only Skill"
assert_contains "Test 10d (with --ticket)" "${OUT_10D2}" "Ticket: MEN-5678"
echo "[ok] Test 10d"

# Test 10e: Skills from review tickets are still included by default (they're still in-progress).
echo ""
echo "==> Test 10e: Include review ticket skills by default"
${DOCMGR} meta update --ticket MEN-4242 --field Status --value review --root "${DOCS_ROOT}" >/dev/null
OUT_10E="$(${DOCMGR} skill show --skill websocket --root "${DOCS_ROOT}")"
assert_contains "Test 10e" "${OUT_10E}" "Title: WebSocket Management"
echo "[ok] Test 10e"

# Test 11: Verify skill list filters work together
echo ""
echo "==> Test 11: Combined filters (ticket + topic)"
OUT_11="$(${DOCMGR} skill list --ticket MEN-4242 --topics backend --root "${DOCS_ROOT}")"
assert_contains "Test 11" "${OUT_11}" "Skill: API Design"
echo "[ok] Test 11"

# Test 12: Verify file filter works with ticket filter
echo ""
echo "==> Test 12: Combined filters (ticket + file)"
OUT_12="$(${DOCMGR} skill list --ticket MEN-4242 --file backend/chat/api/register.go --root "${DOCS_ROOT}")"
assert_contains "Test 12" "${OUT_12}" "Skill: API Design"
echo "[ok] Test 12"

# Test 13: Export a ticket skill plan
echo ""
echo "==> Test 13: Export skill plan (websocket-management)"
OUT_13="$(${DOCMGR} skill export websocket-management --ticket MEN-4242 --root "${DOCS_ROOT}" --out "${ROOT_DIR}/exports")"
assert_contains "Test 13" "${OUT_13}" "Exported skill to"
if [[ ! -f "${ROOT_DIR}/exports/websocket-management.skill" ]]; then
  dump_output "Test 13 (missing export file)" "${OUT_13}"
  exit 1
fi
echo "[ok] Test 13"

# Test 14: Import the exported skill into workspace
echo ""
echo "==> Test 14: Import exported skill into workspace"
OUT_14="$(${DOCMGR} skill import "${ROOT_DIR}/exports/websocket-management.skill" --root "${DOCS_ROOT}")"
assert_contains "Test 14" "${OUT_14}" "Imported skill plan"
if [[ ! -f "${DOCS_ROOT}/skills/websocket-management/skill.yaml" ]]; then
  dump_output "Test 14 (missing imported plan)" "${OUT_14}"
  exit 1
fi
echo "[ok] Test 14"

echo ""
echo "==> Skills smoke tests completed successfully!"
