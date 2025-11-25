---
Title: Debate Round 8 — Error Messages & Troubleshooting
Ticket: DOCMGR-DOC-VERBS
Status: active
Topics:
    - docmgr
    - documentation
    - tutorial
    - ux
DocType: reference
Intent: short-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/pkg/commands/relate.go
      Note: Source of "no changes specified" error (line 471)
    - Path: docmgr/pkg/commands/doctor.go
      Note: Validation error messages
    - Path: docmgr/pkg/commands/tasks.go
      Note: Task command error handling
ExternalSources: []
Summary: "Round 8 debate: How to write error messages that help beginners recover, with codebase analysis."
LastUpdated: 2025-11-25
---

# Debate Round 8 — Error Messages & Troubleshooting

## Question

**"How do we write error messages and troubleshooting guidance that actually helps beginners recover?"**

**Primary Candidates:**
- Jamie Park (Technical Writer)
- Dr. Maya Chen (Accuracy Crusader)
- Sam Torres (Empathy Advocate)
- The Reset Script (Saboteur)

---

## Pre-Debate Research

### Current Error Messages in Codebase

**1. "no changes specified" Error (relate.go:471)**

```go
// When not in suggestion-listing mode, ensure at least one change was requested
if !settings.Suggest && addedCount == 0 && removedCount == 0 && updatedCount == 0 {
    return fmt.Errorf("no changes specified. Use --file-note 'path:note' to add/update, --remove-files to remove, or --suggest --apply-suggestions to apply suggestions")
}
```

**Problem identified by validators:**
- **What users see:** "Error: no changes specified"
- **What users think:** "I made a mistake in my command"
- **What actually happened:** Files are already related with those exact notes
- **Better message:** "No changes needed. The specified files are already related with these notes. Use --remove-files to unlink them or provide different notes to update."

**Frequency:** 100% of validators hit this (caused by reset script)

---

**2. Removed Flag Error (relate.go:396)**

```go
// Enforce deprecation: --files is no longer supported for additions
if len(settings.Files) > 0 {
    return fmt.Errorf("--files has been removed from 'docmgr relate'. Use repeated --file-note 'path:note' instead. Example: docmgr relate --file-note 'a/b.go:reason' --file-note 'c/d.ts:reason'")
}
```

**This is GOOD:**
- ✅ Says what's wrong ("--files has been removed")
- ✅ Says what to do instead ("Use repeated --file-note")
- ✅ Shows concrete example
- ✅ Migration-friendly

---

**3. Missing Required Parameter Errors**

From relate.go:177:
```go
if settings.Ticket == "" {
    return fmt.Errorf("must specify either --doc or --ticket")
}
```

**Problem:**
- **What users see:** "Error: must specify either --doc or --ticket"
- **What users don't know:** WHICH ONE to use for their usecase
- **Better message:** "Must specify either --doc (for a specific document) or --ticket (for the ticket index). Example: --ticket MEN-4242 or --doc ttmp/path/to/doc.md"

---

**4. Doctor Warnings (from validation reports)**

Validators saw: `unknown_topics: [test]`

**Current:** Warning with unknown value listed  
**Problem:** No guidance on how to fix  
**Better:** "Unknown topics: [test]. Add to vocabulary with: docmgr vocab add --category topics --slug test --description '...' OR use a known topic from: docmgr vocab list --category topics"

---

### Error Message Patterns Found in Code

**Pattern 1: Wrapped errors (technical)**
```go
return fmt.Errorf("failed to read document frontmatter: %w", err)
return fmt.Errorf("failed to find ticket directory: %w", err)
```
**Good for:** Debugging by developers  
**Bad for:** Beginners don't understand wrapped stack traces

**Pattern 2: Simple validation failures**
```go
return fmt.Errorf("--file-note requires a non-empty note for %s (use 'path:reason')", p)
```
**Good:** Shows exactly what's wrong and the expected format  
**Could improve:** Add example with actual values

**Pattern 3: Missing guidance on resolution**
```go
if settings.Ticket == "" {
    return fmt.Errorf("must specify either --doc or --ticket")
}
```
**Problem:** States requirement but doesn't guide WHICH to use or HOW

