---
Title: Jamie's commentary on Glazed documentation style guide
Ticket: DOCMGR-DOC-VERBS
Status: active
Topics:
    - docmgr
    - documentation
    - style-guide
DocType: working-note
Intent: long-term
Owners:
    - jamie-park
RelatedFiles:
    - Path: docmgr/pkg/doc/docmgr-how-to-use.md
      Note: Tutorial to evaluate against these guidelines
ExternalSources:
    - glaze help how-to-write-good-documentation-pages — Source style guide from Glazed project
Summary: "Technical writer's analysis of Glazed style guide and how docmgr tutorial measures up."
LastUpdated: 2025-11-25
---

# Jamie's Commentary on Glazed Style Guide

## Context

I reviewed `glaze help how-to-write-good-documentation-pages` to see if we can adopt these standards for docmgr documentation. This is excellent guidance—let me analyze how our tutorial measures up and what we should change.

---

## Glazed Style Guide Summary

### Core Principles (from Glazed)

1. **Clarity** — Simple, direct, unambiguous. Avoid jargon or explain it.
2. **Accuracy** — Code examples tested and current.
3. **Conciseness** — Direct and to the point. Eliminate wordiness.
4. **Completeness** — Cover thoroughly but stay on-topic.
5. **Audience-Centric** — Frame in context of user's goal and problem.

### Key Guidelines

**Structure:**
- YAML frontmatter (metadata)
- H1 matches frontmatter Title
- **H2 sections start with topic-focused paragraph** (explains concept, not just "this section covers...")
- Short paragraphs (one idea each)
- Bulleted lists for scannability

**Code Examples:**
- Minimal and focused (one concept per example)
- Comments explain WHY, not WHAT
- Runnable (copy-paste works)
- Show expected output

**Style:**
- Active voice ("The framework provides..." not "...is provided by...")
- Consistent terminology
- Professional, helpful, direct tone
- Developer audience (familiar with language, new to framework)

---

## How docmgr Tutorial Measures Up

### ✅ What We're Doing Well

**1. Clear Structure:**
- ✅ YAML frontmatter present
- ✅ H1 matches Title
- ✅ Progressive parts (Essentials → Workflows → Power User → Reference)
- ✅ Bulleted lists and code blocks

**2. Runnable Examples:**
```bash
docmgr ticket create-ticket --ticket MEN-4242 \
  --title "Normalize chat API paths" \
  --topics chat,backend
```
- ✅ Copy-paste works
- ✅ Realistic values (not FOO-123)
- ✅ Shows full commands with flags

**3. Expected Output Shown:**
```
root=/path/to/ttmp vocabulary=/path/vocabulary.yaml tickets=0 docs=0
```
- ✅ Tutorial shows what users should see
- ✅ Helps users verify success

**4. Audience-Centric:**
- ✅ "Quick Navigation" box lets users choose path based on use case
- ✅ Progressive disclosure (beginners read Part 1, power users jump to Part 3)

---

### ❌ Where We're Violating Guidelines

#### Violation 1: Section Introductions (HIGH IMPACT)

**Glazed guideline:**
> Every major H2 section must begin with a single, topic-focused paragraph that explains the core concept, not just describes what the section contains.

**How we violate it:**

**Bad example from our tutorial (Section 7):**
```markdown
## 7. Relating Files to Docs [INTERMEDIATE]

### The Workflow
**When to relate files:**
1. **During design** — Identify which code files...
```

**Problem:** Jumps straight into "When to relate" without explaining WHAT relating does or WHY it matters.

**Better (following Glazed guideline):**
```markdown
## 7. Relating Files to Docs [INTERMEDIATE]

Bidirectional linking between documentation and code is one of docmgr's most powerful features. By relating code files to docs with explanatory notes, you create a navigation map that answers two critical questions: "What's the design for this code file?" (code review context) and "Which code implements this design?" (implementation reference). The `docmgr doc relate` command manages these relationships in frontmatter, while `docmgr doc search --file` provides instant reverse lookup.

### The Workflow
[Continue with when/how...]
```

**Our tutorial DOES have some good intros:**

**Good example (Section 6):**
> Metadata (frontmatter) defines how docs are discovered, filtered, and validated. docmgr provides the `meta update` command to modify frontmatter fields programmatically, ensuring valid YAML syntax, consistent formatting, and automated timestamp updates.

**This follows the guideline!** Explains WHAT metadata is, WHY it matters, HOW docmgr handles it.

**But we're inconsistent.** Some sections have good intros (like Section 6), others jump straight into instructions (like Section 7).

**Impact:** Users get confused about PURPOSE before learning HOW.

