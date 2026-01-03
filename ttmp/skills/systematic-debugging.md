---
Title: "Systematic Debugging"
DocType: skill
Topics:
  - debugging
  - troubleshooting
  - quality
WhatFor: |
  Enforces four-phase debugging process (Root Cause → Pattern Analysis →
  Hypothesis → Implementation) to prevent random fixes and ensure bugs are
  understood before being addressed.
WhenToUse: |
  Use when encountering any bug, test failure, or unexpected behavior,
  before proposing fixes. Use ESPECIALLY under time pressure or after
  multiple failed fix attempts.
Status: active
Intent: long-term
---

# Skill: Systematic Debugging

## Overview

Random fixes waste time and create new bugs. Quick patches mask underlying issues.

**Core principle:** ALWAYS find root cause before attempting fixes. Symptom fixes are failure.

**Violating the letter of this process is violating the spirit of debugging.**

## When to Use

Use for ANY technical issue:
- Test failures
- Bugs in production
- Unexpected behavior
- Performance problems
- Build failures
- Integration issues

**Use this ESPECIALLY when:**
- Under time pressure (emergencies make guessing tempting)
- "Just one quick fix" seems obvious
- You've already tried multiple fixes
- Previous fix didn't work
- You don't fully understand the issue

**Don't skip when:**
- Issue seems simple (simple bugs have root causes too)
- You're in a hurry (rushing guarantees rework)
- Manager wants it fixed NOW (systematic is faster than thrashing)

## The Iron Law

```
NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST
```

If you haven't completed Phase 1, you cannot propose fixes.

## The Four Phases

You MUST complete each phase before proceeding to the next.

### Phase 1: Root Cause Investigation

**BEFORE attempting ANY fix:**

1. **Read Error Messages Carefully**
   - Don't skip past errors or warnings
   - They often contain the exact solution
   - Read stack traces completely
   - Note line numbers, file paths, error codes

2. **Reproduce Consistently**
   - Can you trigger it reliably?
   - What are the exact steps?
   - Does it happen every time?
   - If not reproducible → gather more data, don't guess

3. **Check Recent Changes**
   - What changed that could cause this?
   - Git diff, recent commits
   - New dependencies, config changes
   - Environmental differences

4. **Gather Evidence**
   - Add logging at component boundaries
   - Trace data flow through the system
   - Verify assumptions with print statements
   - Document what you find

5. **Document Findings**

   ```bash
   TICKET="<current-ticket>"
   
   # Create or update investigation log
   docmgr doc add --ticket $TICKET --doc-type reference \
     --title "Bug Investigation: <brief-description>"
   
   # Or append to existing reference doc in your editor
   ```

### Phase 2: Pattern Analysis

**Find the pattern before fixing:**

1. **Find Working Examples**
   - Locate similar working code in same codebase
   - What works that's similar to what's broken?

2. **Compare Against References**
   - If implementing pattern, read reference implementation COMPLETELY
   - Don't skim - read every line
   - Understand the pattern fully before applying

3. **Identify Differences**
   - What's different between working and broken?
   - List every difference, however small
   - Don't assume "that can't matter"

4. **Understand Dependencies**
   - What other components does this need?
   - What settings, config, environment?
   - What assumptions does it make?

### Phase 3: Hypothesis and Testing

**Scientific method:**

1. **Form Single Hypothesis**
   - State clearly: "I think X is the root cause because Y"
   - Write it down in your investigation log
   - Be specific, not vague

2. **Test Minimally**
   - Make the SMALLEST possible change to test hypothesis
   - One variable at a time
   - Don't fix multiple things at once

3. **Verify Before Continuing**
   - Did it work? Yes → Phase 4
   - Didn't work? Form NEW hypothesis
   - DON'T add more fixes on top

4. **When You Don't Know**
   - Say "I don't understand X"
   - Don't pretend to know
   - Ask for help
   - Research more

### Phase 4: Implementation

**Fix the root cause, not the symptom:**

1. **Create Failing Test Case**
   - Use `test-driven-development` skill
   - Write test that reproduces bug
   - Verify test fails with bug present

2. **Implement Single Fix**
   - Address the root cause identified
   - ONE change at a time
   - No "while I'm here" improvements
   - No bundled refactoring

3. **Verify Fix**
   - Test passes now?
   - No other tests broken?
   - Issue actually resolved?

