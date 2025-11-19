---
Title: Intern Guide — Closing Workflow, Status/Intent, and LLM UX
Ticket: DOCMGR-CLOSE
Status: active
Topics:
    - docmgr
    - workflow
    - ux
    - automation
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Intern-friendly reference for implementing and using `ticket close`, status vocabulary, and dual-mode output in docmgr, with links to code and debates.
LastUpdated: 2025-11-19T15:00:34.331348291-05:00
---

# Intern Guide — Closing Workflow, Status/Intent, and LLM UX

## Purpose

This guide gives you the complete context to implement and use the new closing workflow in `docmgr`:
- Add and use `docmgr ticket close` (atomic close: update Status, optional Intent, changelog, LastUpdated)
- Treat Status as vocabulary‑guided; keep Intent vocabulary‑controlled
- Prefer human‑first output; enable structured output (`--with-glaze-output`) where it matters

It also links the most relevant code files so you can jump straight to implementation locations.

## Background

docmgr organizes documentation by ticket workspaces under `ttmp/` with YAML frontmatter. Important fields:
- `Status` — Workflow state (free-form historically)
- `Intent` — Longevity (vocabulary‑controlled; default typically `long-term`)
- `RelatedFiles` — Code references with optional notes

We ran 5 debate rounds and a synthesis. The conclusions:
- Introduce high-level `ticket close` under the `ticket/` command group
- Make the close operation atomic and human-first by default
- Add structured output where high value (lists, status, and `ticket close`)
- Guide Status via vocabulary (warn, don’t fail); Intent remains vocabulary‑controlled

## Command: `docmgr ticket close`

- Location (to add): `cmd/docmgr/cmds/ticket/close.go` (or similar under ticket/ group)
- Behavior:
  - Preconditions (minimal): check ticket exists; optionally check "all tasks done" (warn if not)
  - Update Status: default to `complete` (override via `--status`)
  - Update Intent: optional `--intent` (defaults from config or omitted)
  - Append a changelog entry (short, with reason)
  - Update `LastUpdated`
- Output:
  - Human (default): concise success message with the changes
  - Structured (optional): `--with-glaze-output --output json` returns a single JSON object: `{ ticket, all_tasks_done, operations: { status_updated, intent_updated, changelog_updated } }`

### Suggested flags

- `--status complete` — override default status for close
- `--intent long-term` — set intent explicitly during close
- `--changelog-entry "..."` — optional message; defaults to a standard "Ticket closed"
- `--with-glaze-output --output json` — structured output for LLMs/automation

## Status & Intent Guidance

- Recommended Status values (team-extensible): `draft, active, review, complete, archived`
- Suggested transitions (not enforced): `draft → active → review → complete → archived`; `review → active`; `complete → active` (reopen) should warn
- Intent stays vocabulary-controlled (default from `.ttmp.yaml` or `long-term`)

## Dual‑Mode Output

- Human-first output remains the default across commands
- Use `--with-glaze-output --output json|csv|yaml|table` where:
  - You need stable, parseable results (CI, LLM orchestration)
  - The volume is high (lists, search results, status)
  - You want a single atomic JSON object summarizing an operation (e.g., `ticket close`)

## Quick Examples

Human:
```bash
# Close a ticket (human-friendly)
docmgr ticket close --ticket DOCMGR-CLOSE
```

Structured:
```bash
# Close a ticket (structured output)
docmgr ticket close --ticket DOCMGR-CLOSE --with-glaze-output --output json | jq
```

Status vocabulary (managed by team):
```bash
# See current vocabulary entries
docmgr vocab list --with-glaze-output --output json

# Add a status value (if we extend vocabulary)
docmgr vocab add --category status --slug review --description "Ready for review"
```

## Implementation Roadmap

1) Add `ticket close` command
- Wire under `ticket/` in `cmd/docmgr/cmds/root.go`
- Implement CLI file in `cmd/docmgr/cmds/ticket/`
- Orchestrate changes via existing functions (`meta update`, `changelog update`)

2) Structured output for `ticket close`
- Implement `RunIntoGlazeProcessor` to emit `{ ticket, all_tasks_done, operations: {...} }`
- Keep `Run` for human output

3) Status vocabulary (warnings)
- Extend `vocabulary.yaml` with a `status:` section (draft/active/review/complete/archived as seeds)
- Update `doctor` to warn on unknown Status

4) UX improvements (later, optional)
- On `tasks check`, print actionable suggestion when all tasks complete
- Consider `--with-glaze-output` for `tasks check` to expose `all_tasks_done`

## Related Files (code)

- `pkg/commands/create_ticket.go` — sets ticket defaults (Status, Intent)
- `pkg/commands/add.go` — doc creation; inherits Status/Intent, templates
- `pkg/commands/meta_update.go` — updates fields; used by `ticket close` wrapper
- `pkg/commands/tasks.go` — tasks list/check; place for actionable suggestions
- `pkg/commands/list_tickets.go` — `countTasksInTicket`: open/done counts
- `pkg/commands/status.go` — displays Status and staleness
- `pkg/models/document.go` — Document fields, Vocabulary types
- `internal/workspace/config.go` — config defaults for Intent/root
- `cmd/docmgr/cmds/root.go` — command tree wiring (`ticket`, `doc`, etc.)
- `ttmp/vocabulary.yaml` — current vocabulary file

## Reference Documents

- Synthesis: `reference/06-debate-synthesis-closing-workflow-status-intent-and-llm-ux.md`
- Rounds 1–5: see `analysis/12-…16-…` under `DOCMGR-CODE-REVIEW`
- CLI verbs mapping: `DOCMGR-DOC-VERBS/.../reference/01-cli-verbs-mapping-old-vs-new.md`

## FAQ

- Q: Do I need structured output for everything?
  - A: No. LLMs generally handle human output. Use structured output when parsing is valuable (large lists, CI, orchestration, and atomic ops like close).
- Q: Should `ticket close` enforce "all tasks done"?
  - A: Prefer warning + explicit override for flexibility. Enforced policies can come later if the team wants it.
- Q: What about `ticket reopen` / `ticket archive`?
  - A: Good follow-ups. Start with `close` and gather feedback.
