---
Title: Frontmatter Walkthrough
Ticket: DOCMGR-REFACTOR
Status: review
Topics:
    - tooling
DocType: working-note
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Testing new internal/documents helpers
LastUpdated: 2025-11-18T20:46:06.748225903-05:00
---


---
Title: Frontmatter Walkthrough
Ticket: DOCMGR-REFACTOR
Status: draft
Topics:
  - tooling
DocType: working-note
Intent: short-term
Owners:
  - manuel
RelatedFiles: []
ExternalSources: []
Summary: >
  Testing new internal/documents helpers
LastUpdated: 2025-11-18
---

# Frontmatter Walkthrough

## Summary

- Exercised the new `internal/documents` helpers by creating two docs via `doc add`, updating their metadata with `meta update`, and wiring RelatedFiles + changelog entries through the CLI.
- Added contextual error wrapping across high-traffic listing/meta commands and fixed the silent config warning bug.
- Verified the Cobra hierarchy end-to-end with `go run ./cmd/docmgr` invocations (`help`, `ticket tickets`, `doc docs`, `tasks list`).

## Notes

- CLI runs:
  - `go run ./cmd/docmgr help how-to-use`
  - `go run ./cmd/docmgr ticket tickets --root ttmp`
  - `go run ./cmd/docmgr doc docs --ticket DOCMGR-REFACTOR --root ttmp`
  - `go run ./cmd/docmgr tasks list --ticket DOCMGR-REFACTOR --root ttmp`
  - `go run ./cmd/docmgr doc search --ticket DOCMGR-REFACTOR --root ttmp --query docmgr`
  - `go run ./cmd/docmgr doctor --root ttmp --ticket DOCMGR-REFACTOR`
- Tasks CLI tightening:
  - Wrapped every `tasks` subcommand with contextual error handling (`gp.AddRow`, file writes, setting parsing).
  - `go run ./cmd/docmgr changelog update ...` + `doc relate` to record `pkg/commands/tasks.go` work.
- Search + doctor error context:
  - Added contextual error messages for search results + suggestions, plus all doctor issue rows.
  - Recorded file links via `doc relate` and changelog entries referencing `pkg/commands/search.go` and `pkg/commands/doctor.go`.
- Doc creation & metadata:
  - `doc add --doc-type design-doc --title "Error Context Rollout"` → status flipped to `review` with `meta update`.
  - `doc add --doc-type playbooks --title "CLI Regression Checklist"` → summary updated via `meta update`.
  - `doc relate` + `changelog update` pointed at `internal/workspace/config.go`, `pkg/commands/list_docs.go`, and friends.
- Tests: `go test ./...` clean; manual CLI invocations produced expected tabular output without regressions.
- Re-ran `go test ./...` after tasks + search/doctor edits.

## Decisions

- Prioritize wrapping error context on listing/meta/import/layout commands first since they are exercised in every workflow; leave lower-traffic doctor/tasks refinements for a follow-up pass.
- Keep emitting warnings (instead of hard failures) for malformed `.ttmp.yaml` inside `ResolveRoot`, but ensure the warning includes the actual parse error.

## Next Steps

- Continue Round 4 work by tightening error contexts in remaining commands (`tasks`, `doctor`, `search`) and exploring a shared helper for repeated `gp.AddRow` error handling.
- Expand structured changelog/task automation once the remaining error-handling refactors land.
