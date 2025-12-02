---
Title: Doctor auto-fix context and onboarding
Ticket: DOCMGR-DOCTOR-AUTOFIX
Status: active
Topics:
  - yaml
  - diagnostics
  - ux
DocType: analysis
Intent: short-term
Owners: []
Summary: Full context to implement doctor suggest/auto-fix using existing frontmatter fix heuristics and diagnostics stack.
LastUpdated: 2025-12-02
---

# Doctor auto-fix context and onboarding

This note is for a new contributor to implement auto-fix/suggest-fixes in `docmgr doctor`. It summarizes the current behavior, where the fix engine lives, and what to wire up.

## Current state
- Doctor reports frontmatter and schema issues via taxonomies and rules but does not suggest/apply fixes. Related files are checked only during ticket scans (index) and not in `doctor --doc`.
- `validate frontmatter` already supports `--suggest-fixes` and `--auto-fix` (with `.bak` backups) and re-parses after applying fixes. It suppresses error taxonomy on success and renders fresh taxonomy on failed re-parse.
- Single-file doctor mode exists (`doctor --doc`): runs parse + required field checks + optional Status/Topics + vocab warnings, but no fixes.

## Fix engine (reuse this)
- **File:** `pkg/commands/validate_frontmatter.go`
  - `generateFixes`: orchestrates fixes; uses:
    - `normalizeDelimiters` (adds closing `---`, wraps if missing start, handles variants, splits fm/body)
    - `scrubStrayDelimiters` (drops delimiter-like lines inside frontmatter)
    - `peelTrailingNonKeyLines` (moves non-`:` lines from end of frontmatter to body)
    - `frontmatter.PreprocessYAML` (quotes unsafe scalars)
  - `applyAutoFix`: writes `.bak` and rewrites the file; re-parses to confirm success; on success prints “Frontmatter auto-fixed”.
  - Schema reuse: calls `doc.Validate()` to emit required-field schema taxonomies.
- **Helpers:** `pkg/frontmatter/frontmatter.go` (`NeedsQuoting`, `QuoteValue`, `PreprocessYAML`).
- **Parser:** `internal/documents/frontmatter.go` (position-aware parsing, snippets, problem classification).

## Diagnostics wiring
- **Taxonomies:** `pkg/diagnostics/docmgrctx/frontmatter.go` (context carries `Fixes []string`).
- **Rules:** `pkg/diagnostics/docmgrrules/frontmatter_rules.go` (renders line/col/snippet/problem + suggested fixes; actions include `docmgr validate frontmatter` and `docmgr help yaml-frontmatter-validation`).
- **Doctor:** `pkg/commands/doctor.go`
  - Ticket scan: checks index frontmatter, required fields, optional Status/Topics, vocab, related-file existence.
  - Single-file mode: parse + schema + vocab; no related-file or fixes.

## Goals for this ticket
- Add doctor flags for suggest/auto-fix (decide scope: per-file with `--doc`, per-ticket with `--ticket`/`--all`?).
- Reuse `generateFixes` and friends; ensure taxonomies include `Fixes` for doctor output.
- Apply fixes safely (backups, opt-in).
- Optional: add related-file check to `doctor --doc`.
- Update rules/docs/smokes to cover doctor’s new behavior.

## Suggested implementation steps
1. Extract/share fix helpers if needed (or call into `validate_frontmatter` helpers directly).
2. Add doctor flags (e.g., `--suggest-fixes`, `--auto-fix`) and thread through single-file/ticket paths.
3. In the doctor frontmatter error paths, attach fixes and optionally apply auto-fix (mirroring `validate frontmatter` flow).
4. On auto-fix success, suppress the error taxonomy; on failure, render the fresh taxonomy from re-parse.
5. Add related-file check in `doctor --doc` (optional).
6. Tests: unit tests for doctor auto-fix invocation; smokes to cover suggest/auto-fix in doctor.
7. Docs: update `pkg/doc/docmgr-diagnostics-and-rules.md`, CLI guides, and `docmgr help yaml-frontmatter-validation`.

## Files to read/edit
- `pkg/commands/validate_frontmatter.go` — fix generator/auto-fix orchestration.
- `pkg/frontmatter/frontmatter.go` — quoting/preprocess.
- `internal/documents/frontmatter.go` — parser + snippets/classification.
- `pkg/diagnostics/docmgrctx/frontmatter.go` — taxonomy context with `Fixes`.
- `pkg/diagnostics/docmgrrules/frontmatter_rules.go` — rule rendering.
- `pkg/commands/doctor.go` — add flags, integrate fixes, consider related-file check in `--doc`.
