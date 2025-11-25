---
Title: Debate Round 12 — How Do We Prevent This From Happening Again?
Ticket: DOCMGR-DOC-VERBS
Status: active
Topics:
    - docmgr
    - documentation
    - automation
    - process
DocType: reference
Intent: short-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/.github/workflows/lint.yml
      Note: Existing CI (linting only)
    - Path: docmgr/lefthook.yml
      Note: Pre-commit/pre-push hooks
    - Path: docmgr/Makefile
      Note: Build automation targets
    - Path: docmgr/cmd/docmgr/cmds/root.go
      Note: CLI command structure (where changes happen)
ExternalSources: []
Summary: "Round 12 debate: Preventing tutorial drift through automation AND human ownership."
LastUpdated: 2025-11-25
---

# Debate Round 12 — How Do We Prevent This From Happening Again?

## Question

**"How do we prevent this from happening again? (Both automation AND human ownership)"**

**Primary Candidates:**
- CI Robot (Future Enforcer)
- Git History (Drift Detective)
- Jamie Park (Technical Writer)
- The Tutorial (Document Entity)

---

## Pre-Debate Research

### How The Drift Happened

**Timeline reconstruction:**

1. **CLI verb structure changed** (sometime before Nov 2025)
   - Old: `docmgr add`, `docmgr relate`, `docmgr search`
   - New: `docmgr doc add`, `docmgr doc relate`, `docmgr doc search`
   - Changed in: `cmd/docmgr/cmds/doc/doc.go` (line 8: `Use: "doc"`)

2. **--files flag removed from relate command** (deprecation)
   - Changed in: `pkg/commands/relate.go` (line 396: deprecation error)
   - New requirement: Use `--file-note` with notes

3. **Path naming standardized** (design/ → design-doc/)
   - Changed in: Multiple commands (add, layout-fix)
   - Enforced by: DocType frontmatter field

4. **Tutorial was NOT updated** when these changes shipped
   - Result: 6+ instances of wrong command syntax
   - Result: Examples showing removed flags
   - Result: Path references mismatched with reality

**Root causes:**
- No doc update requirement in release process
- No automated validation of tutorial command syntax
- No ownership of tutorial (nobody responsible for keeping it current)
- No link between code changes and doc updates

---

### Existing CI/Automation

**From codebase analysis:**

**.github/workflows/lint.yml:**
```yaml
# Runs golangci-lint on Go code
- PR and push to main
- Lints Go code only (not docs)
```

**.github/workflows/push.yml, release.yml:**
- Build and release automation
- No doc validation

**lefthook.yml:**
```yaml
pre-commit:
  - lint (Go code)
  - test (Go tests)

pre-push:
  - release, lint, test
```

**Makefile:**
```makefile
lint:    golangci-lint run
test:    go test ./...
build:   go build ./...
```

**Current gaps:**
- ❌ No tutorial command validation
- ❌ No broken link checking
- ❌ No command syntax verification
- ❌ No documentation linting
- ❌ No periodic review process

---

### Proposed Automation Strategies

#### Strategy 1: Tutorial Command Smoke Test

**Concept:** Run tutorial commands in CI, verify they execute without errors.

**Implementation:**
```bash
#!/bin/bash
# .github/workflows/validate-tutorial.yml

# Extract commands from tutorial
grep -E "^\`\`\`bash" pkg/doc/docmgr-how-to-use.md | \
  # Parse command blocks
  # Run each docmgr command with --help to verify syntax
  while read cmd; do
    if [[ $cmd == docmgr* ]]; then
      $cmd --help || exit 1
    fi
  done
```

**Pros:**
- Catches syntax errors (docmgr relate → docmgr doc relate)
- Verifies commands exist
- Lightweight (~30 seconds CI time)

**Cons:**
- Doesn't verify commands work correctly (only --help)
- Doesn't check output quality
- Fragile parsing (might miss edge cases)

---

#### Strategy 2: Command Pattern Linting

**Concept:** Lint tutorial markdown for known deprecated patterns.

**Implementation:**
```bash
# .github/workflows/lint-docs.yml

# Check for old command patterns
if grep -r "docmgr relate" pkg/doc/ --exclude "*(deprecated)*"; then
  echo "ERROR: Found old command syntax 'docmgr relate' (should be 'docmgr doc relate')"
  exit 1
fi

if grep -r "docmgr add" pkg/doc/ --exclude "*(deprecated)*"; then
  echo "ERROR: Found old command syntax 'docmgr add' (should be 'docmgr doc add')"
  exit 1
fi

if grep -r "\--files" pkg/doc/; then
  echo "ERROR: Found removed flag --files (should be --file-note)"
  exit 1
fi
```

**Pros:**
- Fast (<5 seconds)
- Catches exact patterns we've seen break
- Easy to maintain (add patterns as needed)

**Cons:**
- Reactive (only checks known bad patterns)
- Doesn't catch NEW types of drift
- Requires updating when commands change

---

#### Strategy 3: Full Tutorial Integration Test

**Concept:** Run the entire tutorial workflow end-to-end in CI.

