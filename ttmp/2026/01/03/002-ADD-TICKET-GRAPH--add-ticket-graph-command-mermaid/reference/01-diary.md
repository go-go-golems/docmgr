---
Title: Diary
Ticket: 002-ADD-TICKET-GRAPH
Status: active
Topics:
    - docmgr
    - cli
    - tooling
    - diagnostics
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T15:01:48.927111734-05:00
---

# Diary

## Goal

This diary captures the research and writing process for designing a new `docmgr` command that generates a **Mermaid graph** for a ticket: the ticket’s documents plus the code files referenced via `RelatedFiles`, with an option to expand the graph transitively. It records the discovery trail (files read, commands run) and highlights tricky semantics and review risks.

## Step 1: Create the ticket and set up the docs that will drive the feature

This step created the docmgr ticket workspace and the two documents I’ll maintain while researching and specifying the “ticket graph” command. The output of this step is the scaffolding: a diary for the process and a reference guide that will become the exhaustive design/implementation write-up.

The overarching goal is to propose a command that can be implemented cleanly in the current `docmgr` architecture: cobra + glazed dual-mode commands, workspace discovery, and the in-memory SQLite index (`Workspace.InitIndex` + `Workspace.QueryDocs`). I will keep the spec grounded in what the current index can already answer (docs by ticket, related_files reverse lookup) and what would need to change for deeper transitive graphs.

### What I did
- Created ticket `002-ADD-TICKET-GRAPH`:
  - `GOWORK=off go run ./cmd/docmgr ticket create-ticket --ticket 002-ADD-TICKET-GRAPH --title "Add ticket graph command (Mermaid)" --topics docmgr,cli,tooling,diagnostics`
- Created two reference docs under the ticket:
  - `reference/01-diary.md` (this file)
  - `reference/02-ticket-graph-mermaid-design-and-implementation-guide.md` (the exhaustive guide)

### Why
- The ticket provides a durable home for the spec, diary trail, and future implementation notes.
- Having the guide doc in place early helps me write continuously while researching (instead of doing a big write-up at the end).

### What worked
- Ticket creation and doc creation succeeded under the repo’s `.ttmp.yaml` configuration (`docmgr/ttmp` root).

### What didn't work
- N/A.

### What I learned
- This repo still requires running `docmgr` as `cd docmgr && GOWORK=off go run ./cmd/docmgr ...` due to `go.work` toolchain constraints at the repo root.

### What was tricky to build
- N/A (scaffolding step).

### What warrants a second pair of eyes
- N/A (no code changes yet).

### What should be done in the future
- N/A (this step is just scaffolding).

### Code review instructions
- N/A.

### Technical details
- Ticket root:
  - `docmgr/ttmp/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid/`

## Step 2: Inspect existing ticket/doc query APIs and identify the right integration point

This step mapped where a “ticket graph” command should live in the CLI tree and what internal APIs it should rely on. The main outcome is a concrete implementation strategy: implement a new glazed dual-mode command under `docmgr ticket`, backed by `Workspace.QueryDocs` and the existing ticket index discovery helper patterns.

I focused on two questions:

1) How do we reliably locate a ticket and enumerate the Markdown documents “in the ticket”?
2) How do we turn those documents’ `RelatedFiles` into a stable set of file nodes and optionally expand the graph transitively?

### What I did
- Read the ticket command wiring (cobra attach points):
  - `docmgr/cmd/docmgr/cmds/ticket/ticket.go`
  - `docmgr/cmd/docmgr/cmds/ticket/list.go`
  - `docmgr/cmd/docmgr/cmds/ticket/create.go`
- Read document enumeration patterns that already use `Workspace.QueryDocs`:
  - `docmgr/pkg/commands/list_docs.go`
  - `docmgr/pkg/commands/list_tickets.go` (especially `queryTicketIndexDocs`)
- Re-anchored on the workspace/index/query primitives used everywhere:
  - `docmgr/internal/workspace/workspace.go`
  - `docmgr/internal/workspace/index_builder.go`
  - `docmgr/internal/workspace/query_docs.go`
  - `docmgr/internal/workspace/query_docs_sql.go`
  - `docmgr/internal/workspace/sqlite_schema.go`
- Re-checked how path normalization works for “file-ish” nodes:
  - `docmgr/internal/paths/resolver.go`

### Why
- A “ticket graph” command should not re-walk the filesystem or hand-parse frontmatter. The repo already has a canonical index-backed API (`Workspace.QueryDocs`) designed for consistent skip rules, parse error handling, and reverse lookup semantics.

