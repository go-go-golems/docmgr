---
Title: Design — UX improvements (scriptability and path resolver)
Ticket: DOCMGR-UX
Status: active
Topics:
    - tooling
    - ux
    - cli
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/pkg/commands/list_docs.go
      Note: Docs columns and filtering
    - Path: docmgr/pkg/commands/list_tickets.go
      Note: Tickets columns and row assembly
    - Path: docmgr/pkg/commands/tasks.go
      Note: Tasks listing and mutations
    - Path: docmgr/pkg/commands/vocab_list.go
      Note: Vocabulary listing and columns
ExternalSources: []
Summary: Design updated (doc-type test)
LastUpdated: 2025-11-05T14:44:55.696297566-05:00
---




# Design — UX improvements (scriptability and path resolver)

## Executive Summary
Make docmgr easier to script by removing friction around common operations while leveraging existing Glazed flags and outputs.

## Problem Statement
Scripting common flows (get ticket dir, get index.md path, get a doc path) currently requires:
1) Remembering `--with-glaze-output` and choosing an output format
2) Parsing JSON/CSV
3) Mapping names to file paths externally

This is error-prone and verbose for frequent operations.

## Proposed Solution
MVP additions:

1) Metadata defaulting for index
   - `docmgr meta update --ticket T --field Owners --value "manuel"` → targets `index.md` if `--doc` omitted

## Design Decisions
– Keep table output as default for humans; require explicit flags for machine modes
– Prefer existing Glazed primitives (`--select`, `--fields`, `--filter`, `--output`) for scripting
– Fail fast with actionable errors when ticket/doc not found

## Alternatives Considered
– Rely solely on `--with-glaze-output` + JSON and ask users to install `jq`
  Rejected: raises barrier for simple shell usage; CSV parsing is brittle.
– Auto-detect piping to switch to machine output
  Rejected: implicit behavior can surprise users; prefer explicit flags.

## Implementation Plan
1) Update `meta update` to default to index when `--ticket` is set and `--doc` omitted
2) Tests and docs: usage examples and help text

## Additional UX enhancement proposals (beyond MVP)

1) Default Glaze activation
- Auto-enable Glazed when any Glazed flag is present (`--select`, `--fields`, `--filter`, `--output=json|yaml|csv|tsv`) or when `--output != table`.
- Keep `--with-glaze-output` for explicit control.
- Acceptance:
  - `list docs --output json` emits JSON without the toggle.
  - `list docs --select path` emits values without the toggle.
  (Note: Glazed-level; out of scope for docmgr changes.)

2) Shorthand flags
- `--json`, `--yaml`, `--csv`, `--tsv` map to `--output json|yaml|csv|tsv` (+ enable Glaze).
- `--paths` maps to `--select path` (+ enable Glaze).
- Acceptance: Shorthands match long-form outputs byte-for-byte.
  (Note: Glazed-level; out of scope for docmgr changes.)

3) Stable column contracts (document + guarantee)
- Where columns are defined today:
  - `docmgr/pkg/commands/list_tickets.go` → `RunIntoGlazeProcessor`: fields `ticket,title,status,topics,path,last_updated`
  - `docmgr/pkg/commands/list_docs.go`    → `RunIntoGlazeProcessor`: fields `ticket,doc_type,title,status,topics,path,last_updated`
  - `docmgr/pkg/commands/tasks.go`        → `TasksListCommand.RunIntoGlazeProcessor`: fields `index,checked,text,file`
  - `docmgr/pkg/commands/vocab_list.go`   → `RunIntoGlazeProcessor`: fields `category,slug,description`
- Proposed docmgr changes:
  - Introduce field-name constants in `docmgr/pkg/models/fields.go` (or `pkg/commands/constants.go`) and use in all row builders to avoid drift.
  - Add unit tests asserting headers for each command (CSV with headers) match the documented list.
  - Update user docs and `--help` long text to include an explicit "Columns" section (see item 4).
- Acceptance:
  - CSV headers for tickets/docs/tasks/vocab match constants and docs; tests cover regressions.

4) Schema discovery in help
- Where to edit:
  - `docmgr/pkg/commands/list_tickets.go` (`NewListTicketsCommand`),
  - `docmgr/pkg/commands/list_docs.go` (`NewListDocsCommand`),
  - `docmgr/pkg/commands/tasks.go` (`NewTasksListCommand`),
  - `docmgr/pkg/commands/vocab_list.go` (`NewVocabListCommand`).
- Proposed docmgr changes:
  - Extend `cmds.WithLong` to include a "Columns:" block listing the stable fields for each command, sourced from shared constants (item 3) to keep help in sync with code.
  - Optionally add a `--print-columns` boolean flag per command that prints the columns and exits (thin wrapper; not Glazed-level).
- Acceptance:
  - `docmgr list docs --help` shows a "Columns" section with `ticket,doc_type,title,status,topics,path,last_updated`.
  - If implemented, `docmgr list docs --print-columns` emits the same list and exits 0.

5) Help examples for scripting
- Where to edit:
  - Same files as item 4; add examples to `cmds.WithLong` under an "Examples (scriptable):" section.
- Proposed examples to embed:
  - Paths only: `--with-glaze-output --select path`
  - CSV no headers: `--with-glaze-output --output csv --with-headers=false --fields path`
  - TSV subset: `--with-glaze-output --output tsv --fields doc_type,title,path`
  - Templated lines: `--with-glaze-output --select-template '{{.doc_type}} {{.title}}: {{.path}}' --select _0`
- Acceptance:
  - `--help` includes copy/paste examples for each list command; examples verified in tests/docs.

6) Quiet/machine mode
- Scope: docmgr-level (not Glazed). Applies to bare-mode messages and result-empty signaling.
- Proposed docmgr changes:
  - Add flags to affected commands (`tickets`, `docs`, `tasks list`, `vocab list`): `--quiet` (bool), `--require-results` (bool).
  - In `RunIntoGlazeProcessor`, count added rows (local counter). If `--require-results` and count==0, return a well-typed error to yield non-zero exit.
  - In `Run` (bare mode), honor `--quiet` by suppressing human-readable prints; Glazed mode is already quiet by default (data-only).
  - Centralize helpers in `docmgr/pkg/commands/internal/util.go` (e.g., `ResultCounter` and `ShouldBeQuiet(pl)`), imported by list commands.
- Files to touch:
  - `docmgr/pkg/commands/list_tickets.go`, `list_docs.go`, `tasks.go` (TasksListCommand), `vocab_list.go` (add flags, row counting, quiet checks).
- Acceptance:
  - With `--require-results`, exit non-zero when no rows match filters.
  - With `--quiet`, bare-mode prints are suppressed; Glazed output unaffected.

8) Shell completions
- Provide completions for `--ticket` and `--doc_type` (bash/zsh/fish).
  (Note: Glazed-level for completion plumbing; docmgr contributes candidates.)


## Acceptance checklist (for proposals)
- Unit/integration tests added
- `--help` updated with examples and columns
- Docs updated (how-to + embedded help)
- Backward compatibility validated (flags, outputs)

## Open Questions
– Should `--output path` support multiple columns (e.g., `--fields path,title`)? Probably out of scope for MVP.
– Collision rules for `--doc` when multiple docs share the same title?

## References
– `docmgr/cmd/docmgr/main.go`
– `glazed/pkg/doc/tutorials/build-first-command.md`
