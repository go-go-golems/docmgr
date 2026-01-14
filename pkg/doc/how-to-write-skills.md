---
Title: How to Write Skills
Slug: how-to-write-skills
Short: Guide for creating, structuring, and maintaining skill documents that teach LLMs disciplined workflows.
Topics:
- docmgr
- skills
- documentation
- writing
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# How to Write Skills

## Overview

Skills are structured markdown documents that teach LLMs (and humans) to follow disciplined workflows. Unlike general documentation that describes what exists, skills prescribe how to work—they're executable playbooks that enforce best practices like test-driven development, systematic debugging, and structured design processes. A well-written skill transforms "maybe I should write tests" into "you MUST write a failing test first, watch it fail, then implement." This document teaches you how to create effective skills that LLMs will follow reliably.

Skills in docmgr are plan files (`skill.yaml`) that live under `ttmp/skills/` or a ticket’s `skills/` folder. The plan’s `skill` metadata includes `what_for` (what the skill accomplishes) and `when_to_use` (when to apply it). These fields, combined with clear structure and strong enforcement language, ensure skills are discoverable and consistently applied.

**What you'll learn:**
- The anatomy of a skill document (frontmatter, structure, enforcement patterns)
- How to write trigger conditions that LLMs can match reliably
- Enforcement techniques that prevent workflow skipping
- A complete example skill from start to finish
- Testing and iteration strategies

---

## When to Create a Skill

Create a skill when you have a workflow that:
- **Should be mandatory** — Not a suggestion, but a required process (TDD, code review, systematic debugging)
- **Is repeatable** — Same steps apply across multiple situations
- **Has failure modes** — Common mistakes or shortcuts you want to prevent
- **Requires discipline** — Easy to skip or do incorrectly without structure

**Good candidates:**
- Test-driven development (prevents code before tests)
- Systematic debugging (prevents random fixes)
- Design brainstorming (prevents jumping to code)
- Code review checklists (prevents incomplete reviews)
- Documentation standards (prevents inconsistent docs)

**Poor candidates:**
- One-off procedures (use a playbook document instead)
- Highly variable processes (too much context-dependence)
- Simple reminders (just add to a checklist)

---

## Skill Plan Contract (skill.yaml)

Skills in docmgr use a `skill.yaml` plan file with explicit metadata and sources:

```yaml
skill:
  name: test-driven-development
  title: Test-Driven Development
  description: Enforces the RED-GREEN-REFACTOR cycle.
  what_for: Ensure every function has a failing test before implementation.
  when_to_use: Use when implementing features or refactoring behavior.
  topics: [testing, tdd, quality]
  license: Proprietary
  compatibility: Requires go test tooling.

sources:
  - type: file
    path: backend/testing/framework.md
    output: references/testing-framework.md
    strip-frontmatter: true
    append_to_body: false

  - type: binary-help
    binary: glaze
    topic: help-system
    output: references/glaze-help-system.md
    wrap: markdown

output:
  skill_dir_name: test-driven-development
  skill_md:
    include_index: true
    index_title: References
```

**Key fields explained:**

- **`skill.what_for`**: Explains what the skill accomplishes. Keep it concise (2-3 sentences). Focus on outcomes and benefits, not process steps.

- **`skill.when_to_use`**: The trigger condition that helps LLMs (and humans) decide when to apply this skill. Use clear "use when" language: "Use when implementing any feature", "Use when encountering any bug", "Use when starting creative work".

- **`skill.topics`**: Enable filtering with `docmgr skill list --topics testing`. Choose topics that match how developers think about the domain.

- **`sources`**: Declares explicit files or binary help output that should be packaged into the skill. `docmgr skill list --file` and `--dir` filter against `file` sources.
- **`sources[].append_to_body`**: When true, the resolved content is appended into the main SKILL.md body (in order) before the references index, and the source output file is not written. When any append-to-body content exists, the auto-generated intro/WhatFor/WhenToUse sections are suppressed to avoid duplicate headers. If the appended content already starts with a `# Title`, the exporter also skips the generated title to prevent duplication.