4. **Document the Fix**

   ```bash
   docmgr changelog update --ticket $TICKET \
     --entry "Fixed <bug>: root cause was <X>, solution was <Y>" \
     --file-note "path/to/fixed.go:Applied fix for <bug>"
   ```

5. **If Fix Doesn't Work**
   - STOP
   - Count: How many fixes have you tried?
   - If < 3: Return to Phase 1, re-analyze with new information
   - **If ≥ 3: STOP and question the architecture**
   - DON'T attempt Fix #4 without architectural discussion

### If 3+ Fixes Failed: Question Architecture

**Pattern indicating architectural problem:**
- Each fix reveals new shared state/coupling/problem in different place
- Fixes require "massive refactoring" to implement
- Each fix creates new symptoms elsewhere

**STOP and question fundamentals:**
- Is this pattern fundamentally sound?
- Are we "sticking with it through sheer inertia"?
- Should we refactor architecture vs. continue fixing symptoms?

**Discuss with your human partner before attempting more fixes.**

This is NOT a failed hypothesis - this is a wrong architecture.

## Red Flags

If you catch yourself thinking:

| Thought | Reality |
|---------|---------|
| "Quick fix for now, investigate later" | Later never comes. Investigate now. |
| "Just try changing X and see if it works" | Guessing wastes time. Investigate first. |
| "Add multiple changes, run tests" | Can't isolate what worked. One change at a time. |
| "Skip the test, I'll manually verify" | Manual verification doesn't prevent regression. |
| "It's probably X, let me fix that" | "Probably" = guessing. Investigate to confirm. |
| "I don't fully understand but this might work" | Don't fix what you don't understand. |
| "One more fix attempt" (after 2+ failures) | 3+ failures = architectural problem. Stop. |

**ALL of these mean: STOP. Return to Phase 1.**

**If 3+ fixes failed:** Question the architecture, don't try again.

## Verification Checklist

Before marking debugging complete:

- [ ] Completed Phase 1 (root cause investigation)
- [ ] Completed Phase 2 (pattern analysis)
- [ ] Completed Phase 3 (hypothesis and testing)
- [ ] Created failing test reproducing bug
- [ ] Implemented single fix addressing root cause
- [ ] Verified fix works (test passes)
- [ ] No other tests broken
- [ ] Documented investigation and fix in changelog
- [ ] Related fixed files to docs

Can't check all boxes? You haven't found root cause. Return to Phase 1.

**Checklist discipline:**

```bash
TICKET="<current-ticket>"
docmgr task add --ticket $TICKET --text "Complete Phase 1: Root cause investigation"
docmgr task add --ticket $TICKET --text "Complete Phase 2: Pattern analysis"
docmgr task add --ticket $TICKET --text "Complete Phase 3: Hypothesis testing"
docmgr task add --ticket $TICKET --text "Create failing test reproducing bug"
docmgr task add --ticket $TICKET --text "Implement fix addressing root cause"
docmgr task add --ticket $TICKET --text "Verify fix works, no tests broken"
docmgr task add --ticket $TICKET --text "Document investigation and fix"
```

## Integration with Other Skills

**Related skills:**
- `test-driven-development` — For creating failing test case (Phase 4, Step 1)
- `documenting-as-you-code` — For documenting investigation and fix

**Use before:**
- Proposing any fix
- Writing any debug-related code

## Example: Quick Reference

```bash
# Phase 1: Investigate
git log --oneline -10                    # Recent changes
git diff HEAD~5..HEAD -- path/to/file    # What changed
go test -v ./...                         # Reproduce

# Document findings
docmgr doc add --ticket $TICKET --doc-type reference \
  --title "Bug Investigation: Auth failure"

# Phase 2-3: Analyze and hypothesize
# (work in investigation doc, not shown)

# Phase 4: Fix and document
docmgr changelog update --ticket $TICKET \
  --entry "Fixed auth bug: JWT expiry not validated" \
  --file-note "backend/auth/jwt.go:Added expiry check"

docmgr task check --ticket $TICKET --id <investigation-task>
```

## Skill Type

**Rigid** — Follow all four phases. No fixes without Phase 1 complete.

## Provenance

Adapted from Superpowers' systematic debugging skill:
- [`superpowers/skills/systematic-debugging/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/systematic-debugging/SKILL.md)