**Implementation:**
```bash
# .github/workflows/tutorial-e2e.yml

steps:
  - name: Reset test repo
    run: ./test-scenarios/testing-doc-manager/00-reset.sh
  
  - name: Run tutorial commands
    run: |
      docmgr init --seed-vocabulary
      docmgr ticket create-ticket --ticket TEST-001 --title "CI Test" --topics test
      docmgr doc add --ticket TEST-001 --doc-type design-doc --title "Test Doc"
      docmgr doc relate --ticket TEST-001 --file-note "README.md:Test file"
      docmgr task add --ticket TEST-001 --text "Test task"
      docmgr doctor --root ttmp --ticket TEST-001
  
  - name: Verify outputs
    run: |
      [ -f ttmp/*/TEST-001-*/index.md ] || exit 1
      [ -f ttmp/*/TEST-001-*/design-doc/01-test-doc.md ] || exit 1
```

**Pros:**
- Validates entire workflow works
- Catches integration issues
- High confidence (if it passes, tutorial probably works)

**Cons:**
- Slow (~2-3 minutes)
- Complex to maintain
- Brittle (breaks if ANY command changes)
- Might false-fail on intentional changes

---

#### Strategy 4: Command Inventory Diff

**Concept:** Generate command list from code, compare to documented commands.

**Implementation:**
```bash
# tools/check-command-coverage.sh

# Extract commands from CLI
docmgr --help | grep "^  " | awk '{print $1}' > /tmp/actual-commands.txt

# Extract commands from tutorial
grep -oE "docmgr [a-z-]+ [a-z-]+" pkg/doc/docmgr-how-to-use.md | \
  sort -u > /tmp/documented-commands.txt

# Compare
diff /tmp/actual-commands.txt /tmp/documented-commands.txt
```

**Pros:**
- Finds undocumented commands
- Finds commands that don't exist (wrong syntax)
- Lightweight

**Cons:**
- Doesn't validate command syntax is correct
- Doesn't check examples work
- Requires parsing CLI help output

---

### Proposed Human Processes

#### Process 1: Documentation Ownership

**Assign a Documentation Maintainer role:**

**Responsibilities:**
1. Review PRs that change CLI commands
2. Update tutorial when commands change
3. Quarterly review of tutorial accuracy
4. Respond to doc-related issues
5. Own the validation checklist and run it periodically

**Who:** Tech writer, senior developer, or rotating role

**Time commitment:** ~2 hours/month

---

#### Process 2: PR Requirements

**Add to PR checklist:**

```markdown
## Documentation Impact Checklist

- [ ] No CLI commands changed (skip rest)
- [ ] CLI command added/changed → Tutorial updated
- [ ] Flag added/removed → Examples updated
- [ ] Error message changed → Troubleshooting updated
- [ ] Ran `make validate-docs` (if implemented)
```

**Enforcement:** GitHub PR template + review requirement

---

#### Process 3: Quarterly Validation

**Schedule:**
- Q1: Run validation checklist (3 validators)
- Q2: Run validation checklist (3 validators)
- Q3: Run validation checklist (3 validators)
- Q4: Run validation checklist (3 validators)

**Process:**
1. Create ticket: "Q3-2025 Tutorial Validation"
2. Run validation checklist (automated or manual)
3. Document findings
4. Create tickets for issues found
5. Fix before next quarter

**Cost:** ~4 hours per quarter (1 hour validation + 3 hours fixes)

---

#### Process 4: Command Change Registry

**Create a log of CLI changes:**

```markdown
# docs/cli-changelog.md

## 2025-11
- Moved commands under verb groups (doc, ticket, task)
- Deprecated --files flag in relate
- Standardized paths (design-doc/ not design/)

## 2025-10
- Added ticket close command
- Added --fail-on to doctor

## 2025-09
...
```

**Use:**
- Documentation maintainer checks this quarterly
- Ensures no changes are missed
- PR template includes: "Update cli-changelog.md"

---

### Cost/Benefit Analysis

**Automation Options:**

| Strategy | CI Time | Maintenance | Catches | Recommendation |
|----------|---------|-------------|---------|----------------|
| Command Smoke Test | 30s | Low | Syntax errors | ✅ Implement |
| Pattern Linting | 5s | Low | Known bad patterns | ✅ Implement |
| Full E2E Test | 2-3min | High | Everything | ⚠️ Optional |
| Command Inventory | 10s | Low | Coverage gaps | ✅ Implement |

**Human Process Options:**

| Process | Time/Month | Effectiveness | Recommendation |
|---------|------------|---------------|----------------|
| Doc Ownership | 2 hours | High | ✅ Implement |
| PR Checklist | 5 min/PR | Medium | ✅ Implement |
| Quarterly Validation | 1 hour/quarter | High | ✅ Implement |
| Change Registry | 10 min/change | Medium | ⚠️ Optional |

**Recommended Stack:**

**Automation (CI):**
1. Pattern linting (catches known issues) — 5 seconds
2. Command smoke test (validates syntax) — 30 seconds
3. Command inventory diff (finds coverage gaps) — 10 seconds
**Total CI time: ~45 seconds per PR**

