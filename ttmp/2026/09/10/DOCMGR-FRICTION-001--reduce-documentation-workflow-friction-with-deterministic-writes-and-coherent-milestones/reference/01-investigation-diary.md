---
Title: Investigation diary
Ticket: DOCMGR-FRICTION-001
Status: active
Topics:
    - docmgr
    - cli
    - usability
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://internal/documents/frontmatter.go
      Note: Lossless body framing, commit 91e0603
    - Path: repo://internal/documents/stability_test.go
      Note: Ten-cycle, CRLF and unknown metadata regressions
    - Path: repo://internal/documents/write.go
      Note: Atomic no-op writer and permission preservation
    - Path: repo://pkg/commands/add.go
      Note: Explicit creation separator
    - Path: repo://pkg/commands/changelog_entries.go
      Note: Pure canonical entry builder
    - Path: repo://pkg/commands/changelog_stability_test.go
      Note: Boundary regression matrix
    - Path: repo://pkg/commands/create_ticket.go
      Note: Explicit index creation separator
    - Path: repo://pkg/models/document.go
      Note: Unknown YAML node retention
    - Path: repo://ttmp/2026/09/10/DOCMGR-FRICTION-001--reduce-documentation-workflow-friction-with-deterministic-writes-and-coherent-milestones/design-doc/01-intern-guide-to-deterministic-document-mutation-and-milestone-workflows.md
      Note: Primary research handoff
    - Path: repo://ttmp/2026/09/10/DOCMGR-FRICTION-001--reduce-documentation-workflow-friction-with-deterministic-writes-and-coherent-milestones/scripts/03-render-guides.cjs
      Note: Reproducible diagram and print export
    - Path: repo://ttmp/2026/09/10/DOCMGR-FRICTION-001--reduce-documentation-workflow-friction-with-deterministic-writes-and-coherent-milestones/sources/remarkable-upload.log
      Note: Verified delivery result
ExternalSources: []
Summary: Research chronology and review instructions for docmgr workflow friction.
LastUpdated: 2026-09-10T15:37:22Z
WhatFor: Preserve inspected evidence and explain the implementation handoff.
WhenToUse: Resuming this ticket or reviewing the reproductions.
---


# Investigation diary

## Goal

Explain session friction through inspected docmgr source and reproducible experiments, then deliver an intern guide without modifying product behavior.

## Step 1: Investigate persistence and lifecycle behavior

Created this ticket and the cross-linked skills ticket, inspected the relevant command/persistence APIs and converted a frozen copy of the motivating Pi session. Wrote the intern guide after reproducing the whitespace behavior rather than attributing every administrative failure to docmgr on the basis of memory.

### Prompt Context

**User prompt (verbatim):**

> ok, create a new docmgr ticket in ~/code/wesen/go-go-golems/docmgr itself about improving docmgr, and another one about improving the skills in ~/.pi/agent/skills (which is a git repo) itself.
>
> In both cases, Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.
>
> Cross reference the things if necessary. Do go into the details of this session in particular, reference which skills and other guidance was wrong specifically, etc...
>
> [REMINDER] Output a <summary>...</summary> block at the VERY END of your response. This is mandatory.

**Assistant interpretation:** Create two researched implementation handoffs and upload them; do not implement the proposed product/skill changes yet.

**Inferred user intent:** Make the session's avoidable work actionable for a new contributor while retaining useful correctness checks.

### What I did

- Read `AGENT.md`, document persistence, changelog, close, relation, task and HTTP boundaries.
- Built a local pinned CLI and ran the two ticket-contained reproducers in temporary directories.
- Ran existing focused document/changelog tests.
- Wrote the design guide with source maps, API sketches, decisions, phases and regression criteria.

### Why

The original EOF failures alone do not identify every root cause. A source-level reproducer separates formatting defects from excessive evidence-handling procedure.

### What worked

The frontmatter reproducer observed leading newline counts 1/2/3/4; changelog append produced two trailing newlines. The close probe observed repeated changes/history and an error after status had already become complete. Existing focused tests passed, showing the missing regression coverage. No live workspace was used for destructive fault injection.

### What didn't work

An initial discovery grep included an invalid literal newline regex and reported `rg: the literal "\n" is not allowed in a regex`; corrected the search to explicit symbols. Skills-repository status initially reported a missing `ttmp` root; initialized it as requested. The close probe's exit 1 is an intentionally injected failure, not a failed investigation. Final staging rejected trailing tab padding in the human `go version -m` capture; removed only that display whitespace. The first commit hook ran the full Go tests successfully, then errcheck rejected `01-reproduce-writes.go:25:20: Error return value of os.RemoveAll is not checked`. Replaced the unchecked deferred cleanup with a checked helper call; no product code or experiment policy changed.

### What I learned

