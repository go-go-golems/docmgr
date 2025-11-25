---
Title: Full Tutorial Validation Review
Ticket: MEN-3083
Status: active
Topics:
    - docmgr
    - tutorial
    - documentation
DocType: working-note
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "Comprehensive review of docmgr 'how-to-use' tutorial: findings, clarity issues, and improvement recommendations."
LastUpdated: 2025-11-25T11:10:00-05:00
---

# Tutorial Validation Review: docmgr how-to-use

**Date:** 2025-11-25  
**Validator:** dumdum AI  
**Ticket:** MEN-3083  
**Status:** Beginner validation complete

---

## Executive Summary 

I followed the "Beginner tutorial validation checklist" by reading `docmgr help how-to-use` (Part 1: Essentials) and executing the recommended workflow. **The tutorial is MOSTLY CLEAR** — I was able to understand and follow 80% of it without re-reading sections. However, I found several clarity issues and one critical workflow problem that should be addressed.

**Overall Assessment:** The tutorial is good quality and achieves its goal of introducing newcomers to docmgr. With the improvements below, it would be excellent.

---

## Part 1: Essentials — Clarity Assessment

### ✅ What Works Well

1. **Glossary Section (Key Concepts)**
   - Clearly defines: Ticket, Ticket workspace, Docs root, Frontmatter, RelatedFiles, Vocabulary
   - Perfect placement early in the document
   - Prevents confusion about jargon

2. **Quick Navigation Box**
   - Clear visual separation
   - Helps readers choose their path based on use case
   - Good UX

3. **Section Progression**
   - Prerequisites → Setup → First Ticket → Add Docs → Search
   - Linear, logical flow
   - Each section builds on the previous one

4. **Code Examples**
   - Concrete, runnable commands
   - Show expected output
   - Easy to copy/paste

5. **Best Practices Callouts**
   - "Smart Default" note about documents inheriting ticket properties is helpful
   - Prevents confusion about repetition

### ⚠️ Issues Found

#### Issue #1: Unclear Distinction Between `doc add` and `doc create` (MINOR)
**Location:** Section 4, "Add Documents"  
**Problem:** The tutorial only mentions `docmgr doc add` but doesn't clearly state this is the ONLY way to create new docs. I wondered: "Is there a `docmgr doc create` command?" No significant time lost, but a note like "Use `docmgr doc add` (not `create`) to add new documents" would eliminate ambiguity.

**Recommendation:** Add clarifying note: "Note: Use `docmgr doc add` (the only command for creating new documents). `docmgr doc create` is for something different."

#### Issue #2: File Paths in Examples Are Fake (MINOR)
**Location:** Throughout (e.g., `backend/api/register.go`, `web/src/store/api/chatApi.ts`)  
**Problem:** The example files don't exist. When I followed Step 3 of the checklist and tried to run `docmgr doc relate` with these paths, it failed because the files weren't in the repo. This is not the tutorial's fault (checklist problem), but the tutorial doesn't warn that these are placeholder paths.

**Recommendation:** Add a note: "Note: The file paths in examples (`backend/api/register.go`, etc.) are placeholders. When you use `docmgr doc relate`, substitute your actual file paths."

#### Issue #3: `--file-note` Format Isn't Crystal Clear (MINOR)
**Location:** Section 7, "Relating Files to Docs"  
**Problem:** The format `--file-note "path:reason"` is shown but the colon separator isn't explicitly called out. When I first read it, I parsed it as two separate fields but the significance of the colon took a moment to register.

**Recommendation:** Make it explicit: "Use `--file-note "path:reason"` where `path` is the file path and `reason` is a short note explaining why this file matters. The colon (`:`) separates the two."

#### Issue #4: RelatedFiles YAML Structure Isn't Explained (MEDIUM)
**Location:** Section 14, "Advanced: RelatedFiles with notes and ignores"  
**Problem:** The tutorial shows the YAML output:
```yaml
RelatedFiles:
    - path: backend/api/register.go
      note: Registers API routes
```
But it doesn't explain: "Is this lowercase or capitalized?" (I found it's `path`/`note` in formatted output but `Path`/`Note` in the actual files I viewed). This inconsistency is confusing.

**Recommendation:** Clarify: "docmgr supports both `Path`/`Note` (capitalized) and `path`/`note` (lowercase) formats in YAML. Use `docmgr doc relate` to maintain consistency; avoid hand-editing RelatedFiles."