### What worked
- There is a clean place to attach the command: add `graph` as a subcommand under `docmgr ticket` alongside `create`, `list`, `rename`, etc.
- Ticket lookup can follow the existing `list tickets` model: find the ticket’s `index.md` via `QueryDocs` with `DocType=index`, then derive the ticket directory from the index path.

### What didn't work
- N/A (research step).

### What I learned
- `list tickets` already has a helper (`queryTicketIndexDocs`) that:
  - discovers the workspace,
  - initializes the index,
  - queries `DocType=index` docs,
  - derives the absolute ticket directory.
  The ticket graph command can either reuse this function (if we move it to a shared package) or replicate the same “index.md for ticket” resolution pattern (as `doc relate` does today).

### What was tricky to build
- The definition of “files in a ticket” is ambiguous unless we specify it carefully:
  - ticket documents (markdown files) under the ticket directory,
  - control docs (README/tasks/changelog) are tagged as control docs but still real markdown under the ticket,
  - archived docs live under `archive/` and may or may not be included,
  - “files” could also mean code files referenced via `RelatedFiles`.
  The final spec needs explicit defaults and flags for inclusion/exclusion.

### What warrants a second pair of eyes
- Decide (and document) the intended default scope for graph expansion:
  - ticket-only graph vs repo-wide “knowledge graph” expansion via reverse lookup.
  This impacts performance, usability, and surprise factor.

### What should be done in the future
- N/A for this research step (implementation work will come in later steps).

### Code review instructions
- N/A (no code changes in this step).

### Technical details
- Candidate command placement: `docmgr ticket graph`
- Candidate data source: `Workspace.QueryDocs` with `ScopeTicket` + `IncludeErrors=false` + `IncludeBody=false`

## Step 3: Draft the ticket graph semantics and the “transitive expansion” contract

This step turned the vague feature request (“graph of all the files in a ticket and related files”) into an explicit contract with clear defaults and predictable behavior. The main deliverable is a precise definition of nodes, edges, and what “transitive” means in the context of docmgr (which only knows about doc↔file links, not code↔code dependencies).

I chose to treat the base graph as bipartite: ticket documents (markdown) and related code files (from `RelatedFiles`). “Transitive” expansion then naturally means: from those file nodes, discover *other docs* (optionally across the whole repo) that also reference those files, and optionally continue outward in BFS layers.

### What I did
- Defined a base graph (depth 0) and a transitive expansion model (depth ≥ 1).
- Identified the data queries needed at each expansion step and how to batch them with existing `QueryDocs` semantics (OR matching for `RelatedFile`).
- Identified the output constraints for Mermaid:
  - stable node IDs,
  - safe labels,
  - reasonable edge labeling (notes can be long),
  - avoiding “graph explosion” via max-nodes / max-edges.

### Why
- Without a precise “transitive” definition, the command could easily become either useless (no expansion) or dangerous (repo-wide explosion).
- Mermaid syntax is simple, but it’s very easy to generate invalid graphs if you don’t treat node IDs and labels as a formal output contract.

### What worked
- A BFS-by-layer model maps cleanly to docmgr’s existing index capabilities:
  - enumerate docs in ticket (ScopeTicket),
  - gather related files from those docs,
  - repo-scope query for “docs referencing any of these files” (ScopeRepo + RelatedFile filter).

### What didn't work
- N/A (design step).

### What I learned
- `Workspace.QueryDocs` already implements OR semantics for `RelatedFile` filters, which is exactly what we need for “batch expand N files at a time”.

### What was tricky to build
- Preventing runaway expansion while still providing a useful transitive feature:
  - you need a visited set for both docs and files,
  - you need limits and/or depth controls,
  - and you need to decide whether to expand from newly discovered docs’ related files (which can balloon quickly).

### What warrants a second pair of eyes
- The exact default behavior for transitive expansion:
  - Should depth>0 default to “include external docs but do not expand their files” (safer),
  - or “fully expand outward” (more powerful but riskier)?

### What should be done in the future
- N/A (implementation will address performance + limits).

### Code review instructions
- N/A.

### Technical details
- Proposed defaults:
  - `--depth 0` (ticket-only docs + their related files)
  - `--scope repo` only when depth>0 is explicitly set
  - `--max-nodes` / `--max-edges` required for large expansions

## Step 4: Write the exhaustive guide and attach core implementation references

