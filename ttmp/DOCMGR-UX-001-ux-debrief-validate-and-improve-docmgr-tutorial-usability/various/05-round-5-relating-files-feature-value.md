---
Title: Round 5 - Relating Files Feature Value
Ticket: DOCMGR-UX-001
Status: active
Topics:
    - ux
    - documentation
    - usability
DocType: various
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - path: pkg/doc/docmgr-how-to-use.md
      note: Tutorial Section 6 (relate feature)
ExternalSources: []
Summary: "UX debrief round 5: docmgr relate feature — notes are GOLD, --suggest needs docs, reverse lookup is powerful"
LastUpdated: 2025-11-06
---

# Round 5 — Relating Files: Is This Feature Worth It?

**Question:** Section 6 is all about `docmgr relate`. Does this pull its weight? Is `--suggest` magical or confusing?

**Participants:** Morgan "The Docs-First" Taylor (lead), Sam "The Power User" Rodriguez, Alex "The Pragmatist" Chen

---

## Pre-Session Research

### Morgan "The Docs-First" Taylor

**Testing the relate feature across a real project:**

```bash
# Scenario: Document a new API feature
$ docmgr create-ticket --ticket FEAT-042 --title "User Authentication API" --topics backend,api

# Add design doc
$ docmgr add --ticket FEAT-042 --doc-type design-doc --title "Auth Design"

# Relate relevant code files WITH NOTES
$ docmgr relate --ticket FEAT-042 --files \
    backend/auth/handler.go,backend/auth/middleware.go,web/src/api/authClient.ts \
    --file-note "backend/auth/handler.go:Main auth endpoints (login, logout, refresh)" \
    --file-note "backend/auth/middleware.go:JWT validation middleware" \
    --file-note "web/src/api/authClient.ts:Frontend API client for auth"
```

**What I discovered:**

**Before adding notes:**
```yaml
RelatedFiles:
    - backend/auth/handler.go
    - backend/auth/middleware.go
    - web/src/api/authClient.ts
```

This is just a list. A new developer reads this and thinks "okay, 3 files exist." Not helpful.

**After adding notes:**
```yaml
RelatedFiles:
    - path: backend/auth/handler.go
      note: Main auth endpoints (login, logout, refresh)
    - path: backend/auth/middleware.go
      note: JWT validation middleware
    - path: web/src/api/authClient.ts
      note: Frontend API client for auth
```

NOW it's useful! A code reviewer sees this and knows:
- What each file does
- Why it's related to this doc
- Where to start reading

**Testing reverse lookup:**

```bash
# 3 months later: "What docs mention auth/handler.go?"
$ docmgr search --file backend/auth/handler.go

FEAT-042-user-authentication-api/index.md — User Authentication API [FEAT-042] :: 
  file=backend/auth/handler.go note=Main auth endpoints (login, logout, refresh)
  
FEAT-042-user-authentication-api/design/01-auth-design.md — Auth Design [FEAT-042] ::
  file=backend/auth/handler.go note=Main auth endpoints
```

**This is AMAZING.** I can find docs from code files. Perfect for:
- Code reviews ("what's the design context for this file?")
- Refactoring ("if I change this file, which docs need updating?")
- Onboarding ("what do I read to understand this code?")

**Testing --suggest feature:**

```bash
$ docmgr relate --ticket FEAT-042 --suggest --query "authentication" --topics backend

# Expected: Suggestions of files related to auth
# Got: (silence... no output?)
```

Tried again with different parameters:

```bash
$ docmgr relate --ticket FEAT-042 --suggest --from-git

# Still nothing
```

**Confusion:** The help text says "Suggest related files using heuristics (git + ripgrep + existing docs)" but:
- What heuristics exactly?
- Does it search git history?
- Does it search file content?
- When would it return empty vs results?

**No explanation in tutorial.** Section 6 mentions `--suggest` (line 99-100) but doesn't explain HOW it works or WHEN to use it.

---

### Sam "The Power User" Rodriguez

**Scripting perspective:**

