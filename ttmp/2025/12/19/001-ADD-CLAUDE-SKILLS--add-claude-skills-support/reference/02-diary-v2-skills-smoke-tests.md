---
Title: 'Diary (v2): Skills smoke tests'
Ticket: 001-ADD-CLAUDE-SKILLS
Status: active
Topics:
    - features
    - skills
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-19T13:34:17.25026493-05:00
---

# Diary (v2): Skills smoke tests

## Goal

Capture the step-by-step work of making the skills smoke tests **reliable and conceptually correct**, including what failed, why it failed, and what contracts we’re adopting for `DOCMGR_PATH`, cwd, and `--root`.

## Context

We added a skills feature to docmgr (DocType=skill + WhatFor/WhenToUse + new CLI verbs). We then added an end-to-end smoke test script in `test-scenarios/testing-doc-manager/20-skills-smoke.sh`.

The first iterations of the smoke tests failed in ways that strongly suggest “test harness contract” issues (cwd/root resolution) rather than indexing logic bugs.

## Quick Reference

### Key lesson: cwd matters unless root is explicit

- docmgr discovers its config (e.g. `.ttmp.yaml`) by walking from the **current working directory** upward.
- If tests invoke docmgr from the “wrong” cwd, `--root ttmp` will point at the wrong docs root.
- **Fix**: in scenario tests, prefer `--root "${REPO}/ttmp"` (absolute) to make the test deterministic.

### Smoke test contract (target)

- **Binary contract**: scenarios should run against a pinned binary (fast + stable), or if using `go run`, ensure the process cwd remains `${REPO}` or pass absolute `--root`.
- **Assertions**: smoke tests should assert discoverability + hydration + reverse lookup, not exact formatting.

## Usage Examples

### Run only the skills smoke test (recommended)

From the source repo:

```bash
cd /home/manuel/workspaces/2025-12-19/add-docmgr-skills/docmgr && \
  DOCMGR_PATH="$(command -v docmgr)" \
  bash test-scenarios/testing-doc-manager/20-skills-smoke.sh /tmp/docmgr-scenario
```

### Run with `go run` (dev-only, slower)

Use a `DOCMGR_PATH` that does *not* change cwd away from the scenario repo, or make the script pass absolute `--root` everywhere.

## Related

- `analysis/02-skills-smoke-tests-failures-so-far-structured-approach.md`
- `test-scenarios/testing-doc-manager/20-skills-smoke.sh`
- `pkg/commands/skill_list.go`
- `pkg/commands/skill_show.go`

## Step 1: Run skills smoke test with a pinned built binary

We switched to a pinned binary (`/tmp/docmgr-scenario-local`) to remove the “go run + cwd mismatch” ambiguity entirely. This immediately made the scenario runs reproducible: the test runner controls exactly which `docmgr` is executed and we avoid module/cwd complications.

### What I did
- Built a pinned binary: `go build -o /tmp/docmgr-scenario-local ./cmd/docmgr`
- Ran the standard scenario setup steps (reset/mock repo/init ticket/create docs)
- Ran `test-scenarios/testing-doc-manager/20-skills-smoke.sh`

### What I learned
- The earlier failures were consistent with **test harness contract** issues (cwd/root), not necessarily broken indexing.
- Even when the binary is correct, smoke tests can still pass “silently” unless we add assertions.

### What warrants a second pair of eyes
- The smoke test should not rely on formatting stability; assertions should be minimal and robust.

## Step 2: Fix seed vocabulary to include docType=skill

The scenario uses `docmgr init --seed-vocabulary`, which seeds a minimal vocabulary in the new workspace. That seed did not include `skill`, so the smoke test was forced to mutate vocabulary via `docmgr vocab add` to get skills recognized reliably.

This step fixes the root cause: **the default seed vocabulary now includes `docType: skill`**.

**Commit (code):** f015d9c — "Init: seed skill docType in default vocabulary"

### What I did
- Updated `pkg/commands/init.go` `seedDefaultVocabulary()` to add:
  - `skill` — “Skill documentation (what it's for and when to use it)”

### What I learned
- Scenario suites that call `init --seed-vocabulary` are effectively asserting what a “fresh workspace” supports.
- If we add new docTypes, we must update seed defaults, otherwise new workspaces require manual `vocab add`.

## Step 3: Make the skills smoke test deterministic and assert results

We hardened the smoke test so it actually *fails* if skills aren’t indexed or the commands return empty output. We also made root usage deterministic by passing an absolute `--root` pointing at the scenario repo’s docs root.

**Commit (code):** 621b7c4 — "Test: make skills smoke deterministic and assert results"

### What I did
- Updated `test-scenarios/testing-doc-manager/20-skills-smoke.sh` to:
  - use `DOCS_ROOT="${REPO}/ttmp"` and pass `--root "${DOCS_ROOT}"` consistently
  - **assert** expected outputs via `grep -q` (list/show + file/dir filters + JSON output)
  - assert that seeded vocabulary contains `docTypes: skill`
- Fixed `test-scenarios/testing-doc-manager/run-all.sh` step ordering (step 19 before step 20)

### What worked
- Re-running the full setup + `20-skills-smoke.sh` passed end-to-end with the pinned binary.

### What was tricky to build
- Avoiding “green but empty output” runs: we needed explicit assertions, not just command exit codes.
