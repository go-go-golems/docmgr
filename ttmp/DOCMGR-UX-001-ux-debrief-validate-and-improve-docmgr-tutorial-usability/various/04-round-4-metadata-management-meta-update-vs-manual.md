---
Title: Round 4 - Metadata Management meta update vs Manual
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
      note: Tutorial Section 5 (meta update)
ExternalSources: []
Summary: "UX debrief round 4: meta update vs manual editing — CLI is verbose for single-field updates, excels at bulk operations"
LastUpdated: 2025-11-06
---

# Round 4 — Metadata Management: `docmgr meta update` vs Manual Editing

**Question:** Section 5 shows `docmgr meta update` for frontmatter. Is this better than just opening the file? When would you use which?

**Participants:** Alex "The Pragmatist" Chen (lead), Sam "The Power User" Rodriguez, Morgan "The Docs-First" Taylor

---

## Pre-Session Research

### Alex "The Pragmatist" Chen

**Testing both approaches:**

**Scenario 1: Update one field on one doc**

```bash
# CLI approach
$ docmgr meta update --doc ttmp/T-001-test/design/01-my-design.md --field Summary --value "Updated summary"
# Character count: 97 characters
# Time: 0.056s

# Manual approach
# Open file in editor
# Edit Summary line in frontmatter, save
# Character count: Just editing one line
# Time: ~5 seconds (open, find line, edit, save)
```

**Scenario 2: Update 3 fields on one doc (Summary, Status, Owners)**

```bash
# CLI approach (3 commands)
$ docmgr meta update --doc ttmp/T-001-test/design/01-my-design.md --field Summary --value "New summary"
$ docmgr meta update --doc ttmp/T-001-test/design/01-my-design.md --field Status --value review
$ docmgr meta update --doc ttmp/T-001-test/design/01-my-design.md --field Owners --value "alex,manuel"
# Total: 3 commands, ~300 characters
# Time: 0.168s (3 × 0.056s)

# Manual approach
# Open file in editor
# Edit 3 lines in frontmatter, save
# Time: ~6 seconds (same open cost, slightly more editing)
```

**Scenario 3: Update one field on ALL design-docs in a ticket**

```bash
# CLI approach
$ docmgr meta update --ticket T-001 --doc-type design-doc --field Status --value review
# One command, updates 5 docs
# Time: 0.112s

# Manual approach
# Open each file in editor, edit Status line, save
# Time: 30-60 seconds for 5 docs (repetitive)
```

**Analysis:**

| Scenario | CLI Better? | Manual Better? | Winner |
|----------|-------------|----------------|--------|
| 1 field, 1 doc | ❌ (97 chars, verbose) | ✅ (faster to just edit) | **Manual** |
| 3 fields, 1 doc | ❌ (3 commands, repetitive path) | ✅ (open once, edit 3 lines) | **Manual** |
| 1 field, 5+ docs | ✅ (one command) | ❌ (repetitive) | **CLI** |
| Scripted/CI | ✅ (automatable) | ❌ (manual) | **CLI** |

**Verdict:** The CLI is for BULK operations or AUTOMATION. For quick single-doc edits, just open the damn file.

**Time measurement:**
- CLI overhead per command: ~0.05s
- Manual edit overhead: ~5s (once), then ~1s per additional field
- Break-even point: ~3 docs

---

### Sam "The Power User" Rodriguez

**Scripting perspective:**

```bash
# Scenario: Update all stale docs to trigger review
$ docmgr search --updated-since "30 days ago" --with-glaze-output --output json | \
  jq -r '.[] | .path' | \
  while read doc; do
    docmgr meta update --doc "$doc" --field Status --value review
  done

# This is POWERFUL for automation
```

**What meta update is REALLY for:**

1. **Bulk operations** — Update 10+ docs at once
2. **Validation** — CLI can validate field values (unknown topics warn in doctor)
3. **Automation** — Scripts, CI, cron jobs
4. **Consistency** — Ensures proper YAML syntax

**What it's NOT for:**

1. **Quick edits** — Opening file is faster
2. **Complex changes** — Multiple fields on one doc
3. **Exploratory work** — When you're not sure what to change

**Power user patterns:**

```bash
# Pattern 1: Bulk status update
$ docmgr meta update --ticket MEN-4242 --doc-type design-doc --field Status --value complete

# Pattern 2: Sync owners from ticket to all docs
$ OWNERS=$(docmgr list tickets --ticket MEN-4242 --with-glaze-output --output json | jq -r '.[0].owners')
$ docmgr meta update --ticket MEN-4242 --field Owners --value "$OWNERS"

# Pattern 3: Add topic to all docs in directory
$ for doc in ttmp/MEN-4242-*/design/*.md; do
    docmgr meta update --doc "$doc" --field Topics --value "backend,api"
  done
```