The writer already uses per-file temporary replacement. The missing guarantee is stable framing and multi-file lifecycle consistency, not the complete absence of atomic file writes. `ApplyRelatedFilesUpdate` already has a no-op guard, so the report must not claim all mutations rewrite unchanged content.

### What was tricky to build

Distinguishing raw authored Markdown from serializer-owned separators, and choosing a deterministic close failure that does not depend on Unix permission assumptions. A directory at the changelog path reliably proves partial close behavior.

### What warrants a second pair of eyes

Review the proposed body-preservation contract, unknown metadata and permission behavior, operation-ID semantics and the difference between recoverability and atomic visibility. No new milestone API exists yet.

### What should be done in the future

Implement the open phases in tasks.md after the design is accepted. Keep this research ticket active rather than confusing guide delivery with implemented fixes.

### Code review instructions

Start with the design guide and `sources/source-map.txt`. Re-run scripts 01/02 from the docmgr root using a locally built binary. Compare JSON evidence and check that no product source changed.

### Technical details

Source pin: `710f31384d2d3242bfb84bf37cd06d2868a63bd3`. Session evidence is owned by SKILLS-FRICTION-001; both guides reference that frozen snapshot and the original project commit. Delivery/render receipts are stored under this ticket's sources directory. The guide and diary bundle was successfully uploaded to `/ai/2026/09/10/DOCMGR-FRICTION-001` and its listing verified. Both diagrams were inspected; the second was narrowed for print readability. Print copies add lowercase `title` metadata for the uploader while preserving canonical docmgr frontmatter. The uploaded diary is retained as `sources/print/diary.md`; this final delivery note was recorded afterward without requiring a recursive re-upload.

## Step 2: Establish byte-stable persistence contracts

Implementation began with the lowest-level reproduced defects, rather than teaching skills to repair generated files. The serializer now owns only delimiter termination; callers explicitly supply creation spacing. Changelog boundaries are canonical without trimming authored Markdown spaces.

The original research captures and uploaded snapshots remain historical evidence. New tests establish the changed behavior, including ten no-op cycles, CRLF body preservation, unknown metadata and file modes.

### Prompt Context

**User prompt (verbatim):**

> let's actually build it all. commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill).
>
> [REMINDER] Output a <summary>...</summary> block at the VERY END of your response. This is mandatory.

**Assistant interpretation:** Implement both researched tickets with focused commits and detailed investigation entries.

**Inferred user intent:** Replace the identified friction with tested tooling and coherent instructions, not another proposal.

### What I did
- Added a pure document serializer and an atomic, permission-preserving write-if-changed primitive.
- Retained unknown YAML fields through an inline node map; comments/style remain outside the fidelity guarantee.
- Audited creation versus metadata-update call sites; only create-ticket and doc-add need initial separators.
- Extracted a deterministic changelog builder, preserving internal Markdown and sorting file notes.
- Added regression tests using the standard testing library.

### Why
- Metadata-neutral cycles must not grow the body or replace unchanged files.
- Atomic single-file replacement is useful but must not imply multi-file consistency.

### What worked
- `go test ./internal/documents ./pkg/commands ./pkg/models` passes.
- Ten cycles retain identical bytes, sentinel mtime and mode; CRLF body and custom nested YAML survive metadata edits.

### What didn't work
- The first test command returned `go: updates to go.mod needed; to update it: go mod tidy` because the new tests imported testify, which this module does not directly use. Rewrote these tests with the standard library rather than adding a test-only dependency; the same focused command then passed.

### What I learned
- YAML formatting and Markdown fidelity need separate contracts. Canonical YAML can lose comments while retaining unknown values and exact body bytes.

### What was tricky to build
- Removing the extra serializer newline alone changes new-document presentation. An explicit CreationBody helper fixes that boundary without guessing whether existing leading blank lines were authored.
- Replacing an existing file via CreateTemp used to impose temporary-file permissions. The shared writer now preserves target permissions and rejects symlink/non-regular targets.

### What warrants a second pair of eyes
- Unknown YAML nodes with aliases and unusual scalar styles; first-write metadata normalization is intentionally not byte-identical YAML.
- File replacement remains subject to uncooperative filesystem races; no multi-file or adversarial filesystem guarantee is claimed.

### What should be done in the future
- Build cooperative operation locking, recoverable projections, shared close behavior, milestone and resume APIs in the next steps.

### Code review instructions
- Start at `internal/documents/frontmatter.go`, `write.go` and `stability_test.go`; then inspect changelog builder/tests and the two creation callers.
- Run the focused command above and the repository commit hooks. `format_file` is not exposed; the documented equivalent `gofmt` is used.

### Technical details
- Metadata framing uses LF; supplied body bytes, including CRLF and hard-break spaces, are untouched.
- New files default to 0644; existing permission bits survive replacement. Identical canonical bytes are not rewritten.