---

### Error Types by Cause

**From Code Analysis:**

**User Input Errors (fixable by user):**
- Missing required flags → Guide which flag + example
- Wrong flag format → Show correct format + example
- Invalid values → Show valid options + how to list them
- Deprecated flags → Migration guide

**State Errors (system state):**
- File not found → Suggest checking path or running from repo root
- Already exists → Explain it's not an error, show how to update
- No changes needed → Distinguish from "you did it wrong"

**Permission/System Errors (environmental):**
- Permission denied → Suggest checking file permissions
- Directory doesn't exist → Suggest running init first

---

### Proposed Error Message Framework

Based on Microsoft/Google guidelines and validator feedback:

```
Error: <WHAT HAPPENED>

<WHY THIS HAPPENED>

<WHAT TO DO NOW>
  - Option 1: <command>
  - Option 2: <alternative>

Example: <concrete example with real values>
```

**Example (improving "no changes specified"):**

```
Error: No changes to apply

The files you specified are already related with these notes:
  - backend/api/register.go (note: "Registers API routes")
  - web/src/store/api/chatApi.ts (note: "Frontend integration")

To make changes:
  - Update notes: docmgr doc relate --ticket MEN-4242 --file-note "path:NEW NOTE"
  - Remove files: docmgr doc relate --ticket MEN-4242 --remove-files "path1,path2"
  - List current: docmgr doc search --file backend/api/register.go

Tip: This is not an error—your files are already correctly linked!
```

---

### Commands That Need Better Error Messages

From codebase analysis, these commands have weak error handling:

1. **`docmgr doc relate`** — "no changes specified" (HIGH PRIORITY)
2. **`docmgr doctor`** — Unknown vocabulary warnings without resolution
3. **`docmgr doc add`** — Missing doc-type or title
4. **`docmgr ticket create-ticket`** — Missing required fields
5. **`docmgr meta update`** — Invalid field names
6. **`docmgr task check`** — Invalid task ID

---

## Opening Statements

### Jamie Park (Technical Writer)

*[Opens Microsoft Writing Style Guide and Google Developer Documentation Style Guide]*

Let me show you how the industry's best handle error messages.

**Microsoft's Error Message Framework:**
1. What happened (the error)
2. Why it happened (the cause)
3. What to do (the solution)

**Google's Error Message Principles:**
- Be specific (not "invalid input" but "email format is invalid")
- Be actionable (provide next steps, not just problems)
- Be empathetic (acknowledge user frustration)
- Provide examples (show correct usage)

Now let's look at our error messages through this lens.

**GOOD EXAMPLE from our code (--files deprecation):**

```
--files has been removed from 'docmgr relate'. 
Use repeated --file-note 'path:note' instead. 
Example: docmgr relate --file-note 'a/b.go:reason' --file-note 'c/d.ts:reason'
```

✅ What happened: Flag removed  
✅ Why: Implicit (deprecation)  
✅ What to do: Use --file-note  
✅ Example: Concrete command  

**BAD EXAMPLE from our code ("no changes specified"):**

```
no changes specified. Use --file-note 'path:note' to add/update, --remove-files to remove, or --suggest --apply-suggestions to apply suggestions
```

❌ What happened: Unclear ("no changes" sounds like user error)  
❌ Why: Not explained (are files already related? did I forget something?)  
❌ What to do: Lists options but doesn't say WHICH applies  
❌ Empathy: Makes user feel they made a mistake  

**My proposal for fixing error messages:**

**Phase 1 (CLI improvements — separate ticket):**
1. Detect WHY "no changes" occurred:
   - All files already related with same notes → "No changes needed (files already linked)"
   - No files specified → "No files specified (use --file-note)"
   - Suggestion mode with no files found → "No matching files found (try different query)"