#### Issue #5: Vocabulary Isn't Enforced But Not Crystal Clear Why (MINOR)
**Location:** Section 2, "First-Time Setup"  
**Problem:** The tutorial says "Vocabulary... used for validation warnings" and "not enforced — you can use any topics." This is good, but the WHY isn't explained. I had to infer: "So I can use unknown topics and they'll just trigger warnings?"

**Recommendation:** Expand the explanation: "Vocabulary is a guideline, not a constraint. Unknown topics trigger warnings from `docmgr doctor` but don't prevent you from using them. This flexibility lets teams evolve their vocabulary over time."

#### Issue #6: Doctor Warnings Section Incomplete (MEDIUM)
**Location:** Section 11, "Validation with Doctor"  
**Problem:** The tutorial lists "Unknown topic/docType/intent" as a warning but doesn't explain how to resolve each one:
- Unknown topic → `docmgr vocab add --category topics --slug X`
- Unknown docType → `docmgr vocab add --category docTypes --slug X`
- Unknown intent → `docmgr vocab add --category intent --slug X`

The tutorial doesn't make this actionable.

**Recommendation:** Add a subsection "Resolving Doctor Warnings" with specific commands for each warning type.

#### Issue #7: Task vs. Changelog Distinction Could Be Sharper (MINOR)
**Location:** Sections 10 (Tasks) and 8 (Changelog)  
**Problem:** When should I use `task add` vs. `changelog update`? The tutorial describes both but doesn't clearly contrast:
- **Tasks:** Actionable to-dos, checkbox-based, tracked daily
- **Changelog:** Historical record of completed work, milestone-focused

**Recommendation:** Add a comparison table:

| Feature | Tasks | Changelog |
|---------|-------|-----------|
| Purpose | Track to-dos | Record history |
| Format | Checkboxes | Timestamped entries |
| When to use | Before/during work | After completing work |
| Typical lifespan | Cycle with ticket | Permanent record |

---

## Part 2: Everyday Workflows — Assessment

### ✅ What Works Well

1. **Metadata Management (Section 6)**
   - Clear examples of `--ticket` vs. `--doc` usage
   - Bulk operations example is practical
   - Good explanation of when to use each approach

2. **Changelog Recording (Section 8)**
   - Good emphasis on file notes being required
   - Examples show both minimal and full usage
   - Timestamp automation is helpful

3. **Doctor Validation (Section 9)**
   - Clear list of what it checks
   - `.docmgrignore` pattern shown
   - Common warnings listed

### ⚠️ Issues Found

#### Issue #8: `--suggest` Flag Isn't Explained (MEDIUM)
**Location:** Section 8, "Changelog Record"  
**Problem:** The tutorial shows:
```bash
docmgr changelog update --ticket MEN-4242 --suggest --apply-suggestions --query WebSocket
```
But `--suggest` isn't explained in the tutorial. What does it do? I had to infer: "It suggests related files based on query/topics, then applies them automatically if I add `--apply-suggestions`."

**Recommendation:** Add a subsection explaining `--suggest`:
"The `--suggest` flag tells docmgr to search for related files based on your `--query` and `--topics`. Add `--apply-suggestions` to accept suggestions automatically, or omit it to review and choose which suggestions to apply."

#### Issue #9: Relate Suggestions Workflow Isn't Clear (MEDIUM)
**Location:** Section 14, "Advanced: RelatedFiles"  
**Problem:** The tutorial shows:
```bash
docmgr doc relate --ticket MEN-3083 --suggest --apply-suggestions --query timeline --topics conversation,events
```
But the tutorial earlier emphasized "file notes are required" yet this command seems to auto-apply suggestions. Do the suggestions include notes? What if I don't like the suggestions? It's unclear.

**Recommendation:** Clarify the suggestion workflow:
- Show what `--suggest` prints (preview of suggestions)
- Explain that suggestions auto-generate notes
- Show how to review suggestions before applying them
- Example: "Run without `--apply-suggestions` first to see suggestions, then decide."

#### Issue #10: Multi-Step Workflows Need More Concrete Examples (MEDIUM)
**Location:** Section 14, "Index Playbook"  
**Problem:** The "Index Playbook" gives 3 steps but doesn't show them in a complete workflow. I had to mentally stitch them together.

**Recommendation:** Provide a complete end-to-end example like:
```bash
# Step 1: Relate files
docmgr doc relate --ticket MEN-4242 \
  --file-note "backend/api/register.go:API normalization" \
  --file-note "web/src/store/api/chatApi.ts:Frontend integration"

# Step 2: Update summary
docmgr meta update --ticket MEN-4242 \
  --field Summary \
  --value "Normalize API paths; update WebSocket lifecycle"

# Step 3: Validate
docmgr doctor --ticket MEN-4242 --stale-after 30 --fail-on error
```

