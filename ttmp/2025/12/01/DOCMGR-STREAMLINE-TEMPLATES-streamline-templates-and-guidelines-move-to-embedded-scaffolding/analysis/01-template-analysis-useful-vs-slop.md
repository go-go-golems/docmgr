---
Title: Template Analysis - Useful vs Slop
Ticket: DOCMGR-STREAMLINE-TEMPLATES
Status: active
Topics:
    - docmgr
    - templates
    - guidelines
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Analysis of which templates add value vs just add boilerplate slop
LastUpdated: 2025-12-01T14:30:01.435728217-05:00
---

# Template Analysis - Useful vs Slop

## Purpose

This analysis evaluates each template in `ttmp/_templates/` to determine:
1. Which templates provide genuine scaffolding value
2. Which templates just add boilerplate/comments that users immediately delete
3. Recommendations for streamlining the template system

## Methodology

For each template, we:
- Created a test document using `docmgr doc add`
- Examined the generated output
- Evaluated whether the template sections are:
  - **Useful**: Provide structure that guides writing
  - **Slop**: Just HTML comments that users delete immediately
  - **Mixed**: Some useful sections, some slop

## Template Analysis

### ✅ Useful Templates (Keep)

#### `design-doc.md`
**Verdict: USEFUL** - Provides clear structure for design documents

**Analysis:**
- Executive Summary, Problem Statement, Proposed Solution, Design Decisions, Alternatives Considered, Implementation Plan, Open Questions, References
- These sections guide the writer through a complete design process
- Comments are minimal and helpful
- Structure matches best practices for design docs

**Recommendation:** Keep as-is, this is a good template.

#### `reference.md`
**Verdict: USEFUL** - Minimal but helpful structure

**Analysis:**
- Goal, Context, Quick Reference, Usage Examples, Related
- Sections are focused and purposeful
- Comments are brief and actionable
- Good balance of structure without over-prescription

**Recommendation:** Keep as-is.

#### `playbook.md`
**Verdict: USEFUL** - Clear operational structure

**Analysis:**
- Purpose, Environment Assumptions, Commands, Exit Criteria, Notes
- Structure matches operational runbook needs
- Comments guide what to fill in each section
- Useful for creating actionable playbooks

**Recommendation:** Keep as-is.

#### `code-review.md`
**Verdict: USEFUL** - Structured review format

**Analysis:**
- Summary, Context, Files Reviewed, Findings (Strengths/Issues), Decisions & Follow-ups, References
- Good structure for capturing review outcomes
- Bullet points guide the reviewer
- Useful for maintaining review records

**Recommendation:** Keep as-is.

### ⚠️ Mixed Templates (Streamline)

#### `index.md`
**Verdict: MIXED** - Some useful structure, some slop

**Analysis:**
- **Useful:** Overview, Status, Tasks, Changelog sections
- **Slop:** 
  - "Key Links" section that just says "See frontmatter RelatedFiles field" - redundant
  - "Topics" section with `{{TOPICS_LIST}}` placeholder - frontmatter already has this
  - "Structure" section listing directories - this is boilerplate that rarely changes
- The template includes too much boilerplate that doesn't help writers

**Recommendation:** 
- Remove "Key Links" section (redundant with frontmatter)
- Remove "Topics" section (already in frontmatter)
- Keep "Structure" section but make it more concise
- Focus on Overview, Status, Tasks, Changelog

#### `working-note.md`
**Verdict: MIXED** - Simple structure, minimal slop

**Analysis:**
- Summary, Notes, Decisions, Next Steps
- Structure is fine, but comments are minimal
- Could be even simpler - just headings

**Recommendation:** Keep but remove HTML comments, just use headings.

#### `tutorial.md`
**Verdict: MIXED** - Good structure but verbose

**Analysis:**
- Overview, Prerequisites, Step-by-Step Guide, Verification, Troubleshooting, Related Resources
- Structure is good for tutorials
- Comments are helpful but could be more concise
- Step-by-step structure with numbered steps would be better

**Recommendation:** Keep structure, streamline comments.

### ❌ Slop Templates (Remove or Simplify)

#### `log.md`
**Verdict: SLOP** - Just placeholder comments

**Analysis:**
- Only contains: `<!-- Log entries in reverse chronological order (newest first) -->`
- Two placeholder entries with `{{DATE}}` and comments
- This is just noise - users will delete immediately
- Logs should be free-form, not templated

**Recommendation:** 
- Remove template entirely
- Fall back to minimal: `# {{TITLE}}\n\n<!-- Log entries -->`

#### `script.md`
**Verdict: SLOP** - Over-structured for simple scripts

**Analysis:**
- Purpose, Usage, Implementation, Notes sections
- For throwaway scripts, this is overkill
- Users creating scripts don't need this structure
- Comments add no value

**Recommendation:**
- Simplify to minimal template or remove
- Scripts are throwaway - they don't need structure

#### `task-list.md`
**Verdict: SLOP** - Redundant with tasks.md

**Analysis:**
- Contains: Tasks section with checkboxes, Completed section, Notes
- This duplicates `tasks.md` which already exists in every ticket
- No need for a separate doc-type for this
- Just adds confusion

**Recommendation:**
- Remove this template entirely
- Users should use `tasks.md` in ticket root, not create separate task-list docs

## Summary

### Templates to Keep (6)
- `design-doc.md` - Excellent structure
- `reference.md` - Good minimal structure  
- `playbook.md` - Useful operational structure
- `code-review.md` - Good review structure
- `tutorial.md` - Keep but streamline
- `working-note.md` - Keep but simplify

### Templates to Streamline (1)
- `index.md` - Remove redundant sections

### Templates to Remove/Simplify (3)
- `log.md` - Remove template, use minimal fallback
- `script.md` - Simplify or remove
- `task-list.md` - Remove entirely (redundant with tasks.md)

## Recommendations

1. **Move templates to embedded FS** - Templates should be embedded in binary for scaffolding, not live in repo
2. **Remove slop templates** - Delete `log.md`, `script.md`, `task-list.md` templates
3. **Streamline mixed templates** - Clean up `index.md`, simplify `working-note.md`, `tutorial.md`
4. **Keep useful templates** - `design-doc.md`, `reference.md`, `playbook.md`, `code-review.md` are valuable
5. **Guidelines are useful** - Keep all guidelines, they provide value without adding slop to documents

## Testing Notes

When testing with `go run ./cmd/docmgr doc add`:
- Templates are loaded from filesystem first (`ttmp/_templates/`)
- Then fallback to embedded strings in `internal/templates/templates.go`
- This means filesystem templates override embedded ones
- For docmgr's own repo, we should use embedded templates, not filesystem ones
- Filesystem templates should only be for user customization

## Template File Flag

**Finding:** There is **no flag** to specify an explicit template file when creating documents.

The `docmgr doc add` command does not support a `--template` or `--template-file` flag. Templates are automatically selected based on `--doc-type`:
1. First checks filesystem: `ttmp/_templates/<doc-type>.md`
2. Falls back to embedded: `internal/templates/templates.go` TemplateContent map

**Recommendation:** This is fine - explicit template files would add complexity without much benefit. The doc-type-based selection is sufficient.

## Next Steps

1. Create embedded FS structure for templates and guidelines
2. Move useful templates to embedded FS
3. Remove slop templates
4. Update `LoadTemplate` to prefer embedded, allow filesystem override
5. Update `docmgr init` to scaffold templates/guidelines but not use them for docmgr itself
6. Test that docmgr uses embedded templates when filesystem ones don't exist
