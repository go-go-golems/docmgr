---
Title: Bug Report: skill show name matching issues
Ticket: 004-BETTER-SKILL-SHOWING
Status: active
Topics:
    - skills
    - ux
    - cli
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-19T19:37:03.313515848-05:00
---

# Bug Report: skill show name matching issues

## Goal

Document the UX friction encountered when trying to load skills using `docmgr skill show`, so we can improve the command's name-matching behavior to be more intuitive.

## Context

While testing the 3 newly created skills (`test-driven-development.md`, `systematic-debugging.md`, `documenting-as-you-code.md`), I encountered several friction points when trying to load skills using `docmgr skill show`. The command's current name-matching behavior doesn't match user expectations, requiring exact knowledge of the Title field format.

This bug report captures the specific command attempts, their results, and recommendations for improving the UX.

## Issues Encountered

### Issue 1: Positional argument not supported

**What I tried:**
```bash
docmgr skill show test-driven-development
```

**What happened:**
```
Error: Too many arguments
```

**Expected behavior:**
The skill name should be accepted as a positional argument (like `git show <ref>` or `kubectl get pod <name>`).

**Current workaround:**
Must use `--skill` flag:
```bash
docmgr skill show --skill "test-driven-development"
```

**Severity:** Minor annoyance — adds typing overhead but workaround is clear from help text.

---

### Issue 2: Filename matching doesn't work

**What I tried:**
```bash
docmgr skill show --skill "test-driven-development"
```

**Reasoning:** The skill file is named `test-driven-development.md`, so I expected the filename (without extension) to work.

**What happened:**
```
Error: no skills found matching "test-driven-development"
```

**Expected behavior:**
The command should match against the filename (slug) in addition to the Title field, since filenames are visible in file browsers and `docmgr skill list` output.

**Current workaround:**
Must match against the Title field from frontmatter:
```bash
docmgr skill show --skill "Skill: Test-Driven Development"  # Exact title
# OR
docmgr skill show --skill "Test-Driven"                      # Partial title match
```

**Severity:** Moderate — creates confusion because users naturally think in terms of filenames when browsing the filesystem.

---

### Issue 3: "Skill: " prefix requirement unclear

**What I tried:**
```bash
docmgr skill show --skill "test-driven-development"   # Failed
docmgr skill show --skill "Test-Driven Development"   # Would likely fail
```

**What worked:**
```bash
docmgr skill show --skill "Skill: Test-Driven Development"  # Exact title
docmgr skill show --skill "Test-Driven"                      # Partial match
```

**Observation:** All test skill titles include a "Skill: " prefix (e.g., `Title: "Skill: Test-Driven Development"`). This prefix is helpful for document lists but creates friction when trying to load a skill by name.

**Expected behavior:**
The command should either:
1. Strip "Skill: " prefix from titles when matching (transparent to user), OR
2. Match against both the full title and a normalized version without the prefix

**Current workaround:**
Use partial match that doesn't include "Skill: " prefix:
```bash
docmgr skill show --skill "Test-Driven"
```

**Severity:** Moderate — requires users to understand the Title field format convention.

---

## Summary of Working Commands

After trial and error, these patterns work:

```bash
# ✅ Works: Exact title match
docmgr skill show --skill "Skill: Test-Driven Development"

# ✅ Works: Partial title match (skips prefix)
docmgr skill show --skill "Test-Driven"
docmgr skill show --skill "Systematic"
docmgr skill show --skill "Documenting"

# ❌ Doesn't work: Positional argument
docmgr skill show test-driven-development

# ❌ Doesn't work: Filename-based matching
docmgr skill show --skill "test-driven-development"

# ❌ Doesn't work: Title without "Skill: " prefix
docmgr skill show --skill "Test-Driven Development"
```

## Recommendations

### Recommendation 1: Support positional argument

**Change:**
```bash
# Current (requires flag)
docmgr skill show --skill "name"

# Proposed (positional + flag both work)
docmgr skill show "name"
docmgr skill show --skill "name"  # Still supported
```