---

## Part 3: Power User Features — Assessment

### ✅ What Works Well

1. **Structured Output Examples (Section 11)**
   - JSON/CSV examples are clear
   - Stable field names documented
   - Good automation patterns shown

2. **CI Integration Examples (Section 12)**
   - GitHub Actions example is runnable
   - Pre-commit hook example is practical
   - Makefile integration is useful

### ⚠️ Issues Found

#### Issue #11: "Available Output Formats" Missing Detail (MINOR)
**Location:** Section 11, Structured Output  
**Problem:** The tutorial lists output formats but doesn't explain when to use each:
- `json` — For automation/parsing
- `csv` — For spreadsheets
- `tsv` — For tab-separated data
- `yaml` — For configs/CI
- `table` — For humans

The tutorial doesn't guide which to pick.

**Recommendation:** Add guidance: "Pick your format based on your use case:
- **json** for CI scripts and API integrations
- **csv** for spreadsheets and data analysis
- **yaml** for configuration files and deployment automation
- **table** for human-readable output (terminal or reports)
- **tsv** rarely needed unless specifically required"

#### Issue #12: Field Selection Examples Are Advanced (MEDIUM)
**Location:** Section 11, "Field Selection Examples"  
**Problem:** The `--select-template` example is complex:
```bash
docmgr list docs --with-glaze-output \
  --select-template '{{.doc_type}}: {{.title}}' --select _0
```
What does `--select _0` mean? The template syntax isn't explained.

**Recommendation:** Add a beginner example first:
```bash
# Simple: just get paths
docmgr list docs --with-glaze-output --select path

# Intermediate: filter columns
docmgr list docs --with-glaze-output --output csv \
  --fields doc_type,title,path
```
Then show the advanced template example with an explanation: "`{{.field}}` is template syntax; `_0` is a placeholder for output."

---

## Part 4: Reference — Assessment

### ✅ What Works Well

1. **Command Reference (Section 13)**
   - Clear examples for each command
   - Good coverage of common tasks

2. **Vocabulary Management (Section 15)**
   - Shows how to add custom topics/doc types
   - Clear explanation of vocabulary purpose

### ⚠️ Issues Found

#### Issue #13: Workflow Recommendations Could Be More Specific (MINOR)
**Location:** Section 17, "Tips and Best Practices"  
**Problem:** Recommendations like "Relate files with notes (always)" are good but lack detail about WHEN in the workflow:
- At design time?
- After implementation?
- Before code review?

**Recommendation:** Make it prescriptive:
"Relate files in three phases:
1. **Design phase:** Identify which files will implement the design
2. **Implementation phase:** After coding, relate the actual files used
3. **Code review phase:** Before review, verify all key files are related with clear notes for reviewers"

#### Issue #14: Shell Gotchas Section Too Short (MINOR)
**Location:** Section 17, "Shell Gotchas"  
**Problem:** Only mentions parentheses and tab completion. But there are other gotchas:
- Paths with spaces
- Special characters in notes
- Quotes in file notes

**Recommendation:** Expand with examples:
```bash
# Paths with spaces: quote them
--file-note "path/to/my file.go:reason"

# Special characters in notes: escape or quote
--file-note "path.go:Reason with 'quotes' and special chars"

# Ticket directories with parentheses
cd "ttmp/MEN-XXXX-name-\(with-parens\)"
```

---

## Critical Issue: Reset Script vs. Fresh Tutorial Follow

**Issue #15: The reset script pre-executes tutorial steps (HIGH)**

**Location:** Checklist Step 2-3  
**Problem:** The `02-reset-and-recreate-repo.sh` script automatically runs:
- `docmgr init`
- `docmgr ticket create-ticket`
- `docmgr doc add`
- `docmgr doc relate` (with files)
- `docmgr task add`
- `docmgr changelog update`
- `docmgr doctor`

This means when a beginner follows the tutorial, the repo is already populated. They can't actually run the commands fresh; they get "no changes specified" or other errors.

**Impact:** Beginners can't follow the tutorial step-by-step because the repo is pre-configured. They learn ABOUT the commands but don't actually RUN them.

