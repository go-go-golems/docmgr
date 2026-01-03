---
Title: "Diary: PR creation postmortem (prescribe)"
Ticket: 002-ADD-TICKET-GRAPH
Status: active
Topics:
    - docmgr
    - tooling
    - git
    - docs
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/docmgr/cmds/ticket/graph.go
      Note: CLI entrypoint for the new graph command described in the PR
    - Path: pkg/commands/ticket_graph.go
      Note: Core implementation for ticket graph generation and expansion
    - Path: pkg/commands/ticket_graph_test.go
      Note: Graph fixtures and tests validating output and expansion behavior
    - Path: Makefile
      Note: Exports `GOWORK=off`, which surfaced in PR generation copy
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T17:20:09-05:00
WhatFor: ""
WhenToUse: ""
---

# Diary: PR creation postmortem (prescribe)

## Goal

Capture what went wrong while trying to generate and submit the `docmgr` PR via `prescribe`, what work had to be done manually (git spelunking, base verification, YAML repair, toolchain diagnosis), and what `prescribe` could provide to reduce this friction.

## Step 1: Initialize a session and generate a PR description (first attempt)

This step attempted to follow a “happy path”: initialize a `prescribe` session from the current branch diff, then generate a PR title/body YAML suitable for `prescribe create`. The immediate impact was learning that the generated output can be “YAML-looking” but still fail the tool’s YAML parse/struct unmarshal, which blocks the “use last generated” workflow.

### What I did
- Initialized a session:
  - `prescribe session init --save --target origin/main --title "ticket graph: add graph command + export formats" --description "Add a ticket graph command (mermaid), transitive expansion, export helpers, and accompanying docs/tests."`
- Verified session stats:
  - `prescribe session show` reported `total_files: 17` and `token_count: 38360`.
- Generated PR YAML:
  - `PINOCCHIO_PROFILE=gemini-2.5-pro prescribe generate --ai-api-type gemini --ai-engine gemini-2.5-pro --stream --output-file .pr-builder/generated-pr.md`

### Why
- The goal was to let `prescribe` drive the full sequence: diff → model → parse YAML → `prescribe create` → `gh pr create`.

### What worked
- `prescribe session init` produced a sensible file set for this branch.
- The LLM output was written to `.pr-builder/generated-pr.md` reliably even when parsing failed.

### What didn't work
- The generated YAML failed `prescribe` parsing with:
  - `failed to parse PR YAML: yaml: mapping values are not allowed in this context`
- Because parsing failed, `.pr-builder/last-generated-pr.yaml` was not written (so `prescribe create --use-last` had nothing to consume).

### What I learned
- `--output-file` is a useful “raw capture”, but it’s not a substitute for a successful parse because `create --use-last` depends on `.pr-builder/last-generated-pr.yaml`.

### What was tricky to build
- The error message didn’t include a line/column pointer into the returned YAML, so diagnosing *why* it didn’t parse required manual inspection.

### What warrants a second pair of eyes
- Confirm the intended PR YAML schema for `prescribe create` (string vs sequence types, and whether code fences are allowed).

### What should be done in the future
- Add a “schema enforcement” option or preset that reliably produces parseable YAML (see recommendations below).

### Code review instructions
- Inspect `.pr-builder/generated-pr.md` for the raw output and verify whether it conforms to the expected `title/body/changelog/release_notes` schema.

## Step 2: Iterate prompts and hit a stable failure mode (type mismatch for `changelog`)

This step retried generation with a prompt that explicitly demanded a fixed YAML shape. The impact was discovering a more concrete, repeatable failure: the model produced a YAML sequence for `changelog` while `prescribe` expects a string, causing unmarshalling to fail.

### What I did
- Re-ran generation with a stricter prompt (“ONLY YAML; exact keys and types”) and captured the raw output in `.pr-builder/generated-pr.md`.
- Observed `prescribe` parse failure:
  - `failed to parse PR YAML: yaml: unmarshal errors: line 27: cannot unmarshal !!seq into string`

### Why
- We needed a parseable `.pr-builder/last-generated-pr.yaml` to proceed to `prescribe create`.

### What worked
- The failure mode became explicit: YAML was syntactically valid, but semantically incompatible with `prescribe`’s expected struct types.

### What didn't work
- The model kept emitting:
  - triple-backtick fenced YAML (` ```yaml ... ``` `), and
  - `changelog:` as a YAML list (sequence).

### What I learned
- Even when a model can “follow instructions”, it’s still common for it to drift into markdown-fenced output and to choose list-y structures unless the tool aggressively enforces a schema.

### What was tricky to build
- It’s surprisingly hard to “prompt away” type mismatches reliably without tool-level schema validation.

### What warrants a second pair of eyes
- Confirm whether `prescribe` should accept `changelog` as either a string or a list of strings (and normalize), since both are reasonable representations.

### What should be done in the future
- Add post-processing in `prescribe`:
  - strip code fences automatically, and
  - if `changelog` is a list, join with newlines (or error with a helpful fix-it hint and a proposed normalized value).

### Code review instructions
- Ensure any final PR YAML uses `changelog: |` as a multiline string, not a YAML list.

### Technical details: full broken YAML (raw model output)

This is the exact YAML that failed parsing due to `changelog` being a sequence (and being wrapped in code fences):