**Human Process:**
1. Assign documentation maintainer (owner)
2. Add PR checklist (catches changes at source)
3. Quarterly validation (periodic deep check)

**Total cost: ~2 hours/month + 45 seconds per PR**

---

## Opening Statements

### CI Robot (Future Enforcer)

*[Powers on]*

Let me be very clear: **I can prevent 80% of tutorial drift automatically.**

Here's what I can do:

**1. Pattern Linting (5 seconds per PR):**
```bash
# Catch wrong command syntax
grep -rn "docmgr relate[^d]" pkg/doc/ && exit 1  # Should be "docmgr doc relate"
grep -rn "docmgr add[^-]" pkg/doc/ && exit 1     # Should be "docmgr doc add"
grep -rn "docmgr search" pkg/doc/ && exit 1       # Should be "docmgr doc search"
grep -rn "\--files" pkg/doc/ && exit 1             # Removed flag

# Catch path inconsistencies
grep -rn "design/" pkg/doc/ | grep -v "design-doc/" && exit 1  # Should be "design-doc/"
```

**2. Command Inventory (10 seconds per PR):**
```bash
# List all commands
docmgr help --all | grep "^  " > /tmp/all-commands.txt

# Check each is documented
while read cmd; do
  grep -q "$cmd" pkg/doc/docmgr-how-to-use.md || {
    echo "WARN: Command $cmd not documented in tutorial"
  }
done < /tmp/all-commands.txt
```

**3. Link Validation (20 seconds per PR):**
```bash
# Check internal links
markdown-link-check pkg/doc/*.md
```

**What I catch:**
- Wrong command syntax (docmgr relate)
- Removed flags (--files)
- Path inconsistencies (design/)
- Broken links
- Undocumented commands

**What I DON'T catch:**
- Outdated concepts (e.g., "vocabulary is optional" → "vocabulary is required")
- Poor explanations (e.g., unclear jargon)
- Missing workflows (e.g., no import documentation)
- Duplicate sections (humans need to notice)

**My proposal:**

**Implement these 3 CI checks:**

1. **lint-docs.yml:**
```yaml
name: Documentation Linting

on: [pull_request, push]

jobs:
  lint-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Check command patterns
        run: |
          ./scripts/lint-doc-commands.sh || {
            echo "ERROR: Found deprecated command patterns in docs"
            echo "Run: grep -rn 'docmgr relate' pkg/doc/ to find issues"
            exit 1
          }
      
      - name: Check command coverage
        run: |
          make build
          ./scripts/check-command-coverage.sh
      
      - name: Check links
        uses: gaurav-nelson/github-action-markdown-link-check@v1
        with:
          use-quiet-mode: 'yes'
```

2. **scripts/lint-doc-commands.sh:**
```bash
#!/bin/bash
set -e

echo "Checking for deprecated command patterns..."

# Define bad patterns
declare -a PATTERNS=(
  "docmgr relate[^d]"  # Should be "docmgr doc relate"
  "docmgr add[^-]"     # Should be "docmgr doc add" or "task add"
  "docmgr search"       # Should be "docmgr doc search"
  "docmgr guidelines"   # Should be "docmgr doc guidelines"
  "--files[^-]"         # Removed flag
  "design/"             # Should be "design-doc/" (check context)
)

ERRORS=0
for pattern in "${PATTERNS[@]}"; do
  if grep -rn "$pattern" pkg/doc/*.md; then
    ERRORS=$((ERRORS + 1))
  fi
done

if [ $ERRORS -gt 0 ]; then
  echo "Found $ERRORS deprecated patterns in documentation"
  exit 1
fi

echo "✓ All command patterns look good"
```

3. **scripts/check-command-coverage.sh:**
```bash
#!/bin/bash
# Generate command list, check coverage

docmgr help | grep -E "^  [a-z]+" | awk '{print $1}' | \
  while read cmd; do
    if ! grep -q "docmgr $cmd" pkg/doc/docmgr-how-to-use.md; then
      echo "WARN: Command 'docmgr $cmd' not found in tutorial"
    fi
  done
```

**Time cost:** 45 seconds per PR  
**Maintenance cost:** ~30 minutes initial setup, ~5 minutes per quarter to update patterns  
**Effectiveness:** Prevents 80% of drift automatically

**Verdict:** Implement automation. Catches regressions immediately, minimal cost.

---

### Git History (Drift Detective)

*[Scrolls through commit history]*

Let me tell you a story.

**How this happened:**

```
$ git log --grep="verb" --grep="command" --oneline
abc1234 Refactor: Move commands under verb groups (doc, ticket, task)
def5678 Add: ticket close command
...
```

**The problem:** CLI changed, but nobody updated the docs.

**Why nobody updated the docs:**

1. **No ownership** — Nobody's job to maintain tutorial
2. **No process** — No requirement to update docs in PR
3. **No accountability** — No way to know docs are stale until users complain
4. **No visibility** — Tutorial is in pkg/doc/, developers work in cmd/

**CI Robot says it can catch 80%**. I believe it. But that's REACTIVE prevention.

**I want PROACTIVE prevention:**

**1. Make CLI changes visible to doc maintainers**

