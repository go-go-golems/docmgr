---
Title: Phase 3 — Status vocabulary and guidance
Ticket: DOCMGR-PHASE3
Status: active
Topics:
    - tools
    - ux
    - documentation
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/pkg/commands/doctor.go
      Note: Warns on unknown status; validates topics/intent; staleness & prefixes
    - Path: docmgr/pkg/commands/ticket_close.go
      Note: Sets status on close; interacts with status vocabulary
    - Path: docmgr/pkg/doc/docmgr-how-to-use.md
      Note: Docs include closing workflow and vocabulary guidance
    - Path: docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/13-debate-round-2-new-verbs-and-command-patterns.md
      Note: Verb patterns and command design
    - Path: docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/14-debate-round-3-status-and-intent-lifecycle-transitions.md
      Note: Round 3 decisions
    - Path: docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/15-debate-round-4-automation-vs-manual.md
      Note: Automation boundaries and explicit commands
    - Path: docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/16-debate-round-5-llm-usage-patterns.md
      Note: LLM usage patterns and structured outputs
    - Path: docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/reference/06-debate-synthesis-closing-workflow-status-intent-and-llm-ux.md
      Note: Phase 3 synthesis and plan
    - Path: docmgr/ttmp/vocabulary.yaml
      Note: 'Seeds status vocabulary: draft'
ExternalSources: []
Summary: 'Phase 3: add status vocabulary seeds, doctor warnings, and guidance; document transitions (warnings only).'
LastUpdated: 2025-11-19T16:30:09.525066593-05:00
---



# Phase 3 — Status vocabulary and guidance

## Overview

This ticket implements Phase 3 from the synthesis: introduce a `status` vocabulary (team‑extensible), warn on unknown values in `doctor`, and document suggested (non‑enforced) transitions. The goal is consistency and discoverability without rigidity: teams can add custom statuses, `doctor` helps catch typos, and documentation shows best‑practice transitions.

## Key Links

- Debate synthesis: `docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/reference/06-debate-synthesis-closing-workflow-status-intent-and-llm-ux.md`
- Round 3 (Status/Intent lifecycle): `docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/14-debate-round-3-status-and-intent-lifecycle-transitions.md`
- Round 4 (Automation vs Manual): `docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/15-debate-round-4-automation-vs-manual.md`
- Round 5 (LLM usage patterns): `docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/16-debate-round-5-llm-usage-patterns.md`
- CLI guide (help): `docmgr/pkg/doc/docmgr-how-to-use.md`
- Vocabulary seeds: `docmgr/ttmp/vocabulary.yaml`

## Status

Current status: **active**

## Scope

- Add `status` category to `vocabulary.yaml` with seeds: `draft, active, review, complete, archived`
- Update `doctor` to warn on unknown Status (do not fail)
- Document suggested transitions (non‑enforced):
  - `draft → active → review → complete → archived`
  - `review → active` (send back)
  - `complete → active` (reopen) — warn as unusual
- Ensure discoverability:
  - `docmgr vocab list --category status`
  - Helpful warning message shows valid values

## Topics

- tools
- ux
- documentation

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Next Steps (Execution)

1) (Optional) Explore exposing suggested transitions in structured outputs (e.g., `ticket close` JSON)

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
