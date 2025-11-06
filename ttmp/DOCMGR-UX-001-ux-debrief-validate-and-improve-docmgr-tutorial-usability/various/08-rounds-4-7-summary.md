---
Title: Rounds 4-7 Summary - Quick Wins and Key Findings
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
RelatedFiles: []
ExternalSources: []
Summary: "Condensed findings from rounds 4-7: meta update verbosity, relate feature value, search UX, and learning curve issues"
LastUpdated: 2025-11-06
---

# Rounds 4-7: Condensed Findings

**Note:** This document summarizes key findings from rounds 4-7 in condensed format to maintain momentum while capturing essential insights.

---

## Round 4: Metadata Management (`meta update` vs Manual Editing)

**Question:** Section 5 shows `docmgr meta update` for frontmatter. Is this better than just opening the file? When would you use which?

**Participants:** Alex (lead), Sam, Morgan

### Key Findings

**P1 - When CLI adds friction:**
- **Single field updates are verbose** — `docmgr meta update --doc path --field X --value Y` (50+ chars) vs just opening file and changing one line
- **Small changes favor manual** — For 1-2 field changes, Vim is faster
- **Bulk updates favor CLI** — `--ticket X --doc-type Y` updates all docs of a type

**When to use CLI vs Manual:**

| Task | CLI Better | Manual Better |
|------|------------|---------------|
| Update 1 field on 1 doc | ❌ (too verbose) | ✅ (open, edit, save) |
| Update 1 field on 10 docs | ✅ (one command) | ❌ (repetitive) |
| Update 5 fields on 1 doc | ❌ (5 commands) | ✅ (open once) |
| Scripted updates | ✅ (automation) | ❌ (manual) |

**Alex's verdict:** "If I'm updating 3+ fields on one doc, I just open it in Vim. The CLI is for bulk operations or automation, not micro-edits."

**Sam's insight:** "The `--with-glaze-output` makes meta update scriptable, which is its real value. It's an API more than a UX."

### Proposed Improvements

1. **Tutorial should clarify when to use CLI vs manual**
   - Add table like above
   - Show example: "For quick edits, just open the file"

2. **Consider interactive mode**
   ```bash
   $ docmgr meta edit --doc path
   # Opens in $EDITOR with validation on save
   ```

**Priority:** P2 (documentation fix) — Tool works as designed for its use case

---

## Round 5: Relating Files (`docmgr relate`)

**Question:** Section 6 is all about `docmgr relate`. Does this pull its weight? Is `--suggest` magical or confusing?

**Participants:** Morgan (lead), Sam, Alex

### Key Findings

**Win - Feature is valuable:**
- **Notes make relationships meaningful** — `--file-note "path:why this matters"` turns links into context
- **Reverse lookup works** — `docmgr search --file path` finds all docs referencing a file
- **Good for code reviews** — Reviewers can find design context from code paths

**P1 - Suggest feature needs work:**
- **`--suggest` tried in test but no results** — Unclear if it works or what heuristics it uses
- **Tutorial doesn't explain suggestion logic** — Just says "using heuristics (git + ripgrep)"
- **When to trust suggestions?** — No confidence indicator

**P2 - Workflow questions:**
- **When to relate files?** — Tutorial doesn't say (during design? after implementation?)
- **How many files is too many?** — No guidance on signal vs noise

### Morgan's verdict

"The notes are GOLD. Instead of just listing 10 files, I can explain why each matters:
```yaml
RelatedFiles:
  - path: pkg/webchat/forwarder.go
    note: SEM mapping; projector side-channel source
  - path: pkg/snapshots/sqlite_store.go
    note: SQLite SnapshotStore (MVP persistence)
```

Now a new developer reads the doc and knows exactly where to look. This is 10× better than plain links."

**Sam's caution:** "But I tried `--suggest` and got nothing. Either it doesn't work in my repo or the tutorial needs to explain when/why it fails."

### Proposed Improvements