This step produced the main deliverable for the ticket: a verbose design-and-implementation guide for building `docmgr ticket graph`. The document covers CLI design, graph semantics, Mermaid output contracts, transitive expansion algorithms, and a concrete testing plan. It is intentionally written to be “implementation-ready”: someone can pick it up and start coding with clear seams and guardrails.

I also related the most important code files (workspace query/index + path resolver + ticket command wiring) to the guide document so future readers can jump from code to spec using `docmgr doc search --file ...`.

### What I did
- Wrote the full guide:
  - `reference/02-ticket-graph-mermaid-design-and-implementation-guide.md`
- Related core files to the guide (index/query/paths + ticket attach point):
  - `internal/workspace/query_docs.go`
  - `internal/workspace/query_docs_sql.go`
  - `internal/workspace/index_builder.go`
  - `internal/workspace/sqlite_schema.go`
  - `internal/paths/resolver.go`
  - `pkg/commands/list_tickets.go`
  - `cmd/docmgr/cmds/ticket/ticket.go`

### Why
- The request was for a “similarly exhaustive” document to the earlier search guide, with diagrams, pseudocode, API contract notes, and extension guidance.

### What worked
- The guide can be read as:
  - a spec (what the command should do),
  - an implementation plan (where to add code and what APIs to call),
  - and a safety checklist (limits + escaping + scope controls).

### What didn't work
- N/A.

### What I learned
- Mermaid output is deceptively strict: you need a clear ID/label escaping policy up front or you will generate graphs that fail to render for real-world paths and notes.

### What was tricky to build
- Defining “transitive expansion” in a docmgr-native way (doc↔file links only) while keeping the feature safe and bounded for large workspaces.

### What warrants a second pair of eyes
- The proposed default interactions between `--depth`, `--scope`, and `--expand-files` (this is the most UX- and performance-sensitive part of the spec).

### What should be done in the future
- When implementing, add a small scenario fixture with two tickets referencing one shared file to validate repo-scope depth expansion without relying on a huge real workspace.

### Code review instructions
- Start in `docmgr/ttmp/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid/reference/02-ticket-graph-mermaid-design-and-implementation-guide.md`.
- Follow the “Where this command should live” section and mirror patterns used by `docmgr ticket list` and `docmgr doc search`.

### Technical details
- Guide doc path:
  - `docmgr/ttmp/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid/reference/02-ticket-graph-mermaid-design-and-implementation-guide.md`

## Step 5: Upload diary + ticket graph guide to reMarkable

This step published the two main documents for this ticket (the diary and the ticket-graph implementation guide) to the reMarkable device for offline reading and annotation. The uploader converts Markdown to PDF (stripping YAML frontmatter) and then uploads via `rmapi` into a dated folder under `ai/YYYY/MM/DD/`, mirroring the ticket directory structure to avoid collisions.

### What I did
- Ran a dry-run upload to confirm the on-device destination and the exact commands.
- Uploaded:
  - `reference/01-diary.md` → `01-diary.pdf`
  - `reference/02-ticket-graph-mermaid-design-and-implementation-guide.md` → `02-ticket-graph-mermaid-design-and-implementation-guide.pdf`

### Why
- The guide is long and benefits from the reMarkable “read/annotate” workflow.

### What worked
- The final upload succeeded and both PDFs landed under:
  - `ai/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid/reference/`

### What didn't work
- A first attempt to upload the guide timed out in this environment after successfully uploading the diary PDF. Re-running the upload for the guide alone with a longer command timeout succeeded.

### What I learned
- For large markdown documents, it’s worth uploading each file independently if you hit timeouts, because the conversion+upload pipeline is per-file and you can avoid repeating work.

### What was tricky to build
- N/A (publish step; tooling already exists).

### What warrants a second pair of eyes
- N/A.

### What should be done in the future
- N/A.

### Code review instructions
- N/A.

### Technical details
- Ticket dir:
  - `/home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid`
- Commands:
  - Dry run:
    - `python3 /home/manuel/.local/bin/remarkable_upload.py --ticket-dir /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid --mirror-ticket-structure --dry-run /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid/reference/01-diary.md /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid/reference/02-ticket-graph-mermaid-design-and-implementation-guide.md`
  - Upload:
    - `python3 /home/manuel/.local/bin/remarkable_upload.py --ticket-dir /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid --mirror-ticket-structure /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid/reference/01-diary.md /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid/reference/02-ticket-graph-mermaid-design-and-implementation-guide.md`

