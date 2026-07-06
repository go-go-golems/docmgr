---
name: diary
description: "Write and maintain an implementation diary: step-by-step narrative of what changed, why, what worked, what failed (with exact errors/commands), what was tricky, and how to review/validate. Use when a user asks to create/update a “Diary” doc, record work for a ticket, document debugging outcomes, or produce code review instructions and future follow-ups."
---

# Diary

## Overview

Produce concise, structured diary steps that capture the implementation journey (including failures) and make review and continuation straightforward.

## Create and Maintain a Diary (docmgr)

### Create the diary document

```bash
docmgr doc add \
  --ticket TICKET-ID \
  --doc-type reference \
  --title "Diary"
```

### Relate files (always use absolute paths)

```bash
docmgr doc relate \
  --doc ttmp/.../reference/01-diary.md \
  --file-note "/abs/path/to/file.go:Why this file matters"
```

docmgr stores the path in anchored form (`repo://...`, `abs://...`, ...); keep passing absolute paths — they are anchored automatically. A `--file-note` without a `:` separator errors and exits 1.

### Update changelog when a step completes

```bash
docmgr changelog update --ticket TICKET-ID \
  --entry "Step N: ... (commit <hash>)" \
  --file-note "/abs/path/to/file.go:Reason"
```

## Recommended Working Loop (keeps code ↔ docs consistent)

Use this loop for any non-trivial change:

```bash
# 1) Implement + format + test (capture failures verbatim)
gofmt -w path/to/file.go && go test ./... -count=1

# 2) Commit code (focused)
git add path/to/file.go && git commit -m "..." && git rev-parse HEAD

# 3) Update ticket bookkeeping (use the stable task ID from 'docmgr task list',
#    e.g. [v0sv]; 1-based positions also work)
docmgr task check --ticket TICKET-ID --id v0sv

# 4) Update diary + relate files + update changelog (include commit hash)
docmgr doc relate --doc ttmp/.../reference/01-diary.md \
  --file-note "/abs/path/to/file.go:What changed (commit <hash>)"
docmgr changelog update --ticket TICKET-ID \
  --entry "Step N: ... (commit <hash>)" \
  --file-note "/abs/path/to/file.go:Reason"

# 5) Commit docs changes
git add ttmp/... && git commit -m "Diary: record Step N"
```

## Diary Step Format (strict)

When asked to “write/update the diary”, output Markdown that follows this structure:

```markdown
# Diary

## Goal

Brief statement of what this diary captures.

## Step N: [Descriptive Step Name]

Write 1–2 short prose paragraphs here (required). Explain intent, what changed at a high level, and what it unlocked.

### Prompt Context

**User prompt (verbatim):** "<paste exactly what the user typed>"

**Assistant interpretation:** <your interpretation of what the user is asking for (paraphrase)>

**Inferred user intent:** <what outcome the user is trying to achieve / why they asked>

**Commit (code):** <hash> — "<message>"   # if code changed

### What I did
- Concrete actions (files changed, commands run)

### Why
- Motivation / rationale

### What worked
- Outcomes that moved the work forward

### What didn't work
- Failures/blockers with exact errors, versions, and commands

### What I learned
- Insights, corrected assumptions, patterns

### What was tricky to build
- Sharp edges: invariants, ordering constraints, normalization quirks, deadlocks, etc.

### What warrants a second pair of eyes
- Review-critical correctness/perf/concurrency risks

### What should be done in the future
- Only follow-ups implied by this step; if nothing: `N/A`

### Code review instructions
- Where to start (files + key symbols)
- How to validate (commands/tests)

### Technical details
- Reference material (snippets, schema/contracts, command examples)
```

## Writing Rules (strict)

- Start each step with 1–2 short prose paragraphs (required).
- Always include a `Prompt Context` section for each step.
  - `User prompt (verbatim)` must be the exact user text (no paraphrasing, no normalization) the **first time** that prompt appears in the diary.
  - If the same user prompt is reused for multiple steps, only include it verbatim once (the first time). In later steps, replace it with a short pointer like: `User prompt (verbatim): (see Step N)` or `User prompt (verbatim): (same as Step N)`.
  - `Assistant interpretation` is your concise paraphrase of the request.
  - `Inferred user intent` captures the user’s underlying goal/outcome.
- Record failures immediately; keep errors/versions/commands verbatim.
- For `What was tricky to build`, explain the underlying cause, the symptoms you observed, and how you approached a solution. If you found a solution, state the exact steps you took.
- Always include these sections for any code/behavior change: `What was tricky to build`, `What warrants a second pair of eyes`, `What should be done in the future`.
- Do not add backwards-compatibility shims unless the ticket/spec requires it; document behavior changes and update tests/docs instead.
- Relate any file you modified and any file that materially shaped decisions; always use absolute paths in `--file-note`.

## Reference

Load `references/diary.md` for the full diary rationale, patterns, and a worked example entry.
