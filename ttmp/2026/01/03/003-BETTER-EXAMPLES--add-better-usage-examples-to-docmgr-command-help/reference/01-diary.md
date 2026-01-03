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
