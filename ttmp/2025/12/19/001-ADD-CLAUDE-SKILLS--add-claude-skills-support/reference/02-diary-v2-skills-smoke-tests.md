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
RelatedFiles:
    - Path: pkg/commands/skill_list.go
      Note: Implements skill list used by smoke tests
    - Path: pkg/commands/skill_show.go
      Note: Implements skill show used by smoke tests
    - Path: test-scenarios/testing-doc-manager/20-skills-smoke.sh
      Note: Skills smoke test script under iteration
    - Path: test-scenarios/testing-doc-manager/run-all.sh
      Note: Scenario runner wiring for step 20
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