**Fix needed:** Audit all H2 sections, add topic-focused intro paragraphs where missing.

---

#### Violation 2: Jargon Without Explanation (MEDIUM IMPACT)

**Glazed guideline:**
> Avoid jargon where possible, or explain it clearly if it's necessary.

**How we violate it:**

Terms used BEFORE defined:
- "frontmatter" (used in Section 2, defined in glossary at Section 5)
- "docs root" (used in Section 2, defined in glossary)
- "ticket workspace" (used frequently, only defined in glossary)

**Why this matters:** Beginners hit jargon, get confused, have to search for definition.

**From validation reports:**
> "I kept wondering: what's a ticket workspace? Is it different from a ticket? Confusion."

**Good news:** We HAVE a glossary (Section 5: Key Concepts).

**Bad news:** It's AFTER we use the terms in Sections 1-4.

**Fix needed:** Either:
1. Move glossary to Section 1 (before first use), OR
2. Add inline definitions: "ticket workspace (directory containing all docs related to a ticket)"

**Glazed approach:** Define at first use.

---

#### Violation 3: Wordiness (MEDIUM IMPACT)

**Glazed guideline:**
> Be direct and to the point. Eliminate wordiness and focus on delivering information efficiently.

**How we violate it:**

**Wordy example (Section 14):**
```markdown
Bidirectional linking between documentation and code is one of docmgr's most powerful features. By relating code files to docs with explanatory notes, you create a navigation map that answers two critical questions: "What's the design for this code file?" (code review context) and "Which code implements this design?" (implementation reference). The `docmgr doc relate` command manages these relationships in frontmatter, while `docmgr doc search --file` provides instant reverse lookup from any code file to its related documentation.
```

**Word count:** 74 words (one sentence!)

**Glazed recommendation:** Break into shorter sentences.

**Better:**
```markdown
Bidirectional linking connects documentation to code. When you relate files to docs with notes, you create a navigation map. During code review, ask: "What's the design for this file?" Run `docmgr doc search --file path` to find the design doc. During implementation, ask: "Which code implements this design?" Check the doc's RelatedFiles.
```

**Word count:** 56 words, 5 shorter sentences. More scannable.

**Impact:** Long sentences slow comprehension, especially for non-native English speakers.

**Fix needed:** Audit long paragraphs (>60 words), break into shorter sentences.

---

#### Violation 4: Code Comments Explain "What" Not "Why" (LOW IMPACT)

**Glazed guideline:**
> Use comments to explain the why, not the what.
> - Bad: `// Create a new row`
> - Good: `// Use types.MRP to ensure type-safe key-value pairs`

**Our tutorial uses bash commands, not Go code, so this is less relevant.**

But we CAN apply the principle:

**Current (explains what):**
```bash
# Full-text search
docmgr doc search --query "WebSocket"

# Filter by metadata
docmgr doc search --query "API" --topics backend
```

**Better (explains why):**
```bash
# Find all docs mentioning WebSocket (full-text search)
docmgr doc search --query "WebSocket"

# Narrow search to backend API docs only (metadata filtering)
docmgr doc search --query "API" --topics backend --doc-type design-doc
```

**Impact:** Helps users understand WHEN to use each pattern.

**Fix needed:** Review all code comments, rewrite to explain WHY/WHEN, not just WHAT.

---

#### Violation 5: Examples Not Always Minimal (LOW IMPACT)

**Glazed guideline:**
> Keep examples minimal and focused. Remove all boilerplate or irrelevant logic.

**Our tutorial mostly does this well.** But some examples are complex:

**Complex example (Section 11):**
```bash
docmgr doc search --updated-since "60 days ago" --with-glaze-output --output json | \
  jq -r '.[] | .path' | \
  while read doc; do
    docmgr meta update --doc "$doc" --field Status --value "needs-review"
  done
```

**Issue:** This teaches THREE concepts:
1. Search with time filters
2. Glaze JSON output
3. Shell scripting loop
4. jq parsing

**Glazed recommendation:** One concept per example.

**Better (break into steps):**
```bash
# Step 1: Find stale docs (JSON output)
docmgr doc search --updated-since "60 days ago" \
  --with-glaze-output --output json > stale-docs.json

# Step 2: Extract paths with jq
cat stale-docs.json | jq -r '.[] | .path' > paths.txt

# Step 3: Update each doc
while read doc; do
  docmgr meta update --doc "$doc" --field Status --value "needs-review"
done < paths.txt
```

**Impact:** Easier to understand, users can test each step.

**Fix needed:** Break complex examples into step-by-step.

---