**Recommendation:** Modify the reset script to create a SKELETON repo (just `docmgr init` and basic directory structure) so beginners can run the tutorial commands fresh. Or, create a separate "follow-along" setup script that doesn't pre-run the tutorial commands.

---

## Answers to Validation Questions

### Q1: On-disk layout for newly created ticket
The ticket creates a directory at `ttmp/YYYY/MM/DD/TICKET-SLUG/` with:
- Subdirectories for each doc type: `design-doc/`, `reference/`, `playbooks/`, `scripts/`, `sources/`, `archive/`, `various/`, `.meta/`
- Base files: `index.md` (ticket overview), `tasks.md` (todo checklist), `changelog.md` (history)
- Numeric prefixes (01-, 02-, etc.) for docs within each subdirectory; base files exempt

### Q2: Command to create design doc and where it appears
Command: `docmgr doc add --ticket TICKET --doc-type design-doc --title "Title"`  
Location: `ttmp/YYYY/MM/DD/TICKET-SLUG/design-doc/NN-slug.md` (where NN is the auto-incremented prefix)

### Q3: How to relate multiple files in one invocation
Repeat `--file-note` flags in a single command:
```bash
docmgr doc relate --ticket TICKET \
  --file-note "path1:reason1" \
  --file-note "path2:reason2"
```

### Q4: CLI verbs for to-dos vs. progress
- **To-dos (tasks.md):** `docmgr task add|check|edit|remove|list`
- **Progress (changelog.md):** `docmgr changelog update`

### Q5: Warning from doctor and follow-up action
Warning: `unknown_topics: [test]` (topic not in vocabulary.yaml)  
Follow-up: Either use a known topic from `docmgr vocab list --category topics`, or add the topic with `docmgr vocab add --category topics --slug test --description "..."`

### Q6: Steps to append changelog entry with file notes
```bash
docmgr changelog update --ticket TICKET \
  --entry "Your changelog text" \
  --file-note "path1:reason1" \
  --file-note "path2:reason2"
```

### Q7: How to relate files to a specific subdocument
Use `--doc` instead of `--ticket`:
```bash
docmgr doc relate --doc ttmp/YYYY/MM/DD/TICKET-SLUG/design-doc/01-file.md \
  --file-note "code/path.go:reason"
```

### Q8: How to learn acceptable topic/status values
Run: `docmgr vocab list --category topics` or `docmgr vocab list --category status`  
Or: `docmgr vocab list` to see everything

---

## Overall Recommendations (Priority Order)

### High Priority
1. **Fix the reset script workflow** — Separate skeleton setup from full demo so beginners can run tutorial commands fresh
2. **Add doctor warnings resolution guide** — Show exact commands to fix each warning type
3. **Clarify `--suggest` workflow** — Explain what suggestions look like and when to apply them
4. **Expand task vs. changelog comparison** — Show a side-by-side table

### Medium Priority
5. Add examples of relating files to subdocuments (not just ticket index)
6. Explain `--file-note` colon format explicitly
7. Add complete end-to-end workflow examples
8. Clarify RelatedFiles YAML case sensitivity
9. Expand shell gotchas with more examples

### Low Priority
10. Add output format guidance (when to use json vs. csv vs. yaml)
11. Clarify when in the workflow to relate files (design vs. implementation vs. review)
12. Mention unknown topics are flexible/not enforced earlier

---

## Conclusion

The tutorial is **solid and well-structured**. A newcomer (even one who's a bit "dumdum") can follow 80% of it without confusion. The issues found are mostly clarity gaps, not fundamental problems. With the high-priority fixes above, this tutorial would be excellent for onboarding new docmgr users.

**Quality Score: 8/10**
- Strengths: Clear structure, good examples, logical progression, helpful callouts
- Weaknesses: Some unclear sections, no task/changelog comparison, reset script interferes with fresh learning

---

## Files Referenced During Validation

- `/tmp/test-git-repo/ttmp/2025/11/25/MEN-3083-tutorial-validation-ticket/index.md` (created and populated)
- `/tmp/test-git-repo/ttmp/2025/11/25/MEN-3083-tutorial-validation-ticket/design-doc/02-placeholder-design-context.md` (design doc created)
- `/tmp/test-git-repo/ttmp/2025/11/25/MEN-3083-tutorial-validation-ticket/tasks.md` (task added)
- `/tmp/test-git-repo/ttmp/2025/11/25/MEN-3083-tutorial-validation-ticket/changelog.md` (entry appended)

Command outputs and validation logs saved to `/tmp/docmgr-validation-logs/docmgr-run-*.log`