## Step 6: Create implementation tasks and prepare to implement `docmgr ticket graph`

This step translated the guide into an actionable task list inside the ticket workspace and set up the expected commit discipline for the implementation phase. The main outcome is a scoped checklist that matches the incremental implementation plan: wire the command, implement depth=0, then add transitive expansion, tests, and validation runs.

**Commit (docs):** 1a7d7bbec00ca8e34a4509cbd8b7002c6776c9d6 — "Docs: add ticket graph spec and tasks"

### What I did
- Read `~/.cursor/commands/git-commit-instructions.md` and committed to following the “diff → stage specific files → commit message → record hash” workflow for each implementation step.
- Added ticket tasks using `docmgr task add` and removed the scaffold placeholder task:
  - Added tasks covering CLI contract, skeleton wiring, depth=0 graph, transitive expansion, tests, validation, and doc/remarkable updates.
  - Removed the default “Add tasks here” task via `docmgr task remove`.

### Why
- The user explicitly requested: create tasks in the ticket, commit in individual steps, validate by running the graph command itself, and keep an implementation diary.

### What worked
- The tasks now reflect the intended implementation order and provide checkboxes that can be updated after each commit.

### What didn't work
- N/A.

### What I learned
- `docmgr task remove` is the clean way to remove scaffold placeholder tasks instead of editing `tasks.md` by hand.

### What was tricky to build
- N/A.

### What warrants a second pair of eyes
- N/A (no code yet).

### What should be done in the future
- As we implement, each commit should check off tasks incrementally and update this diary with:
  - exact validation commands,
  - the commit hash,
  - and the key review risks introduced in that step.

### Code review instructions
- Start by reading the current tasks list:
  - `docmgr/ttmp/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid/tasks.md`

### Technical details
- Commands used:
  - `cd docmgr && GOWORK=off go run ./cmd/docmgr task add --ticket 002-ADD-TICKET-GRAPH --text "..."`
  - `cd docmgr && GOWORK=off go run ./cmd/docmgr task remove --ticket 002-ADD-TICKET-GRAPH --id 1`

## Step 7: Implement `docmgr ticket graph` (depth=0 Mermaid output)

This step introduced the first working version of the new command: `docmgr ticket graph`. It outputs a Mermaid flowchart representing the bipartite graph between ticket docs and the code files referenced via `RelatedFiles`. The command supports a minimal but useful CLI contract (ticket selection, basic formatting controls, include/exclude knobs) and can also emit a structured edge list in glazed mode.

I also fixed a practical development workflow blocker: the docmgr repo lives inside a larger workspace that contains a parent `go.work`. Pre-commit hooks run `make test` and `make lint`, and those commands were unintentionally picking up the parent `go.work` and failing. Exporting `GOWORK=off` from the docmgr `Makefile` makes the repo self-contained for `go test` and `golangci-lint` during hooks.

**Commit (code):** e473c1c494d1a1c44b317d49eb42ce5db32ef892 — "Ticket: add ticket graph command (mermaid, depth 0)"

### What I did
- Added cobra wiring for the new command:
  - `cmd/docmgr/cmds/ticket/graph.go`
  - `cmd/docmgr/cmds/ticket/ticket.go`
- Implemented the depth=0 graph builder + Mermaid renderer:
  - `pkg/commands/ticket_graph.go`
- Ensured docmgr’s `make test`/`make lint` don’t accidentally use the parent workspace `go.work`:
  - `Makefile` now exports `GOWORK=off`

### Why
- The base ticket graph (no transitive expansion) is the smallest “end-to-end useful” slice: it exercises workspace discovery, doc enumeration, `RelatedFiles` normalization, and Mermaid rendering.
- Fixing the hook environment was necessary to support the user request to commit in small steps while keeping pre-commit validation enabled.

### What worked
- `docmgr ticket graph` produces a valid Mermaid graph for both:
  - `--ticket 002-ADD-TICKET-GRAPH`
  - `--ticket 001-ADD-DOCMGR-UI` (acts as a “historical ticket” check)
- Pre-commit hooks now pass because `make test` and `make lint` run with `GOWORK=off`.

### What didn't work
- The first commit attempt failed because pre-commit hooks ran in an environment that picked up the parent `go.work`, causing `go test`/`golangci-lint` to error. The `Makefile` export fix resolved this.

### What I learned
- When a repo is nested inside a larger mono-workspace, `go` will happily discover and use a parent `go.work` unless you explicitly disable it. This can break “self-contained module” repos in CI hooks.