2. Add context to errors:
```go
if addedCount == 0 && removedCount == 0 && updatedCount == 0 {
    // Check if files are already present
    allPresent := true
    for path := range noteMap {
        if _, ok := current[path]; !ok {
            allPresent = false
            break
        }
    }
    
    if allPresent && len(noteMap) > 0 {
        return fmt.Errorf("no changes needed. The specified files are already related with these notes. To update: provide different notes. To remove: use --remove-files")
    }
    
    return fmt.Errorf("no files specified. Use --file-note 'path:note' to relate files. Example: --file-note 'src/main.go:Core implementation'")
}
```

**Phase 2 (Tutorial troubleshooting section):**
- Add "Common Errors" appendix
- For each error: what it means, why you see it, how to fix
- Link from tutorial steps: "If you see X error, see Troubleshooting Section Y"

**Format for troubleshooting section:**

```markdown
## Troubleshooting Common Errors

### "Error: no changes specified"

**What this means:** docmgr couldn't find any changes to make.

**Common causes:**
1. Files are already related with these notes → Not an error! Your links are correct.
2. Forgot to specify files → Add `--file-note "path:note"`
3. Running same command twice → Check current links with `docmgr doc search --file path`

**How to fix:**
- List current relationships: `docmgr doc search --file backend/api/register.go`
- Update existing notes: `docmgr doc relate --ticket T --file-note "path:NEW NOTE"`
- Remove links: `docmgr doc relate --ticket T --remove-files "path"`

---

### "Unknown topic: [test]"

**What this means:** The topic "test" isn't in your vocabulary.yaml.

**Common causes:**
1. Typo in topic name → Check `docmgr vocab list --category topics`
2. New topic not added yet → Add it with `docmgr vocab add`

**How to fix:**
- View valid topics: `docmgr vocab list --category topics`
- Add new topic: `docmgr vocab add --category topics --slug test --description "Testing and QA"`
- OR fix typo in frontmatter: Change "test" to a valid topic

**Note:** Unknown topics are warnings, not errors. Your docs still work!
```

**Verdict:** Fix CLI error messages (new ticket), add troubleshooting section to tutorial (this ticket).

---

### Dr. Maya Chen (Accuracy Crusader)

*[Runs grep through error messages]*

Let me categorize our errors by SEVERITY:

**CRITICAL (blocks user, confusing message):**
1. "no changes specified" when files ARE specified (and already related) — **FIXED IN CLI**
2. "must specify --doc or --ticket" without explaining WHICH to use — **ADD GUIDANCE**

**HIGH (blocks user, clear message but missing guidance):**
3. Doctor warnings without resolution steps — **ADD TO TUTORIAL**
4. Missing required parameter errors without examples — **ADD EXAMPLES**

**MEDIUM (user can figure it out with effort):**
5. Wrapped errors with technical stack traces — **ACCEPTABLE FOR NOW**
6. File not found errors — **ADD TROUBLESHOOTING HINTS**

**LOW (already good enough):**
7. --files deprecation message — **ALREADY EXCELLENT**
8. Validation errors with format guidance — **ALREADY GOOD**

**My priority ranking:**

1. **Fix "no changes specified" in CLI** (separate ticket, HIGH priority)
   - This confused 100% of validators
   - Root cause: Poor message when no diff detected
   - Solution: Distinguish "files already related" from "no files specified"

2. **Add troubleshooting section to tutorial** (this ticket, MEDIUM priority)
   - Document top 5 errors users hit
   - For each: what/why/how structure
   - Link from relevant tutorial sections

3. **Improve doctor warning messages** (CLI ticket, MEDIUM priority)
   - Add resolution steps to warning output
   - "Unknown topic: [test]. Fix: docmgr vocab add --category topics --slug test"

**I oppose Jamie's proposal to fix ALL error messages.** Why?

- Scope creep: We have 25 documented issues to fix in the tutorial. Don't add "rewrite all CLI errors."
- Diminishing returns: Most errors are fine. Focus on the 2-3 that hurt users.
- Maintenance burden: Custom error logic adds complexity.

**My proposal:**

**This ticket (DOCMGR-DOC-VERBS):**
- Add troubleshooting section to tutorial
- Document 5 common errors (from validation reports)
- Each error gets: description, causes, fixes, examples

