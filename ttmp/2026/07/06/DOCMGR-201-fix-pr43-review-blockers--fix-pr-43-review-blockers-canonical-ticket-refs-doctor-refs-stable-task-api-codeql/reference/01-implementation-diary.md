---
Title: Implementation diary
Ticket: DOCMGR-201-fix-pr43-review-blockers
Status: active
Topics:
    - docmgr
    - cli
    - ux
    - tooling
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources:
    - https://github.com/go-go-golems/docmgr/pull/43
Summary: Chronological implementation diary for fixing PR 43 review blockers found in DOCMGR-200.
LastUpdated: 2026-07-06T11:40:00-04:00
WhatFor: "Follow the DOCMGR-201 implementation work and reproduce its validation steps."
WhenToUse: "Use when reviewing or continuing the PR 43 blocker fixes."
---

# Diary

## Goal

Capture the implementation journey for DOCMGR-201: fixing the PR 43 review blockers found in DOCMGR-200, keeping the code changes, experiments, validation, and review instructions tied to the ticket.

## Step 1: Ticket setup and blocker triage

Created a focused follow-up ticket for the PR 43 review findings so the blocker fixes are separate from the original DOCMGR-200 review deliverable. The ticket now has a diary, an analysis/fix-plan document, and explicit task rows for the three correctness blockers, CodeQL triage, and validation.

This step also preserves the current branch state: the DOCMGR-200 review artifacts were committed first, then DOCMGR-201 was created cleanly on top of the same PR branch.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to address the issues, keep a diary as you work, commit at appropriate intervals."

**Assistant interpretation:** Create a new docmgr ticket for the PR 43 blocker fixes, maintain a diary while implementing, and make focused commits as work completes.

**Inferred user intent:** Turn the review findings into tracked implementation work that can be committed and reviewed independently.

**Commit (code):** N/A — setup/documentation step in progress.

### What I did
- Committed the prior DOCMGR-200 PR review artifacts as `6da50a9` (`DOCMGR-200: add PR 43 review report`) so new work starts from a clean baseline.
- Created ticket `DOCMGR-201-fix-pr43-review-blockers` with title `Fix PR 43 review blockers: canonical ticket refs, doctor refs, stable task API, CodeQL`.
- Added this diary document and `analysis/01-pr-43-blockers-fix-plan.md`.
- Added tasks for:
  - canonical `doc add` ticket metadata;
  - forgiving `doctor --ticket` references;
  - stable task refs through HTTP/UI;
  - CodeQL path alert triage;
  - regression tests and validation.

### Why
- The review findings are actionable code changes and should not be mixed into the DOCMGR-200 review-report commit.
- A separate ticket gives reviewers a compact scope and a place to record validation failures while fixing the branch.

### What worked
- Ticket creation and document scaffolding succeeded.
- The new task IDs are stable-marker IDs, which is useful because one of the fixes will exercise that same stable-ID model.

### What didn't work
- N/A for setup.

### What I learned
- The branch already contained one uncommitted review-report bundle; committing it first avoided mixing review documentation with blocker-fix code.

### What was tricky to build
- The main ordering constraint was repository hygiene: committing DOCMGR-200 first prevents future commits from bundling old review artifacts with new code fixes.

### What warrants a second pair of eyes
- Confirm that DOCMGR-201 is the desired ticket ID and scope before the blocker fixes land.

### What should be done in the future
- Implement the blockers in small commits, checking off the ticket tasks as each fix is validated.

### Code review instructions
- Start from the DOCMGR-200 report at `ttmp/2026/07/05/DOCMGR-200-.../analysis/02-pr-43-code-review-and-project-review.md`.
- Review this ticket's `analysis/01-pr-43-blockers-fix-plan.md` before code changes.

### Technical details
- Setup commands used:
  - `git add ttmp/2026/07/05/DOCMGR-200-... && git commit -m "DOCMGR-200: add PR 43 review report"`
  - `go run -tags sqlite_fts5 ./cmd/docmgr ticket create --root ttmp --ticket DOCMGR-201-fix-pr43-review-blockers --title "Fix PR 43 review blockers: canonical ticket refs, doctor refs, stable task API, CodeQL" --topics docmgr,cli,ux,tooling`
  - `go run -tags sqlite_fts5 ./cmd/docmgr doc add --root ttmp --ticket DOCMGR-201-fix-pr43-review-blockers --doc-type reference --title "Implementation diary"`
  - `go run -tags sqlite_fts5 ./cmd/docmgr doc add --root ttmp --ticket DOCMGR-201-fix-pr43-review-blockers --doc-type analysis --title "PR 43 blockers fix plan"`