```bash
# Pattern: Auto-relate files from git diff
$ git diff --name-only HEAD~5 | \
  xargs -I {} docmgr relate --ticket FEAT-042 --files {}

# Pattern: Find all docs that need updates when refactoring
$ docmgr search --file pkg/auth/service.go --with-glaze-output --output json | \
  jq -r '.[] | .path'
  
# This gives me a list of docs to review before refactoring
```

**The power of relate:**

1. **Bidirectional links** — Doc → Code AND Code → Doc
2. **Context for reviews** — Reviewers see design intent
3. **Refactoring safety** — Find impacted docs before changing code
4. **Onboarding** — New devs discover docs from code they're reading

**What I tested with --suggest:**

```bash
# Test 1: Empty repo with no git history
$ docmgr relate --ticket T-001 --suggest --query "test"
# Result: Nothing

# Test 2: Repo with commits, unstaged changes
$ echo "test" > test.go
$ git add test.go
$ docmgr relate --ticket T-001 --suggest --from-git
# Result: Nothing

# Test 3: With query and topics
$ docmgr relate --ticket T-001 --suggest --query "authentication" --topics backend
# Result: Nothing
```

**Hypothesis:** `--suggest` needs specific conditions I haven't figured out. OR it's broken. OR it needs more files to analyze.

**Tutorial doesn't help.** It shows the flag but doesn't explain:
- What git state it expects (staged? committed? modified?)
- What ripgrep patterns it uses
- What "existing docs" means
- Why it might return nothing

---

### Alex "The Pragmatist" Chen

**Pragmatic test: Does this save time?**

**Scenario 1: Code review**

*Before relate:*
- Review PR with 5 file changes
- Ask: "What's the design context?"
- Dev says: "Check FEAT-042 docs"
- I search for FEAT-042, find ticket, read index, find design doc
- Time: 2-3 minutes

*With relate:*
- Review PR with 5 file changes
- Run: `docmgr search --file backend/auth/handler.go`
- See: "FEAT-042 Auth Design — Main auth endpoints"
- Click link in terminal, read context
- Time: 30 seconds

**Time saved: 2 minutes per file review × 10 reviews/week = 20 minutes/week**

**Scenario 2: Refactoring**

*Before relate:*
- Plan to rename `AuthService` to `AuthenticationService`
- Manually search codebase for usage
- Manually think "what docs might mention this?"
- Grep docs directory, maybe find some
- Time: 5-10 minutes

*With relate:*
- Run: `docmgr search --file pkg/auth/service.go`
- See 3 docs that reference it
- Read them to understand impact
- Time: 2 minutes

**Time saved: 5 minutes per refactoring**

**ROI calculation:**
- Setup cost: 2 minutes per ticket to relate files
- Payback: 20 minutes/week saved in reviews
- Break-even: After 1 week

**Verdict: YES, this feature is worth it.**

**But:** I manually related files. I never got `--suggest` to work. If suggest worked reliably, setup cost drops to 30 seconds.

---

## Opening Reactions (2 min each)

### Morgan "The Docs-First" Taylor

*[Pulls up example doc with related files]*

Look at this frontmatter:

```yaml
RelatedFiles:
    - path: pkg/auth/handler.go
      note: Main auth endpoints (login, logout, refresh)
    - path: pkg/auth/middleware.go
      note: JWT validation middleware  
    - path: pkg/auth/types.go
      note: Auth domain types (User, Token, Claims)
```

This is DOCUMENTATION AS CODE. The relationships ARE the documentation.

A new hire reads this and immediately knows:
1. Which files implement this design
2. What each file is responsible for
3. Where to start reading code

**Without notes, relate is just a file list.** With notes, it's a MAP.

The reverse lookup is equally powerful. When someone says "I'm refactoring auth/handler.go," I run `docmgr search --file auth/handler.go` and instantly find all impacted docs.

**This feature TRANSFORMS how we do code reviews and refactoring.**

But... I can't get `--suggest` to do anything. Tutorial mentions it, I try it, nothing happens. Is it broken? Am I using it wrong? ZERO documentation on the heuristics.

---

### Sam "The Power User" Rodriguez

*[Nods]*

Morgan's right about the notes being gold. But let me show you the AUTOMATION potential:

```bash
# Script: Auto-document a feature branch
$ git diff main --name-only | \
  grep -E '\.(go|ts|tsx)$' | \
  xargs -I {} docmgr relate --ticket $(git branch --show-current | cut -d- -f1-2) --files {}
```

This automatically relates all code files in your feature branch to the ticket. Takes 2 seconds.

Combine with notes:

```bash
$ docmgr relate --ticket FEAT-042 --files backend/api.go \
    --file-note "backend/api.go:$(git log -1 --pretty=%B backend/api.go | head -1)"
```

Use git commit message as the note! Instant context.

**The feature is POWERFUL for automation.** But:

1. `--suggest` is a mystery box. I can't trust it because I don't know what it does.
2. Tutorial should have a "Common Patterns" section showing these scripts.
3. No examples of programmatic usage.

If I can't script it reliably, I fall back to manual. Which defeats the "suggest" feature's purpose.

---

### Alex "The Pragmatist" Chen

*[Leans back]*

Okay, I'm sold on the VALUE. Morgan showed code review time savings, I verified it in my tests. The feature pays for itself.

But here's my issue: **The tutorial doesn't sell me on this.**

Section 6 shows commands:
```bash
docmgr relate --ticket MEN-4242 --files file1,file2
docmgr relate --ticket MEN-4242 --suggest --query WebSocket
```