- **`output`**: Controls export naming and how `SKILL.md` is generated.

---

## Skill Document Structure

The document body follows a consistent structure that makes skills easy to understand and apply. Each section serves a specific purpose in guiding the LLM's behavior.

### Recommended Structure

```markdown
# [Skill Name]

## Overview
[2-3 sentences: what this skill does and why it matters]

## When to Use
[Clear trigger conditions with examples]

## The Iron Law (or Core Principle)
[The non-negotiable rule this skill enforces]

## The Process
[Step-by-step workflow with concrete actions]

## Red Flags
[Common rationalizations and why they don't hold]

## Verification Checklist
[Items that must be checked before marking complete]

## Integration
[Other skills this requires or references]

## Examples
[Good vs bad examples showing the pattern]
```

**Why this structure:**
- **Overview**: Quick understanding before committing to read the whole skill
- **When to Use**: Explicit matching criteria (prevents "I'm not sure if this applies")
- **Iron Law**: Sets expectation that this is mandatory, not optional
- **The Process**: Concrete steps prevent ambiguity
- **Red Flags**: Anticipates and blocks common shortcuts
- **Verification**: Ensures completeness before moving on
- **Integration**: Creates workflow chains (brainstorming → planning → execution)
- **Examples**: Shows what good/bad looks like in practice

---

## Enforcement Techniques

Skills need strong enforcement language to prevent LLMs from skipping steps. These techniques come directly from analyzing Superpowers' most effective skills.

### 1. Use `<EXTREMELY_IMPORTANT>` Tags

Wrap critical rules in XML-style tags that signal high importance:

```markdown
<EXTREMELY_IMPORTANT>
If you write production code before writing a failing test, DELETE the code
and start over. No exceptions. Don't keep it as "reference."
</EXTREMELY_IMPORTANT>
```

### 2. State "Iron Laws"

Lead with a non-negotiable rule in a prominent code block:

```markdown
## The Iron Law

```
NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST
```

Write code before the test? Delete it. Start over.
```

### 3. Include Red Flags Tables

Anticipate common rationalizations and provide counter-arguments:

```markdown
## Red Flags

These thoughts mean STOP—you're rationalizing:

| Thought | Reality |
|---------|---------|
| "This is too simple to test" | Simple code breaks. Test takes 30 seconds. |
| "I'll test after to verify it works" | Tests passing immediately prove nothing. |
| "Tests after achieve same goals" | Tests-after = "what does this do?" Tests-first = "what should this do?" |
```

### 4. Provide Verification Checklists

Create checklists that must be completed before marking work done:

```markdown
## Verification Checklist

Before marking work complete:

- [ ] Every new function/method has a test
- [ ] Watched each test fail before implementing
- [ ] Each test failed for expected reason
- [ ] All tests pass
- [ ] Output pristine (no errors, warnings)
```

**Important:** When you include a checklist, note in the skill that the LLM should convert checklist items to docmgr tasks:

```markdown
**Checklist discipline:** Convert each checklist item into a docmgr task:
- `docmgr task add --ticket <TICKET> --text "<checklist item>"`
```

### 5. Use Strong Modal Language

Don't suggest—mandate:

- ❌ "You should write tests first"
- ✅ "You MUST write tests first"
- ❌ "Consider checking for skills"
- ✅ "You MUST check for skills BEFORE ANY RESPONSE"
- ❌ "It's good practice to..."
- ✅ "NEVER skip this step"

---

## Skill Types: Rigid vs Flexible

Not all skills need the same enforcement level. Classify your skill to set expectations:

**Rigid Skills** (TDD, debugging, code review):
- Follow exactly as written
- No adaptation allowed
- Strong enforcement language
- Verification checklists mandatory
- Example: "Write code before test? Delete it. Start over."