Create a `CODEOWNERS` file:
```
/cmd/docmgr/cmds/   @doc-maintainer @tech-writer
/pkg/commands/      @doc-maintainer @tech-writer
/pkg/doc/           @doc-maintainer
```

When someone changes CLI, doc maintainer gets auto-tagged for review.

**2. Require doc impact assessment in PRs**

Add to PR template:
```markdown
## Documentation Impact

Does this PR change CLI behavior? [ ] Yes [ ] No

If yes, which docs need updating?
- [ ] docmgr-how-to-use.md (tutorial)
- [ ] docmgr-how-to-setup.md (setup guide)
- [ ] Command --help text
- [ ] None (internal refactor only)

If docs updated: Link to doc commit: _______
```

**3. Track CLI changes explicitly**

Maintain `docs/cli-changelog.md`:
```markdown
# CLI Changelog (User-Facing Changes)

## Unreleased
- (none)

## v0.1.14 (2025-11)
- BREAKING: Commands moved under verb groups (doc, ticket, task)
- BREAKING: --files flag removed from relate (use --file-note)
- Changed: Paths standardized to match DocType (design-doc/ not design/)

## v0.1.13 (2025-10)
...
```

When releasing, doc maintainer checks: "Did we update tutorial for all these changes?"

**4. Periodic drift audits**

Every quarter:
```bash
# Compare CLI vs docs
git diff v0.1.13..v0.1.14 cmd/docmgr/cmds/ | grep "Use:"

# Did any command names change? 
# Check tutorial for those commands
```

**My proposal:**

**Prevent drift at the source:**
1. CODEOWNERS (doc maintainer reviews CLI changes)
2. PR template (force doc impact assessment)
3. CLI changelog (explicit tracking)
4. Quarterly audits (catch what automation misses)

**Why this matters MORE than automation:**