## How Glazed Guidelines Apply to docmgr Tutorial

### Principle 1: Clarity

**Current status: MOSTLY GOOD**

Strengths:
- Commands explained clearly
- Progressive structure
- Examples are concrete

**Needs work:**
- Jargon used before defined (move glossary earlier)
- Some sections lack topic-focused intros
- Long sentences in places

---

### Principle 2: Accuracy

**Current status: CRITICAL ISSUES**

This is WHY we're having the debate!

Problems:
- Wrong command syntax (docmgr relate)
- Removed flags (--files)
- Path inconsistencies (design/ vs design-doc/)

**Action:** Round 3 triage already prioritized these as CRITICAL.

---

### Principle 3: Conciseness

**Current status: NEEDS WORK**

Problems:
- Some paragraphs >70 words (hard to scan)
- Duplicate sections (changelog appears 3x)
- Could trim without losing information

**Action:** 
- Round 2 decided: consolidate duplicates in Phase 2
- Add to Phase 2: Audit for wordiness

---

### Principle 4: Completeness

**Current status: GAPS FOUND**

Missing:
- Task editing commands
- Shell completion
- Command aliasing
- Import workflows
- Maintenance commands

**Action:** Round 4 identified these, prioritized additions.

---

### Principle 5: Audience-Centric

**Current status: GOOD**

Strengths:
- "Quick Navigation" box (choose your path)
- Progressive disclosure (beginner → power user)
- Use cases explained ("During code review...")
- Real-world examples (not toy examples)

**Could improve:**
- Add "When to use this" for each command
- More troubleshooting (when things go wrong)

---

## Recommended Actions Based on Glazed Guidelines

### HIGH Priority (Phase 1):

**1. Add topic-focused intros to all H2 sections:**

Audit sections, add concept-explaining paragraphs where missing.

Example template:
```markdown
## N. [Section Title]

[1-2 sentences explaining WHAT this is and WHY it matters, 
followed by 1 sentence on HOW docmgr handles it.]

### [Subsection with instructions]
```

**Estimated effort:** 30-45 minutes to review all sections  
**Impact:** HIGH — helps users understand PURPOSE before learning mechanics

---

**2. Move glossary earlier (before first jargon use):**

Current: Section 5 (Key Concepts)  
Better: Section 1 or 2 (before "ticket workspace" appears)

**Or:** Add inline definitions at first use:
```markdown
Create your first **ticket** (a unit of work with an identifier like MEN-4242) 
in a **ticket workspace** (directory containing all docs for that ticket).
```

**Estimated effort:** 20 minutes  
**Impact:** HIGH — prevents confusion from undefined jargon

---

### MEDIUM Priority (Phase 2):

**3. Break long sentences (>50 words):**

Audit and split into multiple shorter sentences.

**Estimated effort:** 1 hour  
**Impact:** MEDIUM — improves scannability

---

**4. Improve code comments (explain WHY/WHEN):**

Audit all bash comments, rewrite to explain use case not just syntax.

**Estimated effort:** 30 minutes  
**Impact:** MEDIUM — helps users choose right command pattern

---

**5. Break complex examples into steps:**

Identify multi-concept examples (like the jq pipeline), split into step-by-step.

**Estimated effort:** 45 minutes  
**Impact:** MEDIUM — easier to learn and test

---

## Proposed docmgr Style Guide (Based on Glazed)

### Adapted Principles

**1. Clarity**
- Active voice ("Run docmgr init" not "docmgr init should be run")
- Define jargon at first use or link to glossary
- Short sentences (<50 words)

**2. Accuracy**
- All commands tested in fresh environment
- Output shown matches reality
- Command syntax matches current CLI version
- Updated when CLI changes (see Round 12 prevention strategy)

**3. Conciseness**
- No duplicate sections (single source of truth)
- Trim wordiness without losing context
- Use tables for comparisons
- Use bullets for lists

**4. Completeness**
- Cover entire workflow (init → create → ... → close)
- Document all everyday commands
- Mention advanced commands (with pointer to --help)
- Include troubleshooting

**5. Audience-Centric**
- Frame as user goals ("Find docs about X", "Track progress on ticket")
- Explain WHEN to use each command
- Progressive disclosure (beginner → advanced)
- Real-world examples (not toy data)

---

### Structure Guidelines

**Every H2 section MUST have:**

1. **Topic-focused intro paragraph** (1-3 sentences):
   - WHAT: What is this concept?
   - WHY: Why does it matter?
   - HOW: How does docmgr handle it?

2. **Subsections with instructions** (H3):
   - Clear headings
   - Step-by-step when needed
   - Examples with expected output