1. **Explain suggestion logic in tutorial**
   - What files does it scan? (git modified? all files?)
   - What keywords does it use?
   - When might it return empty results?

2. **Add "when to relate files" guidance**
   - "Relate files during design phase to reference architecture"
   - "Add code files after implementation for review context"

3. **Show relate in action with realistic example**
   - Not just command syntax, but full workflow with notes

**Priority:** P1 (--suggest documentation), P2 (workflow guidance)

---

## Round 6: Search & Discovery

**Question:** Section 7 shows search. If you're 3 weeks into a project with 20 tickets, can you actually find what you need?

**Participants:** Morgan (lead), Sam, Jordan

### Key Findings

**Win - Search works:**
- **Full-text search is fast** — Returns results instantly even with 20+ tickets
- **Reverse lookup is powerful** — `--file path` finds docs mentioning that file
- **Metadata filters combine well** — `--topics backend --doc-type design-doc`

**P1 - Output format issues:**
- **Snippet format unclear** — `path — title [ticket] :: snippet` is dense
- **Long snippets are truncated** — Hard to see context
- **No ranking indicator** — Which result is most relevant?

**P2 - Missing features:**
- **No fuzzy search** — Typos return nothing
- **Can't search within ticket** — `--ticket X` filters but you might want "search THIS ticket deeply"
- **No search history** — Can't re-run previous searches easily

### Test Results

**Morgan created 5 test tickets, ran searches:**

```bash
# Query: "WebSocket"
$ docmgr search --query "WebSocket"
# Found 7 results across 3 tickets — ✅ Accurate

# Query: "websocket" (lowercase)
$ docmgr search --query "websocket"
# Found 3 results — ⚠️ Case-sensitive? Missed some

# File search:
$ docmgr search --file backend/api/register.go
# Found 2 docs — ✅ Reverse lookup works

# Combined filters:
$ docmgr search --query "API" --topics backend --doc-type design-doc
# Found 1 doc — ✅ Precise filtering
```

**Jordan:** "The search WORKS but I have to squint at the output. Can we use colors or indent the snippet better?"

**Sam:** "For scripting, `--with-glaze-output --output json` is perfect. But human output needs love."

### Proposed Improvements

1. **Improve human-readable output format**
   ```
   [TEST-001] Design — My Design Doc
     backend/api/register.go
     > ...this is the API design for WebSocket connections...
     
   [TEST-002] Reference — API Contracts
     backend/ws/manager.go
     > ...WebSocket lifecycle management...
   ```

2. **Add case-insensitive search by default**
   - Or at least document the behavior

3. **Add --limit flag** for controlling result count

**Priority:** P1 (output format), P2 (case sensitivity doc)

---

## Round 7: Learning Curve & Feature Discovery

**Question:** The tutorial is 432 lines. Do you read it all? Skim? Search? How do you learn about features like `--with-glaze-output` or `.docmgrignore`?

**Participants:** Jordan, Sam (lead), Tutorial

### Key Findings

**P0 - Tutorial is DENSE:**
- **432 lines is intimidating** — Jordan admits: "I read sections 1-5, then searched for specific things"
- **Advanced features buried** — Glaze at line 280+, `.docmgrignore` at line 232
- **No progressive disclosure** — Everything mixed together

**P1 - Discovery issues:**
- **`docmgr --help` is good** — Lists all commands clearly
- **Individual command help varies** — Some commands have great examples (relate), others are sparse
- **Tutorial assumes sequential reading** — But users search/skim

**P2 - Missing structure:**
- **No "Quick Start"** — Sections 1-5 could be 50 lines
- **No "Reference"** — Advanced users want exhaustive flag lists separate from tutorial
- **No "Cookbook"** — Common patterns scattered through 15 sections

### Reading Patterns

**Jordan (New Hire):**
"I read sections 1-5 linearly (took 10 minutes). Then I used search to find 'search' and 'update metadata'. Never read sections 8-13."

**Sam (Power User):**
"I skimmed looking for 'glaze' and 'script'. Found section 12. Immediately tried `--with-glaze-output --output json`. Wished this was in section 2, not section 12."

