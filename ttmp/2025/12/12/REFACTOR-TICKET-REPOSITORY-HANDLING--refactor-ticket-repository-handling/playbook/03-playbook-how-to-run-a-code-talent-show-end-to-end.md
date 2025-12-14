---
Title: 'Playbook: How to Run a Code Talent Show (end-to-end)'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/04-code-review-contestant-1-dj-skippy-skip-policy.md
      Note: Example proper code review writeup
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/playbook/01-test-playbook-contestant-1-dj-skippy-skip-policy.md
      Note: 'Candidate #1 execution playbook'
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/playbook/02-test-playbook-contestant-2-ingrid-the-indexer-index-builder-initindex.md
      Note: 'Candidate #2 execution playbook'
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/16-talent-show-candidates-code-performance-review.md
      Note: Candidate roster reference
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/18-the-jury-panel-judge-personas-and-judging-criteria.md
      Note: Jury panel definition
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/19-jury-deliberation-contestant-1-dj-skippy-skip-policy.md
      Note: Example jury deliberation transcript
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T20:31:25.16930234-05:00
---


# 🎭 Playbook: How to Run a Code Talent Show (end-to-end)

## Purpose

This playbook describes the complete "code talent show" workflow—a creative code review methodology that transforms dry technical evaluation into an engaging narrative while maintaining rigorous standards.

**What we're doing:** Taking refactored subsystems and treating them as personified contestants in a talent competition, complete with background stories, signature moves, live performances, and a panel of expert judges who deliberate using actual runtime evidence.

**Why this works better than traditional code review:**

1. **Engagement:** People remember "DJ Skippy's signature move" better than "the segment boundary check in line 80"
2. **Systematic:** Each judge represents a distinct evaluation dimension (robustness, simplicity, spec adherence, maintainability)
3. **Evidence-based:** Judges reference actual test output, not hunches
4. **Memorable:** The deliberation format forces articulation of trade-offs
5. **Fun:** Because code review shouldn't feel like a chore

**The workflow:**
- Scout candidates from the codebase (personify code as contestants)
- Set up observability (make behavior inspectable via tests/CLI)
- Select a jury with explicit criteria
- Run each candidate's "show" (execute real tests, capture output)
- Conduct deliberation (judges debate using evidence)
- Produce artifacts (deliberation transcript + actionable code review)

**The payoff:** You get both entertainment AND rigor. The jury debates are fun to read, but they're grounded in actual code snippets, test results, and spec requirements. The final code review is actionable because it emerged from a structured evaluation process.

## Environment Assumptions

- You can run `docmgr` (system or local binary) to create/update ticket docs.
- You can run Go tests (`go test`) for the repo.
- You are working from within the repo checkout.
- For scenario tests, you have bash + python3 available.

Optional but recommended:

- You have a local build location (e.g. `/tmp`) for binaries and scenario outputs.

## Commands

```bash
# All commands assume you are in the repo root:
cd /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr

# Build a local binary for scenario runs (optional but recommended)
go build -o /tmp/docmgr-local ./cmd/docmgr

# Run the full scenario suite against your local binary (integration safety net)
DOCMGR_PATH=/tmp/docmgr-local bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario-local

# Run a candidate’s performance-stage tests (example: contestant #1 DJ Skippy)
go test ./internal/workspace -run '^TestSkipPolicy_' -count=1 -v

# Run a candidate’s baseline tests (example: contestant #1 unit tests)
go test ./internal/workspace -run 'TestDefaultIngestSkipDir|TestComputePathTags_' -count=1 -v

# Run ingestion end-to-end unit test (example: contestant #2 Ingrid)
go test ./internal/workspace -run TestWorkspaceInitIndex_IngestsDocsTopicsAndRelatedFiles -count=1 -v
```

## Exit Criteria

You have produced (and linked into the ticket) the following artifacts:

1. Candidate roster document (who is in the show, what they perform, what to run).
2. One playbook per candidate (how to run their show in the current codebase).
3. A judge panel document defining personas + criteria.
4. For each candidate run:
   - captured runtime output (test output and/or exported sqlite snapshot),
   - deliberation transcript (jury discussion),
   - final code review writeup (spec mapping + risks + follow-ups).

Plus:

- Ticket `index.md` links to all of the above.
- Ticket `changelog.md` has entries for major artifacts and verdicts.

## Notes and Meta-Commentary

### On the nature of personification

**Why does this work?** Human brains are wired for stories, not abstractions. "The skip policy" is boring. "DJ Skippy, who grew up in chaos and became the canonical gatekeeper" is memorable.

**The psychology:** When you personify code:
- You're forced to articulate its PURPOSE (why does DJ Skippy exist?)
- You identify its PERSONALITY (what makes skip policy unique vs other code?)
- You create EMPATHY (we want DJ Skippy to succeed!)

**Side benefit:** It makes code review collaborative instead of adversarial. Nobody wants to "fail" DJ Skippy—we all want to help them improve their performance.

### On judge dynamics

**The magic happens in Round 2 (cross-examination).** That's where you discover:
- Implicit assumptions you didn't know you were making
- Trade-offs you were resolving unconsciously
- Where "good enough" actually is on the spectrum

**Example:** Murphy and Ockham's nil-check debate revealed the REAL issue wasn't "should we add a check?" but "should we document WHY there's no check?" That insight only emerged through debate.

**Pro tip:** If your deliberation doesn't have at least one heated debate, your judges aren't different enough. Make their personalities more extreme!

### On the relationship to specifications

**With a spec:** The Oracle becomes the most powerful judge because there's an objective truth to reference. Debates can be settled by "what does §6.1 actually say?"

**Without a spec:** Oracle becomes weaker (no authority) and Ada becomes stronger (craft becomes the main standard). Adjust weights accordingly.

**This refactor had an excellent spec** (Design Spec §6 was precise), which made the Oracle's job easy. In projects without specs, you might need a different judge—perhaps "Judge User" who represents end-user needs.

### On time investment