3. **Navigation aids:**
   - "See also" links to related sections
   - Cross-references to deeper coverage
   - Forward pointers ("We'll cover X in Section Y")

**Example template:**
```markdown
## 7. Relating Files to Docs [INTERMEDIATE]

Bidirectional linking connects documentation to code. By relating files with notes, you create a navigation map: from code to design docs (code review) and from docs to implementation (reference). This is docmgr's most powerful feature for maintaining context between code and documentation.

### Basic Usage

[Commands and examples]

### When to Relate Files

[Guidance on timing]

### Writing Good Notes

[How to write useful notes]
```

---

### Code Example Guidelines

**Every bash example should:**

1. **Be copy-paste runnable** (with realistic values):
   ```bash
   # ✅ Good
   docmgr ticket create-ticket --ticket MEN-4242 --title "..."
   
   # ❌ Bad
   docmgr ticket create-ticket --ticket YOUR-TICKET --title YOUR-TITLE
   ```

2. **Explain use case in comment** (WHY/WHEN, not WHAT):
   ```bash
   # ✅ Good: Find design context during code review
   docmgr doc search --file backend/api/register.go
   
   # ❌ Bad: Search for file
   docmgr doc search --file backend/api/register.go
   ```

3. **Show expected output** (so users can verify):
   ```bash
   docmgr status --summary-only
   # Expected: root=/path/ttmp tickets=1 docs=3
   ```