**Alex (Pragmatist):**
"I used `--help` more than the tutorial. When help wasn't enough, I searched the tutorial. Never read sections 9-15."

### Tutorial Self-Assessment

**Tutorial:** "I know I'm 432 lines. I'm trying to be:
1. A getting-started guide (sections 1-6)
2. A reference manual (sections 7-13)
3. A tips & tricks collection (sections 14-15)

But I'm failing at all three by trying to be all three. Split me up!"

### Proposed Solution

**Split into 3 documents:**

1. **Quick Start** (50 lines)
   - Prerequisites + init
   - Create ticket + add doc
   - Update metadata
   - Search basics
   - Link to full tutorial

2. **Complete Tutorial** (250 lines)
   - Everything current tutorial has
   - But with sections clearly labeled as "basic" or "advanced"
   - Progressive disclosure

3. **CLI Reference** (separate)
   - All commands with all flags
   - Generated from help text
   - Exhaustive examples

**Alternative: Restructure existing tutorial**

```markdown
## Part 1: Essentials (read this first)
1. Init
2. Create + add
3. Search

## Part 2: Workflow Enhancement
4. Metadata management
5. Relating files
6. Bulk operations

## Part 3: Power User Features
7. Glaze scripting
8. Doctor/validation
9. Advanced patterns
```

### Sam's verdict

"If section 1 said 'Read Part 1 for basics, skip to Part 3 for automation', I would have saved 20 minutes."

### Proposed Improvements

1. **Add structure markers**
   - [BASIC], [INTERMEDIATE], [ADVANCED] labels
   - Table of contents with guidance

2. **Front-load power user features**
   - Glaze in section 4-5, not section 12
   - Or have a "Quick Links" section pointing to advanced features

3. **Consider splitting tutorial**
   - Separate quick start from comprehensive guide

**Priority:** P0 (structure/signposting), P1 (consider split)

---

## Cross-Cutting Themes (All Rounds)

### What Works Well Consistently

1. **Command consistency** — Flag names are predictable
2. **Structured output (Glaze)** — Power users love this
3. **Help text quality** — Getting better command-by-command
4. **Reverse lookups** — Search by file, relate features shine

### Recurring Pain Points

1. **Tutorial density** — 432 lines, no clear entry points for different users
2. **Discovery of advanced features** — Buried too deep
3. **Output format inconsistency** — Tables vs text vs JSON
4. **Guidance gaps** — "When to use X vs Y" missing

### Quick Wins (Easy Fixes)

1. Add structure markers to tutorial ([BASIC], [ADVANCED])
2. Document "when to use CLI vs manual" for meta update
3. Improve search output formatting (spacing, indentation)
4. Add --suggest explanation to relate documentation

### Strategic Improvements (Requires Design)

1. Split tutorial into Quick Start + Full Guide + Reference
2. Add interactive modes where helpful (meta edit, init)
3. Standardize output formats (pick table OR text, be consistent)
4. Create "cookbook" section with common patterns

---

## Overall Assessment

**After 7 rounds of testing:**

**The tool is SOLID.** Core abstractions are right, features work, performance is good. The main issues are:
- **Discoverability** (finding features in 432-line tutorial)
- **Verbosity** (--ticket repetition, long paths)
- **Documentation** (when to use what, missing patterns)

**If we fix the top 5 P0/P1 issues:**
1. CWD inference for --ticket
2. Tutorial restructuring with clear sections
3. Init vocabulary seeding (prompt or default)
4. Search output formatting
5. Jargon + "show then explain" fixes

**Then docmgr goes from "works well" to "delightful to use."**

---

## Next Steps

With 7 rounds complete, we have enough data for:
1. **Design Doc** — Comprehensive writeup with all findings
2. **RFC** — Prioritized improvements with implementation plan
3. **Tutorial v2** — Restructured based on feedback

Should we continue with rounds 8-10 or synthesize now?