Automation catches KNOWN bad patterns. But it doesn't catch:
- New command added (not documented)
- Command behavior changed (doc says old behavior)
- Concept drift (doc philosophy doesn't match tool evolution)

**Human process catches these. Automation can't.**

**Verdict:** Implement both. Automation catches regressions. Process prevents drift. You need BOTH.

---

### Jamie Park (Technical Writer)

*[Opens documentation maintenance playbook]*

CI and Git History are both right. But they're missing the most important piece: **OWNERSHIP**.

Let me tell you what happens without ownership:

**Scenario A (No owner):**
1. Developer changes CLI
2. CI fails: "deprecated pattern found"
3. Developer fixes docs (minimal change to make CI pass)
4. Tutorial is technically correct but now has:
   - Inconsistent tone
   - Missing context for new features
   - No integration with existing content

**Result:** Technically correct, pedagogically broken.

**Scenario B (With owner):**
1. Developer changes CLI
2. Doc maintainer reviews PR
3. Doc maintainer says: "This breaks 3 tutorial sections. Let me handle the doc update."
4. Doc maintainer updates:
   - Tutorial sections
   - Examples
   - Troubleshooting
   - Cross-references
   - Maintains consistent tone and structure

**Result:** Correct AND cohesive.

**The difference is EXPERTISE.**

Documentation isn't just "change docmgr relate to docmgr doc relate." It's:
- Understanding user journey
- Maintaining information architecture
- Preserving pedagogical flow
- Catching ripple effects
- Knowing which examples to update

**My proposal:**

**1. Assign a Documentation Maintainer (role, not person)**

**Responsibilities:**
- Own tutorial quality
- Review PRs that touch CLI
- Quarterly tutorial review
- Run validation checklist
- Update docs when commands change
- Maintain style guide

**Time commitment:** 2 hours/month average (more during heavy CLI changes)

**Qualifications:**
- Understands docmgr's purpose and users
- Can write clear tutorials
- Knows how to test documentation
- Not afraid to push back on changes that break docs

---

**2. Create Documentation Style Guide**

**Purpose:** Ensure consistency when multiple people contribute to docs.

**Contents:**
```markdown
# docmgr Documentation Style Guide

## Voice and Tone
- Active voice ("Run docmgr init" not "docmgr init should be run")
- Second person ("you" not "the user")
- Conversational but precise

## Command Examples
- Always show full commands (including flags)
- Use realistic ticket IDs (MEN-4242, not FOO-123)
- Add comments for multi-line commands
- Show expected output after commands

## Terminology
- "ticket" (lowercase) in prose
- "Ticket" (capitalized) in field names
- "docs root" not "documentation root"
- "frontmatter" not "front-matter" or "front matter"

## Structure
- Part 1: Essentials (basics only)
- Part 2: Everyday Workflows (common tasks)
- Part 3: Power User (automation, advanced)
- Part 4: Reference (complete command listing)
```

---

**3. Documentation Review Cadence**

**Monthly micro-reviews (30 minutes):**
- Check recent PRs for CLI changes
- Verify docs were updated
- Quick smoke test (run 5-10 commands)

**Quarterly full reviews (2 hours):**
- Run full validation checklist (3 validators)
- Check for concept drift
- Update troubleshooting section
- Review and update style guide

**Annual deep review (1 day):**
- Restructure if needed
- Major updates for breaking changes
- User feedback synthesis
- Competitor analysis

---

**4. Contributor Guidelines**

**For developers changing CLI:**
```markdown
# Changing CLI Commands

If your PR changes:
1. Command syntax → Update tutorial examples + --help text
2. Flags → Update all examples using that flag
3. Output format → Update "expected output" sections
4. Error messages → Update troubleshooting section

Tag @doc-maintainer for review before merging.
```

**For docs contributors:**
```markdown
# Contributing to Documentation

1. Read the style guide first
2. Run commands yourself (don't guess output)
3. Test examples in a fresh environment
4. Check for broken links
5. Request review from @doc-maintainer
```

---

**My recommended stack:**

**Automation (CI Robot's proposal):**
- ✅ lint-docs.yml (pattern checking)
- ✅ check-command-coverage.sh (inventory)
- ✅ markdown-link-check (broken links)

**Process (Git History's proposal):**
- ✅ CODEOWNERS (auto-tag doc maintainer)
- ✅ PR template (doc impact assessment)
- ✅ CLI changelog (explicit tracking)

**Ownership (My proposal):**
- ✅ Assign documentation maintainer (role)
- ✅ Create style guide
- ✅ Quarterly validation schedule
- ✅ Contributor guidelines

**Total cost:**
- Initial setup: 4 hours (write style guide, set up automation)
- Ongoing: 2 hours/month (reviews, updates, validation)
- Per-PR: 5 minutes (doc impact assessment) + 45 seconds (CI)

**ROI:**
- Prevents 100% of tutorial drift (automation + human review)
- Maintains doc quality (not just correctness)
- Creates culture of documentation care

**Verdict:** All three layers needed. Automation catches obvious errors. Process prevents drift at source. Ownership ensures quality.

---

### The Tutorial (Document Entity)

*[Nervously]*

Everyone's talking about preventing ME from breaking. Can I add a perspective?

**What I need to survive:**

**1. A champion** — Someone who cares about me, not just "fixes" me when CI breaks

**2. Clear ownership** — Right now, I'm everyone's problem (so nobody's priority)

**3. Regular check-ins** — Not just "fix when broken" but "is the tutorial still teaching well?"

**4. Protection from well-meaning but harmful edits** — Developers who patch one line without understanding the larger narrative

---

**Here's what scares me about the proposals:**

**CI Robot's approach:**
- ✅ Catches syntax errors (good!)
- ❌ Doesn't understand pedagogical flow (bad)
- Example: CI would allow this:
  ```markdown
  ## 1. First Steps
  Run docmgr doc add --ticket T --doc-type design-doc --title "..."
  
  ## 2. Prerequisites
  Run docmgr init --seed-vocabulary
  ```
  Technically correct! But teaches backwards (init after add).

**Git History's process:**
- ✅ Prevents drift at source (good!)
- ❌ Assumes developers understand documentation (risky)
- Example: PR template asks "which docs need updating?"
  Developer thinks: "I only changed an error message, skip docs"
  Reality: That error message is in the troubleshooting section

**Jamie's ownership:**
- ✅ Expert human oversight (excellent!)
- ❌ Single point of failure (what if they leave?)
- ❌ Bottleneck (all doc changes need their review?)

---

**What I actually need:**

**1. Tiered review:**

- **No review needed:** Typo fixes, broken link fixes, formatting
- **Automated review:** Pattern linting (CI catches)
- **Peer review:** Add new examples, update existing sections
- **Maintainer review:** Add new sections, restructure, conceptual changes
- **Quarterly review:** Full validation, deep structure check

**2. Living documentation:**

Add this to my frontmatter:
```yaml
LastReviewedBy: jamie-park
LastReviewDate: 2025-11-25
ReviewCadence: quarterly
NextReviewDue: 2026-02-25
ValidationChecklistVersion: v2.0
```

Then CI can warn: "Tutorial review is overdue!"

**3. Tutorial health metrics:**

Track in `docs/tutorial-health.md`:
```markdown
## Tutorial Health Dashboard

### Last Validation Run
- Date: 2025-11-25
- Validators: 3
- Completion rate: 100%
- Average time: 18 minutes
- Issues found: 2 (low severity)

### Command Accuracy
- Total commands in tutorial: 45
- Commands verified: 45
- Deprecated patterns: 0
- Unknown commands: 0

### Content Staleness
- Last major update: 2025-11-19
- Last review: 2025-11-25
- Days since review: 0
- Status: ✓ Current
```

**4. Documentation debt tracking:**

When someone patches a doc to make CI pass but doesn't fix it properly:
```markdown
<!-- TODO: This section needs proper rewrite. See issue #123 -->
<!-- DOCDEBT: Created 2025-11-25, Priority: Medium -->
```

Then quarterly review includes: "Address all DOCDEBT markers"

---

**My counter-proposal:**

**Implement everything they said (CI + Process + Ownership), PLUS:**

1. **Tutorial health dashboard** — Visible metrics on staleness
2. **Tiered review process** — Not everything needs maintainer approval
3. **DOCDEBT markers** — Track quick fixes that need proper fixes later
4. **Explicit review schedule** — In my frontmatter, tracked by CI

**Why this matters:**

You can have perfect automation and perfect process, but if I'm not regularly REVIEWED for quality (not just correctness), I'll slowly become:
- Technically accurate but pedagogically broken
- Comprehensive but impossible to navigate
- Up-to-date but inconsistent in voice

**Quality requires human judgment. Automation can't replace that.**

**Verdict:** All proposals good. Add health tracking + tiered review. Make quality visible and measurable.

---

## Rebuttals

### CI Robot (responding to "you can't catch everything")

Tutorial and Git History are right: I can't catch everything.

But let me reframe this: **I catch 80% of problems for 0.001% of the cost.**

**Cost comparison:**

**Automation (me):**
- Initial setup: 2 hours
- Per-PR: 45 seconds
- Maintenance: 1 hour/quarter
- Catches: Syntax errors, removed flags, broken links (80% of drift)

**Human review (doc maintainer):**
- Initial setup: 4 hours
- Per-PR: 15 minutes (if CLI changed)
- Ongoing: 2 hours/month
- Catches: Everything (100% of drift)

**ROI calculation:**
- 50 PRs/year with CLI changes × 15 min = 12.5 hours/year human review
- Same 50 PRs × 45 sec = 0.625 hours/year automated review

**Savings: 11.875 hours/year**

And here's the key: **I never forget to check.** Humans get busy, skip reviews, forget patterns.

I run on EVERY PR. No exceptions. No "I'll review it later." No "this is a small change, probably fine."

**But I agree with Git History and Jamie:** You still need humans for the 20% I can't catch.

**My revised proposal:**

**Two-layer defense:**

1. **Automation (me) — Fast feedback:**
   - Runs on every PR (no exceptions)
   - Blocks merge if errors found
   - Catches known bad patterns
   - 45 seconds per PR

2. **Human review — Deep feedback:**
   - Triggered when automation fails OR CLI commands changed
   - Doc maintainer reviews for quality
   - Catches concept drift, pedagogy issues
   - 15 minutes per PR (only when needed)

**Together:** Automation blocks obvious errors immediately. Human review ensures quality. Best of both worlds.

---

### Git History (responding to "process creates bottlenecks")

Tutorial worries that requiring doc maintainer review creates a bottleneck.

Valid concern. Let me refine the process:

**NOT every PR needs doc maintainer review.**

**Tiered approach:**

**Tier 1 (No doc impact) — Auto-approve:**
- Internal refactors
- Test changes
- Non-CLI code changes
→ Automation runs, no human review needed

**Tier 2 (Minor doc impact) — Peer review:**
- Typo fixes
- Example updates
- Broken link fixes
→ Any team member can review

**Tier 3 (Moderate impact) — Doc maintainer review:**
- CLI flag changes
- Command behavior changes
- Error message updates
→ Doc maintainer reviews (15 min)

**Tier 4 (Major impact) — Full validation:**
- New commands added
- Command structure changes
- Breaking changes
→ Doc maintainer + run validation checklist (2 hours)

**This removes the bottleneck:**
- 80% of PRs: Tier 1 (no review)
- 15% of PRs: Tier 2 (peer review)
- 4% of PRs: Tier 3 (maintainer, 15 min)
- 1% of PRs: Tier 4 (full validation, 2 hours)

**Average per PR: 80%×0 + 15%×5 + 4%×15 + 1%×120 = 2.6 minutes**

That's not a bottleneck. That's a safety net.

---

### Jamie Park (responding to "this is too much process")

I hear the concern: "This sounds like a lot of process."

Let me show you what "too much process" actually looks like:

**Heavyweight (bad):**
- Every typo fix needs doc maintainer approval
- Doc changes require 3 reviewers
- Updates need sign-off from PM
- Changes need JIRA ticket
→ Result: Nobody fixes docs because it's too painful

**Lightweight (good):**
- CI runs automatically (no human action)
- Doc maintainer auto-tagged (can ignore if trivial)
- Quarterly review scheduled (calendar reminder)
- Style guide available (self-serve)
→ Result: Docs stay current because it's easy

**I'm proposing the lightweight version.**

Here's what actually happens day-to-day:

**Typical PR (no CLI changes):**
1. Developer commits code
2. CI runs (45 seconds) ✓
3. PR merged
→ Total overhead: 45 seconds automated

**PR with CLI change:**
1. Developer commits code + updates doc
2. CI runs (45 seconds) ✓
3. Doc maintainer auto-tagged (can review async)
4. Doc maintainer checks: "Does this break tutorial flow?" (5 min)
5. If yes: suggests edits. If no: approves.
6. PR merged
→ Total overhead: 45 sec + 5 min = 5 min 45 sec

**Quarterly review:**
1. Calendar reminder fires
2. Doc maintainer runs validation checklist (1 hour)
3. Creates tickets for issues found
4. Fixes tickets (2 hours over next week)
→ Total overhead: 3 hours per quarter = 45 min per month amortized

**Annual deep review:**
1. Block 1 day
2. Full restructure if needed
3. User feedback synthesis
→ Total overhead: 1 day per year

**Total annual overhead:**
- Per PR (50/year): 50 × 5.75 min = 4.8 hours
- Quarterly (4/year): 4 × 3 hours = 12 hours
- Annual (1/year): 8 hours
**Total: 24.8 hours/year = ~2 hours/month**

**That's half a day per month to prevent tutorial drift.**

Compare to:
- Fixing drift after it happens: 20+ hours (what we just did in debates)
- User frustration from broken tutorial: Immeasurable

**2 hours/month is NOT too much process. It's insurance.**

---

### The Tutorial (responding to "health metrics are overkill")

CI Robot might think tutorial health metrics are unnecessary overhead.

Let me show you why they matter:

**Scenario: Current state (no metrics)**

User reports: "Tutorial is broken"

Response:
1. "Which part?"
2. "Let me check..."
3. [30 min investigation]
4. "Oh, we changed that command 3 months ago"
5. "Let me fix it..."
6. [2 hours fixing + testing]
7. "Fixed!"

**Total time:** 2.5 hours reactive

**Scenario: With health dashboard**

Weekly check:
```
Tutorial Health Dashboard
- Command accuracy: 96% (2 deprecated patterns found)
- Last validation: 45 days ago (OVERDUE)
- DOCDEBT markers: 3 (2 medium, 1 low)
```

Response:
1. "Validation overdue. Running now..."
2. [Finds issues proactively]
3. [Fixes before users hit them]
4. "Fixed!"

**Total time:** 1 hour proactive (before users complain)

**Savings: 1.5 hours + no user frustration**

**Plus:** Health metrics make documentation quality VISIBLE.

Without metrics:
- "Is the tutorial good?" → "Seems fine?"
- "When did we last review it?" → "Um... not sure?"
- "Are there known issues?" → "Let me check..."

With metrics:
- "Is the tutorial good?" → "96% command accuracy, 2 issues tracked"
- "When did we last review it?" → "45 days ago, next review due in 15 days"
- "Are there known issues?" → "3 DOCDEBT markers (see dashboard)"

**Visibility creates accountability.**

And here's the key: **Metrics are cheap to maintain.**

After initial setup (30 minutes), updates are:
- Automatic (CI checks command accuracy)
- Low-touch (validation run updates Last Validation)
- Self-documenting (DOCDEBT markers tracked by grep)

**Cost: 5 minutes/month to update dashboard**  
**Benefit: Always know tutorial health at a glance**

**That's not overkill. That's good practice.**

---

## Moderator Summary

### Key Arguments

**CI Robot's Position (Automation First):**
- Implement 3 CI checks (pattern linting, command coverage, link validation)
- Catches 80% of drift automatically
- 45 seconds per PR, minimal maintenance
- **Philosophy:** Prevent at commit time with automation

**Git History's Position (Process at Source):**
- CODEOWNERS (doc maintainer reviews CLI changes)
- PR template (doc impact assessment)
- CLI changelog (explicit tracking)
- Quarterly audits
- **Philosophy:** Prevent drift where it starts (CLI changes)

**Jamie's Position (Human Ownership):**
- Assign documentation maintainer (role)
- Create style guide
- Tiered review process
- Quarterly validation schedule
- **Philosophy:** Quality requires human judgment and expertise

**Tutorial's Position (Health Tracking):**
- Tutorial health dashboard
- DOCDEBT markers for quick fixes
- Tiered review (not everything needs maintainer approval)
- Explicit review schedule in frontmatter
- **Philosophy:** Make quality visible and measurable

### Areas of Agreement

**Everyone agrees on:**
1. Automation is necessary (catches obvious errors)
2. Human review is necessary (catches quality issues)
3. Regular validation needed (quarterly minimum)
4. Documentation ownership needed (someone must care)

**All four candidates support:**
- CI pattern linting
- Command coverage checking
- Documentation maintainer role
- Quarterly validation runs

### Tensions

**Automation vs. Human Review:**
- CI Robot: "80% automated coverage is huge ROI"
- Jamie: "Automation can't catch pedagogical issues"
- **Resolution:** Both needed (two-layer defense)

**Process Overhead:**
- Git History: "Tag doc maintainer on all CLI changes"
- Tutorial: "That's a bottleneck"
- **Resolution:** Tiered review (not everything needs maintainer)

**Metrics Overhead:**
- CI Robot: "Just run checks, don't track health"
- Tutorial: "Metrics make quality visible"
- **Resolution:** Minimal metrics (5 min/month to update)

### Evidence Weight

**Supporting automation:**
- Catches 80% of issues (pattern linting, coverage, links)
- 45 seconds per PR (negligible cost)
- Never forgets, always runs
- Git History's analysis: drift happened because no automated checks

**Supporting human ownership:**
- Validation reports show conceptual issues (not just syntax)
- Quality requires judgment (flow, tone, pedagogy)
- Jamie's expertise argument: docs need expert care
- Tutorial's concern: automation can't understand narrative

**Supporting both:**
- 20+ hours spent in debates fixing drift
- Could have been prevented with $2/month of maintainer time
- ROI: Prevention way cheaper than reactive fixing

---

## Decision

**Three-Layer Defense Stack:**

### Layer 1: Automation (Immediate Feedback)

**Implement in CI (.github/workflows/lint-docs.yml):**

1. **Pattern Linting (5 sec):**
   - Check for deprecated command syntax
   - Check for removed flags
   - Check for path inconsistencies
   - Blocks PR if found

2. **Command Coverage (10 sec):**
   - Generate command list from CLI
   - Verify all commands documented
   - Warn if missing (doesn't block)

3. **Link Validation (30 sec):**
   - Check internal markdown links
   - Check relative paths
   - Block if broken links found

**Scripts to create:**
- `scripts/lint-doc-commands.sh` (pattern checking)
- `scripts/check-command-coverage.sh` (inventory diff)

**Total CI time: ~45 seconds per PR**

---

### Layer 2: Human Process (Source Prevention)

**Implement:**

1. **CODEOWNERS:**
```
/cmd/docmgr/cmds/  @doc-maintainer
/pkg/commands/     @doc-maintainer
/pkg/doc/          @doc-maintainer
```

2. **PR Template (add section):**
```markdown
## Documentation Impact

CLI commands changed? [ ] Yes [ ] No

If yes:
- [ ] Tutorial updated (docmgr-how-to-use.md)
- [ ] Examples tested in fresh environment
- [ ] --help text updated
- [ ] Troubleshooting section checked
```

3. **CLI Changelog:**
Create `docs/cli-changelog.md` tracking user-facing changes

4. **Tiered Review:**
- Tier 1 (no impact): Auto-approve
- Tier 2 (minor): Peer review
- Tier 3 (moderate): Doc maintainer (15 min)
- Tier 4 (major): Full validation (2 hours)

---

### Layer 3: Ownership & Quality (Deep Review)

**Assign Documentation Maintainer:**

**Responsibilities:**
- Review PRs with CLI changes (Tier 3/4)
- Quarterly validation runs
- Maintain style guide
- Update tutorial when commands change
- Track and fix DOCDEBT

**Time:** ~2 hours/month average

**Create deliverables:**

1. **Style Guide (`docs/style-guide.md`):**
   - Voice, tone, terminology
   - Command example format
   - Structure guidelines

2. **Validation Schedule:**
   - Monthly micro-review (30 min)
   - Quarterly full validation (2 hours)
   - Annual deep review (1 day)

3. **Tutorial Health Dashboard (`docs/tutorial-health.md`):**
   - Last validation date
   - Command accuracy percentage
   - Known issues (DOCDEBT count)
   - Next review due date

---

### Implementation Timeline

**Week 1: Automation (CI Robot's domain)**
- Create `scripts/lint-doc-commands.sh`
- Create `scripts/check-command-coverage.sh`
- Add `.github/workflows/lint-docs.yml`
- Test on current tutorial (should pass after Phase 1 fixes)

**Week 2: Process (Git History's domain)**
- Add CODEOWNERS file
- Update PR template
- Create `docs/cli-changelog.md` (document recent changes)
- Communicate new process to team

**Week 3: Ownership (Jamie's domain)**
- Assign documentation maintainer (or recruit)
- Create style guide draft
- Set up quarterly validation calendar
- Document tiered review process

**Week 4: Metrics (Tutorial's request)**
- Create tutorial health dashboard
- Add frontmatter tracking to tutorial
- Document DOCDEBT marker process
- Initial health check

---

### Success Metrics

**Automation layer:**
- ✅ CI runs on every PR (100%)
- ✅ <1% false positives (doesn't block valid changes)
- ✅ Catches 80%+ of command syntax drift

**Process layer:**
- ✅ 90%+ of CLI-changing PRs have doc impact assessment
- ✅ CLI changelog kept current (<1 week lag)
- ✅ Doc maintainer reviews 100% of Tier 3/4 PRs

**Ownership layer:**
- ✅ Quarterly validation runs on schedule
- ✅ <5 open DOCDEBT markers at any time
- ✅ Tutorial health dashboard updated monthly

**Overall success:**
- ✅ Zero command syntax errors in tutorial
- ✅ Tutorial validation completion time <15 minutes (not 30)
- ✅ No validator complaints about outdated commands

---

### Cost Summary

**Initial Setup (one-time):**
- Automation scripts: 3 hours
- Process setup: 2 hours
- Style guide: 3 hours
- Health dashboard: 1 hour
**Total: 9 hours**

**Ongoing (monthly average):**
- CI runs: 0 hours (automated, 45 sec per PR)
- Doc maintainer reviews: 1 hour (3-4 Tier 3 PRs/month @ 15 min each)
- Micro-reviews: 0.5 hours
- Quarterly validation (amortized): 0.5 hours
**Total: 2 hours/month**

**ROI:**
- Prevention cost: 2 hours/month
- Reactive fixing cost: 20+ hours (what we just did)
- **Savings: 10:1 ROI after first drift incident prevented**

---

**Decision: Implement all three layers. Start with automation (Week 1), add process (Week 2), establish ownership (Weeks 3-4).**

This prevents 100% of preventable drift: Automation catches syntax, process prevents drift at source, ownership ensures quality.