**Total time for DJ Skippy (Contestant #1):**
- Scouting and personification: 30 minutes
- Performance-stage test creation: 3 hours
- Running the show: 5 minutes
- Jury deliberation writeup: 2 hours
- Code review writeup: 1 hour
- Bookkeeping (relate/changelog): 15 minutes

**Total: ~7 hours** for one contestant.

**Was it worth it?** For a 97-line file? Maybe not if you ONLY got a code review. But we also got:
- Performance tests that serve as documentation
- A review format that future engineers can reuse
- Insights about judge trade-offs that inform future design
- Memorable narratives that make the codebase more approachable

**Rule of thumb:** Talent shows take 3-5x longer than traditional code review, but produce 5-10x more valuable artifacts. Use for code that matters.

### On reusability

**The best part:** Once you have the format, subsequent contestants are faster:
- Judges are already defined
- Performance-stage test pattern is established
- Deliberation format is template-izable
- Bookkeeping hygiene is routine

Contestant #2 (Ingrid the Indexer) will probably take 4 hours instead of 7. Contestant #8 might take only 2 hours.

### On adaptation

**You don't need to follow this exactly.** Adapt to your team:
- Fewer judges if your team is small (3 is minimum for good coverage)
- Different weights if your org has different priorities
- Skip performance-stage tests if baseline unit tests are already narrative
- Skip deliberation transcripts if you only want code reviews (but you'll miss the fun!)

**The core idea is universal:** Code review should be evidence-based, multi-dimensional, and memorable.

### On the future (the JURY-AST-INSTRUMENTATION vision)

Right now, judges use existing tests and manual code inspection. The future vision:
- Judges write YAML describing what to observe
- AST tooling auto-instruments code with observability hooks
- Build with `-overlay` produces "jury mode" binary
- Judges watch ACTUAL RUNTIME DECISIONS, not just test results

**Example jury-mode output:**
```
🎭 JURY OBSERVE [SKIP] ❌ SKIP  .meta  reason: implementation metadata
🎭 JURY OBSERVE [TAGS] tasks.md  tags: [is_control_doc]
```

But even without that tooling, the talent show format works TODAY with just:
- Good tests
- Manual code reading
- Structured deliberation

---

## Related

- **Candidate Roster:** `reference/16-talent-show-candidates-code-performance-review.md`
- **Judge Panel:** `reference/18-the-jury-panel-judge-personas-and-judging-criteria.md`  
- **Example Playbooks:**
  - `playbook/01-test-playbook-contestant-1-dj-skippy-skip-policy.md`
  - `playbook/02-test-playbook-contestant-2-ingrid-the-indexer-index-builder-initindex.md`
- **Example Deliberation:** `reference/19-jury-deliberation-contestant-1-dj-skippy-skip-policy.md`
- **Example Code Review:** `analysis/04-code-review-contestant-1-dj-skippy-skip-policy.md`
- **Future Vision:** `JURY-AST-INSTRUMENTATION` ticket (YAML-driven observability)

---

**Remember:** The goal isn't to make code review take longer. The goal is to make code review so engaging that people WANT to do it thoroughly. When code review becomes storytelling, quality improves because people pay attention. 🎭✨

## Step-by-step process (recommended workflow)

### Phase 1: Candidate Scouting 🔍 (The Talent Scout)

**Goal:** Transform boring subsystems into compelling characters with backstories and talents.

**The mindset:** You're a talent scout walking through a codebase looking for *star potential*. Not every piece of code makes a good contestant. You want candidates with:
- **A clear purpose** (they solve one problem well)
- **Observable behavior** (you can watch them work)
- **Interesting edge cases** (they have "signature moves")
- **Real impact** (they matter to the system's correctness)

**The scouting process:**

1. **Read the design spec** (if one exists) and identify the major subsystems
   - Example from this refactor: The spec mentions skip policies (§6), indexing (§7), querying (§10), normalization (§7.3)
   - Each becomes a potential contestant

2. **Walk through the implementation** and identify self-contained units
   - Look for files/packages with clear boundaries
   - Prefer units with 1-3 files (not sprawling systems)
   - Example: `skip_policy.go` is 97 lines, does one thing, perfect contestant!

3. **Personify each candidate** by giving them:
   - **A stage name** (e.g., "DJ Skippy the Bouncer")
   - **A backstory** (why they exist, what problem they solved)
   - **A talent** (their core capability)
   - **A signature move** (the trickiest thing they do correctly)
   - **A performance act** (what they'll demonstrate live)

4. **Define what "good performance" looks like:**
   - What tests should pass?
   - What edge cases must be handled?
   - What output proves correctness?
   - What would failure look like?

**Example from DJ Skippy:**

```markdown
### Contestant #1: "The Bouncer" (Skip Policy)
**Real Name:** skip_policy.go
**Stage Name:** DJ Skippy
**Hometown:** internal/workspace

**Background Story:**
DJ Skippy grew up in chaos where different commands had different ideas 
about which files to skip. After years of inconsistency, they became 
THE canonical gatekeeper...

**Talent:** Path-segment boundary detection
**Signature Move:** Distinguishing /archive/ from /myarchive/

**Performance Act:** The Directory Gauntlet
- Skip .meta/ and _*/ correctly
- Tag archive/scripts/sources with segment-boundary safety
- Recognize control docs only when sibling index.md exists
```

**Anti-patterns to avoid:**

- ❌ Picking contestants that are too large (entire packages with 20 files)
- ❌ Picking trivial code that has no interesting edge cases
- ❌ Forgetting to define what "performance" means (how will we watch them?)
- ❌ Picking code that has no tests (you'll have to write observability first)

**Time estimate:** 1-2 hours for a roster of 8 contestants

**Artifact to create:**

```bash
docmgr doc add \
  --ticket YOUR-TICKET \
  --doc-type reference \
  --title "Talent Show Candidates - Code Performance Review"

# Write the roster with personifications and performance definitions
# Relate key source files to the roster doc
```

**Existing reference (this ticket):**
- `reference/16-talent-show-candidates-code-performance-review.md`

### Phase 2: Observability Setup 🔬 (The Stage Manager)

**Goal:** Make the contestants' behavior observable so judges can watch them perform in real-time, not just read static code.

**The philosophy:** Code review shouldn't be theoretical. Judges need to see actual execution—what inputs were processed, what decisions were made, what outputs were produced. "Works on my machine" is not good enough; we need reproducible evidence.

**The observability ladder (start at the top, work down):**

#### Tier 1: Existing Unit Tests (The Easy Win)
**What:** Tests that already exist, have assertions, and can be run with `go test`  
**When to use:** Always start here. If tests exist and pass, you have baseline correctness.  
**Example:**
```bash
go test ./internal/workspace -run TestDefaultIngestSkipDir -v
```

**Pros:** No setup needed, fast, covers isolated behavior  
**Cons:** May not show the full picture (integration gaps)

#### Tier 2: Performance-Stage Tests (The Show)
**What:** Enhanced tests that output human-readable performance commentary with emojis, difficulty ratings, and explanations  
**When to use:** When you want judges to *understand* behavior, not just see "PASS"  
**Example:**

```bash
go test ./internal/workspace -run '^TestSkipPolicy_' -v

# Output:
# 🎪 ACT 2: The Segment Boundary Challenge
# ✅ ⭐⭐⭐ Advanced  myarchive/doc.md
#    Expected tags: archived=false
#    Explanation: False positive avoidance: 'myarchive' is not '/archive/'
#    ✓ JUDGMENT: CORRECT
```

**Pros:** Judges can read test output like a script, see reasoning  
**Cons:** Requires writing enhanced tests (but they double as great documentation)

**How to create:** Add a `*_performance_test.go` file that:
- Wraps existing behavior in narrative structure
- Adds difficulty ratings (⭐ to ⭐⭐⭐)
- Includes emojis for quick scanning (✅/❌)
- Prints explanations for each case
- Outputs summary statistics

```go
// Pattern: Performance stage test
func TestCandidate_ActName(t *testing.T) {
    fmt.Println("\n🎪 ACT 1: The Act Name")
    fmt.Println("═══════════════════")
    fmt.Println("What this act demonstrates...")
    
    testCases := []struct{
        input      string
        expected   bool
        difficulty string
        explanation string
    }{
        {input: "edge_case", expected: true, 
         difficulty: "⭐⭐⭐ Advanced",
         explanation: "Why this is tricky..."},
    }
    
    for _, tc := range testCases {
        result := FunctionUnderTest(tc.input)
        emoji := emojiFor(result == tc.expected)
        fmt.Printf("%s %s  %s\n", emoji, tc.difficulty, tc.input)
        fmt.Printf("   Expected: %v\n", tc.expected)
        fmt.Printf("   Actual: %v\n", result)
        fmt.Printf("   Explanation: %s\n", tc.explanation)
        // ... assertions
    }
}
```

#### Tier 3: Scenario/Integration Tests (The Full Show)
**What:** Bash scripts that run the actual CLI against a mock workspace  
**When to use:** To prove end-to-end behavior, not just unit correctness  
**Example:**

```bash
go build -o /tmp/docmgr-local ./cmd/docmgr
DOCMGR_PATH=/tmp/docmgr-local \
  bash test-scenarios/testing-doc-manager/run-all.sh /tmp/output
```

**Pros:** Tests the real user experience, catches wiring issues  
**Cons:** Slower, harder to debug, may not be granular enough

#### Tier 4: Exported Artifacts (The Physical Evidence)
**What:** Commands that produce inspectable files (SQLite DBs, JSON reports, logs)  
**When to use:** When you need to examine state that's not visible in test output  
**Example:**

```bash
go run ./cmd/docmgr workspace export-sqlite --out /tmp/workspace.db --force
sqlite3 /tmp/workspace.db "SELECT * FROM docs WHERE is_control_doc = 1;"
```

**Pros:** Judges can inspect at leisure, share artifacts  
**Cons:** Indirect observation (see the result, not the decision process)

**Important reality check for this repo:**
- The `workspace export-sqlite` verb IS wired and works ✅
- The `workspace analyze-skip` verb EXISTS but is NOT wired yet ⚠️
- Most contestants rely on Tier 1-2 (unit + performance tests) 👍

**Pro tip:** Don't add new CLI commands mid-review just for observability. If unit tests + performance tests + export-sqlite aren't enough, you probably picked a contestant that's too hard to evaluate.

### Phase 3: Select a Jury ⚖️ (The Casting Director)

**Goal:** Make evaluation trade-offs explicit by personifying different review perspectives as distinct judges.

**Why we personify judges:** Every code review involves trade-offs (simplicity vs. robustness, spec adherence vs. pragmatism). By giving each perspective a personality and voice, we make these trade-offs visible and force articulation of why we value one dimension over another.

**The core insight:** There's no such thing as "objectively good code." There's code that:
- Won't break (Murphy cares)
- Is easy to understand (Ockham cares)
- Matches requirements (Oracle cares)
- Will be maintainable in 5 years (Ada cares)

All four matter, but they sometimes conflict! The jury format forces us to debate those conflicts explicitly.

**Recommended panel composition:**

#### The Four-Judge Panel (Balanced Coverage)

1. **🔨 Judge Murphy - The Pessimist** (25% weight)
   - **Focus:** Robustness, edge cases, production failures, error handling
   - **Personality:** Battle-scarred veteran who's debugged production at 3 AM
   - **Catchphrase:** "But what if the disk is full?"
   - **What they review:** Error paths, nil checks, panic potential, defensive code

2. **🗡️ Judge Ockham - The Minimalist** (25% weight)
   - **Focus:** Simplicity, absence of over-engineering, clarity
   - **Personality:** Medieval monk with a razor, hates unnecessary complexity
   - **Catchphrase:** "Could this be simpler?"
   - **What they review:** Line count, abstraction layers, helper justification

3. **📜 The Spec Oracle - The Purist** (25% weight)
   - **Focus:** Specification adherence, completeness, semantic correctness
   - **Personality:** Mystical being that embodies the design document
   - **Catchphrase:** "The Oracle observes divergence at §6.1"
   - **What they review:** Requirement coverage, spec citations, behavior matching intent

4. **💎 Judge Ada - The Craftsperson** (25% weight)
   - **Focus:** Implementation quality, maintainability, testability, aesthetics
   - **Personality:** Ada Lovelace returned, believes code is poetry
   - **Catchphrase:** "This code shall outlive us all—write it accordingly"
   - **What they review:** Naming, structure, test quality, long-term maintainability

**Why this panel works:**
- **Murphy vs Ockham** creates productive tension (safety vs simplicity)
- **Oracle vs Ada** balances "does it match spec?" vs "is it maintainable?"
- Equal weighting (25% each) forces consensus—no judge can dominate

**Alternative panels** (use when appropriate):

- **Security-focused panel:** Add "Judge Schneier" (crypto/security expert) at 20%, reduce others to 20% each
- **Performance-focused panel:** Add "Judge Knuth" (algorithms/efficiency) at 20%
- **Accessibility-focused panel:** Add "Judge Berners-Lee" (usability/API design) at 20%

**Creating your jury panel:**

```bash
docmgr doc add \
  --ticket YOUR-TICKET \
  --doc-type reference \
  --title "The Jury Panel: Judge Personas and Judging Criteria"

# For each judge, write:
# - Full name and title
# - Origin story (why they have this perspective)
# - Personality traits (what annoys them, what makes them happy)
# - Judging criteria (with % weights)
# - Catchphrases (for consistency in deliberation)
```

**Time estimate:** 1-2 hours to develop personas with depth

**Existing reference (this ticket):**
- `reference/18-the-jury-panel-judge-personas-and-judging-criteria.md`

**Pro tip:** Make judges have *opinions* and *biases*. Murphy should be genuinely paranoid. Ockham should get physically pained by complexity. Oracle should speak in formal third-person. Ada should wax poetic about craftsmanship. This makes deliberations fun to read AND forces you to inhabit different evaluation mindsets.

### Phase 4: Run the Candidate's Show 🎬 (The Performance)

**Goal:** Execute the contestant's code under controlled conditions and capture comprehensive evidence of their behavior.

**The performance philosophy:** A talent show isn't just about "can you do the thing?" It's about "show us HOW you do the thing, and make it impressive." We want to see:
- Basic skills executed flawlessly
- Edge cases handled gracefully
- The signature move performed perfectly
- Everything documented with running commentary

**Pre-show checklist:**

Before running any candidate's show, create their playbook:

```bash
docmgr doc add \
  --ticket YOUR-TICKET \
  --doc-type playbook \
  --title "Test Playbook: Contestant #N Name (Subsystem)"
```

The playbook should document:
- What files/functions are being tested
- What commands to run (exact bash commands, copy-pasteable)
- What output to expect
- What "success" looks like
- What runtime evidence will be captured

**Running the show (step-by-step):**

#### Act 1: The Baseline Tests (Prove Basic Correctness)

Run existing unit tests to establish baseline:

```bash
# Example for DJ Skippy:
go test ./internal/workspace \
  -run 'TestDefaultIngestSkipDir|TestComputePathTags_' \
  -count=1 -v

# What judges want to see:
# - All tests pass ✓
# - Edge cases covered (check test names)
# - Fast execution (<1s typically)
```

**Capture this output** (copy-paste or redirect to file):
```bash
go test ./internal/workspace -run TestDefaultIngestSkipDir -v 2>&1 | tee baseline-tests.log
```

#### Act 2: The Performance Stage (Demonstrate Mastery)

Run performance-stage tests that provide narrative commentary:

```bash
# Example for DJ Skippy:
go test ./internal/workspace -run '^TestSkipPolicy_' -count=1 -v

# What judges want to see:
# 🎪 ACT titles with clear goals
# ⭐ Difficulty ratings showing edge case coverage
# ✅/❌ Visual pass/fail indicators
# Explanations of WHY cases are tricky
# 📊 Performance summaries with statistics
# 📋 JSON reports for programmatic analysis
```

**Key observation:** If you see:
```
🏆 DJ SKIPPY: FLAWLESS PERFORMANCE!
```

That's good! But judges will still dig into the code to see HOW that flawlessness was achieved.

#### Act 3: The Integration Reality Check (End-to-End Behavior)

Run scenario tests or real CLI commands:

```bash
# Build a local binary
go build -o /tmp/docmgr-local ./cmd/docmgr

# Run full scenario suite
DOCMGR_PATH=/tmp/docmgr-local \
  bash test-scenarios/testing-doc-manager/run-all.sh /tmp/scenario-output

# Or run specific scenarios
bash test-scenarios/testing-doc-manager/19-export-sqlite.sh /tmp/scenario-output
```

**What this proves:** The candidate works not just in isolation, but integrated with the full system.

**Current reality for this refactor:**
- Most contestants aren't wired into normal verbs yet
- `export-sqlite` IS wired, so it proves the ingestion → indexing → export path
- Scenario suite includes `19-export-sqlite.sh` which exercises the full Workspace backend

#### Act 4: The Artifact Inspection (Physical Evidence)

If the candidate produces artifacts, inspect them:

```bash
# Example: Export SQLite and inspect
go run ./cmd/docmgr workspace export-sqlite \
  --out /tmp/workspace.db --root ttmp --force

# Judges can now inspect the artifact:
sqlite3 /tmp/workspace.db <<SQL
SELECT COUNT(*) as skipped 
FROM docs 
WHERE path LIKE '%/.meta/%' OR path LIKE '%/_templates/%';
-- Expected: 0 (they should be skipped at ingestion!)

SELECT COUNT(*) as control_docs
FROM docs
WHERE is_control_doc = 1;
-- Expected: ≥3 (index.md, tasks.md, changelog.md per ticket)

SELECT path, is_archived_path, is_scripts_path, is_control_doc
FROM docs
LIMIT 20;
-- Judge can manually inspect tag assignments
SQL
```

**What this proves:** Not just that tests pass, but that the RIGHT DATA ended up in the RIGHT STATE.

**Capturing evidence for judges:**

Create a performance report file:

```bash
cat > /tmp/dj-skippy-evidence.txt <<EOF
=== CONTESTANT: DJ Skippy (Skip Policy) ===

=== ACT 1: Baseline Unit Tests ===
$(go test ./internal/workspace -run TestDefaultIngestSkipDir -v 2>&1)

=== ACT 2: Performance Stage ===
$(go test ./internal/workspace -run '^TestSkipPolicy_' -v 2>&1)

=== ACT 3: Integration (Export SQLite) ===
$(go run ./cmd/docmgr workspace export-sqlite --out /tmp/workspace.db --force 2>&1)

=== ACT 4: Artifact Inspection ===
Skipped paths in DB: $(sqlite3 /tmp/workspace.db "SELECT COUNT(*) FROM docs WHERE path LIKE '%/.meta/%' OR path LIKE '%/_templates/%';")
Control docs found: $(sqlite3 /tmp/workspace.db "SELECT COUNT(*) FROM docs WHERE is_control_doc = 1;")

=== VERDICT ===
All acts completed successfully.
Performance duration: $(date)
EOF
```

**Time estimate:** 15-30 minutes per contestant (longer if writing new performance tests)

**Existing references (this ticket):**
- `playbook/01-test-playbook-contestant-1-dj-skippy-skip-policy.md`
- `playbook/02-test-playbook-contestant-2-ingrid-the-indexer-index-builder-initindex.md`

**Common pitfalls:**
- ❌ Only running tests without looking at code
- ❌ Assuming "tests pass" means "correctly implements spec"
- ❌ Not capturing output (can't review what you can't see)
- ❌ Running tests without `-v` flag (miss important details)

### Phase 5: Jury Deliberation 🎙️ (The Main Event)

**Goal:** Transform raw test output and code inspection into a structured evaluation that surfaces trade-offs, debates alternatives, and produces actionable recommendations.

**The deliberation format:** Three rounds where judges speak with increasing synthesis:

#### Round 1: Individual Assessments (5 minutes per judge)

**Format:** Each judge speaks independently, focusing on THEIR criteria:

```markdown
**Judge Murphy Speaks:**

*Murphy opens skip_policy.go with suspicious eyes*

"Let me look for error handling..."

[Murphy uses grep to search for error/panic/nil]

"I see only one error check: line 77 in hasSiblingIndex. It does os.Stat 
and returns err == nil. What if permission denied?"

[Murphy examines the logic]

"Wait... treating permission denied as 'not found' is actually CORRECT here. 
Conservative fallback. Good."

[Murphy reviews test output]

"The performance tests cover the myarchive vs archive edge case. That's 
a production bug I've seen before. Excellent coverage."

**Murphy's Score: 8.5/10**
- Edge cases: 9/10
- Error handling: 9/10  
- Production readiness: 8/10
- Robustness: 9/10
```

**Key elements of individual assessment:**
1. **Tool use:** Judges actually grep/search/inspect code
2. **Evidence citation:** Reference line numbers, test output, specific cases
3. **Reasoning:** Explain WHY something is good/bad
4. **Preliminary score:** On their criteria (weighted)

**Each judge must:**
- Review the actual code (not just test results)
- Reference specific line numbers or functions
- Connect observations to their judging criteria
- Provide a preliminary score with justification

#### Round 2: Cross-Examination (10 minutes total)

**Format:** Judges debate trade-offs and challenge each other's assessments:

```markdown
**Murphy vs. Ockham:**

Murphy: "You gave this a perfect 10, but there's NO NIL CHECK on the 
DirEntry parameter!"

Ockham: "Adding a nil check is UNNECESSARY COMPLEXITY. The stdlib 
guarantees filepath.WalkDir never passes nil."

Murphy: "But what if the stdlib has a bug—"

Ockham: "Should we also check if true equals true? You draw the line 
at stdlib contracts; I draw it at paranoia."

Oracle: *glowing* "The specification does not mandate nil checks. The 
contract with filepath.WalkDir is external. WITHIN SPEC."

Murphy: *grudgingly* "Fine. But a COMMENT would help..."

Ockham: "On THAT, we agree."
```

**Key elements of cross-examination:**
1. **Identify conflicts:** Find where judges disagree
2. **Surface trade-offs:** Make implicit decisions explicit
3. **Force articulation:** Make judges defend their positions
4. **Reach compromise:** Find common ground or agree to disagree

**Common debate patterns:**
- **Murphy vs Ockham:** Safety vs simplicity (defensive checks vs minimal code)
- **Oracle vs Ada:** Spec compliance vs maintainability (letter of law vs spirit)
- **Ada vs Ockham:** Comments vs self-documenting code
- **Murphy vs Oracle:** Who's responsible for edge cases (impl vs spec)

**The magic:** Through debate, you discover:
- Where your assumptions differ
- What trade-offs are being made implicitly
- Where documentation is needed
- What "good enough" actually means

#### Round 3: Final Scoring and Consensus (5 minutes)

**Format:** Judges revise scores based on debate and reach consensus:

```markdown
**Murphy's Final Score: 8.75/10** (raised from 8.5 after debate convinced 
him that nil check is unnecessary)

**Ockham's Final Score: 9.75/10** (lowered from 10 after agreeing that 
WHY comments serve simplicity's deeper purpose)

**Oracle's Final Score: 9.5/10** (perfect spec adherence, minor doc gaps)

**Ada's Final Score: 9.75/10** (beautiful code, needs WHY comments)

**Aggregate: 9.45/10 → 🏆 GOLDEN BUZZER**

**Consensus Statement:**
"This implementation is production-ready with exceptional quality. The 
only recommendation is adding WHY comments at key decision points."

**Specific Follow-Ups:**
1. Add comment explaining filepath.WalkDir contract assumption
2. Add comment explaining os.Stat error semantics  
3. Add comment explaining segment boundary rationale
4. Consider citing Decision 6, Decision 7 from spec
```

**The consensus must include:**
- Aggregate score (weighted average)
- Final verdict (Golden Buzzer / Pass / Needs Work / Fail)
- Consensus statement (what all judges agree on)
- Specific, actionable recommendations (not vague advice)

**Time estimate:** 1-2 hours for complete deliberation per contestant

**Document structure:**

```bash
docmgr doc add \
  --ticket YOUR-TICKET \
  --doc-type reference \
  --title "Jury Deliberation: Contestant #N Name"

# Structure:
# - Performance summary (test results)
# - Preliminary research (each judge's initial code inspection)
# - Round 1: Individual assessments (5 min each)
# - Round 2: Cross-examination (debates)
# - Round 3: Final scores and consensus
# - Appendix: Code snippets referenced
```

**Existing reference (this ticket):**
- `reference/19-jury-deliberation-contestant-1-dj-skippy-skip-policy.md`

**Pro tips:**
- Make judges reference ACTUAL lines of code (not "the error handling looks good")
- Use real test output quotes ("Act 2: 7/7 cases passed including 3 edge cases")
- Let judges change their minds during debate (that's growth!)
- Don't force consensus—if judges genuinely disagree, document the disagreement

### Phase 6: Proper Code Review Writeup 📋 (The Official Record)

**Goal:** Translate the entertaining deliberation into a professional, actionable code review that could be submitted as a PR review or design checkpoint.

**Why both formats?** The deliberation is fun and memorable (people will actually read it!), but organizations need traditional code review artifacts for records, compliance, and handoff to new engineers. Think of the deliberation as the "director's commentary" and the code review as the "official release."

**The translation process:**

The deliberation gave you:
- Specific code concerns from each judge
- Trade-off discussions
- Aggregate scoring
- Recommended improvements

Now distill that into professional format:

#### Required Sections

**1. Executive Summary (The TL;DR)**

Start with the verdict so busy people can stop reading:

```markdown
## Executive Summary

**Verdict:** ✅ SHIP (with minor documentation improvements)
**Score:** 9.45/10 (aggregate across 4 evaluation dimensions)
**Critical Issues:** None
**Recommended Follow-ups:** Add 4 WHY comments (30-minute task)

This implementation is production-ready and demonstrates exceptional code 
quality. The skip policy logic is correct, well-tested, and adheres 
perfectly to Specification §6.
```

**2. Scope (What Was Reviewed)**

Be precise about boundaries:

```markdown
## Scope

**Files reviewed:**
- internal/workspace/skip_policy.go (97 lines)
- internal/workspace/skip_policy_test.go (119 lines)  
- internal/workspace/skip_policy_performance_test.go (568 lines)

**Functions reviewed:**
- DefaultIngestSkipDir (directory skip predicate)
- ComputePathTags (tag computation)
- containsPathSegment (helper, segment-boundary check)
- hasSiblingIndex (helper, control-doc detection)
- isControlDocBase (helper, name matching)

**NOT in scope:**
- Integration with index_builder.go (separate review)
- Performance characteristics at scale (deferred)
```

**3. Specification Mapping (Requirements Coverage)**

Tie implementation to spec sections:

```markdown
## Specification Mapping

This implementation fulfills Design Spec §6: Canonical Skip Rules.

| Requirement | Spec Section | Implementation | Status |
|-------------|--------------|----------------|--------|
| Skip .meta/ entirely | §6.1 | Line 28-30 | ✅ Complete |
| Skip _*/ entirely | §6.1 | Line 31-33 | ✅ Complete |
| Tag archive/ paths | §6.1 | Line 48 | ✅ Complete |
| Tag scripts/ paths | §6.1 | Line 49 | ✅ Complete |
| Tag control docs | §6.2 | Line 55-57 | ✅ Complete |
| Segment-boundary safety | Implied | Line 80-94 | ✅ Exceeds spec |
```

**4. Runtime Behavior Evidence (What We Observed)**

Ground claims in actual execution:

```markdown
## Runtime Behavior Evidence

### Unit Tests (Baseline Correctness)
- TestDefaultIngestSkipDir: 7/7 cases passed
- TestComputePathTags_*: All segment checks passed
- Execution time: <0.1s

### Performance Stage Tests (Edge Case Verification)
- Act 1 (Classic Skip): 6/6 cases, 100% accuracy
- Act 2 (Segment Boundaries): 7/7 cases including 3 advanced edge cases
  - Correctly distinguishes "myarchive" from "/archive/"
  - Correctly distinguishes "scripts-old" from "/scripts/"
- Act 3 (Control Docs): 5/5 cases, sibling-index logic verified
- Grand Finale: 15/15 realistic paths correctly classified

### Integration Evidence (Export SQLite)
- workspace export-sqlite command succeeds
- Exported DB query: 0 docs under .meta/ or _*/ (skip works)
- Exported DB query: Control docs properly tagged
```

**5. Strengths (What's Good)**

```markdown
## Strengths

1. **Simplicity:** 97 lines total, zero abstraction layers, readable by juniors
2. **Correctness:** 100% test pass rate, handles all specified skip rules
3. **Edge case handling:** Segment-boundary logic prevents subtle false positives
4. **Testability:** Pure functions, comprehensive test coverage
5. **Spec adherence:** Perfect mapping to Design Spec §6
6. **Performance:** Sub-millisecond execution, no allocations in hot path
```

**6. Risks / Concerns (What to Watch)**

Be honest about gaps:

```markdown
## Risks and Concerns

### Low Risk
- **Missing nil check on DirEntry:** Acceptable because filepath.WalkDir 
  contract guarantees non-nil. However, lacks comment explaining this 
  assumption.
  - Impact: Low (stdlib contract is stable)
  - Mitigation: Add comment documenting contract dependency

### Observability Gaps
- **hasSiblingIndex error handling:** os.Stat errors treated same as 
  "not found". Correct behavior, but undocumented.
  - Impact: Medium (debugging confusion if permission denied)
  - Mitigation: Add comment explaining conservative fallback

### Documentation Gaps
- **Missing WHY comments:** Code is clear on WHAT, but design rationale 
  is not explained inline.
  - Impact: Medium (future maintainers may "simplify" incorrectly)
  - Mitigation: Add 4 targeted comments (see recommendations)
```

**7. Recommendations (Concrete, Actionable)**

Give diffs, not vague advice:

```markdown
## Recommendations

### Priority 1: Add WHY Comments (30-minute task)

Add comment explaining contract assumption:
```go
// DefaultIngestSkipDir is the canonical ingest-time directory skip predicate.
//
// Spec: §6.1 (directories).
// - Always skip `.meta/` entirely (Decision 6).
// - Always skip underscore dirs (`_*/`) entirely (Decision 7).
//
// Note: Assumes d is non-nil per filepath.WalkDir contract.
func DefaultIngestSkipDir(_ string, d fs.DirEntry) bool {
```

[Continue with 3 more specific comment additions...]

### Priority 2: Consider Spec Citation Enhancement (optional, 15 minutes)

Currently cites "§6" but not decision numbers. Could add:
```go
if name == ".meta" {
	return true  // Decision 6: skip .meta/ entirely
}
```

This is minor but helps spec traceability.
```

**8. Ship Checklist (Final Verification)**

```markdown
## Ship Checklist

Before merging to main:
- [ ] All tests pass (baseline + performance stage)
- [ ] Scenario tests pass (integration proof)
- [ ] WHY comments added per recommendations
- [ ] Code review approved by 2+ engineers
- [ ] RelatedFiles updated in design spec doc
- [ ] Changelog updated with ship entry
```

**Time estimate:** 1-2 hours to write comprehensive code review

**Document creation:**

```bash
docmgr doc add \
  --ticket YOUR-TICKET \
  --doc-type analysis \
  --title "Code Review: Contestant #N Name"

# Relate the code under review + spec + tests
docmgr doc relate --doc ttmp/.../analysis/XX-code-review.md \
  --file-note "/path/to/code.go:Code under review" \
  --file-note "/path/to/spec.md:Specification authority"
```

**Existing reference (this ticket):**
- `analysis/04-code-review-contestant-1-dj-skippy-skip-policy.md`

**The golden rule:** Every criticism must include either:
1. A specific code reference (line number or function name), OR
2. A concrete recommendation (ideally a diff)

Vague comments like "error handling could be better" are not allowed. Instead: "Line 77: os.Stat error is treated as 'not found'. Add comment explaining this is intentional conservative fallback."

**Pro tip from this run:**
When judges debated Murphy vs Ockham on nil checks, the debate itself surfaced the REAL issue: not "should we add the check?" but "should we document WHY we didn't?" The code review captures that insight: "Add comment explaining filepath.WalkDir contract assumption."

## Common Pitfalls (Learn from Our Mistakes)

### Pitfall #1: "The tests pass, ship it!"

**The trap:** Seeing green tests and assuming that means correct implementation.

**Why it fails:** Tests might pass but:
- Cover the wrong cases (test what's easy, not what's critical)
- Have wrong expectations (test encodes a bug)
- Miss integration issues (unit tests pass, integration fails)

**How we avoided it:** Judges reviewed BOTH test output AND the actual code. Oracle verified that tests actually checked spec requirements. Murphy looked for edge cases NOT covered by tests.

**Example:** DJ Skippy's tests checked `myarchive` vs `archive`—an edge case that many test suites would miss because "it's just a substring check, what could go wrong?"

### Pitfall #2: Forgetting to personify judges properly

**The trap:** Writing deliberation as "the code looks good" instead of inhabiting distinct judge perspectives.

**Why it fails:** You miss the trade-offs! If all judges just say "yep, looks fine," you haven't actually done multi-dimensional evaluation.

**How we avoided it:** Each judge had a distinct voice:
- Murphy ALWAYS looked for error paths first
- Ockham counted lines and checked for abstractions
- Oracle cross-referenced spec sections systematically
- Ada examined naming and maintainability

**Pro tip:** If your judges all agree immediately, you probably didn't develop their personalities enough. Real code reviews involve trade-offs and disagreement!

### Pitfall #3: Vague recommendations

**The trap:** Ending with "add better error handling" or "improve documentation."

**Why it fails:** Vague advice doesn't get acted on. What specific errors? Which functions need docs?

**How we avoided it:** Every recommendation in the final code review included:
- Specific line numbers
- Exact diffs for comments
- Concrete time estimates ("30-minute task")

**Example:**
- ❌ Bad: "Add more comments"
- ✅ Good: "Add comment at line 26 explaining filepath.WalkDir contract assumption"

### Pitfall #4: Overusing the format for simple code

**The trap:** Running a full talent show for a 10-line function.

**Why it fails:** Talent shows work for "performance pieces"—code with interesting behavior, edge cases, and trade-offs. Not everything needs this treatment.

**When to use talent shows:**
- ✅ Multi-file subsystems with clear boundaries
- ✅ Code with tricky edge cases or algorithm complexity
- ✅ Code implementing important spec requirements
- ✅ Code that's critical to system correctness

**When NOT to use:**
- ❌ Trivial getters/setters
- ❌ Boilerplate configuration
- ❌ Code with no interesting edge cases
- ❌ Code with no test coverage (fix that first!)

**Example:** DJ Skippy (97 lines, segment-boundary logic, spec §6 implementation) = perfect candidate. A simple config parser? Just do regular code review.

### Pitfall #5: Building observability tools during review

**The trap:** "We need a `workspace analyze-skip` command to really see this work!"

**Why it fails:** You're reviewing EXISTING code, not building new features. Adding observability tools mid-review expands scope and delays the actual review.

**How we avoided it:** We STARTED to write `workspace analyze-skip`, then realized: "Wait, the performance tests already show everything. We don't need the CLI verb yet."

**The rule:** Use existing tests + artifacts. Only build new observability if:
1. No tests exist at all, AND
2. The behavior is genuinely unobservable otherwise

In this refactor, `export-sqlite` was enough for integration proof.

### Pitfall #6: Not documenting judge bias

**The trap:** Treating aggregate score as objective truth.

**Why it fails:** Different organizations value different things. A startup might weight Ockham (simplicity) at 40% and Murphy (robustness) at 15%. A bank might invert that.

**How we avoided it:** Our judge panel doc explicitly states:
- Each judge's focus (robustness, simplicity, spec, maintainability)
- Weight for each (25% each, balanced)
- Scoring rubric (1-10 scale, what each score means)

**Pro tip:** If your team consistently disagrees with talent show verdicts, adjust judge weights. If your CTO always says "too complex," increase Ockham's weight to 35%.

### Pitfall #7: Skipping the bookkeeping

**The trap:** Writing great deliberations but forgetting to `docmgr doc relate` and update changelogs.

**Why it fails:** Three months later, someone reads the deliberation and asks "which code was this reviewing?" and you can't remember.

**How we avoided it:** After EVERY new document:

```bash
# Relate immediately
docmgr doc relate --doc ttmp/.../new-doc.md \
  --file-note "/path/to/code.go:What this relates to"

# Update changelog
docmgr changelog update --ticket YOUR-TICKET \
  --entry "Added deliberation for contestant #N, verdict: GOLDEN BUZZER"

# Update ticket index
docmgr doc relate --ticket YOUR-TICKET \
  --file-note "ttmp/.../new-doc.md:Brief description"
```

The last step (updating ticket index) makes documents discoverable from the ticket front page.

## Complete Example: DJ Skippy's Journey (What We Actually Did)

Here's the real workflow we executed for Contestant #1:

### Week 1: Scouting and Setup

```bash
# 1. Created talent show candidates document
docmgr doc add --ticket REFACTOR-TICKET-REPOSITORY-HANDLING \
  --doc-type reference \
  --title "Talent Show Candidates - Code Performance Review"

# Wrote 8 contestant profiles, including DJ Skippy:
# - Background story (consolidates inconsistent skip logic)
# - Talent (path-segment boundary detection)
# - Signature move (myarchive vs /archive/)
# - Performance definition (what tests to run)

# 2. Created judge panel
docmgr doc add --ticket REFACTOR-TICKET-REPOSITORY-HANDLING \
  --doc-type reference \
  --title "The Jury Panel: Judge Personas and Judging Criteria"

# Defined 4 judges with 25% weight each:
# - Murphy (robustness), Ockham (simplicity), Oracle (spec), Ada (craft)
```

### Week 2: Building the Performance Stage

```bash
# 3. Created performance-stage tests
# Wrote skip_policy_performance_test.go with:
# - TestSkipPolicy_Act1_TheClassicSkip
# - TestSkipPolicy_Act2_TheSegmentBoundaryChallenge  
# - TestSkipPolicy_Act3_TheControlDocRecognition
# - TestSkipPolicy_GrandFinale_TheFullDirectoryTree

# Each test outputs narrative commentary:
go test ./internal/workspace -run '^TestSkipPolicy_' -v
# Output:
# 🎪 ACT 1: The Classic Skip
# ❌ ⭐ Basic  .meta
#    Decision: SKIP
#    ✓ JUDGMENT: CORRECT
```

### Week 3: The Performance

```bash
# 4. Ran DJ Skippy's show
go test ./internal/workspace -run '^TestSkipPolicy_' -v > dj-skippy-performance.log

# Results:
# - Act 1: 6/6 passed
# - Act 2: 7/7 passed (including 3 advanced edge cases)
# - Act 3: 5/5 passed  
# - Grand Finale: 15/15 paths correct
# - Verdict: 🏆 FLAWLESS PERFORMANCE

# 5. Created playbook documenting how to reproduce
docmgr doc add --ticket REFACTOR-TICKET-REPOSITORY-HANDLING \
  --doc-type playbook \
  --title "Test Playbook: Contestant #1 DJ Skippy"

# Documented exact commands, expected output, verification steps
```

### Week 4: The Deliberation

```bash
# 6. Conducted jury deliberation
docmgr doc add --ticket REFACTOR-TICKET-REPOSITORY-HANDLING \
  --doc-type reference \
  --title "Jury Deliberation: Contestant #1 DJ Skippy"

# Wrote:
# - Preliminary research (each judge examines code with grep/tools)
# - Round 1: Individual assessments (Murphy: 8.5, Ockham: 10, Oracle: 9.5, Ada: 9.75)
# - Round 2: Cross-examination (Murphy vs Ockham debate on nil checks!)
# - Round 3: Consensus (aggregate: 9.45/10, GOLDEN BUZZER verdict)

# 7. Translated to proper code review
docmgr doc add --ticket REFACTOR-TICKET-REPOSITORY-HANDLING \
  --doc-type analysis \
  --title "Code Review: Contestant #1 DJ Skippy"

# Wrote professional code review with:
# - Executive summary (verdict + score)
# - Spec mapping table
# - Runtime evidence section
# - Concrete recommendations with diffs
# - Ship checklist
```

### Throughout: Bookkeeping

```bash
# After each document, related files:
docmgr doc relate --doc ttmp/.../deliberation.md \
  --file-note "/abs/path/to/skip_policy.go:Code under review" \
  --file-note "/abs/path/to/spec.md:Spec authority §6"

# After major milestones, updated changelog:
docmgr changelog update --ticket REFACTOR-TICKET-REPOSITORY-HANDLING \
  --entry "Completed jury deliberation for DJ Skippy: 9.45/10, GOLDEN BUZZER"

# Updated ticket index so docs are discoverable:
docmgr doc relate --ticket REFACTOR-TICKET-REPOSITORY-HANDLING \
  --file-note "ttmp/.../deliberation.md:Jury verdict for contestant #1"
```

---

## Workflow Cheatsheet (Copy-Paste Ready)

```bash
# === PHASE 1: SCOUTING ===
docmgr doc add --ticket TICKET --doc-type reference \
  --title "Talent Show Candidates"
# Write roster, relate source files

# === PHASE 2: OBSERVABILITY ===  
# Write performance-stage tests in *_performance_test.go
# Add narrative output (🎪 ACT headers, ⭐ difficulty, ✅ judgments)

# === PHASE 3: JURY SELECTION ===
docmgr doc add --ticket TICKET --doc-type reference \
  --title "The Jury Panel"
# Define 4 judges with distinct criteria

# === PHASE 4: RUN THE SHOW ===
go test ./path/to/package -run '^TestCandidate_' -v | tee performance.log

# Create playbook:
docmgr doc add --ticket TICKET --doc-type playbook \
  --title "Test Playbook: Contestant #N Name"

# === PHASE 5: DELIBERATION ===
docmgr doc add --ticket TICKET --doc-type reference \
  --title "Jury Deliberation: Contestant #N"
# Write: research → Round 1 → Round 2 → Round 3 → verdict

# === PHASE 6: CODE REVIEW ===
docmgr doc add --ticket TICKET --doc-type analysis \
  --title "Code Review: Contestant #N"
# Write: summary → scope → spec → evidence → strengths → risks → recommendations

# === BOOKKEEPING (after each doc) ===
docmgr doc relate --doc ttmp/.../doc.md --file-note "/path:note"
docmgr doc relate --ticket TICKET --file-note "ttmp/.../doc.md:description"
docmgr changelog update --ticket TICKET --entry "milestone description"
```

---

## When to Use This Process (and When Not To)

### ✅ Great fit for:
- **Major refactors** with explicit specifications (like this Workspace refactor)
- **Critical subsystems** where correctness is non-negotiable (security, data integrity)
- **Teaching moments** where you want junior engineers to learn evaluation techniques
- **Contentious code** where different engineers have strong opinions
- **Post-mortems** reviewing code that caused production issues

### ❌ Overkill for:
- Trivial bug fixes (one-line changes)
- Straightforward features with obvious implementations
- Code with no tests (write tests first!)
- Urgent hotfixes (talent shows take time; save for post-mortem)
- Preliminary/prototype code (wait until it's stabilized)

### The sweet spot:
**Use talent shows for the 20% of code that's 80% of the risk.** The skip policy is 97 lines, but it determines "what is a document" for the entire system. Getting it wrong would be catastrophic. That's worth the talent show treatment.

For routine code, regular PR review is fine.

---

## Success Metrics (How to Know It Worked)

After running a talent show, you should have:

### Deliverables
- [ ] Candidate roster with 4-8 contestants
- [ ] Judge panel with 4 distinct personas
- [ ] One playbook per contestant (how to run their show)
- [ ] One deliberation per contestant reviewed
- [ ] One code review per contestant reviewed
- [ ] All docs linked via RelatedFiles
- [ ] Ticket index references all artifacts
- [ ] Changelog entries for major milestones

### Outcomes
- [ ] At least one interesting debate between judges (signals trade-off identification)
- [ ] At least one judge changed their score during cross-examination (signals learning)
- [ ] Concrete recommendations with line numbers (signals actionability)
- [ ] A verdict you're comfortable defending to stakeholders
- [ ] New engineers can understand the review by reading deliberation → code review

### Team Benefits
- [ ] Team members reference contestants by stage name ("Did we check DJ Skippy?")
- [ ] Debates from deliberation influence future design decisions
- [ ] Talent show becomes a shared language for discussing code quality
- [ ] Engineers WANT to read the review (engagement)

**If you achieve 80% of the above, the talent show succeeded!**

---

## Recommended docmgr hygiene (The Boring But Important Part)

After creating any artifact doc, immediately wire it into docmgr's knowledge graph:

```bash
# Step 1: Relate key files TO the doc
docmgr doc relate --doc ttmp/.../your-new-doc.md \
  --file-note "/abs/path/to/file.go:Why this file matters" \
  --file-note "/abs/path/to/spec.md:Authority for requirements"

# Step 2: Relate the doc TO the ticket index  
docmgr doc relate --ticket YOUR-TICKET \
  --file-note "ttmp/.../your-new-doc.md:Brief description"

# Step 3: Update changelog
docmgr changelog update --ticket YOUR-TICKET \
  --entry "Milestone description" \
  --file-note "ttmp/.../your-new-doc.md:Artifact created"
```

**Why this matters:** Three months from now, when someone asks "where's the code review for the skip policy?" they can:
1. Start at the ticket index
2. See links to all artifacts
3. Follow RelatedFiles to see what code was reviewed
4. Check changelog to understand the timeline

Without this hygiene, even great documentation becomes "lost knowledge."

---

## Final Thoughts: Why This Works

**The insight:** Code review is both science AND art. Traditional reviews focus on the science (correctness, performance, edge cases) but neglect the art (maintainability, clarity, beauty).

**The talent show format forces you to engage with both:**
- Judges are grounded in evidence (test output, code snippets)
- But they also debate aesthetics (Murphy vs Ockham on simplicity)
- The verdict emerges from synthesis, not checklist completion

**The best part:** This format is MEMORABLE. Six months from now, engineers will remember:
- "DJ Skippy's signature move is the segment boundary check"
- "Murphy and Ockham fought about nil checks"
- "The Oracle cited §6.1 to settle the debate"

They won't remember "PR #1337 had good test coverage." But they WILL remember the talent show.

**Use this when code review feels like a chore and you want it to feel like storytelling instead.** 🎭