**The problem:**

```bash
# This is TOO VERBOSE for a single-field update:
$ docmgr meta update --doc ttmp/MEN-4242-normalize-chat-api/design/01-path-normalization.md --field Summary --value "Path normalization strategy for chat API"
# 168 characters!

# vs just:
# Open the file in your editor, edit one line, done
```

---

### Morgan "The Docs-First" Taylor

**Team workflow perspective:**

**When I use CLI:**
- End of sprint: `docmgr meta update --ticket PROJ-042 --doc-type design-doc --field Status --value complete`
- Reorganization: `docmgr meta update --ticket PROJ-042 --field Topics --value "backend,database,migration"`
- Standards enforcement: Update all docs to add missing owners

**When I use manual:**
- Writing a doc: Edit frontmatter as I write
- Quick fixes: Typo in summary, update status on one doc
- Exploratory: Not sure what fields need changing

**The hybrid approach:**

```bash
# Use CLI to find docs needing updates
$ docmgr search --query "authentication" --topics backend --doc-type design-doc

# Then manually edit the ones that need complex changes
# Use CLI for bulk status changes
```

**Team observation:**

Out of 20 developers on my team:
- 15 use manual editing exclusively (don't know about `meta update`)
- 3 use CLI for bulk operations
- 2 (including me) use hybrid approach

**The CLI isn't DISCOVERABLE.** Tutorial shows it but doesn't explain WHEN to use it.

---

## Opening Reactions (2 min each)

### Alex "The Pragmatist" Chen

*[Pulls up timing spreadsheet]*

I ran the tests. CLI vs manual. Here's the truth: **For single-doc, single-field updates, the CLI is overkill.**

Look at this command:
```
docmgr meta update --doc ttmp/T-001-test/design/01-my-design.md --field Summary --value "New summary"
```

That's 97 characters. I have to type the full path, the field name, the value. And if I make a typo in the path? Error. Start over.

Compare to:
```
# Just open the file, find Summary: line, change it, save
```

Manual editing is **10× faster** for this use case.

BUT — and this is a big but — when I need to update 5 docs? 10 docs? The CLI is a LIFESAVER. One command:
```
docmgr meta update --ticket T-001 --doc-type design-doc --field Status --value review
```

Boom. 5 docs updated in 0.1 seconds.

**My verdict:** The CLI is for automation and bulk ops. Tutorial should SAY that.

---

### Sam "The Power User" Rodriguez

*[Leans forward]*

Okay, Alex makes good points about single-doc edits. But you're missing the REAL value of `meta update`: **it's an API, not a UX.**

I don't type those commands manually. I SCRIPT them. Check this:

```bash
# Update all docs older than 30 days
docmgr search --updated-since "30 days ago" --with-glaze-output --output json | \
  jq -r '.[] | .path' | \
  while read doc; do
    docmgr meta update --doc "$doc" --field Status --value "needs-review"
  done
```

Can you do THAT with manual editing? No. The CLI is for AUTOMATION.

But here's where I agree with Alex: **the tutorial presents meta update as a primary workflow** when it should be presented as an automation tool.

Section 5 says "Enrich Metadata" and shows `meta update` commands. It doesn't say "this is for bulk operations" or "use this in scripts." New users think "oh I need to use this command for every metadata change" and it becomes tedious.

**Fix:** Tutorial should have two sections:
1. "Quick Metadata Updates" → "Just edit the file in your editor"
2. "Bulk Metadata Operations" → "Use `meta update` for automation"

---

### Morgan "The Docs-First" Taylor

*[Nods in agreement]*

Both of you are right. And here's the team perspective: **most developers don't even know `meta update` exists.**

I surveyed my team. 15 out of 20 developers manually edit frontmatter. They don't know the CLI option. And you know what? **That's fine!** Because for their use case (editing the doc they're writing), manual is better.

The 3 developers who DO use `meta update`? They're the ones doing bulk operations:
- QA engineer updating all test docs to "verified" status
- Tech lead marking all design docs as "complete" at sprint end
- DevOps person syncing metadata from Jira to docs

**The problem:** Tutorial presents CLI as the PRIMARY method. Section 5 title is "Enrich Metadata" and it's all CLI commands. There's NO mention of "you can also just edit the file."

New developers read this and think "I have to use docmgr meta update" when really they should just open the file and edit it directly.

**My recommendation:** Tutorial should be explicit about the trade-offs.

---

## Deep Dive Discussion (Cross-Talk Enabled)

**Alex:** Let's talk about the path verbosity. `--doc ttmp/TICKET-slug/doc-type/01-file.md` is ridiculous. That's 50+ characters every time.

**Sam:** From Round 3 we already proposed `--file design/01-file.md` with relative paths when `--ticket` is specified. Does that solve it?

**Alex:** YES. That makes it:
```bash
docmgr meta update --ticket T-001 --file design/01-my-design.md --field Summary --value "X"
```
Still verbose, but better.

**Sam:** But even with that, is it faster than just opening the file?

**Alex:** For ONE field? No. For 3+ docs? Yes.

**Morgan:** Can we talk about discoverability? The tutorial doesn't explain when to use CLI vs manual. It just shows CLI commands.

**Alex:** Right! Section 5 should start with:

"**Quick edits:** For updating 1-2 fields on a single doc, just open the file in your editor and edit the frontmatter YAML.

**Bulk operations:** For updating the same field across multiple docs, use `docmgr meta update`."

**Sam:** Agreed. And then show BOTH approaches. Example:

```markdown
### Update Summary on One Doc

**Quick approach (recommended for single docs):**
Open the file and edit:
```yaml
Summary: "Your new summary here"
```

**CLI approach (useful for automation):**
```bash
docmgr meta update --doc path/to/file.md --field Summary --value "Your new summary"
```
```

**Morgan:** I like that! Show the simple way first, then the powerful way.

**Alex:** Wait, let me test something. What if I want to update 3 fields on one doc?

*[types]*

```bash
$ docmgr meta update --doc path --field Summary --value "X"
$ docmgr meta update --doc path --field Status --value "Y"
$ docmgr meta update --doc path --field Owners --value "Z"
```

Three commands! That's absurd. Just open the file once and edit all three lines.

**Sam:** Unless you're scripting it. But yeah, for interactive use, manual wins.

**Morgan:** So the rule is: **CLI for many docs, manual for one doc**. Simple.

**Alex:** And automation. CLI for automation.

**Sam:** Can we add a `--fields` flag for multiple field updates?

```bash
$ docmgr meta update --doc path --fields Summary="X",Status="Y",Owners="Z"
```

**Alex:** That's getting complicated. Just edit the file at that point.

**Morgan:** Agreed. Keep CLI for bulk, manual for complex single-doc edits.

---

## Live Experiments

**Alex:** Let me time the CLI approach.

*[types]*

```bash
$ time (docmgr meta update --doc test.md --field Status --value review)
real: 0.056s
```

**Alex:** The CLI is fast (0.056s). But that's misleading if you're already writing the doc — editing frontmatter while you're in the file takes zero extra time.

**Sam:** Right, and if you're already in the file writing content, editing frontmatter takes 0 seconds. No command needed.

**Morgan:** This is the key insight: **context matters**. 

- Writing a doc? Frontmatter is right there, edit it.
- Running a script? CLI is faster.
- Updating 10 docs? CLI only way.

**Alex:** So the tutorial should explain the CONTEXT, not just the command.

---

## Facilitator Synthesis

### Erin "The Facilitator" Garcia

*[Draws decision tree on whiteboard]*

Okay team, clear consensus emerging:

### Key Themes

1. **CLI excels at bulk operations** — Unanimous agreement
2. **Manual editing better for single-doc, multi-field** — Strong consensus
3. **Tutorial doesn't explain WHEN to use which** — Major gap
4. **Path verbosity makes CLI tedious for single docs** — Pain point
5. **Most developers don't know CLI exists** — Discovery issue

### Pain Points Identified (by severity)

**P1 - Tutorial gaps:**
1. Doesn't explain when to use CLI vs manual
2. Presents CLI as primary method (it's not)
3. No mention of "just edit the file" option

**P2 - CLI verbosity (already identified in Round 3):**
4. Full paths required for `--doc`
5. No multi-field update in one command

**P2 - Documentation:**
6. Automation use cases not highlighted
7. Bulk operation examples sparse

### Wins Celebrated

1. **CLI enables automation** — Scripts, CI, bulk updates possible
2. **Manual editing works** — No tool required for simple edits
3. **Both approaches coexist** — Users can choose
4. **Bulk updates are powerful** — `--ticket X --doc-type Y` updates all at once

### Proposed Improvements

#### Improvement 1: Tutorial — Explain When to Use CLI vs Manual

**Add to Section 5, before first command example:**

```markdown
## 5. Enrich Metadata

### When to Edit Manually vs Use CLI

**Edit files directly (recommended for most cases):**
- ✅ Updating 1-3 fields on a single doc
- ✅ While actively writing a doc
- ✅ Complex metadata changes
- ✅ When you're already in your editor

**Use `docmgr meta update` CLI:**
- ✅ Updating the same field across 3+ docs
- ✅ Automation (scripts, CI/CD)
- ✅ Programmatic updates
- ✅ Validation of field values

### Quick Example: Both Approaches

**Manual approach (editing one doc):**
Open the file in your editor and modify the frontmatter:
```yaml
Summary: "Updated summary here"
Status: review
Owners:
    - alex
    - manuel
```

**CLI approach (bulk update across docs):**
```bash
# Update status on all design-docs for a ticket
docmgr meta update --ticket MEN-4242 --doc-type design-doc \
    --field Status --value review
```

**Rule of thumb:** Manual for one doc, CLI for many docs.
```

**Impact:** Users understand the trade-offs, choose the right tool

---

#### Improvement 2: Add Automation Examples Section

**Add Section 5.5 after basic `meta update` examples:**

```markdown
### Automation Patterns

`meta update` shines in scripts and automation:

**Pattern: Update docs based on search results**
```bash
# Find all stale docs and mark for review
docmgr search --updated-since "60 days ago" --with-glaze-output --output json | \
  jq -r '.[] | .path' | \
  while read doc; do
    docmgr meta update --doc "$doc" --field Status --value "needs-review"
  done
```

**Pattern: Sync metadata from external system**
```bash
# Update owners from Jira ticket
OWNERS=$(curl -s "jira.com/api/ticket/MEN-4242" | jq -r '.assignees | join(",")')
docmgr meta update --ticket MEN-4242 --field Owners --value "$OWNERS"
```

**Pattern: End-of-sprint bulk update**
```bash
# Mark all design docs as complete
docmgr meta update --ticket MEN-4242 --doc-type design-doc \
    --field Status --value complete
```
```

**Impact:** Power users discover automation potential

---

#### Improvement 3: CLI Reference — Add Decision Table

**In tutorial or help text:**

| Your Goal | Use This | Example |
|-----------|----------|---------|
| Edit 1-2 fields on 1 doc | Manual editing | Open file, edit YAML |
| Edit 5 fields on 1 doc | Manual editing | Open file, edit YAML |
| Edit 1 field on 5+ docs | CLI bulk update | `--ticket X --doc-type Y --field Z --value V` |
| Automation / scripts | CLI | Pipe search results to meta update |
| Validation required | CLI | CLI validates topics/doc-types |

**Impact:** Clear guidance on when to use which approach

---

### Action Items

**For Tutorial (docmgr-how-to-use.md):**
- [ ] Add "When to Use CLI vs Manual" section (Improvement 1)
- [ ] Show manual editing example first, then CLI
- [ ] Add decision table (Improvement 3)
- [ ] Add automation patterns section (Improvement 2)

**For CLI (future enhancement, low priority):**
- [ ] Consider `--fields` for multi-field updates (or document that manual is better)
- [ ] Relative `--file` paths (already proposed in Round 3)

**For Next Round:**
- [ ] Round 5: Relating Files feature

---

## Proposed Improvements (Full Detail)

### Change 1: Rewrite Section 5 Opening

**Current (lines 80-88):**

```markdown
## 5. Enrich Metadata

```bash
INDEX_MD="ttmp/MEN-4242-normalize-chat-api-paths-and-websocket-lifecycle/index.md"
docmgr meta update --doc "$INDEX_MD" --field Owners --value "manuel,alex"
docmgr meta update --doc "$INDEX_MD" --field Summary --value "Unify chat HTTP paths..."
...
```
```

**Proposed:**

```markdown
## 5. Managing Metadata

Documents have metadata (frontmatter) that you can update in two ways:

### Manual Editing (Recommended for Single Docs)

The simplest way to update metadata is to open the file and edit the YAML frontmatter:

```yaml
---
Title: My Design Doc
Summary: "Updated summary"  # <-- Edit this
Status: review               # <-- Or this
Owners:
    - alex                   # <-- Add owners
    - manuel
---
```

**When to use:** Editing 1-5 fields on a single doc, or while actively writing.

### CLI Updates (For Bulk Operations)

For updating multiple docs at once, use `docmgr meta update`:

```bash
# Update all design-docs in a ticket
docmgr meta update --ticket MEN-4242 --doc-type design-doc \
    --field Status --value review

# Update specific doc (automation/scripts)
docmgr meta update --doc ttmp/MEN-4242-.../design/01-design.md \
    --field Summary --value "New summary"
```

**When to use:** Bulk updates (3+ docs), automation, scripts, CI/CD.

**Rule of thumb:** Manual for few docs, CLI for many docs.
```

---

## Summary

**What worked:**
- CLI enables powerful bulk operations and automation
- Manual editing is simple and fast for single docs
- Both approaches coexist well
- Bulk updates (`--ticket X --doc-type Y`) are efficient

**What needs fixing (P1):**
- Tutorial doesn't explain when to use CLI vs manual
- Presents CLI as primary method when it's for specific use cases
- No decision guidance for users

**What needs improving (P2):**
- Automation examples sparse
- Decision table missing
- Manual editing not mentioned as valid approach

**Next steps:**
- Add "When to Use" section at start of Section 5
- Show manual approach first, CLI second
- Add automation patterns for power users
- Document decision criteria clearly
