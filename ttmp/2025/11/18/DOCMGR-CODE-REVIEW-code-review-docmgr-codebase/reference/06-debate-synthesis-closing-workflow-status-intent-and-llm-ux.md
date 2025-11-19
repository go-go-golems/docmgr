---
Title: Debate Synthesis — Closing Workflow, Status/Intent, and LLM UX
Ticket: DOCMGR-CODE-REVIEW
Status: active
Topics:
    - docmgr
    - code-review
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Synthesis of debates R1–R5 with decisions, intern onboarding to docmgr concepts, and the proposed workflow improvements (ticket close, status/intent guidance, dual-mode output).
LastUpdated: 2025-11-19T14:46:48.257139557-05:00
---

# Debate Synthesis — Closing Workflow, Status/Intent, and LLM UX

## Executive Summary

- The current “close a ticket” workflow requires 3–5 commands and relies on non-enforced reminders. This creates cognitive load and inconsistent outcomes.
- We will add a high-level command `docmgr ticket close` that atomically updates status, (optionally) intent, and changelog with clear, human-first output and optional structured output for automation.
- Status will become vocabulary-guided (with warnings for unknown values), while Intent remains vocabulary-controlled; transitions are suggested, not enforced.
- For LLMs, complete the dual-mode output pattern only where it pays off most (list/search/status and new `ticket close`). LLMs generally handle human output well; structured output is most valuable for programmatic bulk and list-style data.

## Key Decisions by Round

- Round 1 (Friction): Too many steps to close; reminders aren’t actionable; tasks/status/changelog are disconnected → Add a single, atomic close operation.
- Round 2 (Verbs): Introduce `ticket close` under `ticket/` (consistent with `create-ticket`, `rename`); implement as a thin orchestrator that calls existing capabilities; provide optional structured output.
- Round 3 (Lifecycle):
  - Treat Status as vocabulary-guided (teams can add values; unknowns warn, not fail).
  - Keep Intent vocabulary-controlled; default to `long-term` (configurable in `.ttmp.yaml`).
  - Offer suggested (not enforced) transitions (draft→active→review→complete→archived).
- Round 4 (Automation): Prefer explicit commands with smart defaults over hidden automation. Detect-and-suggest when all tasks are done; consider an opt-in `--auto-close` only later.
- Round 5 (LLM UX): Maintain unified commands with human-first defaults; add structured output where it matters (lists, `ticket close`), and include actionable state (e.g., `all_tasks_done`) when structured mode is enabled.

## The Winning Architecture

- High-level, atomic close action
  - Command: `docmgr ticket close --ticket <ID> [--status complete] [--intent long-term]`.
  - Behavior: Checks preconditions (e.g., all tasks done if required), updates `Status`, optionally `Intent`, appends a changelog entry, updates `LastUpdated`.
  - Output: Human-friendly by default; add `--with-glaze-output --output json` for scripts/LLMs to receive a single JSON object with `operations` and `state`.
- Vocabulary and transitions
  - Vocabulary.yaml gains a `status` category teams can extend; doctor warns on unknown status.
  - Suggested transitions are documented; commands do not hard-enforce them.
- Dual‑mode output philosophy
  - Human-first outputs remain default.
  - Structured output prioritized where volume matters (list/search/status) and for `ticket close` to enable reliable orchestration.

## Implementation Plan (Phased)

1) Phase 1 — Close Command (MVP)
- Add `ticket close` command (under `cmd/docmgr/cmds/ticket/…`).
- Close updates: `Status=complete` (override with `--status`), `Intent=long-term` if provided/desired, add a changelog entry with short explanation, refresh `LastUpdated`.
- Human-first output with clear, concise confirmation.

2) Phase 2 — Structured Output Where It Counts
- Add Glaze output to `ticket close`: `--with-glaze-output --output json` returning a single-row or object with:
  - `ticket`, `all_tasks_done`, and `operations` (status_updated, intent_updated, changelog_updated).
- Ensure existing list/search/status already support `--with-glaze-output` (they do) and document best practices.

3) Phase 3 — Status Vocabulary and Guidance
- Extend `vocabulary.yaml` with `status` examples: `draft, active, review, complete, archived`.
- Update `doctor` to warn on unknown `Status`.
- Document suggested transitions in the guide; keep enforcement soft (warnings only).

4) Phase 4 — Quality of Life (Optional / Later)
- `tasks check` prints an actionable suggestion when all tasks are done (no auto action): "💡 Run: docmgr ticket close --ticket <ID>".
- Consider `tasks check --with-glaze-output` to expose `all_tasks_done` for scripts; keep default output human-friendly.
- Explore opt‑in `--auto-close` with `--yes` for CI/agents once the manual path is proven.

## For the New Intern: Core Concepts in docmgr

- Tickets and Workspaces
  - Each ticket lives under `ttmp/YYYY/MM/DD/<TICKET>-<slug>/` with a standard layout (index.md, tasks.md, changelog.md, design-doc/, reference/, playbooks/, etc.).
- Frontmatter (YAML) Fields
  - `Title, Ticket, Status, Topics, DocType, Intent, Owners, RelatedFiles, ExternalSources, Summary, LastUpdated`.
  - Status: current workflow state (free-form today; moving to vocabulary-guided with warnings).
  - Intent: longevity (vocabulary-controlled; default typically `long-term`).
- Vocabulary
  - Central list of allowed slugs: `topics`, `docTypes`, `intent` (and proposed `status`). Teams can extend via `docmgr vocab add`.
- Everyday Verbs (human-friendly, scriptable)
  - `docmgr ticket create-ticket`, `docmgr doc add`, `docmgr doc search`, `docmgr task list/add/check`, `docmgr meta update`, `docmgr changelog update`, `docmgr doctor`, `docmgr status`.
- Dual‑Mode Output (when to use it)
  - Default output is designed for humans (and LLMs are generally fine with it).
  - Use `--with-glaze-output --output json|csv|yaml|table` for programmatic use or large/structured data (e.g., list/search/status, and the new `ticket close`).

## Quick How‑To (Closing a Ticket)

- Human‑first (recommended default)
  1. Check tasks: `docmgr task list --ticket <ID>`
  2. Close: `docmgr ticket close --ticket <ID>`
  3. Verify status: `docmgr list tickets --ticket <ID>` or `docmgr status`

- Script/LLM‑friendly
  - Single step with structured output:
    - `docmgr ticket close --ticket <ID> --with-glaze-output --output json`
  - Or detect first, then close:
    - `docmgr task list --ticket <ID> --with-glaze-output --output json | jq …`
    - `docmgr ticket close --ticket <ID> --with-glaze-output --output json`

## Guidance: Status & Intent

- Recommended Status set (teams can customize): `draft, active, review, complete, archived`.
- Suggested transitions (not enforced):
  - draft → active → review → complete → archived
  - review → active (send back for changes); complete → active (reopen) is allowed but unusual (warn).
- Intent defaults to `long-term` for ticket index; documents inherit from the ticket unless overridden.

## Open Questions

- Should `ticket close` require “all tasks done” by default, or warn and continue?
- Which additional fields (if any) should `ticket close` update (e.g., Summary hint)?
- Do we provide first‑class `ticket reopen` / `ticket archive` now or later?
- Should `tasks check` gain structured output immediately, or later as needed?

## References

- Debate Round 1 — Workflow Friction
- Debate Round 2 — New Verbs and Command Patterns
- Debate Round 3 — Status and Intent Lifecycle Transitions
- Debate Round 4 — Automation vs Manual
- Debate Round 5 — LLM Usage Patterns
- `docmgr help how-to-use`, `docmgr help cli-guide`