4. **One concept per example** (don't chain unrelated operations):
   ```bash
   # ✅ Good: Focus on one thing
   docmgr doc search --query "API" --topics backend
   
   # ❌ Bad: Too many concepts
   docmgr doc search --query "API" --topics backend --with-glaze-output --output json | jq '.[] | .path' | xargs -I{} docmgr meta update --doc {} --field Status --value review
   ```

---

### Terminology Consistency

**From validation reports, we're inconsistent on:**

| Inconsistent Usage | Correct Form | Rule |
|-------------------|--------------|------|
| "related files" vs "RelatedFiles" | "RelatedFiles" (field), "related files" (concept) | Capitalize field names, lowercase concepts |
| "Ticket" vs "ticket" | "Ticket" (field), "ticket" (concept) | Same rule |
| "docs root" vs "documentation root" vs "ttmp/" | "docs root" | Use shortest clear form |
| "frontmatter" vs "front-matter" vs "front matter" | "frontmatter" | One word, no hyphen |
| "doc-type" vs "DocType" vs "document type" | "DocType" (field), "doc-type" (CLI flag), "document type" (concept) | Context-dependent |

**Fix needed:** Create terminology table in style guide.

---

## How Tutorial Violates Glazed Guidelines (Ranked)

### HIGH Impact Violations (Fix in Phase 1)

1. **Missing topic-focused intros** (multiple sections)
   - Estimate: 5-8 sections need better intros
   - Effort: 45 minutes
   - Impact: Helps users understand PURPOSE

2. **Jargon used before defined**
   - Glossary appears AFTER jargon used
   - Effort: 20 minutes (move glossary or add inline definitions)
   - Impact: Prevents confusion

3. **Accuracy bugs** (wrong commands, flags, paths)
   - Already prioritized as CRITICAL in Round 3
   - Effort: 90 minutes
   - Impact: Makes tutorial CORRECT

---

### MEDIUM Impact Violations (Fix in Phase 2)

4. **Inconsistent terminology**
   - Need style guide table
   - Effort: 30 minutes to create guide + 1 hour to fix inconsistencies
   - Impact: Professionalism, reduces confusion

5. **Long sentences** (>50 words)
   - Several sections have overly complex sentences
   - Effort: 1 hour to audit and split
   - Impact: Scannability

6. **Complex examples** (multi-concept)
   - Some examples teach 3-4 things at once
   - Effort: 45 minutes to break into steps
   - Impact: Learning curve

---

### LOW Impact Violations (Phase 2 or backlog)

7. **Code comments explain "what" not "why"**
   - Not critical but could improve
   - Effort: 30 minutes
   - Impact: Marginal (bash is self-explanatory)

---

## Comparison: Glazed vs docmgr Documentation Philosophy

### Where We Agree

Both prioritize:
- ✅ Runnable examples
- ✅ Active voice
- ✅ Audience-centric framing
- ✅ Clear structure with sections
- ✅ Expected output shown

---

### Where We Differ (Intentionally)

**Glazed emphasis:**
- API reference for developers building with Glazed (library)
- Go code examples (programming interfaces)
- Framework concepts (layers, processors, middlewares)

**docmgr emphasis:**
- Tutorial for users using docmgr (CLI tool)
- Bash command examples (workflow)
- Ticket-driven documentation workflow

**These differences are appropriate!** Different tools, different audiences.

---

### Where We Differ (Unintentionally — Should Fix)

**Glazed:**
- **Consistent:** Topic-focused intro on EVERY H2
- **Concise:** No duplicate sections
- **Structured:** Clear progression

**docmgr tutorial:**
- **Inconsistent:** Some H2s have intros, others don't
- **Duplicative:** Changelog section appears 3x
- **Sprawling:** Part 2 is 480 lines (33% of tutorial)

**These are fixable!** We should adopt Glazed's discipline.

---

## Recommendations for docmgr Style Guide

### Adopt from Glazed (Verbatim):

1. **H2 sections start with topic-focused paragraph** — Explain concept, not just describe contents
2. **Consistent terminology** — Create table, enforce in reviews
3. **Active voice** — Already doing this mostly
4. **Minimal examples** — One concept per example
5. **Show expected output** — Already doing this

---

### Adapt from Glazed (for CLI context):

1. **Code comments explain use case:**
   - Glazed: "Why this Go pattern?"
   - docmgr: "When to use this flag pattern?"

2. **Runnable examples:**
   - Glazed: Go code that compiles
   - docmgr: Bash commands with realistic values (MEN-4242, not FOO)

3. **Frontmatter:**
   - Glazed: Uses Slug, Topics, IsTemplate
   - docmgr: Same, we're consistent here ✅

---

### Add to docmgr Style Guide (Not in Glazed):

1. **Ticket ID format:** Use realistic IDs (MEN-4242, PROJ-001) not placeholders (YOUR-TICKET)

2. **Path examples:** Use realistic paths (backend/api/register.go) and note they're examples:
   > Note: File paths like `backend/api/register.go` are placeholders. Substitute your actual paths.

3. **Multi-line commands:** Use backslash continuation with clear indentation:
   ```bash
   docmgr doc relate --ticket MEN-4242 \
     --file-note "path1:note1" \
     --file-note "path2:note2"
   ```

4. **Output formatting:** Use code blocks for expected output:
   ````markdown
   ```bash
   docmgr status --summary-only
   ```
   
   Expected output:
   ```
   root=/path/ttmp tickets=1 docs=3
   ```
   ````

---

## Action Items for docmgr Tutorial

### Phase 1 (High Impact, Quick Fixes):

**From Glazed guidelines:**
- [ ] Add topic-focused intros to sections lacking them (45 min)
- [ ] Move glossary earlier OR add inline definitions (20 min)
- [ ] Fix accuracy bugs (already in Round 3 CRITICAL list)

**Estimated addition: +10-15 lines (intros), reorganize glossary**

---

### Phase 2 (Polish):

**From Glazed guidelines:**
- [ ] Create terminology consistency table (30 min)
- [ ] Break long sentences >50 words (1 hour)
- [ ] Improve code comments (explain use cases) (30 min)
- [ ] Break complex multi-concept examples into steps (45 min)

**Estimated: 2.5 hours**

---

## Final Assessment

**The Glazed style guide is EXCELLENT** and we should adopt most of it.

**Our tutorial follows ~70% of the guidelines:**
- ✅ Structure, examples, output, audience
- ❌ Topic-focused intros, jargon handling, conciseness

**Quick wins:**
1. Add topic-focused intros (45 min, HIGH impact)
2. Move glossary earlier (20 min, HIGH impact)
3. Fix terminology consistency (1 hour, MEDIUM impact)

**Total: 2 hours of work to dramatically improve adherence to professional documentation standards.**

---

## Proposal for DOCMGR-DOC-VERBS

**Add to Phase 1 checklist:**
- [ ] Audit all H2 sections, add topic-focused intros where missing
- [ ] Move glossary to Section 2 (before first jargon use)
- [ ] Create terminology consistency table in style guide

**Add to Phase 2 checklist:**
- [ ] Audit long sentences, break into shorter
- [ ] Review code comments, explain use cases
- [ ] Break complex examples into steps

**Rationale:** These are professional standards from an adjacent project (Glazed). Adopting them improves tutorial quality beyond just fixing accuracy bugs.

---

## Meta Note

I'm impressed by the Glazed style guide. It's concise, actionable, and includes great before/after examples.

We should:
1. Link to it in our style guide: "See also: glaze help how-to-write-good-documentation-pages"
2. Adopt its principles wholesale (with CLI adaptations)
3. Use it as inspiration for our own examples

**The "topic-focused intro" guideline alone would fix several validation complaints.**

Good documentation principles are universal. Let's use them.

