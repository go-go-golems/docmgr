# Diary — Implementation Diary Workflow

## Overview

A diary is a **step-by-step narrative** of implementation work: what changed, why it changed, what happened (including failures), and what we learned.

**Core value:** structured reflection and learning capture — document the journey, not just the outcome.

The goal is to make:
- code review faster (reviewers know what to check and how to validate), and
- future continuation easier (a new contributor can pick up mid-stream without re-deriving context).

## Why this structure exists (keep the intent)

- **Prose first**: the 1–2 paragraph intro per step is the “human story” that future readers skim.
- **Failures are gold**: exact errors/versions/commands often save hours later; record them.
- **Tricky + second-pair-of-eyes**: forces you to surface the sharp edges and the real review risks.
- **Future work**: captures implied follow-ups (regressions/contracts/architecture constraints) without turning the diary into a wishlist.

## Quick Start (docmgr)

### Create a diary document

```bash
docmgr doc add \
  --ticket TICKET-ID \
  --doc-type reference \
  --title "Diary"
```

### Relate files to the diary

```bash
# Always use absolute paths
docmgr doc relate \
  --doc ttmp/.../reference/01-diary.md \
  --file-note "/abs/path/to/file.go:Why this file matters"
```

### Update changelog when a step is complete

```bash
docmgr changelog update --ticket TICKET-ID \
  --entry "Step N: ..." \
  --file-note "/abs/path/to/file.go:Reason"
```

## Recommended working loop (keeps code ↔ docs consistent)

For any non-trivial change:

```bash
# 1) Implement + format + test (capture failures verbatim)
gofmt -w path/to/file.go && go test ./... -count=1

# 2) Commit code (focused)
git add path/to/file.go && git commit -m "..." && git rev-parse HEAD

# 3) Update ticket bookkeeping
docmgr task check --ticket TICKET-ID --id N

# 4) Update diary + relate files + update changelog (include commit hash)
docmgr doc relate --doc ttmp/.../reference/01-diary.md \
  --file-note "/abs/path/to/file.go:What changed (commit <hash>)"
docmgr changelog update --ticket TICKET-ID \
  --entry "Step N: ... (commit <hash>)" \
  --file-note "/abs/path/to/file.go:Reason"

# 5) Commit docs changes
git add ttmp/... && git commit -m "Diary: record Step N"
```

## Diary format (keep this style)

### Step structure

- Every step starts with **1–2 short prose paragraphs** before subsections (required).
- Steps are numbered sequentially: `Step 1:`, `Step 2:`, etc.

```markdown
# Diary

## Goal

Brief statement of what this diary captures.

## Step N: [Descriptive Step Name]

Write 1–2 short prose paragraphs here (required). Explain intent, what changed at a high level, and what it unlocked.

**Commit (code):** <hash> — "<message>"   # if code changed

### What I did
- Concrete actions (files changed, commands run)

### Why
- Motivation / rationale (why this approach)

### What worked
- Outcomes that moved the work forward

### What didn't work
- Failures/blockers with **exact errors**, versions, and commands

### What I learned
- Extract the insight (patterns, gotchas, corrected assumptions)

### What was tricky to build
- Sharp edges: invariants, ordering constraints, normalization quirks, deadlocks, etc.

### What warrants a second pair of eyes
- The most review-critical semantics/correctness risks/perf/concurrency concerns

### What should be done in the future
- **Only** follow-ups implied by this step:
  - regressions to watch for,
  - contracts to declare/lock down,
  - architecture constraints made visible,
  - missing tests/docs needed to prevent drift.
- Not feature enhancement. If nothing: `N/A`.

### Code review instructions
- Where to start (files + key symbols)
- How to validate (commands/tests)

### Technical details
- Concrete reference material (snippets, schema/contracts, command examples)

### What I'd do differently next time
- Process improvements (optional)
```

## Patterns (short, but keep the intent)

- **Research step**: record what you searched for, what you read, and the decision you made.
- **Implementation step**: record diffs at a high level + the invariants you relied on.
- **Testing/debugging step**: record exact commands and the error output that mattered.

## Writing rules (concise but strict)

- **Be specific**: include file paths, commands, and commit hashes.
- **Document failures immediately** while details are fresh.
- **Always include** these sections for any code/behavior change:
  - `What was tricky to build`
  - `What warrants a second pair of eyes`
  - `What should be done in the future`
- If the ticket/spec says **no backwards compatibility**, do not add shims/flags/fallbacks. Document the behavior change and update tests/docs accordingly.

## File relating rules

- Always use **absolute paths** in `--file-note`.
- Relate any file you modified *and* any file that materially shaped decisions.
- Notes should explain why the file matters for this step (not just “touched”).

## Example diary entry (short but complete)

```markdown
## Step 3: Integrate Atlas via go run in Makefile

This step made Atlas execution self-contained by switching from a preinstalled binary to a `go run` pattern. The impact is reproducible tooling: new environments and CI don’t require out-of-band installation.

**Commit (code):** <hash> — "Build: run atlas via go run"

### What I did
- Updated `backend/Makefile` to replace `ATLAS ?= atlas` with `go run ariga.io/atlas/cmd/atlas@latest`

### Why
- Make Atlas execution reproducible and versioned with the repo

### What worked
- Targets run without requiring a preinstalled `atlas` binary

### What was tricky to build
- Ensuring module downloads and CI caching are acceptable for `go run`

### What warrants a second pair of eyes
- Confirm every Makefile target uses the same atlas invocation (no stragglers)

### What should be done in the future
- If CI becomes flaky/slow due to `go run`, treat it as a tooling contract issue: pin versions and add caching guidance (not a reason to reintroduce local-only dependencies).

### Code review instructions
- Start in `backend/Makefile`, search for `atlas`, then run `make atlas-status`
```
