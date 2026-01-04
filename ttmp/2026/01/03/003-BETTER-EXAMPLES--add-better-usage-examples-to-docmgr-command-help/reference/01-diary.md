---
Title: Diary
Ticket: 003-BETTER-EXAMPLES
Status: active
Topics:
    - docmgr
    - cli
    - docs
    - ux
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/commands/changelog.go
      Note: Adjust reminder text to canonical verbs
    - Path: pkg/commands/relate.go
      Note: Doc relate example changes verified by running relate
    - Path: pkg/commands/search.go
      Note: Verified --query behavior (commit 8ec1c61)
    - Path: pkg/commands/ticket_move.go
      Note: Verified ticket move example + underscore skip (commit 8ec1c61)
    - Path: ttmp/2026/01/03/003-BETTER-EXAMPLES--add-better-usage-examples-to-docmgr-command-help/index.md
      Note: Define ticket overview
    - Path: ttmp/2026/01/03/003-BETTER-EXAMPLES--add-better-usage-examples-to-docmgr-command-help/reference/01-diary.md
      Note: Initialize diary + record Step 1
    - Path: ttmp/2026/01/03/003-BETTER-EXAMPLES--add-better-usage-examples-to-docmgr-command-help/tasks.md
      Note: Define ticket task checklist
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T18:14:20.549352964-05:00
WhatFor: ""
WhenToUse: ""
---




# Diary

## Goal

Track the step-by-step work to add real, copy/paste-ready examples to the long help text of every `docmgr` CLI command.

## Step 1: Bootstrap ticket + find where long help lives

This step created a dedicated ticket workspace so we can safely run “real” CLI examples without polluting other tickets, and established where command long descriptions and parameter definitions are implemented in the codebase.

The key outcome is a concrete audit starting point: `pkg/commands/*.go` uses `cmds.NewCommandDescription(..., cmds.WithLong(...), cmds.WithFlags(...))` for the long help and flag docs, so adding examples is mostly a doc change in those descriptions.

### What I did
- Created ticket `003-BETTER-EXAMPLES` and a `Diary` reference doc.
- Verified docmgr is initialized and the current docs root resolves correctly.
- Listed top-level CLI commands with `docmgr --help`.
- Located command description sources via `rg` (most are in `pkg/commands/*.go`).

### Why
- Keep example testing isolated and reproducible.
- Ensure every example we add to help text matches real flags and behavior.

### What worked
- `go run ./cmd/docmgr status --summary-only` confirms `ttmp/` root resolution is working.
- `docmgr --help` shows the full command surface area we need to cover.

### What didn't work
- Earlier, running `go run` failed with `permission denied` under sandboxed execution when writing to the Go build cache in `~/.cache/go-build/` (resolved once full permissions were enabled).

### What I learned
- In this repo, command examples belong in `cmds.WithLong(...)`, close to the flag definitions, so they’re hard to drift.

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- Ensure every example uses the canonical verbs/subcommands (some legacy aliases may still exist in docs like `create-ticket` vs `ticket create-ticket`).

### What should be done in the future
- Consider a lightweight CI check that fails if any command lacks an `Examples:` section in long help. (Optional.)

### Code review instructions
- Start with `pkg/commands/status.go` and `pkg/commands/create_ticket.go` to see the current long help pattern.
- Validate by running `GOWORK=off go run ./cmd/docmgr --help` and `GOWORK=off go run ./cmd/docmgr <command> --help` for any edited command.

### Technical details
- Status command used: `GOWORK=off go run ./cmd/docmgr status --summary-only`
- Ticket created with: `GOWORK=off go run ./cmd/docmgr ticket create-ticket --ticket 003-BETTER-EXAMPLES --title "Add better usage examples to docmgr command help" --topics docmgr,cli,docs,ux`

## Step 2: Fix initial help examples + validate doc relate

This step started the actual “better examples” work by updating a first slice of command long help strings, focusing on correctness (use real subcommands) and on having copy/paste-ready multi-flag examples. It also removed the suggestion-focused `doc relate` examples per request and replaced them with multi-file relating examples.

**Commit (code):** 8692e86 — "CLI: refresh help examples"

### What I did
- Updated long help examples to use the actual cobra command tree (`docmgr ticket ...`, `docmgr doc ...`), and fixed the README template text emitted by ticket creation.
- Removed `--suggest` examples from `docmgr doc relate` and replaced them with “relate multiple files at once” examples.
- Ran “real” CLI invocations against ticket `003-BETTER-EXAMPLES` to validate `docmgr doc relate` behavior.