```yaml
title: 'feat(ticket): Add `graph` command for Mermaid visualization'
body: >
  Introduces a new `docmgr ticket graph` command to generate Mermaid graphs
  visualizing the relationships between a ticket's documents and their
  referenced files.


  Key features include:

  - **Mermaid Output**: Generates graphs in Mermaid syntax, with options to
  output raw DSL or a pasteable Markdown code block.

  - **Transitive Expansion**: Supports transitive graph expansion via `--depth`
  and `--scope repo` to discover other documents across the repository that
  reference the same files.

  - **Safety Limits**: Includes `--max-nodes` and `--max-edges` flags to
  prevent runaway graph generation during expansion.

  - **Tests & Docs**: Adds comprehensive unit and fixture-based tests for the
  new command, along with extensive documentation and ticket workspace examples.


  Additionally, the Makefile is updated to export `GOWORK=off` to ensure
  local development hooks run correctly within a nested Go workspace.
changelog:
  - '[New] Add `docmgr ticket graph` command to generate a Mermaid graph of a ticket''s documents and related files.'
  - '[New] Implement transitive graph expansion with `--depth` and `--scope` options to discover related documents across the repository.'
  - '[New] Introduce safety limits (`--max-nodes`, `--max-edges`) and query batching for transitive expansion.'
  - '[New] Add comprehensive tests and fixtures for the new graph command, covering output sanitization and transitive expansion logic.'
  - '[New] Add extensive documentation and ticket workspace fixtures for the new feature.'
  - '[Fix] Export `GOWORK=off` in the Makefile to ensure local development hooks run correctly in a nested Go workspace.'
release_notes:
  title: Visualize Ticket Dependencies with `docmgr ticket graph`
  body: >
    You can now generate visual maps of your ticket workspaces with the new `docmgr
    ticket graph` command. This feature creates a Mermaid graph showing all
    documents within a ticket and the code files they reference via
    `RelatedFiles`.


    The command also supports transitive expansion, allowing you to discover
    other documents across the entire repository that are linked to the same
    files, revealing hidden dependencies and relationships.


    Example usage:

    ```sh

    # Generate a graph for a specific ticket

    docmgr ticket graph --ticket MY-TICKET-123

    ```


    The output can be pasted directly into Markdown files to embed visualizations
    in your documentation.
```

## Step 3: Toolchain/workspace friction blocked validation until fixed externally

This step was about validation rather than PR text: running `go test ./...` in `docmgr` initially failed due to Go workspace (`go.work`) version constraints. The impact is that even “basic PR hygiene” (tests before PR) can be blocked by workspace configuration, and the error message points at `go work use` as the fix.

### What I did
- Attempted `go test ./... -count=1` and hit:
  - `go: module ../glazed listed in go.work file requires go >= 1.24.2, but go.work lists go 1.23; ...`
  - `go: module . listed in go.work file requires go >= 1.24.2, but go.work lists go 1.23; ...`
  - `go: module ../pinocchio listed in go.work file requires go >= 1.25.4, but go.work lists go 1.23; ...`
- Confirmed `go.work` was at `go 1.23` and `go env GOVERSION` didn’t match the module requirements.
- After the workspace was fixed externally (running `go work use .`), validation succeeded:
  - `go env GOVERSION` returned `go1.25.4`
  - `go test ./... -count=1` passed.

### Why
- Tests are the fastest sanity check before opening a PR; the workspace mismatch made that impossible until corrected.

### What worked
- Once `go.work` was updated, `go test` ran cleanly in `docmgr`.

### What didn't work
- The initial setup had an inconsistent `go.work` version relative to module requirements.

### What I learned
- In a multi-module workspace, `go.work` is part of the “build contract”; if it’s stale, it blocks even unrelated module testing.

### What was tricky to build
- Diagnosing this required reading multiple `go.mod` / `go.work` values and reconciling the version constraints manually.

### What warrants a second pair of eyes
- Confirm the intended policy: should the repo default to testing with workspace enabled, or should `Makefile` enforce `GOWORK=off`?

### What should be done in the future
- Consider a `docmgr`-local `make test` target that forces a consistent `GOWORK` mode (and prints the selected mode).

## Step 4: What I had to do manually (and what `prescribe` could provide)

This step summarizes the “manual toil” that `prescribe` could have eliminated. The impact is a concrete wishlist: features that would convert this workflow from “research-driven” to “tool-driven”.

### What I did manually
- Verified the correct diff base and change scope:
  - `git fetch origin main`
  - `git log --oneline origin/main..HEAD`
  - `git diff --stat origin/main...HEAD`
- Verified what `prescribe session show` *claims* vs what `.pr-builder/session.yaml` contains (a mismatch in target branch labeling occurred in prior runs).
- Manually inspected `.pr-builder/generated-pr.md` to locate the type mismatch.
- Attempted to craft more restrictive prompts and debug shell quoting issues when passing multiline prompts in `zsh`.

### What `prescribe` could offer (high leverage)
- **Strict schema mode**: `prescribe generate --schema pr.yaml` (or built-in schema) that rejects nonconforming output and re-prompts automatically.
- **Auto-normalization**:
  - Strip markdown code fences automatically.
  - Coerce `changelog` lists into newline-joined strings (or at least provide a fix-it preview).
- **Better parse errors**: line/col diagnostics pointing to the exact YAML token that failed.
- **First-class base selection**:
  - `prescribe session init --target origin/main` should be reflected consistently in `prescribe session show` (and show the resolved commit hash).
  - Include `merge-base`, commit count, and file count vs base in `session show`.
- **Raw vs parsed artifacts**:
  - Always write `raw-response.md` and a `parse-error.txt` alongside the session.
  - Optionally write a “best effort normalized YAML” file even on parse failure.
- **Built-in git summary command**:
  - `prescribe diff summary` to print `git log origin/main..HEAD`, `git diff --stat`, and the file list/token counts without needing separate git calls.