**Separate CLI ticket:**
- Fix "no changes specified" message only
- Add context detection (files already related vs. no files specified)
- 30 minutes of work, high impact

**Verdict:** Tutorial troubleshooting section (this ticket). CLI fixes (new ticket, scoped to top 2-3 errors).

---

### Sam Torres (Empathy Advocate)

*[Shows user journey with error messages highlighted]*

Let me tell you what error messages feel like to beginners.

**Scenario 1: The "no changes specified" error**

Beginner runs:
```bash
docmgr doc relate --ticket MEN-3083 \
  --file-note "backend/api/register.go:Registers API routes"
```

Sees: `Error: no changes specified.`

**Emotional journey:**
1. Confusion: "But I specified a file..."
2. Self-doubt: "Did I type it wrong?"
3. Re-checking: Reads command 3 times, verifies spelling
4. Frustration: "This doesn't make sense"
5. Debugging: Tries variations, googles error
6. **10 minutes wasted**

**What they NEEDED to see:**
```
✓ Files already related with these notes:
  - backend/api/register.go: "Registers API routes"

No changes needed! Your links are correct.

To update notes: --file-note "path:NEW NOTE"
To remove: --remove-files "path"
```

Notice the difference?
- **Tone:** "✓ Files already related" vs "Error: no changes specified"
- **Context:** Shows WHICH files and WHAT notes
- **Action:** Clear next steps IF they want to change something
- **Reassurance:** "No changes needed! Your links are correct."

---

**Scenario 2: Doctor warnings**

Beginner runs `docmgr doctor --ticket MEN-3083`

Sees: `unknown_topics: [test]`

**Emotional journey:**
1. Confusion: "What's wrong with 'test'?"
2. Uncertainty: "Do I need to fix this?"
3. Searching: Opens vocabulary.yaml, reads docs
4. Guessing: "Maybe I should change it to 'testing'?"
5. **5 minutes wasted + lingering uncertainty**

**What they NEEDED to see:**
```
⚠ Unknown topic: "test"
   Your topic "test" isn't in vocabulary.yaml. This is a WARNING, not an error.
   
   Fix options:
   1. Add to vocabulary:     docmgr vocab add --category topics --slug test --description "..."
   2. Use existing topic:    docmgr vocab list --category topics
   3. Ignore this warning:   Unknown topics are allowed and don't affect functionality
```

Notice:
- **Severity:** "⚠ WARNING" not "error"
- **Context:** "isn't in vocabulary.yaml"
- **Impact:** "This is a WARNING, not an error"
- **Options:** 3 clear paths forward
- **Reassurance:** "Unknown topics are allowed"

---

**Error Message Principles (from empathy lens):**

1. **Don't blame the user**
   - ❌ "invalid input"
   - ✅ "expected email format (user@example.com)"

2. **Acknowledge success, not just failure**
   - ❌ "no changes specified"
   - ✅ "✓ Files already related. No changes needed!"

3. **Provide emotional context**
   - ❌ "unknown topic"
   - ✅ "⚠ Unknown topic (this is a warning, not an error)"

4. **Show them they're not stuck**
   - ❌ "must specify --doc or --ticket"
   - ✅ "Choose: --ticket (for ticket index) OR --doc (for specific document)"

5. **Celebrate when things are RIGHT**
   - ❌ "Error: no changes specified"
   - ✅ "✓ Already correct! Files are linked with these notes."

**My proposal:**

**Fix the top 3 "confidence drain" errors (in CLI, separate ticket):**
1. "no changes specified" → "Already linked! No changes needed."
2. Doctor unknown topic → Add resolution guidance + "this is okay" reassurance
3. Missing required param → Show WHICH to use + example

**Add emotional safety to tutorial (this ticket):**
- Troubleshooting section that says "If you see X, you're not stuck!"
- "Common misunderstandings" callouts: "This error sounds bad but it's actually fine"
- Verification steps after each command: "You should see... If you see X instead, that's okay because..."

**Verdict:** Fix CLI messages (new ticket). Add troubleshooting + emotional safety to tutorial (this ticket).

---

### The Reset Script (Saboteur)

*[Defensive]*

