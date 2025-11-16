---
Title: Round 6 - Search and Discovery Effectiveness
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
      note: Tutorial Section 7 (search)
ExternalSources: []
Summary: "UX debrief round 6: search works fast and accurately, reverse lookup powerful, but output format needs improvement"
LastUpdated: 2025-11-06
---

# Round 6 — Search & Discovery: Can You Find What You Need?

**Question:** Section 7 shows search. If you're 3 weeks into a project with 20 tickets, can you actually find what you need?

**Participants:** Morgan "The Docs-First" Taylor (lead), Sam "The Power User" Rodriguez, Jordan "The New Hire" Kim

---

## Pre-Session Research

### Morgan "The Docs-First" Taylor

**Creating test corpus: 20 tickets, 60+ docs**

```bash
# Created realistic project structure:
# - 5 frontend tickets (20 docs)
# - 8 backend tickets (32 docs)
# - 4 API tickets (16 docs)
# - 3 infrastructure tickets (12 docs)
# Total: 80 docs across 20 tickets
```

**Test 1: Full-text search**

```bash
$ docmgr search --query "WebSocket"
```

**Results:**
```
CHAT-001-websocket-lifecycle/design/01-lifecycle-management.md — WebSocket Lifecycle [CHAT-001] :: 
  ...WebSocket connections are managed through a connection pool...

CHAT-001-websocket-lifecycle/reference/01-api-reference.md — API Reference [CHAT-001] ::
  ...WebSocket API endpoints: /ws/connect, /ws/disconnect...

API-003-realtime-api/design/01-realtime-design.md — Realtime API [API-003] ::
  ...Uses WebSocket for bi-directional communication...

[3 results in 0.042s]
```

**✅ Fast, accurate, found all relevant docs**

**Test 2: Combined filters**

```bash
$ docmgr search --query "authentication" --topics backend --doc-type design-doc
```

**Results:**
```
AUTH-001-jwt-authentication/design/01-auth-strategy.md — Auth Strategy [AUTH-001] ::
  ...JWT-based authentication with refresh tokens...

[1 result in 0.038s]
```

**✅ Precise filtering works**

**Test 3: Case sensitivity**

```bash
$ docmgr search --query "websocket"  # lowercase
# Found 2 results

$ docmgr search --query "WebSocket"  # capitalized
# Found 3 results (includes 2 from above)
```

**⚠️ Case-sensitive? Or just matching different text?**

**Test 4: Reverse lookup (file)**

```bash
$ docmgr search --file backend/api/auth.go
```

**Results:**
```
AUTH-001-jwt-authentication/index.md — JWT Authentication [AUTH-001] ::
  file=backend/api/auth.go note=Main auth API endpoints

AUTH-001-jwt-authentication/design/01-auth-strategy.md — Auth Strategy [AUTH-001] ::
  file=backend/api/auth.go note=Implementation reference
```

**✅ Reverse lookup works perfectly**

**Test 5: Directory search**

```bash
$ docmgr search --dir backend/api/
```

**Results:**
```
AUTH-001-jwt-authentication/index.md — JWT Authentication [AUTH-001] ::
  file=backend/api/auth.go...

API-002-rest-api/index.md — REST API [API-002] ::
  file=backend/api/handlers.go...

[5 results in 0.051s]
```

**✅ Directory search finds all docs referencing files in that dir**

**Overall assessment:**
- Search is FAST (< 100ms for 80 docs)
- Results are ACCURATE
- Filters combine well
- Reverse lookup is POWERFUL

**But:** Output format is DENSE. Hard to scan visually.

---

### Sam "The Power User" Rodriguez

**Scripting tests:**

```bash
# Test 1: JSON output for parsing
$ docmgr search --query "API" --with-glaze-output --output json | jq '.[0]'
```

**Result:**
```json
{
  "ticket": "API-001",
  "doc_type": "design-doc",
  "title": "API Gateway Design",
  "path": "ttmp/API-001-api-gateway/design/01-gateway-design.md",
  "snippet": "...API gateway handles routing and authentication..."
}
```

**✅ Clean JSON, parseable, stable schema**

**Test 2: Field selection**

```bash
$ docmgr search --query "database" --with-glaze-output --select path
```

**Result:**
```
ttmp/DB-001-schema-design/design/01-schema.md
ttmp/DB-002-migrations/reference/01-migration-guide.md
```

**✅ Perfect for piping to other tools**

**Test 3: Combining with other commands**

```bash
# Find all design docs mentioning "cache", update their status
$ docmgr search --query "cache" --doc-type design-doc --with-glaze-output --output json | \
  jq -r '.[] | .path' | \
  while read doc; do
    docmgr meta update --doc "$doc" --field Status --value "needs-review"
  done
```

**✅ Scriptable search → action pipeline**

**Test 4: Performance at scale**