### Why
- Some examples in the codebase referenced legacy root-level verbs (`docmgr add`, `docmgr create-ticket`) that are no longer valid in this build.
- The goal of this ticket is for examples to be executable, not aspirational.

### What worked
- `docmgr doc relate --ticket 003-BETTER-EXAMPLES --file-note ...` updates `index.md` RelatedFiles as expected.
- `docmgr doc relate --doc .../reference/01-diary.md --file-note ...` updates the document frontmatter as expected.
- `go test ./... -count=1` passed after the edits.

### What didn't work
- N/A

### What I learned
- The same command implementation can be surfaced under multiple cobra paths (e.g., search), so help examples should prefer the canonical grouping even if aliases exist.

### What was tricky to build
- Keeping examples correct across both the CLI surface and template-generated docs (ticket README template).

### What warrants a second pair of eyes
- Confirm which invocation style we want to standardize on in examples when a command is intentionally duplicated (e.g., `docmgr status` vs `docmgr workspace status`, `docmgr search` alias).

### What should be done in the future
- N/A

### Code review instructions
- Start with `pkg/commands/relate.go`, then skim `pkg/commands/create_ticket.go` and `pkg/commands/add.go`.
- Validate the examples by running:
  - `GOWORK=off go run ./cmd/docmgr doc relate --help`
  - `GOWORK=off go run ./cmd/docmgr ticket create-ticket --help`

## Step 3: Add examples across remaining commands + run real example workflows

This step expanded help examples broadly: group commands (like `docmgr doc`, `docmgr ticket`, etc.) now include an Examples section, and a larger set of leaf commands gained additional, copy/paste-ready examples. The emphasis stayed on “commands that actually work” (correct cobra paths, correct flags, and realistic workflows).

I also ran a battery of real CLI commands against scratch tickets to verify examples end-to-end, including a migration-style `ticket move` and a safe `vocab add` that was reverted immediately after proving the command path and flags.

**Commit (code):** 8ec1c61 — "CLI: add more help examples"

### What I did
- Added/expanded `Examples:` sections across many commands in `pkg/commands/*` (search, doctor, tasks, ticket move/doc move/layout-fix/renumber, import/vocab, etc.).
- Added `Long` + Examples to cobra group commands in `cmd/docmgr/cmds/*` so `docmgr <group> --help` is immediately usable.
- Verified a representative set of examples “for real” using scratch tickets under `ttmp/examples/` and `ttmp/legacy/`.

### Why
- The CLI surface area is large; the fastest way to reduce usage friction is to provide executable examples directly in `--help`.
- Group commands are often where users start; without examples there, discovery is slower.

### What worked
- `docmgr search --query "Scratch Doc" --ticket 003-BETTER-EXAMPLES-SCRATCH2-A` works (positional args do not).
- `docmgr ticket move --ticket 003-BETTER-EXAMPLES-LEGACY2` correctly migrates a ticket from `ttmp/legacy/` into the date-based template.
- `docmgr template validate` works when `<root>/templates/*.templ` exists (validated a throwaway `ttmp/templates/demo.templ`).
- `docmgr validate frontmatter --doc <abs-path> --with-glaze-output --output json` works and returns status `ok`.
- `docmgr vocab add ... --with-glaze-output --output json` works; vocabulary changes were reverted after confirming behavior.

### What didn't work
- Running `docmgr validate frontmatter --doc ttmp/...` while also relying on the default `--root ttmp` produced a bogus `ttmp/ttmp/...` path. The example is now tested using an absolute path to avoid the ambiguity.
- Running `docmgr configure` in this nested repo created an untracked `.ttmp.yaml` (expected behavior when none exists); it should be cleaned up after testing.

### What I learned
- Underscore-prefixed directories under `ttmp/` are intentionally skipped during ingest/indexing; examples should avoid path templates that begin with `_`.

### What was tricky to build
- Keeping examples accurate across commands that run in different modes:
  - some support `--with-glaze-output`,
  - some always emit structured output.

### What warrants a second pair of eyes
- Confirm the intended convention for examples where both root-level and namespaced forms exist (e.g., `docmgr status` vs `docmgr workspace status`), and whether we want to standardize on one in help text.

### What should be done in the future
- Consider adding a lightweight test that asserts every registered command has an `Examples:` section in its long help (including group commands).

### Code review instructions
- Start with `pkg/commands/search.go`, `pkg/commands/tasks.go`, `pkg/commands/ticket_move.go`, and the cobra group defs in `cmd/docmgr/cmds/*`.
- Validate quickly with:
  - `GOWORK=off go run ./cmd/docmgr --help`
  - `GOWORK=off go run ./cmd/docmgr doc --help`
  - `GOWORK=off go run ./cmd/docmgr ticket --help`
