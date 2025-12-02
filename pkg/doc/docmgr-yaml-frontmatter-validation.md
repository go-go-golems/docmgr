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
- Run `docmgr validate frontmatter --doc <path>` to see line/col, snippet, and suggestions.
- For hints: `docmgr validate frontmatter --doc <path> --suggest-fixes`
- To attempt repair: `docmgr validate frontmatter --doc <path> --auto-fix` (creates `<path>.bak`)

## Common issues and fixes
- Missing closing `---`: add a closing delimiter or let `--auto-fix` rewrite it.
- Unquoted colons/hashes: quote the value (`'text: with colon'`) or use block scalars.
- Stray delimiter lines inside frontmatter: remove extra `---` lines; auto-fix will scrub them.
- Plain text inside frontmatter (no `:`): move it into the body or let auto-fix peel trailing non-key lines.

## Commands to remember
- Validate: `docmgr validate frontmatter --doc <file>`
- Suggest: `docmgr validate frontmatter --doc <file> --suggest-fixes`
- Auto-fix: `docmgr validate frontmatter --doc <file> --auto-fix`
- Doctor (workspace scan): `docmgr doctor --ticket <TICKET>` (will report invalid frontmatter and point to validation)

## Implementation references
- Parser + diagnostics: `internal/documents/frontmatter.go`
- Validation verb + auto-fix: `pkg/commands/validate_frontmatter.go`
- Rules rendering guidance: `pkg/diagnostics/docmgrrules/frontmatter_rules.go`
- Quoting/preprocess helpers: `pkg/frontmatter/frontmatter.go`
- Detailed ticket guide: `ttmp/2025/11/29/DOCMGR-YAML-001.../reference/02-frontmatter-healing-and-validation-guide.md`
