---
Title: YAML frontmatter validation and fixes
Slug: yaml-frontmatter-validation
Short: How to diagnose, validate, and auto-fix YAML frontmatter issues in docmgr.
Topics:
- docmgr
- yaml
- validation
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# YAML frontmatter validation and fixes

This page explains how to diagnose and fix YAML/frontmatter issues in docmgr using the built-in tools. It also links to the validation and auto-fix workflows.

## When you see a YAML/frontmatter error
If a command reports `yaml_syntax` or `frontmatter parse` errors:
- **Primary fix path:** `docmgr doctor --ticket <TICKET> --fix` applies safe auto-fixes across all docs in the ticket (creating `<path>.bak` backups), migrates legacy `RelatedFiles` paths to explicit anchors, and re-validates.
- For a single file: run `docmgr validate frontmatter --doc <path>` to see line/col, snippet, and suggestions.
- For hints: `docmgr validate frontmatter --doc <path> --suggest-fixes`
- To attempt single-file repair: `docmgr validate frontmatter --doc <path> --auto-fix` (creates `<path>.bak`)

## Common issues and fixes
- Missing closing `---`: add a closing delimiter or let `doctor --fix` / `--auto-fix` rewrite it.
- Unquoted colons/hashes: quote the value (`'text: with colon'`) or use block scalars.
- Stray delimiter lines inside frontmatter: remove extra `---` lines; auto-fix will scrub them.
- Plain text inside frontmatter (no `:`): move it into the body or let auto-fix peel trailing non-key lines.
- Legacy bare `RelatedFiles` paths: run `docmgr doctor --fix-anchors` to migrate them to anchored paths (`repo://`, `ws://`, `docs://`, `abs://`); see `docmgr help path-anchors`.

## Commands to remember
- Fix a whole ticket (frontmatter + anchors): `docmgr doctor --ticket <TICKET> --fix`
- Anchor migration only: `docmgr doctor --ticket <TICKET> --fix-anchors`
- Validate one file: `docmgr validate frontmatter --doc <file>`
- Suggest: `docmgr validate frontmatter --doc <file> --suggest-fixes`
- Auto-fix one file: `docmgr validate frontmatter --doc <file> --auto-fix`
- Doctor (workspace scan): `docmgr doctor --ticket <TICKET>` (checks all docs in the ticket and points to validation; multi-ticket runs print a per-ticket rollup, add `--details` for everything)

## Implementation references
- For full details, see the technical reference: `docmgr help yaml-frontmatter-validation-reference`
- Parser + diagnostics: `internal/documents/frontmatter.go`
- Validation verb + auto-fix: `pkg/commands/validate_frontmatter.go`
- Rules rendering guidance: `pkg/diagnostics/docmgrrules/frontmatter_rules.go`
- Quoting/preprocess helpers: `pkg/frontmatter/frontmatter.go`
- Detailed ticket guide: `ttmp/2025/11/29/DOCMGR-YAML-001.../reference/02-frontmatter-healing-and-validation-guide.md`
