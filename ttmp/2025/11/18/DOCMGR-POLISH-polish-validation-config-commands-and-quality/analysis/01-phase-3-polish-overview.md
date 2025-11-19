---
Title: Phase 3: Polish Overview
Ticket: DOCMGR-POLISH
Status: active
Topics:
    - docmgr
    - polish
    - validation
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Document validation, config visibility commands, quality improvements
LastUpdated: 2025-11-18T12:25:00.000000000-05:00
---

# Phase 3: Polish Overview

## Goal

Add quality-of-life features that improve user experience: document validation, config introspection, and refined initialization.

## Context

From **Debate Round 10** synthesis:
- **Dependencies**: Builds on Phase 2 refactoring
- **Effort**: 1-2 weeks
- **Focus**: User-facing polish, not internal refactoring

## Tasks (4 items)

### Document Validation (Round 3)

1. ✅ **Add Document.Validate() method**:
   ```go
   func (doc *Document) Validate() error {
       if doc.Title == "" {
           return fmt.Errorf("Title is required")
       }
       if doc.Ticket == "" {
           return fmt.Errorf("Ticket is required")
       }
       if doc.DocType == "" {
           return fmt.Errorf("DocType is required")
       }
       return nil
   }
   ```

   **Usage:**
   - `docmgr add`: Validate on creation (fail if invalid)
   - `docmgr doctor`: Validate all documents, report issues
   - `docmgr search/list`: Skip validation (performance)

**Impact**: Catch user mistakes early (typos, missing fields), better error messages.

### YAML Consolidation (Round 3)

2. ✅ **Consolidate on adrg/frontmatter library**:
   - Remove manual `splitFrontmatter()` in `import_file.go`
   - Use `internal/documents/frontmatter.go` utilities everywhere
   - Eliminates fragile edge case handling

**Impact**: Robust YAML parsing, handles edge cases (--- in body, no frontmatter, etc.).

### Config Visibility (Round 8)

3. ✅ **Implement `docmgr config show` command**:
   ```bash
   $ docmgr config show
   Configuration sources (in precedence order):
     1. --root flag: <not set>
     2. .ttmp.yaml (current dir): <not found>
     3. .ttmp.yaml (parent dirs): found at ../../.ttmp.yaml
     4. .ttmp.yaml (home dir): <not found>
     5. .ttmp.yaml (git root): <not applicable>
     6. Default: ttmp
   
   Active configuration:
     root: /home/user/project/docs
     vocabulary: /home/user/project/docs/vocabulary.yaml
     source: ../../.ttmp.yaml
   ```

**Impact**: Users can debug config issues, understand which config is active.

### Improved Initialization (Round 8)

4. ✅ **Update `docmgr init` to create config file**:
   ```bash
   $ docmgr init
   Created workspace root: ./ttmp
   Created config file: .ttmp.yaml
     root: ttmp
     vocabulary: ttmp/vocabulary.yaml
   
   Hint: Customize paths by editing .ttmp.yaml
   ```

   **Creates:**
   - `ttmp/` directory
   - `.ttmp.yaml` config file with defaults
   - `ttmp/vocabulary.yaml` template

**Impact**: Users know config file exists, can customize easily.

## Success Criteria

- [ ] Invalid documents are caught on creation (`docmgr add` validates)
- [ ] `docmgr doctor` reports validation issues across all documents
- [ ] No manual frontmatter splitting remains (all use library)
- [ ] `docmgr config show` reveals config resolution
- [ ] `docmgr init` creates complete workspace with config

## References

- [Debate Round 3: YAML Robustness](../../DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/03-debate-round-03-yaml-processing-robustness.md)
- [Debate Round 8: Configuration Design](../../DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/08-debate-round-08-configuration-and-path-resolution-design.md)

## Implementation Order

**Week 1:**
1. Add `Document.Validate()` method
2. Integrate into `add` command (validate on creation)
3. Integrate into `doctor` command (report issues)

**Week 2:**
4. Implement `docmgr config show` command
5. Update `docmgr init` to create config file
6. Clean up manual frontmatter splitting

## Timeline

**Estimated effort**: 1-2 weeks
**Priority**: LOW (polish, not urgent)
**Dependencies**: Phase 2 (utilities exist), Phase 1 (error context)

## Future Enhancements (Deferred)

From Round 3 discussions:
- ❓ Enum validation (`Status`, `Intent` must be valid values)
- ❓ Pattern validation (ticket format `[A-Z]+-[0-9]+`)
- ❓ Unknown field detection (warn about possible typos)
- ❓ JSON Schema validation

These can be added later if needed.