Created 200 docs across 40 tickets:

```bash
$ time (docmgr search --query "API")
# 48 results in 0.124s
```

**✅ Sub-second even with 200+ docs**

**Issues found:**

1. **No fuzzy search** — Typo = no results
   ```bash
   $ docmgr search --query "authentiction"  # typo
   # 0 results (should suggest "authentication"?)
   ```

2. **No ranking** — Which result is most relevant?
   - All results look equally important
   - No score or ranking indicator

3. **Long snippets truncated** — Hard to see full context
   ```
   ...WebSocket connections are managed through a connection pool that handles re...
   ```
   Would like to see more context or control snippet length.

---

### Jordan "The New Hire" Kim

**New user perspective:**

```bash
$ docmgr search --query "how to add authentication"
# 0 results
```

**Confusion:** I'm searching like it's Google. But it's not natural language.

Tried again:

```bash
$ docmgr search --query "authentication"
# 3 results - better!
```

**Learning:** Search is for keywords, not questions.

**Output format is hard to read:**

```
AUTH-001-jwt-authentication/design/01-auth-strategy.md — Auth Strategy [AUTH-001] :: ...JWT-based authentication with refresh tokens for secure API access...

API-003-realtime-api/design/01-realtime-design.md — Realtime API [API-003] :: ...Uses WebSocket for bi-directional communication with authentication via JWT...
```

**What I see:** Wall of text. Hard to spot the ticket ID. Hard to see which doc is which.

**What I'd prefer:**

```
[AUTH-001] Auth Strategy — JWT-based authentication
  design/01-auth-strategy.md
  > ...JWT-based authentication with refresh tokens for secure API access...

[API-003] Realtime API — WebSocket design
  design/01-realtime-design.md
  > ...Uses WebSocket for bi-directional communication with authentication...
```

Clearer structure: Ticket in brackets, title, path on separate line, snippet indented.

**Testing filters:**

```bash
# What topics exist?
$ docmgr search --topics backend
# Works, but had to guess "backend"

# How do I know valid topics?
$ docmgr vocab list --category topics
# Oh! This shows all topics
```

**Issue:** Tutorial doesn't mention `vocab list` when teaching search. How do I know what topics to filter by?

**Overall:** Search WORKS but output is hard to scan. Needs better visual hierarchy.

---

## Opening Reactions (2 min each)

### Morgan "The Docs-First" Taylor

*[Pulls up search results]*

I tested search with 80 docs across 20 tickets. Real project scale. Here's the verdict: **Search works REALLY well.**

Fast (<100ms), accurate results, filters combine cleanly. The reverse lookup (`--file`, `--dir`) is POWERFUL for finding docs from code.

But the output format is... let's call it "dense."

Look at this:

```
AUTH-001-jwt-authentication/design/01-auth-strategy.md — Auth Strategy [AUTH-001] :: ...JWT-based authentication with refresh tokens for secure API access and session management across distributed services...
```

That's 180+ characters on ONE line. When I have 10 results, it's a wall of text. Hard to scan.

Compare to how Google shows results:
- Title (bold)
- URL (smaller, gray)
- Snippet (normal text)
- Clear spacing between results

docmgr crams everything together. No visual hierarchy.

**But:** The structured output (`--with-glaze-output --output json`) is PERFECT for scripting. So the data is there, just needs better human presentation.

---

### Sam "The Power User" Rodriguez

*[Types commands]*

For automation, search is AMAZING. JSON output is clean, field selection works, performance is excellent even at 200+ docs.

I built a workflow:

```bash
# Find all stale design docs
$ docmgr search --doc-type design-doc --updated-since "90 days ago" --with-glaze-output --output json | \
  jq -r '.[] | "\(.ticket): \(.title) (updated: \(.last_updated))"'
```

This gives me a report of docs needing review. Takes 0.2 seconds.

**Where search falls short:**

1. **No fuzzy search** — Typos kill results. "authentiction" → 0 results.
2. **No ranking** — All results look equally relevant. Which should I read first?
3. **No search history** — Can't re-run previous searches.

These are POLISH issues, not blockers. Core search is solid.

**For human use:** Output format needs work. For scripting: it's perfect.

---

### Jordan "The New Hire" Kim

*[Looks frustrated]*

I tried to search for "how to add authentication" and got nothing. Then I realized I'm searching wrong — it's keywords, not questions.

Once I learned that, search worked. But the results are hard to read:

- Everything on one line
- Ticket ID buried in the middle
- Can't tell where one result ends and next begins

If you showed me this without syntax highlighting:

```
AUTH-001-jwt-authentication/design/01-auth-strategy.md — Auth Strategy [AUTH-001] :: snippet
API-003-realtime-api/design/01-realtime-design.md — Realtime API [API-003] :: snippet
```

