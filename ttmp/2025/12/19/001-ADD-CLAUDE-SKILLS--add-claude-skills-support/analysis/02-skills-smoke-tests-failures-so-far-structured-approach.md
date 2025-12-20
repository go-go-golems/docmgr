---
Title: 'Skills smoke tests: failures so far + structured approach'
Ticket: 001-ADD-CLAUDE-SKILLS
Status: active
Topics:
    - features
    - skills
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-19T13:34:17.055035361-05:00
---

# Skills smoke tests: failures so far + structured approach

## Goal

Stop “thrash debugging” and converge on a **conceptually sound** smoke-test approach for the new skills feature, with a clear contract for how the test suite invokes `docmgr` (binary vs `go run`), how root resolution is controlled, and what the smoke tests *assert*.

## What we have so far (artifacts)

- Implemented the skills feature:
  - Data model: `WhatFor`, `WhenToUse` on `models.Document`
  - SQLite index: `docs.what_for`, `docs.when_to_use`
  - Index ingest + query hydration
  - CLI: `docmgr skill list`, `docmgr skill show`
- Added a scenario script:
  - `test-scenarios/testing-doc-manager/20-skills-smoke.sh`
  - Wired into `test-scenarios/testing-doc-manager/run-all.sh`

## What failed (observed symptoms)

### Symptom A: `skill show` reported “Too many arguments”

Root cause: our Cobra/glazed wiring expects the skill name via a **flag** (`--skill`), not a positional argument. The smoke test used:

- `docmgr skill show "API Design"`

…but it must be:

- `docmgr skill show --skill "API Design"`

### Symptom B: Skills files existed, but `skill list`/`skill show` returned no results

We verified the scenario created the files correctly:

- Ticket-level skills:
  - `ttmp/.../MEN-4242--.../skills/01-api-design.md`
  - `ttmp/.../MEN-4242--.../skills/02-websocket-management.md`
- Workspace-level skills:
  - `ttmp/skills/workspace-testing.md`

Frontmatter looked valid (DocType=skill, WhatFor/WhenToUse, etc.).

Yet:

- `docmgr skill list` printed nothing
- `docmgr skill show --skill "API Design"` returned “no skills found matching …”

## The likely root cause (conceptual)

This looks like an **invocation contract bug**, not an indexing bug:

- The scenario scripts assume `${DOCMGR}` runs with **cwd = the mock repo** (`/tmp/docmgr-scenario/acme-chat-app`), because they `cd "${REPO}"` at the top.
- When we tried to use `DOCMGR_PATH="go run ..."` we introduced a wrapper that **changed directories** to the source repo to make `go run` work.
- That means the actual `docmgr` process ran with **cwd = the source repo**, so:
  - relative `--root ttmp` pointed to the *source repo’s* `ttmp/`, not the scenario repo’s `ttmp/`
  - config discovery (`.ttmp.yaml`) and root resolution followed the source repo, not `/tmp/docmgr-scenario/acme-chat-app`

So the skills were “missing” simply because we were querying the wrong docs root.

This is the key lesson: **`docmgr` behavior depends on cwd (config discovery) unless `--root` is absolute and explicit.**

## A structured, sound approach for skills smoke tests

### 1) Define the invocation contract

We need exactly one of these contracts (pick one; don’t mix):

- **Contract (CI / default)**: Build a pinned binary once, run it many times.
  - Pros: fast, stable, no cwd surprises
  - Cons: requires `go build` in the repo before running scenarios

- **Contract (dev-only)**: Use `go run` but *without changing cwd away from the scenario repo*.
  - If we use `go run`, we must ensure the `docmgr` process still runs with cwd=`${REPO}`.
  - Practically: set `DOCMGR_PATH` to a command that does *not* `cd` into the source repo before executing `docmgr`, or force absolute `--root`.

### 2) Make root resolution deterministic for skills tests

For skills specifically, we should make the test robust even if cwd is wrong:

- Define `DOCS_ROOT="${REPO}/ttmp"` (absolute)
- Pass `--root "${DOCS_ROOT}"` on every `docmgr` invocation in the skills script

This eliminates “wrong workspace” bugs and makes the smoke test self-contained.

### 3) Keep smoke-test assertions minimal but meaningful

Smoke tests should assert:

- **Indexing**: newly created skill docs are discoverable via `skill list`
- **Hydration**: `skill list` output includes `what_for` / `when_to_use` (not empty) for at least one skill
- **Reverse lookup**: `skill list --file ...` returns at least one expected skill
- **Show**: `skill show --skill ...` prints the preamble + body for a known skill

Avoid fragile assertions on formatting; prefer searching for a few key strings.

### 4) Decide: should smoke tests rely on vocabulary for `DocType=skill`?

Today, the indexer stores `doc_type` as a string; the query filter matches by string.
So skills should be listable even if vocabulary doesn’t contain `skill` (vocab mostly affects diagnostics/validation, not indexing).

For smoke tests, we can:

- either explicitly add `skill` to vocabulary (safe, but adds mutation)
- or treat vocabulary as orthogonal, and just ensure parsing/indexing works

Recommendation: **don’t mutate vocabulary** in smoke tests unless you’re explicitly testing vocabulary flows.

## Next actions

1. Update `20-skills-smoke.sh` to:
   - use `--skill` for show
   - use an absolute `--root "${REPO}/ttmp"` for all calls
   - remove vocabulary mutation
2. Re-run the skills smoke test using a clean invocation contract.
3. If skills still don’t show up, *then* investigate indexing (ingest + schema + query) with a targeted debug mode.

