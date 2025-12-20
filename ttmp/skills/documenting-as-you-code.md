---
Title: "Documenting as You Code"
DocType: skill
Topics:
  - documentation
  - workflow
  - docmgr
WhatFor: |
  Ensures documentation stays synchronized with code changes by relating
  files immediately after modifications and keeping changelog/tasks current.
  Prevents documentation drift and orphaned docs.
WhenToUse: |
  Use whenever you create, modify, or analyze code files. Use immediately
  after completing any implementation work before moving to the next task.
Status: active
Intent: long-term
---

# Skill: Documenting as You Code

## Overview

Documentation written "later" is documentation never written. This skill enforces relating files to docs and updating changelogs immediately after code changes—not at PR time, not when you remember, but as part of the implementation workflow itself.

**Core principle:** Documentation is part of implementation. Unrelated code changes are incomplete changes.

## When to Use

**Use this skill when:**
- Creating new code files
- Modifying existing code
- Analyzing code for design/reference docs
- Completing any task in a ticket workflow
- Before marking a task as complete

**Use ESPECIALLY when:**
- Working on multi-file changes (easy to forget some)
- In the middle of debugging (important to document findings)
- Making architectural decisions (document rationale immediately)

## The Rule

**After EVERY code change, IMMEDIATELY:**
1. Relate the changed files to relevant docs
2. Update the changelog with what changed
3. Check off any completed tasks

No exceptions. No "I'll do it later." Later never comes.

## The Process

### Step 1: Relate Files to Docs

**Immediately after changing code:**

```bash
TICKET="<current-ticket>"

# Relate to the most specific relevant document
docmgr doc relate --ticket $TICKET --doc-type design-doc \
  --file-note "backend/api/handlers.go:Implemented user authentication endpoints" \
  --file-note "backend/auth/jwt.go:JWT token generation and validation"

# Or relate to ticket index if no specific doc yet
docmgr doc relate --ticket $TICKET \
  --file-note "backend/api/handlers.go:Main API handlers for auth"
```

**Always include notes** explaining WHY each file matters. Notes like "Updated file" are useless. Notes like "Implemented JWT validation logic" help reviewers understand context.

### Step 2: Update Changelog

**Document what changed and why:**

```bash
docmgr changelog update --ticket $TICKET \
  --entry "Implemented JWT authentication with token expiry validation" \
  --file-note "backend/api/handlers.go:Auth endpoints" \
  --file-note "backend/auth/jwt.go:Token logic"
```

**Good changelog entries:**
- State what changed (not just "updated file")
- Explain why if non-obvious
- Reference specific functionality
- Link the files you touched

### Step 3: Update Tasks

**Check off completed work:**

```bash
docmgr task list --ticket $TICKET
docmgr task check --ticket $TICKET --id 1,2
```

**Add new tasks if scope emerged:**

```bash
docmgr task add --ticket $TICKET --text "Add rate limiting to auth endpoints"
```

## Red Flags

These thoughts mean you're deferring documentation:

| Thought | Reality |
|---------|---------|
| "I'll relate files after finishing all tasks" | You'll forget. Do it now. |
| "Let me document everything at the end" | End never comes. Document incrementally. |
| "The code is self-documenting" | Code shows WHAT, not WHY. Document intent. |
| "I'm in the flow, don't want to break it" | Breaking flow prevents forgotten context. |
| "This is a small change" | Small changes accumulate. Document each. |
| "I'll remember what I did" | You won't. Future you will thank present you. |
| "Just one more change before documenting" | Next change and next and next. Stop now. |

**All of these mean: Stop coding. Document what you just did. Then continue.**

## The Workflow Loop

```
1. Implement code change
2. STOP (mandatory)
3. Relate files (docmgr doc relate)
4. Update changelog (docmgr changelog update)
5. Check/add tasks (docmgr task check/add)
6. GOTO 1 (next change)
```

Never skip step 2-5. This loop keeps documentation synchronized.

## Quick Reference

| After... | Run... |
|----------|--------|
| Creating file | `docmgr doc relate --ticket T --file-note "path:purpose"` |
| Modifying file | Same + `docmgr changelog update --ticket T --entry "what changed"` |
| Completing task | `docmgr task check --ticket T --id N` |
| Finding issue | `docmgr task add --ticket T --text "issue description"` |

## Verification Checklist

Before marking a task complete:

- [ ] All modified files related to docs with notes
- [ ] Changelog entry added describing changes
- [ ] Task checked off (if completing task)
- [ ] New tasks added (if scope emerged)
- [ ] Notes explain WHY, not just WHAT

**Checklist discipline:** Track these with docmgr tasks:

```bash
docmgr task add --ticket $TICKET --text "Relate all modified files to docs"
docmgr task add --ticket $TICKET --text "Add changelog entry for changes"
docmgr task add --ticket $TICKET --text "Check off completed task items"
docmgr task add --ticket $TICKET --text "Add new tasks if scope emerged"
```

## Integration with Other Skills

**Use with:**
- `test-driven-development` — Document while following TDD cycle
- `systematic-debugging` — Document findings during debugging phases

**Required after:**
- ANY code change (this is not optional)

## Common Mistakes

**Mistake 1: Batch documentation at PR time**

Problem: You forget what files do and why decisions were made.

Solution: Document immediately after each change. Memory is fresh.

**Mistake 2: Vague file notes**

❌ Bad: `--file-note "backend/api/handlers.go:Updated"`

✅ Good: `--file-note "backend/api/handlers.go:Added JWT validation to auth endpoints"`

**Mistake 3: Skipping changelog for "small changes"**

Problem: Small changes accumulate into undocumented complexity.

Solution: Every code change gets a changelog entry. Every single one.

**Mistake 4: Documentation without context**

Problem: Future developers can't understand why decisions were made.

Solution: Include rationale, constraints, alternatives considered in notes.

## Skill Type

**Rigid** — Follow this workflow exactly. Don't defer documentation.

## Provenance

This skill is docmgr-native but inspired by Superpowers' documentation discipline patterns.