**Benefit:** Matches CLI conventions (git, kubectl, etc.) and reduces typing.

**Implementation:** Add positional argument to cobra command definition.

---

### Recommendation 2: Multi-strategy matching

**Change:** Match skill names using multiple strategies in priority order:

1. **Exact title match** (current behavior)
2. **Title without "Skill: " prefix** (new)
3. **Filename without extension** (new)
4. **Partial title match** (current behavior)

**Examples:**
```bash
# All of these should find "Skill: Test-Driven Development"
docmgr skill show "Skill: Test-Driven Development"  # Strategy 1: exact
docmgr skill show "Test-Driven Development"          # Strategy 2: without prefix
docmgr skill show "test-driven-development"          # Strategy 3: filename
docmgr skill show "Test-Driven"                      # Strategy 4: partial
```

**Benefit:** Users can use whatever identifier is most natural (title, filename, or partial match).

**Implementation:** Update skill resolution logic to try multiple matching strategies.

---

### Recommendation 3: Better error messages

**Current error:**
```
Error: no skills found matching "test-driven-development"
```

**Proposed error:**
```
Error: no skills found matching "test-driven-development"

Available skills:
- Skill: Test-Driven Development (test-driven-development.md)
- Skill: Systematic Debugging (systematic-debugging.md)
- Skill: Documenting as You Code (documenting-as-you-code.md)

Tip: Match against title or filename. Use partial matches like "Test-Driven" or "Systematic".
```

**Benefit:** Users immediately see what's available and how to match.

**Implementation:** On match failure, run `docmgr skill list` and include results in error message.

---

### Recommendation 4: Normalize "Skill: " prefix convention

**Problem:** The "Skill: " prefix in titles is useful for document lists but creates matching friction.

**Option A: Strip prefix transparently**
- Internally strip "Skill: " when matching
- Users never need to know about it

**Option B: Document the convention**
- Update `how-to-write-skills.md` to explain the convention
- Add examples showing both exact and partial matching

**Option C: Make prefix optional**
- Allow skills with or without "Skill: " prefix
- Match against whatever the title actually is

**Recommendation:** Option A (strip transparently) for best UX. If not feasible, Option B with better error messages.

---

## Test Cases for Verification

After implementing improvements, these should all work:

```bash
# Positional argument
docmgr skill show "test-driven-development"
docmgr skill show "Test-Driven Development"
docmgr skill show "Test-Driven"

# Filename matching
docmgr skill show --skill "test-driven-development"
docmgr skill show --skill "systematic-debugging"
docmgr skill show --skill "documenting-as-you-code"

# Title matching (with and without prefix)
docmgr skill show --skill "Skill: Test-Driven Development"
docmgr skill show --skill "Test-Driven Development"

# Partial matching
docmgr skill show --skill "Test-Driven"
docmgr skill show --skill "TDD"  # If title or topics contain "TDD"

# Error messages show helpful hints
docmgr skill show "nonexistent-skill"  # Should list available skills
```

## Usage Examples

### Current workflow (verbose)

```bash
# Discover what's available
docmgr skill list

# Copy exact title from output
docmgr skill show --skill "Skill: Test-Driven Development"
```

### Desired workflow (intuitive)

```bash
# Discover what's available
docmgr skill list

# Use filename or natural name
docmgr skill show test-driven-development
docmgr skill show "Test-Driven Development"
docmgr skill show TDD
```

## Related

- Ticket 003-CREATE-SKILL-PROMPTS: Created the test skills that exposed these issues
- [`cmd/docmgr/cmds/skill/show.go`](https://github.com/go-go-golems/docmgr): Current implementation
- Superpowers comparison: Superpowers uses `superpowers-codex use-skill <name>` with simpler filename-based matching

## Next Steps

1. Implement multi-strategy matching (title, title without prefix, filename, partial)
2. Add positional argument support
3. Improve error messages to show available skills
4. Add test cases to verify all matching strategies
5. Update documentation with examples of different matching approaches