**Flexible Skills** (design patterns, architecture):
- Adapt principles to context
- Less prescriptive language
- Guidelines rather than rules
- Example: "Consider these patterns, choose what fits your context."

Include a note in your skill:

```markdown
## Skill Type

**Rigid** — Follow this process exactly. Don't adapt away discipline.

(or)

**Flexible** — Adapt these principles to your specific context.
```

---

## Complete Example: Test-Driven Development Skill

Here's a complete skill that demonstrates all the techniques we've covered. This example is adapted from Superpowers' TDD skill but using docmgr commands.

```yaml
# skill.yaml
skill:
  name: test-driven-development
  title: Test-Driven Development
  description: Enforces RED-GREEN-REFACTOR cycle for every function.
  what_for: Ensure every function has a failing test before implementation.
  when_to_use: Use when implementing features or refactoring behavior.
  topics: [testing, tdd, quality]

sources:
  - type: file
    path: backend/testing/framework.md
    output: references/testing-framework.md

output:
  skill_dir_name: test-driven-development
```

```markdown
# SKILL.md
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
go test ./backend/retry -v -run TestRetryFailedOperations
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
go test ./backend/retry -v
```

Confirm:
- Test passes
- Other tests still pass
- Output pristine (no errors, warnings)

**Test fails?** Fix code, not test.

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

**All of these mean: Delete code. Start over with TDD.**

## Verification Checklist

Before marking work complete:

- [ ] Every new function/method has a test
- [ ] Watched each test fail before implementing
- [ ] Each test failed for expected reason
- [ ] Wrote minimal code to pass each test
- [ ] All tests pass
- [ ] Output pristine (no errors, warnings)

**Checklist discipline:** Convert each item into a docmgr task:

```bash
docmgr task add --ticket <TICKET> --text "Every new function has a test"
docmgr task add --ticket <TICKET> --text "Watched each test fail first"
# ... etc for each item
```

Track progress:
```bash
docmgr task list --ticket <TICKET>
docmgr task check --ticket <TICKET> --id 1
```

## Integration with Other Skills

**Related skills:**
- `systematic-debugging` — For creating failing test case when fixing bugs
- `verification-before-completion` — Verify fix worked before claiming success

## Skill Type

**Rigid** — Follow this process exactly. Don't adapt away discipline.
```

**Key elements in this example:**
1. **Clear frontmatter** with all required fields
2. **Strong opening** with core principle
3. **Iron Law** in prominent code block
4. **Step-by-step process** with concrete commands
5. **Red flags table** preventing rationalization
6. **Verification checklist** with docmgr task mapping
7. **Examples** showing good code patterns
8. **Integration** referencing other skills

---

## Writing Effective Trigger Conditions

The `WhenToUse` field is your skill's matching criteria. Write it so LLMs can reliably determine if the skill applies.

### Good Trigger Conditions

**Be explicit about circumstances:**
```yaml
WhenToUse: |
  Use when implementing any feature or bugfix, before writing implementation
  code. Also use when refactoring existing code or changing behavior.
```

**Use action-oriented language:**
```yaml
WhenToUse: |
  Use when encountering any bug, test failure, or unexpected behavior,
  before proposing fixes.
```

**Specify the timing:**
```yaml
WhenToUse: |
  Use BEFORE any creative work - creating features, building components,
  adding functionality, or modifying behavior. Activates before coding starts.
```

### Bad Trigger Conditions

❌ Too vague:
```yaml
WhenToUse: Use for quality improvements
```

❌ Too narrow (misses obvious cases):
```yaml
WhenToUse: Use only when fixing P0 production bugs in the auth system
```

❌ Passive language:
```yaml
WhenToUse: Can be helpful when thinking about architecture
```

### Testing Trigger Conditions

After writing your trigger condition, test it against real scenarios:

1. **List recent work items** from your team's backlog
2. **Check if trigger would fire** for items where skill should apply
3. **Check for false positives** (skill triggers when it shouldn't)
4. **Refine language** until matching is reliable

---

## Enforcement Language Patterns

Effective skills use specific language patterns borrowed from Superpowers that have proven to work with LLMs.

### Pattern 1: Mandatory Opening Statement

```markdown
<EXTREMELY_IMPORTANT>
[Critical rule that cannot be skipped]
</EXTREMELY_IMPORTANT>
```

### Pattern 2: Iron Laws

```markdown
## The Iron Law