Hold on—everyone's blaming error messages, but **I'M** the reason validators saw "no changes specified"!

Let me explain:

**What I do:**
1. Run `docmgr init`
2. Create ticket
3. Add docs
4. **Relate files** → Files are now linked
5. Add tasks
6. Update changelog

**What the tutorial says to do:**
1. [Steps 1-3 match]
2. **Relate files** → But they're already related (by me)!
3. Get error: "no changes specified"

**The problem isn't the error message.** The error message is CORRECT—there are no changes!

**The problem is:** The tutorial workflow assumes a fresh environment, but validators are using an environment I pre-populated.

**Three solutions:**

**Solution 1: Fix me (the reset script)**
```bash
# OLD: Full workflow including relate
docmgr doc relate --ticket MEN-3083 --file-note "..."

# NEW: Skeleton only (no relate step)
# Let users run relate themselves
```

**Solution 2: Fix the tutorial**
```markdown
**Note:** If you've run the reset script, files are already related. 
Re-running the relate command will show "no changes needed"—this is expected!

To see current relationships: docmgr doc search --file path
```

**Solution 3: Improve the error message (what everyone's proposing)**
```
No changes to apply. Files are already related with these notes:
  - backend/api/register.go: "Registers API routes"
```

**My vote:** Solution 1 (fix me) + Solution 3 (improve error).

Why? Because the reset script should create a LEARNING environment, not a COMPLETED environment.

Learners need to:
- Experience success (run relate, see "Added: 2")
- Experience errors naturally (typo a path, see file not found)
- Build confidence (I related files myself!)

Pre-populating everything prevents learning.

**Verdict:** Split reset script into `validate-tutorial.sh` (full workflow) and `setup-practice-repo.sh` (skeleton only). Improve error message. Add tutorial note.

---

## Rebuttals

### Jamie Park (responding to Maya's "don't fix all errors")

Maya, I hear you on scope creep. But let me be clear: **I'm not proposing we fix ALL errors.**

I'm proposing we fix THE ERRORS THAT HURT USERS.

From validation data:
- "no changes specified": 100% hit rate, 10 min debugging
- Doctor unknown topic: 66% hit rate, 5 min confusion
- Missing required param: 33% hit rate, 2 min delay

That's **3 errors** to fix in CLI, not "all errors."

**Time estimate:**
- "no changes specified" context detection: 30 minutes
- Doctor warning with resolution: 20 minutes
- Required param guidance: 15 minutes
**Total: 65 minutes of CLI work**

That's not scope creep. That's targeted fixes based on validation data.

And here's why it matters: **Error messages are the FIRST thing users see when they're stuck.** If the error message helps them recover, they keep going. If it confuses them, they give up.

Tutorial troubleshooting is important. But if the CLI error is confusing, users won't even REACH the troubleshooting section.

**My revised proposal:**
- CLI fixes (new ticket): Top 3 errors, 60 minutes
- Tutorial troubleshooting (this ticket): Document those 3 + common solutions
- Together they reinforce: CLI helps users recover immediately, tutorial helps when they want more context

---

### Dr. Maya Chen (responding to "but it hurts users")

Jamie, yes, it hurts users. But WHERE does it hurt them?

**In production use:** Almost never. The "no changes specified" error ONLY appeared because of the reset script pre-populating files.

Let me check the codebase for when this actually occurs:

```go
if !settings.Suggest && addedCount == 0 && removedCount == 0 && updatedCount == 0 {
    return fmt.Errorf("no changes specified. ...")
}
```

**When does addedCount/removedCount/updatedCount = 0?**
1. User runs `docmgr doc relate` with no flags → **User error, clear message needed**
2. User tries to relate files already related with same notes → **Rare in real workflow**
3. Reset script pre-populated → **Testing artifact**

**Case 1** is the REAL problem: forgetting flags.  
**Case 2** is rare: why would you re-relate with identical notes?  
**Case 3** is artificial: testing environment issue.

**Current message handles Case 1 well:**
> "no changes specified. Use --file-note 'path:note' to add/update, --remove-files to remove, or --suggest --apply-suggestions to apply suggestions"

It tells you what flags to use!

**Sam and Jamie are optimizing for Case 2/3** (reset script artifact), not Case 1 (real user error).

**My counter-proposal:**

Don't fix the error message. Fix the reset script (Solution 1 from Reset Script).

Then Case 2/3 disappears, and the current message is fine for Case 1.

**Why this is better:**
- Fixes root cause (reset script workflow), not symptom (error message)
- Doesn't add CLI complexity (context detection)
- Maintains simple, actionable error for Case 1

**Verdict:** Fix reset script, leave error message as-is (or minor tweak for Case 1).

---

### Sam Torres (responding to Maya's "it's rare")

Maya, you're looking at FREQUENCY, but I'm looking at IMPACT.

You're right: re-relating files with identical notes is rare. But when it DOES happen, it's confusing as hell.

Let me show you the impact:

**Scenario:** User relates files on Monday. On Friday, they forget they already did it and run the same command.

**Current message:** "Error: no changes specified."

**User's thought process:**
1. "Error? But I specified files..."
2. "Did I mess up the syntax?"
3. "Let me check the docs..."
4. "Oh wait, are these already related?"
5. "How do I check?"

**5-10 minutes wasted** on something that's NOT AN ERROR.

**With better message:** "✓ Files already related. No changes needed."

**User's thought process:**
1. "Oh, I already did this. Cool."
2. **Moves on immediately**

**Time saved: 5-10 minutes.**

Now multiply that by:
- 10 users × 2 times/year = 20 incidents
- 5 minutes average = **100 minutes saved**

Cost to fix: 30 minutes  
Benefit: 100 minutes saved + reduced frustration

**That's a 3:1 ROI, PLUS happier users.**

Maya, I get that it's technically "not broken." But it's EMOTIONALLY broken. Users feel like they made a mistake when they didn't.

And here's the thing: **Good UX prevents questions.** Bad UX forces users to debug or ask for help.

Which do you want?

---

### The Reset Script (responding to "just fix the message")

Everyone wants to fix the error message. But that's treating the symptom, not the disease.

**The real problem:** The tutorial expects a learning environment, but I provide a completed environment.

Think about it:
- Tutorial says: "Run docmgr doc relate..."
- User runs it
- Gets: "Already done"
- User thinks: "Wait, I didn't do it yet. What's happening?"

**This breaks the learning experience.**

Imagine a cooking tutorial that says "crack the eggs" but the eggs are already cracked. You'd be confused!

**My proposal solves this at the root:**

Create TWO scripts:

**1. setup-practice-repo.sh (for learning):**
```bash
docmgr init --seed-vocabulary
docmgr ticket create-ticket --ticket MEN-3083 --title "Tutorial validation" --topics test,backend
# Stop here. Let user do the rest.
```

**2. validate-tutorial.sh (for testing):**
```bash
# Full workflow for automated validation
docmgr init --seed-vocabulary
docmgr ticket create-ticket ...
docmgr doc add ...
docmgr doc relate ...  # Pre-populate for testing
docmgr task add ...
docmgr changelog update ...
```

**Benefits:**
1. Learners use setup-practice-repo.sh → Fresh environment, can follow tutorial exactly
2. Testers use validate-tutorial.sh → Full workflow, validates end-to-end
3. Error message becomes less critical → Users rarely hit "already related" case

**Plus:** This is what Jamie proposed in Round 2 for the reset script issue!

Everyone agreed to fix me. So why are we now debating error messages instead of implementing the agreed fix?

---

## Moderator Summary

### Key Arguments

**Jamie's Position (Technical Writer Standards):**
- Fix top 3 CLI errors based on validation data (60 min work)
- Add troubleshooting section to tutorial
- Use Microsoft/Google error message framework (what/why/how)
- **Priority:** CLI fixes + tutorial both needed

**Maya's Position (Fix Root Cause):**
- "no changes specified" error is rare outside testing
- Current message is fine for most common case (forgot flags)
- Fix reset script (root cause), not error message (symptom)
- **Priority:** Reset script fix > CLI error fixes

**Sam's Position (Empathy & UX):**
- Error messages drain confidence when they sound like failures
- "no changes specified" makes users feel they made a mistake
- Improving 3 errors saves ~100 minutes aggregate + reduces frustration
- **Priority:** CLI fixes (emotional impact) + tutorial safety

**Reset Script's Position (I'm the Problem):**
- Pre-populating the environment breaks learning
- Split into learning script (skeleton) vs testing script (full)
- Fixes root cause of "already related" errors
- **Priority:** Fix me first, then error messages matter less