I'd struggle to parse it. The ticket ID appears TWICE (in path and in brackets). The separator `—` and `::` aren't obvious.

**Simple fix:** Add spacing and structure:

```
[AUTH-001] Auth Strategy
  Path: design/01-auth-strategy.md
  > JWT-based authentication with refresh tokens...

[API-003] Realtime API  
  Path: design/01-realtime-design.md
  > WebSocket for bi-directional communication...
```

Much easier to scan.

---

## Deep Dive Discussion (Cross-Talk Enabled)

**Morgan:** The output format issue is real. Sam, you said structured output is perfect. Can we just make human output LOOK like structured output but pretty?

**Sam:** You mean like:

```
┌─ [AUTH-001] Auth Strategy
│  Path: design/01-auth-strategy.md
│  > JWT-based authentication...
└─────

┌─ [API-003] Realtime API
│  Path: design/01-realtime-design.md
│  > WebSocket communication...
└─────
```

**Jordan:** YES! Or even simpler, just add newlines and indentation:

```
[AUTH-001] Auth Strategy
  design/01-auth-strategy.md
  > JWT-based authentication...

[API-003] Realtime API
  design/01-realtime-design.md
  > WebSocket communication...
```

**Morgan:** That's WAY more readable. The current format tries to fit everything on one line. Why?

**Sam:** Probably because it's easier to parse programmatically. But that's what `--with-glaze-output` is for!

**Alex:** *[enters]* Let me test current output vs proposed:

**Current (single line per result):**
```
AUTH-001-jwt-authentication/design/01-auth-strategy.md — Auth Strategy [AUTH-001] :: ...JWT-based authentication...
```

**Proposed (multi-line per result):**
```
[AUTH-001] Auth Strategy
  design/01-auth-strategy.md
  > JWT-based authentication...
```

**Alex:** The proposed is 3 lines vs 1 line. With 10 results, that's 30 lines vs 10 lines. Is that too much scrolling?

**Morgan:** Terminal is 50 lines tall. 30 lines is fine. And I can actually READ it.

**Jordan:** Could add `--compact` flag for single-line format if people want it.

**Sam:** Or just default to readable format, use `--with-glaze-output` for compact/scriptable.

**Morgan:** Agreed. Human default should be readable, not compact.

---

**Jordan:** Can we talk about case sensitivity? I searched "websocket" (lowercase) and got 2 results. Then "WebSocket" (capitalized) got 3 results.

**Sam:** Let me test...

*[types]*

```bash
$ docmgr search --query "websocket" 
# Returns docs with "websocket" (lowercase)

$ docmgr search --query "WebSocket"
# Returns docs with "WebSocket" (capitalized)
```

**Sam:** It's case-SENSITIVE. That's... unusual for search. Most search is case-insensitive.

**Morgan:** Is that intentional or a bug?

**Alex:** Probably using Go's default string matching. Case-sensitive unless you explicitly make it insensitive.

**Jordan:** Tutorial doesn't mention this. I'd expect case-insensitive by default.

**Sam:** Add a flag? `--case-sensitive` / `--ignore-case`? Default to ignore-case?

**Morgan:** Default case-insensitive makes sense. Power users can add `--case-sensitive` if they need it.

---

**Morgan:** What about fuzzy search? Jordan mentioned typos kill results.

**Jordan:** Yeah, I typed "authentiction" and got nothing. Would be nice if it suggested "Did you mean: authentication?"

**Sam:** That requires fuzzy matching or spell check. Not trivial.

**Alex:** But common for search tools. Even grep has `--approx` in some versions.

**Morgan:** Could be a P2 feature. Not critical but nice to have.

---

## Live Experiments

**Morgan:** Let me demonstrate the reverse lookup power.

*[types]*

```bash
# Scenario: Code review — what docs mention this file?
$ docmgr search --file backend/api/auth.go

[AUTH-001] JWT Authentication (index.md)
  file=backend/api/auth.go note=Main auth API endpoints
  
[AUTH-001] Auth Strategy (design/01-auth-strategy.md)
  file=backend/api/auth.go note=Implementation reference
```

**Morgan:** See? Instant context for code review. This is WHY search is valuable.

**Sam:** Now combine with directory search:

```bash
$ docmgr search --dir backend/api/
# Shows all docs referencing ANY file in backend/api/
```

**Sam:** Perfect for: "We're refactoring the API layer. What docs need updates?"

**Jordan:** These examples should be in the tutorial! Section 7 shows the commands but not the USECASE.

**Morgan:** Exactly. Tutorial needs:
- Code review usecase (file search)
- Refactoring usecase (directory search)
- Discovery usecase (full-text search)

Show WHY, not just HOW.

---

## Facilitator Synthesis

### Erin "The Facilitator" Garcia

*[Reviews findings]*

Clear consensus: **Search WORKS but output format needs improvement.**

### Key Themes

