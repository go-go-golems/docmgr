---
Title: Schema validation context and plan
Ticket: DOCMGR-FRONTMATTER-SCHEMA
Status: active
Topics:
  - yaml
  - validation
  - diagnostics
DocType: analysis
Intent: short-term
Owners: []
Summary: Context and plan to add field-level schema validation with frontmatter schema taxonomies across doctor and validate frontmatter.
LastUpdated: 2025-12-02
---

# Schema validation context and plan

This document captures what’s needed to add field-level schema validation (beyond required fields) and surface issues via `FrontmatterSchema` taxonomies in both `doctor` and `validate frontmatter`.

## Current state
- Required fields (Title, Ticket, DocType) are validated via `models.Document.Validate()`; doctor and validate frontmatter emit `FrontmatterSchema` taxonomies on failure.
- Optional checks in doctor: missing Status/Topics (warnings), vocab mismatches for Topics/DocType/Intent/Status (warnings via `VocabularyUnknown`).
- No per-field patterns/length/hints; no shared validator table.
- Single-file doctor (`--doc`) runs parse + required + optional + vocab; no related-file check yet.
- Help links and syntax/schema rules are rendered from `pkg/diagnostics/docmgrrules`.

## Goal
Add a shared set of field-level validators with hints and severity, and emit `FrontmatterSchema` taxonomies consistently from doctor (workspace and `--doc`) and `validate frontmatter`.

## Proposed approach
1. **Validator table** (shared): define a slice of validators (field, severity, hint, check func) in a reusable package (e.g., `pkg/validation/frontmatter` or similar).
2. **Integration points:**
   - `validate frontmatter`: after parse, run validators; emit schema taxonomies for failures; keep existing required-field behavior via `doc.Validate()`.
   - `doctor` ticket scan + `--doc`: run the same validators to avoid divergence; keep vocab warnings separate.
3. **Hints and severity:** expand `FrontmatterSchemaContext` to carry a hint if needed; render hints in `FrontmatterSchemaRule`.
4. **Extensible checks:** start with:
   - Status required (warning) and optionally must be in vocab (already handled separately)
   - Topics non-empty (warning)
   - Summary length max (warning/error?)
   - Owners non-empty (optional, warning)
   - LastUpdated parseable (warning)
   - DocType in vocab (currently separate)
   - Intent in vocab (currently separate)
5. **Docs/tests:** update docs to describe schema validators; add unit tests for validator table; add smokes if needed.

## Code map
- Parser: `internal/documents/frontmatter.go` (parse/snippets/classify)
- Schema context: `pkg/diagnostics/docmgrctx/frontmatter.go` (`FrontmatterSchemaContext`)
- Rules: `pkg/diagnostics/docmgrrules/frontmatter_rules.go` (render schema issues)
- Doctor: `pkg/commands/doctor.go` (workspace scan + `validateSingleDoc` for `--doc`)
- Validation verb: `pkg/commands/validate_frontmatter.go` (hook validators after parse)
- Vocab models: `pkg/models/document.go` (struct + `Validate()` for required fields)

## Suggested steps
1) Add validator table with hints/severity.
2) Extend schema context/rule to show hints (optional).
3) Wire validators into `validate frontmatter` and doctor (both paths) with consistent emission.
4) Keep vocab warnings separate but ensure they can show alongside schema.
5) Tests: unit tests for validators; update smokes if needed.
6) Docs: update `pkg/doc/docmgr-diagnostics-and-rules.md` and YAML validation help page to mention schema validators/hints.