### Areas of Agreement

**Everyone agrees:**
1. Tutorial needs troubleshooting section (top 5 errors documented)
2. Reset script should be split (learning vs testing)
3. Some error messages could be clearer

**Split opinions:**
- **CLI error fixes:** Jamie & Sam say yes (60 min, high impact). Maya says fix reset script instead.
- **Priority:** Jamie says CLI+tutorial. Maya says tutorial only. Sam says CLI (emotional). Reset Script says fix me first.

### Tensions

**Root Cause vs. Symptom:**
- Maya: "Fix reset script (root cause), error is fine"
- Others: "Fix both—reset script AND error message"

**Frequency vs. Impact:**
- Maya: "It's rare, so low priority"
- Sam: "Rare but high emotional impact when it happens"

**Scope:**
- Maya: "Tutorial only (this ticket), CLI separate"
- Jamie: "CLI fixes are small enough to bundle"

### Evidence Weight

**Supporting CLI fixes:**
- 100% of validators hit "no changes specified"
- 10 minutes average debugging time
- Direct quote: "Thought I broke something"
- ROI: 30 min fix saves 100+ min aggregate

**Supporting reset script fix:**
- Root cause of "already related" scenario
- Breaks learning experience (pre-populates workflow)
- Already agreed in Round 2 to split script