```
[CAPITALIZED IMPERATIVE STATEMENT]
```

[Consequence if violated]
```

### Pattern 3: Red Flags Tables

```markdown
## Red Flags

These thoughts mean STOP—you're rationalizing:

| Thought | Reality |
|---------|---------|
| "[common excuse]" | [why it doesn't hold] |
| "[another excuse]" | [counter-argument] |
```

### Pattern 4: Verification Gates

```markdown
## Verification Checklist

Before marking work complete:

- [ ] [Required check 1]
- [ ] [Required check 2]

Can't check all boxes? You skipped [workflow]. Start over.
```

### Pattern 5: Direct Commands

Use imperative mood, not suggestions:

- ✅ "Delete the code. Start over."
- ❌ "You might want to consider rewriting."
- ✅ "Run the test. Verify it fails."
- ❌ "It's good practice to run tests."

---

## Common Pitfalls

### Pitfall 1: Suggesting Instead of Requiring

**Problem:** Weak language allows LLMs to skip steps.

❌ Bad:
```markdown
It's a good idea to write tests first. This helps ensure quality.
```

✅ Good:
```markdown
You MUST write tests first. Write code before the test? Delete it. Start over.
```

### Pitfall 2: Missing Red Flags

**Problem:** LLM convinces itself "this time is different."

❌ Bad:
```markdown
## Process
1. Write test
2. Write code
```

✅ Good:
```markdown
## Process
1. Write test
2. Write code

## Red Flags
- "Too simple to test" → Simple code breaks. Test takes 30 seconds.
- "I'll test after" → Tests passing immediately prove nothing.
```

### Pitfall 3: Vague Process Steps

**Problem:** Ambiguity leads to skipped verification.

❌ Bad:
```markdown
1. Write test
2. Make sure it works
3. Write code
```

✅ Good:
```markdown
1. RED: Write failing test
2. Verify RED: Run test, confirm it fails with expected message
3. GREEN: Write minimal code to pass
4. Verify GREEN: Run test, confirm it passes
```

### Pitfall 4: No Verification Checklist

**Problem:** "Done" becomes subjective without explicit completion criteria.

❌ Bad:
```markdown
Follow these steps and you're done.
```

✅ Good:
```markdown
## Verification Checklist

Before marking complete:
- [ ] Watched test fail first
- [ ] Test failed for expected reason
- [ ] All tests pass

Can't check all boxes? Start over.
```

---

## Skill Discovery and Linking

Make your skills discoverable through multiple paths:

### 1. Topics

Choose topics that match how developers think:

```yaml
Topics:
  - testing      # What domain?
  - quality      # What goal?
  - backend      # What layer?
```

Test discovery:
```bash
docmgr skill list --topics testing
docmgr skill list --topics quality,backend
```

### 2. Related Files

Link to code files where the skill applies:

```yaml
RelatedFiles:
  - Path: backend/api/handlers.go
    Note: Main API handlers that should follow TDD
  - Path: backend/api/handlers_test.go
    Note: Example tests showing TDD pattern
```

Test discovery:
```bash
docmgr skill list --file backend/api/handlers.go
docmgr skill list --dir backend/api/
```

### 3. Skill Chaining

Reference other skills explicitly:

```markdown
## Integration

**Related skills:**
- `systematic-debugging` — For creating failing test when fixing bugs
- `code-review` — Review gates after implementation

**Required skills:**
- Use `brainstorming` before starting creative work
- Use `verification-before-completion` after claiming "done"
```

---

## Where to Store Skills

### Workspace-Level Skills

For skills that apply across all tickets:

```
ttmp/skills/
├── test-driven-development/skill.yaml
├── systematic-debugging/skill.yaml
├── code-review/skill.yaml
└── brainstorming/skill.yaml
```

**Frontmatter:** Omit `Ticket` field or use a generic ticket like `000-WORKSPACE-SKILLS`

**When to use:** Process skills, quality standards, team-wide workflows

### Ticket-Level Skills

For skills specific to a feature or domain:

```
ttmp/YYYY/MM/DD/TICKET--slug/skills/
├── auth-implementation/skill.yaml
├── websocket-testing/skill.yaml
└── frontend-component-patterns/skill.yaml
```

**Frontmatter:** Include `Ticket` field

**When to use:** Domain-specific patterns, experimental workflows, ticket-scoped processes

**Note on convention:** `docmgr doc add --doc-type skill` still creates DocType skill docs under the doc-type folder, but `docmgr skill list/show` operate on `skill.yaml` plans. Use the `skills/` folders for plan-based skills.

---

## Migration from DocType Skill Docs

DocType skill documents are still valid workflow docs, but they are no longer used by `docmgr skill list/show`. To migrate a DocType skill into a plan:

1. Create `ttmp/skills/<skill-name>/skill.yaml` (or `<ticket>/skills/<skill-name>/skill.yaml`).
2. Copy `Title` → `skill.title`, `WhatFor` → `skill.what_for`, `WhenToUse` → `skill.when_to_use`, and `Topics` → `skill.topics`.
3. Add `sources` entries for any reference files the skill needs (or move the skill body into the exported `SKILL.md` during `docmgr skill export`).
4. Validate with `docmgr skill show <name>` and export with `docmgr skill export <name> --output-skill dist/<name>.skill`.

---

## Testing Your Skill

After writing a skill, test it before sharing with your team:

### 1. Verify Discoverability

```bash
# Can it be found by topic?
docmgr skill list --topics testing

# Can it be found by related file?
docmgr skill list --file backend/api/handlers.go

# Can it be loaded?
docmgr skill show test-driven-development
```

### 2. Test with an LLM

Paste the skill content into an LLM session and ask it to apply the skill to a real task:

```
I need to implement a retry function for failed API calls.
Use the test-driven-development skill.
```

**Watch for:**
- Does the LLM follow the process correctly?
- Does it try to skip steps? (If yes, strengthen enforcement language)
- Does it understand the verification checklist?
- Does it convert checklist to docmgr tasks?

### 3. Check Against Red Flags

Review your skill's red flags table. For each entry, ask:
- Have I actually seen this rationalization in practice?
- Is the counter-argument clear and convincing?
- Are there other common shortcuts I'm missing?

Add any rationalizations you've encountered.

### 4. Get Team Feedback

Share the skill with teammates:
- Is the trigger condition clear?
- Are the steps concrete enough?
- Do the examples help?
- What shortcuts are they tempted to take? (Add to red flags)

---

## Maintaining Skills

Skills evolve as you discover new failure modes and better practices.

### When to Update a Skill

**Add to red flags when:**
- You or teammates skip a step with a new rationalization
- You notice a pattern of "this time is different" excuses

**Update the process when:**
- You discover a better way to accomplish the same goal
- Tools change (e.g., new testing framework)
- Steps prove unnecessary in practice

**Add examples when:**
- You see confusion about "good vs bad"
- New edge cases emerge

### Versioning Skills

Use changelog entries to track skill changes:

```bash
docmgr changelog update --ticket <TICKET> \
  --entry "Updated TDD skill: added 'Keep as reference' to red flags table" \
  --file-note "ttmp/skills/test-driven-development.md:Updated red flags"
```

If the skill is workspace-level (not ticket-specific), relate it to a workspace documentation ticket.

### Deprecating Skills

When a skill becomes obsolete:

1. Update `Status: archived` in frontmatter
2. Add deprecation notice at the top:
   ```markdown
   > **DEPRECATED:** This skill is obsolete. Use `new-skill-name` instead.
   ```
3. Add changelog entry explaining why

---

## Style Guidelines

Follow these conventions for consistency:

### Voice and Tone

- **Imperative mood**: "Write tests first" (not "You should write tests first")
- **Direct address**: "You MUST" (not "One must" or "Developers should")
- **Active voice**: "The skill enforces TDD" (not "TDD is enforced by the skill")
- **Strong modals**: MUST, NEVER, ALWAYS (not "should", "might", "consider")

### Formatting

- **Code blocks**: Always specify language (```bash, ```go, ```yaml)
- **Commands**: Show exact commands with expected output
- **Tables**: Use for red flags and comparisons
- **Checklists**: Use `- [ ]` format for verification items
- **Emphasis**: Use **bold** for important terms, `code` for literals

