---
Title: "Skill: Test-Driven Development"
DocType: skill
Topics:
  - testing
  - tdd
  - quality
WhatFor: |
  Enforces RED-GREEN-REFACTOR cycle to ensure every function has a test
  that was verified to fail before implementation. Prevents untested code
  and ensures tests actually verify behavior.
WhenToUse: |
  Use when implementing any feature or bugfix, before writing implementation
  code. Also use when refactoring existing code or changing behavior.
Status: active
Intent: long-term
---

# Skill: Test-Driven Development

## Overview

Write the test first. Watch it fail. Write minimal code to pass.

**Core principle:** If you didn't watch the test fail, you don't know if it tests the right thing.

**Violating the letter of the rules is violating the spirit of the rules.**

## When to Use

**Always:**
- New features
- Bug fixes
- Refactoring
- Behavior changes

**Exceptions (ask your human partner):**
- Throwaway prototypes
- Generated code
- Configuration files

Thinking "skip TDD just this once"? Stop. That's rationalization.

## The Iron Law

```
NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST
```

Write code before the test? Delete it. Start over.

**No exceptions:**
- Don't keep it as "reference"
- Don't "adapt" it while writing tests
- Don't look at it
- Delete means delete

Implement fresh from tests. Period.

## The Process: Red-Green-Refactor

### RED - Write Failing Test

Write one minimal test showing what should happen.

**Good example:**
```go
func TestRetryFailedOperations(t *testing.T) {
    attempts := 0
    operation := func() error {
        attempts++
        if attempts < 3 {
            return errors.New("fail")
        }
        return nil
    }

    err := RetryOperation(operation, 3)
    
    assert.NoError(t, err)
    assert.Equal(t, 3, attempts)
}
```

**Requirements:**
- One behavior
- Clear name
- Real code (no mocks unless unavoidable)

### Verify RED - Watch It Fail

**MANDATORY. Never skip.**

```bash
go test ./pkg/retry -v -run TestRetryFailedOperations
```

Confirm:
- Test fails (not errors)
- Failure message is expected
- Fails because feature missing (not typos)

**Test passes?** You're testing existing behavior. Fix test.

**Test errors?** Fix error, re-run until it fails correctly.

### GREEN - Minimal Code

Write simplest code to pass the test.

```go
func RetryOperation(fn func() error, maxAttempts int) error {
    for i := 0; i < maxAttempts; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        if i == maxAttempts-1 {
            return err
        }
    }
    return nil
}
```

Don't add features, refactor other code, or "improve" beyond the test.

### Verify GREEN - Watch It Pass

**MANDATORY.**

```bash
go test ./pkg/retry -v
```

Confirm:
- Test passes
- Other tests still pass
- Output pristine (no errors, warnings)

**Test fails?** Fix code, not test.

**Other tests fail?** Fix now.

### REFACTOR - Clean Up

After green only:
- Remove duplication
- Improve names
- Extract helpers

Keep tests green. Don't add behavior.

### Repeat

Next failing test for next feature.

## Red Flags

These thoughts mean STOP — you're rationalizing:

| Thought | Reality |
|---------|---------|
| "Too simple to test" | Simple code breaks. Test takes 30 seconds. |
| "I'll test after" | Tests passing immediately prove nothing. |
| "Tests after achieve same goals" | Tests-after = "what does this do?" Tests-first = "what should this do?" |
| "Already manually tested" | Ad-hoc ≠ systematic. No record, can't re-run. |
| "Deleting X hours is wasteful" | Sunk cost fallacy. Keeping unverified code is technical debt. |
| "Keep as reference, write tests first" | You'll adapt it. That's testing after. Delete means delete. |
| "TDD is dogmatic, I'm being pragmatic" | TDD IS pragmatic. Debugging after is slower. |

**All of these mean: Delete code. Start over with TDD.**

## Verification Checklist

Before marking work complete:

- [ ] Every new function/method has a test
- [ ] Watched each test fail before implementing
- [ ] Each test failed for expected reason (feature missing, not typo)
- [ ] Wrote minimal code to pass each test
- [ ] All tests pass
- [ ] Output pristine (no errors, warnings)
- [ ] Tests use real code (mocks only if unavoidable)
- [ ] Edge cases and errors covered

Can't check all boxes? You skipped TDD. Start over.

**Checklist discipline:** Convert each item into a docmgr task:

```bash
TICKET="<current-ticket>"
docmgr task add --ticket $TICKET --text "Every new function has a test"
docmgr task add --ticket $TICKET --text "Watched each test fail before implementing"
docmgr task add --ticket $TICKET --text "Each test failed for expected reason"
docmgr task add --ticket $TICKET --text "Wrote minimal code to pass tests"
docmgr task add --ticket $TICKET --text "All tests pass with pristine output"
docmgr task add --ticket $TICKET --text "Tests use real code (minimal mocks)"
docmgr task add --ticket $TICKET --text "Edge cases and errors covered"
```

Track progress:
```bash
docmgr task list --ticket $TICKET
docmgr task check --ticket $TICKET --id 1
```

## Integration with Other Skills

**Related skills:**
- `systematic-debugging` — For creating failing test case when fixing bugs
- `documenting-as-you-code` — Document what you build while following TDD

## Skill Type

**Rigid** — Follow this process exactly. Don't adapt away discipline.

## Provenance

Adapted from Superpowers' TDD skill:
- [`superpowers/skills/test-driven-development/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/test-driven-development/SKILL.md)