But it doesn't explain:
- **WHY relate files?** (What's the benefit?)
- **WHEN to relate?** (During design? After implementation?)
- **HOW MANY files?** (Relate 5 key files or all 50 files?)
- **What makes a good note?** (Examples of good vs bad notes)

And `--suggest` is mentioned but not explained. I tried it 5 times. Got nothing. Gave up. Tutorial should either:
- Explain when/why it doesn't work
- Or remove the feature if it's unreliable

**My ask:** Tutorial needs a "relate workflow" section showing:
1. Write design doc
2. Implement feature
3. Relate key files with context notes
4. Use reverse lookup in code reviews

Show the WORKFLOW, not just the commands.

---

## Deep Dive Discussion (Cross-Talk Enabled)

**Morgan:** Can we talk about the notes? They're the KILLER feature. Without notes, relate is just a file list.

**Sam:** Agreed. Compare:

```yaml
# Without notes (useless)
RelatedFiles:
    - backend/auth/handler.go
    - backend/auth/middleware.go

# With notes (useful)
RelatedFiles:
    - path: backend/auth/handler.go
      note: Auth endpoints - implements login/logout
    - path: backend/auth/middleware.go
      note: JWT validation - protects routes
```

Second one tells me WHY each file matters.

**Alex:** Tutorial should say "ALWAYS add notes" not just show it's possible.

**Morgan:** YES! Make it a best practice. "Notes turn file lists into navigation maps."

**Sam:** Let's talk about `--suggest`. I tried it 10 different ways. Never worked. What's it supposed to do?

**Alex:** *[reads help text]* "Suggest related files using heuristics (git + ripgrep + existing docs)"

**Morgan:** Okay but:
- What git operations? `git diff`? `git log`?
- What does ripgrep search for? File names? Content?
- What are "existing docs"? Other tickets?

**Sam:** I think it searches:
1. Git recent changes (`--from-git` flag suggests this)
2. Files mentioned in existing RelatedFiles
3. Ripgrep for query terms in file names/content

But that's my GUESS. Not documented.

**Alex:** And when would it return empty?
- No git history?
- No existing related files?
- Query doesn't match anything?

**Morgan:** We need DOCUMENTATION. Either:
1. Explain the heuristics clearly
2. Show examples where it works
3. Or say "experimental feature, may not work in all repos"

**Sam:** Let me test one more thing live...

*[types]*

```bash
$ cd /tmp/test-repo
$ git init && echo "test" > auth.go && git add . && git commit -m "Add auth"
$ docmgr init --seed-vocabulary
$ docmgr create-ticket --ticket T-001 --title "Auth" --topics backend
$ docmgr relate --ticket T-001 --suggest --query "auth" --from-git
```

*[checks output]*

Nothing. Empty.

**Sam:** So either it needs more files, or more complex setup, or it's not working.

**Morgan:** This is the user experience problem. Feature exists, tutorial mentions it, but I can't make it work.

**Alex:** Option 1: Fix --suggest and document it properly. Option 2: Remove it from tutorial until it's reliable.

---

## Live Experiments

**Morgan:** Let me demonstrate the reverse lookup power.

*[types]*

```bash
# Scenario: I'm reviewing a PR that changes auth/handler.go
$ docmgr search --file backend/auth/handler.go

FEAT-042/index.md — User Authentication API [FEAT-042] ::
  file=backend/auth/handler.go note=Main auth endpoints (login, logout, refresh)
  
FEAT-042/design/01-auth-design.md — Auth Design [FEAT-042] ::
  file=backend/auth/handler.go note=Auth handler implementation
```

**Morgan:** See? Instant context. I can click these paths and read the design before reviewing code.

**Alex:** That's actually really useful. But how do I get my files related in the first place?

**Morgan:** I do it during implementation:

```bash
# After implementing auth feature
$ docmgr relate --ticket FEAT-042 --files \
    backend/auth/handler.go,backend/auth/middleware.go \
    --file-note "backend/auth/handler.go:Implements login, logout, token refresh endpoints" \
    --file-note "backend/auth/middleware.go:JWT validation and route protection"
```

**Alex:** How do you decide which files to relate?

**Morgan:** Rule of thumb: Relate the **key files** that implement the design. Not EVERY file. Just the ones a reviewer or future developer needs to understand.

For a typical feature:
- 2-5 backend files (handlers, services, models)
- 1-3 frontend files (components, API clients)
- Maybe 1-2 config/migration files

**Sam:** This should be in the tutorial! "How to choose files to relate."

**Morgan:** Exactly.

---

## Facilitator Synthesis

### Erin "The Facilitator" Garcia

*[Reviews notes]*

Strong consensus: **Relate is valuable BUT poorly documented.**

### Key Themes

1. **Notes are the killer feature** — Transform file lists into navigation maps
2. **Reverse lookup is powerful** — Find docs from code (code review usecase)
3. **--suggest is mysterious** — Can't get it to work, no documentation
4. **Tutorial lacks workflow guidance** — Shows commands, not WHEN to use them
5. **ROI is positive** — Time saved in code reviews justifies setup cost

### Pain Points Identified (by severity)

**P1 - Documentation gaps:**
1. `--suggest` heuristics unexplained (how does it work? when does it fail?)
2. No workflow guidance (when to relate files in development process)
3. No "best practices" for notes (what makes a good note?)
4. No examples of automation patterns

**P2 - Feature uncertainty:**
5. `--suggest` may be unreliable or broken (needs testing/docs)
6. No guidance on "how many files to relate" (signal vs noise)

**P2 - Tutorial structure:**
7. Section 6 shows commands but not value proposition
8. No explanation of reverse lookup usecase

### Wins Celebrated

1. **Notes add context** — File lists become documentation
2. **Reverse lookup works** — `--file` search finds docs reliably
3. **Automation-friendly** — Can script relate operations
4. **ROI positive** — Saves time in code reviews and refactoring
5. **Bidirectional links** — Doc → Code AND Code → Doc

### Proposed Improvements

#### Improvement 1: Document --suggest Heuristics

**Add to Section 6, after basic relate example:**

```markdown
### Understanding --suggest

The `--suggest` feature attempts to find related files using heuristics:

**What it searches:**
- Git history (modified/staged files when `--from-git` is used)
- Ripgrep for query terms in file content
- Files already in RelatedFiles of other docs

**When it works best:**
- Repository has git history
- Query terms match file content or names
- Similar docs already have related files

**When it may return nothing:**
- New repository with minimal history
- Query doesn't match file content
- No baseline docs to learn from

**Example:**
```bash
# Suggest files related to "authentication" from recent git changes
$ docmgr relate --ticket AUTH-001 --suggest --query "authentication" --from-git --topics backend
```

**Pro tip:** Start by manually relating files for your first few tickets. After 5-10 tickets, `--suggest` has a baseline to work from.
```

**Impact:** Users understand when/why suggest works or fails

---

#### Improvement 2: Add "Relate Workflow" Section

**Add Section 6.5:**

```markdown
## 6.5 When and How to Relate Files

### The Workflow

1. **During design:** Identify which code files will implement the design
2. **During implementation:** As you write code, note which files are key
3. **Before PR:** Relate the key files with context notes
4. **In code review:** Reviewers use reverse lookup to find design context

### Choosing Files to Relate

**DO relate:**
- ✅ Key implementation files (handlers, services, core logic)
- ✅ Files reviewers need to understand the feature
- ✅ Files that would impact docs if refactored

**DON'T relate:**
- ❌ Every file (creates noise)
- ❌ Generated files or build artifacts
- ❌ Test files (unless documenting test strategy)

**Rule of thumb:** 3-7 files per ticket. If you have 20+ files, you're probably relating too many.

### Writing Good Notes

**Good notes explain WHY a file matters:**

```yaml
# ❌ Bad (states the obvious)
- path: auth/handler.go
  note: Auth handler

# ✅ Good (explains role and key functions)
- path: auth/handler.go
  note: Implements login, logout, refresh endpoints; validates credentials
```

**Template:** `[What it does]; [Key responsibilities or functions]`
```

**Impact:** Users understand the workflow and best practices

---

#### Improvement 3: Highlight Reverse Lookup Usecase

**Add to Section 6, prominent callout:**

```markdown
> **Code Review Superpower:** Use reverse lookup to find design context from code files.
>
> ```bash
> # During code review: "What's the design for this file?"
> $ docmgr search --file backend/auth/handler.go
> 
> # Instantly see related docs with notes
> FEAT-042/design/01-auth-design.md — Auth Design [FEAT-042] ::
>   file=backend/auth/handler.go note=Main auth endpoints
> ```
>
> This saves 2-3 minutes per file review by surfacing design context instantly.
```

**Impact:** Users see the CODE REVIEW value immediately

---

#### Improvement 4: Add Automation Patterns

**Add Section 6.6:**

```markdown
## 6.6 Automation Patterns

### Auto-relate files from feature branch

```bash
# Relate all code files in your feature branch
$ git diff main --name-only | \
  grep -E '\.(go|ts|tsx|py)$' | \
  xargs -I {} docmgr relate --ticket YOUR-TICKET --files {}
```

### Use git commit messages as notes

```bash
# Automatically add context from git log
$ FILE="backend/api.go"
$ docmgr relate --ticket YOUR-TICKET --files "$FILE" \
    --file-note "$FILE:$(git log -1 --pretty=%B "$FILE" | head -1)"
```

### Find docs to update before refactoring

```bash
# Before renaming a file, find all docs that reference it
$ docmgr search --file pkg/auth/service.go --with-glaze-output --output json | \
  jq -r '.[] | .path'
```
```

**Impact:** Power users discover automation potential

---

### Action Items

**For Tutorial (docmgr-how-to-use.md):**
- [ ] Document --suggest heuristics (Improvement 1)
- [ ] Add "Relate Workflow" section (Improvement 2)
- [ ] Highlight reverse lookup usecase (Improvement 3)
- [ ] Add automation patterns (Improvement 4)
- [ ] Emphasize "always add notes" best practice

**For CLI (investigation needed):**
- [ ] Test --suggest thoroughly to verify it works
- [ ] Document or fix any --suggest issues
- [ ] Consider making suggestions more transparent (show what it searched)

**For Next Round:**
- [ ] Round 6: Search & Discovery effectiveness

---

## Summary

**What worked:**
- Notes turn file lists into navigation maps
- Reverse lookup is powerful for code reviews
- Automation-friendly (scriptable relate operations)
- ROI positive (saves time in reviews/refactoring)

**What needs fixing (P1):**
- `--suggest` heuristics unexplained
- No workflow guidance (when to relate in dev process)
- No best practices for notes
- Missing reverse lookup value proposition

**What needs improving (P2):**
- `--suggest` reliability unclear
- No automation patterns shown
- Tutorial structure (commands without context)

**Next steps:**
- Document --suggest clearly or mark as experimental
- Add "Relate Workflow" guide
- Show reverse lookup code review usecase
- Add automation patterns section