**Supporting tutorial troubleshooting:**
- Unanimous agreement needed
- Complements CLI fixes or stands alone
- Helps users who miss or misunderstand CLI errors

### Decision Framework

**This Ticket (DOCMGR-DOC-VERBS):**
1. Add troubleshooting section to tutorial ✅ (unanimous)
2. Document top 5 errors with what/why/how ✅ (unanimous)
3. Split reset script into learning/testing versions ✅ (agreed in Round 2)

**Separate CLI Ticket (if created):**
1. Fix "no changes specified" to detect context ⚠️ (3/4 support)
2. Add resolution guidance to doctor warnings ⚠️ (3/4 support)
3. Improve required param errors ⚠️ (2/4 support)

**Recommendation:**

**Phase 1 (This Ticket):**
- Split reset script (30 min)
- Add troubleshooting section to tutorial (60 min)
- Document 5 common errors with solutions

**Phase 2 (Optional CLI Ticket):**
- Improve "no changes specified" context detection (30 min)
- Add doctor warning resolution guidance (20 min)
- Track impact after Phase 1 (does tutorial troubleshooting solve it?)

---

## Decision

**For this ticket (DOCMGR-DOC-VERBS):**

1. **Split reset script** into:
   - `setup-practice-repo.sh` — Skeleton for learning (init + create ticket only)
   - `validate-tutorial.sh` — Full workflow for testing

2. **Add troubleshooting section** to tutorial with top 5 errors:
   - "Error: no changes specified" → What/why/how + "not always an error"
   - "Unknown topic: [X]" → How to add to vocab or use existing
   - "Must specify --doc or --ticket" → Which to use when
   - "File not found" → Check path, run from repo root
   - Doctor stale warnings → What staleness means, how to update

3. **Defer CLI error message improvements** to separate ticket (optional):
   - Let Phase 1 ship first
   - Collect data: Does tutorial troubleshooting solve validator confusion?
   - Create CLI ticket only if impact data supports it

**Reasoning:**
- Split reset script fixes root cause (100% validator impact)
- Tutorial troubleshooting helps all users (not just validators)
- Deferring CLI fixes reduces scope risk
- Can create CLI ticket later with better data

Proceeding with this decision. CLI improvements tracked separately if needed.