### Section Headers

Use consistent header hierarchy:
- `##` for major sections (Overview, Process, Red Flags)
- `###` for subsections (RED, GREEN, REFACTOR)
- `####` for detailed breakdowns (rarely needed)

---

## Resources and References

### Superpowers Skills (Upstream Reference)

This guide is adapted from Superpowers' skill-writing practices:

- **Writing Skills Skill**: [`superpowers/skills/writing-skills/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/writing-skills/SKILL.md)
- **TDD Skill Example**: [`superpowers/skills/test-driven-development/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/test-driven-development/SKILL.md)
- **Using Superpowers Skill**: [`superpowers/skills/using-superpowers/SKILL.md`](https://github.com/obra/superpowers/blob/main/skills/using-superpowers/SKILL.md)
- **Superpowers Repository**: https://github.com/obra/superpowers

### Related docmgr Documentation

- `docmgr help using-skills` — LLM bootstrap prompt for skills usage
- `docmgr help how-to-use` — General docmgr tutorial
- `docmgr help templates-and-guidelines` — Document templates system

### Internal Tickets

- `001-ADD-CLAUDE-SKILLS` — Skills feature implementation design
- `002-ANALYZE-SUPERPOWERS` — Analysis of Superpowers techniques
- `003-CREATE-SKILL-PROMPTS` — This documentation effort

---

## Quick Reference: Skill Checklist

When creating a new skill, ensure it has:

- [ ] Complete skill.yaml metadata (`skill.name`, `skill.description`, `skill.what_for`, `skill.when_to_use`, `skill.topics`)
- [ ] Clear trigger condition in `WhenToUse`
- [ ] Overview section (2-3 sentences)
- [ ] Iron Law or Core Principle (if rigid skill)
- [ ] Step-by-step process with concrete commands
- [ ] Red Flags table with common rationalizations
- [ ] Verification checklist (if applicable)
- [ ] Examples (good vs bad)
- [ ] Integration section (related skills)
- [ ] Skill type classification (Rigid vs Flexible)
- [ ] Related files linked (with notes)
- [ ] Tested with docmgr commands (`skill list`, `skill show`)
- [ ] Tested with real LLM session

---

## Next Steps

After creating your skill:

1. **Test discovery**:
   ```bash
   docmgr skill list --topics <your-topics>
   docmgr skill show <your-skill-name>
   ```

2. **Test with LLM**: Paste into a real session and verify behavior

3. **Document creation**: Add changelog entry and relate the skill file:
   ```bash
   docmgr changelog update --ticket <TICKET> \
     --entry "Created new skill: <skill-name>" \
     --file-note "<skill-path>.md:New skill for <purpose>"
   ```
   Where `<skill-path>` is typically:
   - workspace-level: `ttmp/skills/<skill-slug>`
   - ticket-level: `ttmp/YYYY/MM/DD/<TICKET>--<slug>/skill/<skill-slug>`

4. **Share with team**: Get feedback on trigger conditions and enforcement

5. **Iterate**: Update based on real usage patterns