### What was tricky to build
- Mermaid output contracts:
  - Node IDs must be Mermaid-safe and stable; I used a short SHA1-based identifier.
  - Labels and edge notes need sanitization (newlines, quotes, `|`) and truncation to avoid invalid syntax and unreadable graphs.

### What warrants a second pair of eyes
- Confirm the current label/escaping policy is sufficient for real-world paths and notes (especially edge labels that can contain punctuation and newlines).
- Confirm whether including control docs by default is the desired UX (currently default true).

### What should be done in the future
- Add transitive expansion and safety limits (depth/scope/batching) as described in the guide, with tests that prevent graph explosion regressions.

### Code review instructions
- Start at `pkg/commands/ticket_graph.go` for the core behavior.
- Verify wiring in `cmd/docmgr/cmds/ticket/ticket.go`.
- Validate locally:
  - `cd docmgr && GOWORK=off go test ./...`
  - `cd docmgr && GOWORK=off go run ./cmd/docmgr ticket graph --ticket 002-ADD-TICKET-GRAPH --format mermaid`

## Step 8: Add repo-scope transitive expansion (depth/scope/batching/limits)

This step extended the ticket graph command to support transitive expansion across the repository: starting from a ticket’s related files, it can now discover other documents (potentially in other tickets) that reference those same files. This is the “doc knowledge graph” style expansion described in the ticket guide, implemented as a bounded BFS over doc↔file edges.

The most important engineering constraint here is safety: repo-wide expansion can balloon quickly. I enforced explicit `--scope repo` for `--depth > 0` and added hard limits (`--max-nodes`, `--max-edges`) plus batched reverse-lookup queries (`--batch-size`) so the algorithm scales predictably.

**Commit (code):** 2ee7273b457631b65c5236a0f3f569f63118f391 — "Ticket: add transitive expansion to ticket graph"

### What I did
- Added CLI flags to control expansion and safety:
  - `--depth`, `--scope`, `--expand-files`, `--batch-size`, `--max-nodes`, `--max-edges`
- Refactored the implementation into a graph builder that:
  - builds the base ticket graph (depth 0),
  - then expands outward layer-by-layer using `Workspace.QueryDocs` with `ScopeRepo` and `Filters.RelatedFile`.

### Why
- The original request explicitly asked for a “transitive graph thing”.
- Docmgr already has robust reverse lookup via the indexed `related_files` table; transitive expansion is a natural reuse of that capability.

### What worked
- Running a repo-scope depth=1 graph for this ticket succeeds and produces a large but valid Mermaid graph:
  - `cd docmgr && GOWORK=off go run ./cmd/docmgr ticket graph --ticket 002-ADD-TICKET-GRAPH --scope repo --depth 1 --expand-files=false --format mermaid`
- Pre-commit hooks still pass (tests + lint).

### What didn't work
- N/A (no new blockers beyond the already-addressed go.work hook issue).

### What I learned
- Using `QueryDocs` with `RelatedFile: []string{...}` is an efficient way to implement “OR of many files” reverse lookup, especially when you batch the query inputs.

### What was tricky to build
- Correctly defining expansion semantics that don’t explode:
  - When `--expand-files=false`, newly discovered docs only connect to the triggering file batch (instead of pulling in all their related files).
  - When `--expand-files=true`, you must use a visited set and strict budgets (`--max-nodes/--max-edges`) to keep the BFS bounded.

### What warrants a second pair of eyes
- Confirm the safety contracts are acceptable UX:
  - hard failure when budgets are exceeded (instead of partial output),
  - requiring `--scope repo` for `--depth > 0`.

### What should be done in the future
- Add targeted tests and a small scenario fixture for transitive expansion so semantics don’t drift (this is the next task).

### Code review instructions
- Start in `pkg/commands/ticket_graph.go`:
  - `buildTicketGraph` (argument validation + wiring),
  - `ticketGraphBuilder.expandTransitive` (BFS + batching + limits),
  - `ticketGraphBuilder.addDocAndEdges` (trigger-file filtering vs expand-files behavior).

## Step 9: Add tests for ticket graph behavior (depth=0, transitive, sanitization)

This step added a small but meaningful test suite around the ticket graph command. The tests focus on the most failure-prone parts of the feature:

- Mermaid label sanitization (newlines, quotes, pipes, brackets).
- Stable node IDs (`shortHash`) so graphs don’t churn across runs.
- A minimal workspace fixture that validates:
  - depth=0 outputs expected doc and file nodes, and
  - repo-scope transitive expansion (depth=1) discovers an “external” ticket document that references the same file.

**Commit (code):** c470912857eaba220f35a9dfc108617aad7a2044 — "Test: add ticket graph fixtures and unit tests"

I intentionally avoided testing “error printing” through the cobra wrapper because the glazed CLI wrapper can call `os.Exit` on errors in some configurations; for the “depth requires repo scope” contract I validate the underlying builder function directly.

### What I did
- Added `pkg/commands/ticket_graph_test.go`:
  - unit tests for `sanitizeMermaidLabel` and `shortHash`
  - end-to-end-ish tests that build a temporary repo (`go.mod` + `ttmp/`) and run `ticket graph` via the same `common.BuildCommand` configuration the real CLI uses.
- Ran:
  - `GOWORK=off go test ./...`

### Why
- The ticket graph command is output-heavy and easy to break subtly (invalid Mermaid, unstable output, incorrect expansion). Tests reduce the risk of regressions as we iterate on transitive expansion and output formatting.

### What worked
- The fixture-based transitive test proves `--scope repo --depth 1` discovers another ticket’s doc via shared `RelatedFiles` entries.

### What didn't work
- A first attempt to test the “depth requires repo scope” error via cobra printed an error and terminated the test run; switching that one check to call `buildTicketGraph(...)` directly avoided the issue.

### What I learned
- For commands built on glazed’s cobra integration, “expected error cases” are sometimes best tested one layer down (command logic) to avoid CLI wrappers that may exit the process.

### What was tricky to build
- Building a minimal workspace fixture required:
  - a `go.mod` so `FindRepositoryRoot()` resolves the repo root, and
  - real code files on disk so `paths.Resolver.Normalize` can produce repo-relative canonical keys.

### What warrants a second pair of eyes
- Confirm that importing `cmd/docmgr/cmds/common` in a `pkg/commands` test is an acceptable pattern in this repo (it avoids duplicating the CLI wiring logic but crosses “cmd → pkg” layering in tests).

### What should be done in the future
- Add a test for budget enforcement (`--max-nodes/--max-edges`) once we settle on whether the command should hard-fail or produce partial output when limits are exceeded.

### Code review instructions
- Start in `pkg/commands/ticket_graph_test.go` and run `GOWORK=off go test ./pkg/commands -run TestTicketGraph -v`.

## Step 10: Upload diary + guide PDFs to reMarkable

This step pushed the latest versions of the ticket diary and the design/implementation guide to the reMarkable so I can review the documents away from the workstation. The upload also serves as a quick “does pandoc still render this cleanly?” check for long Markdown documents.

Because I had already uploaded earlier versions, the first non-`--force` run failed with “entry already exists”; rerunning with `--force` replaced the PDFs successfully.

### What I did
- Dry-run upload to confirm the remote destination paths and PDF names:
  - `python3 /home/manuel/.local/bin/remarkable_upload.py --ticket-dir /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid --mirror-ticket-structure --dry-run ...`
- Upload with overwrite enabled:
  - `python3 /home/manuel/.local/bin/remarkable_upload.py --force --ticket-dir /home/manuel/workspaces/2026-01-03/add-docmgr-webui/docmgr/ttmp/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid --mirror-ticket-structure ...`

### Why
- Keep a readable, annotated copy of the docs on the reMarkable (diary + guide).
- Confirm that the Markdown converts to PDF cleanly (frontmatter stripping + pandoc/xelatex path).

### What worked
- Both PDFs uploaded to:
  - `ai/2026/01/03/002-ADD-TICKET-GRAPH--add-ticket-graph-command-mermaid/reference/`

### What didn't work
- The initial upload without `--force` failed with:
  - `Error: entry already exists (use --force to recreate, --content-only to replace content)`

### What I learned
- When iterating on the same ticket docs, I should expect collisions and either:
  - upload into a new remote ticket root, or
  - explicitly use `--force` once I’ve confirmed the destination is correct via `--dry-run`.

### What was tricky to build
- N/A (no implementation work here; this was a publishing step).

### What warrants a second pair of eyes
- N/A.

### What should be done in the future
- N/A.

### Technical details
- Uploader script: `python3 /home/manuel/.local/bin/remarkable_upload.py`
- Rendering path: `pandoc` → `xelatex` (DejaVu fonts)