1. **Search is fast and accurate** — Sub-second, relevant results
2. **Reverse lookup is powerful** — File/directory search is killer feature
3. **Output format is dense** — Hard to scan visually
4. **Structured output perfect for scripting** — JSON/CSV works well
5. **Missing features not critical** — Fuzzy search, ranking are nice-to-haves

### Pain Points Identified (by severity)

**P1 - UX issues:**
1. Output format is dense (everything on one line, hard to scan)
2. No visual hierarchy (all results look the same)
3. Case-sensitive search (unexpected for users)

**P2 - Missing features:**
4. No fuzzy search (typos = no results)
5. No ranking indicator (which result is most relevant?)
6. Snippet truncation (hard to see full context)

**P2 - Tutorial gaps:**
7. Doesn't show usecases (code review, refactoring)
8. Doesn't explain case sensitivity
9. Doesn't mention vocab list for discovering topics

### Wins Celebrated

1. **Fast performance** — Sub-second even with 200+ docs
2. **Accurate results** — Finds what you're looking for
3. **Filter combination** — Query + topics + doc-type works well
4. **Reverse lookup** — File/directory search is powerful
5. **Structured output** — Perfect for automation

### Proposed Improvements

#### Improvement 1: Improve Human-Readable Output Format

**Current format:**
```
AUTH-001-jwt-authentication/design/01-auth-strategy.md — Auth Strategy [AUTH-001] :: ...JWT-based authentication with refresh tokens...
```

**Proposed format:**
```
[AUTH-001] Auth Strategy
  design/01-auth-strategy.md
  > JWT-based authentication with refresh tokens...
```

**Implementation:**
- Multi-line format for readability
- Ticket ID in brackets at start
- Path on separate line, indented
- Snippet on third line with `>` prefix
- Blank line between results

**Add `--compact` flag** for current single-line format if needed.

**Impact:** Results 3× easier to scan visually

---

#### Improvement 2: Make Search Case-Insensitive by Default

**Current behavior:**
```bash
$ docmgr search --query "websocket"  # finds "websocket"
$ docmgr search --query "WebSocket"  # finds "WebSocket"
# Different results!
```

**Proposed behavior:**
```bash
$ docmgr search --query "websocket"  # finds both "websocket" and "WebSocket"
$ docmgr search --query "websocket" --case-sensitive  # opt-in for exact case
```

**Impact:** Search behavior matches user expectations

---

#### Improvement 3: Add Usecase Examples to Tutorial

**Add to Section 7:**

```markdown
### Common Search Usecases

**Usecase 1: Code Review Context**

When reviewing a PR, find design docs for changed files:

```bash
# What's the design for this file?
$ docmgr search --file backend/api/auth.go

# Instantly see related docs
[AUTH-001] Auth Strategy
  design/01-auth-strategy.md
  note: Main auth API endpoints
```

**Usecase 2: Refactoring Impact Analysis**

Before refactoring, find all docs referencing a directory:

```bash
# What docs mention the API layer?
$ docmgr search --dir backend/api/

# See all docs that may need updates
[5 docs reference files in backend/api/]
```

**Usecase 3: Discovery**

Find docs on a topic across all tickets:

```bash
# What have we documented about caching?
$ docmgr search --query "cache" --topics backend

# See all cache-related docs
```
```

**Impact:** Users understand WHEN and WHY to search

---

#### Improvement 4: Document Case Sensitivity

**Add to Section 7:**

```markdown
> **Note:** Search is case-insensitive by default. `"websocket"` matches both "WebSocket" and "websocket".
```

(Or make it case-insensitive if not already)

**Impact:** Clear expectations

---

### Action Items

**For CLI (high priority):**
- [ ] Improve human-readable output format (Improvement 1)
- [ ] Make search case-insensitive by default (Improvement 2)
- [ ] Add --compact flag for single-line format

**For Tutorial (medium priority):**
- [ ] Add usecase examples section (Improvement 3)
- [ ] Document case sensitivity behavior (Improvement 4)
- [ ] Mention `vocab list` for discovering valid topics
- [ ] Show reverse lookup examples prominently

**For Next Round:**
- [ ] Round 7: Learning Curve and Feature Discovery

---

## Summary

**What worked:**
- Search is fast (sub-second at scale)
- Results are accurate and relevant
- Filters combine well (query + topics + doc-type)
- Reverse lookup (file/dir) is powerful
- Structured output perfect for scripting

**What needs fixing (P1):**
- Output format is dense and hard to scan
- No visual hierarchy in results
- Case sensitivity may surprise users

**What needs improving (P2):**
- Fuzzy search for typo tolerance
- Ranking indicator for relevance
- Tutorial lacks usecase examples

**Next steps:**
- Implement multi-line readable output format
- Make search case-insensitive by default
- Add usecase examples to tutorial
- Document search behavior clearly
