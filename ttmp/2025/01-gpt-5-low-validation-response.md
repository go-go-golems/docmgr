---
Title: Tutorial validation responses (gpt-5 low)
Ticket: MEN-3083
Status: active
Topics:
    - docmgr
    - tutorial
DocType: working-note
Intent: long-term
Owners:
    - gpt-5 low
RelatedFiles: []
ExternalSources: []
Summary: "Full responses to the beginner tutorial validation checklist, authored by gpt-5 low."
LastUpdated: 2025-11-24
---

# Tutorial validation responses (gpt-5 low)

- I skimmed the tutorial output and reset the practice repo. I then ran all tutorial commands manually in a clean repo and validated with doctor. One issue: the manual `doc relate` step returned “no changes specified” even though flags were provided; the scripted run succeeded with the same flags. I logged this in the working note and stopped retrying as per your rule: I think I'm stuck, let's TOUCH GRASS.

## Answers to the checklist questions
- Describe the on-disk layout for a newly created ticket:
  - A ticket directory is created under `ttmp/YYYY/MM/DD/TICKET-slug/` with `index.md`, `README.md`, `tasks.md`, `changelog.md`, plus subdirs like `design/` (or `design-doc/`), `reference/`, `playbooks/`, `scripts/`, `sources/`, `.meta/`, `various/`, `archive/`.
- Name the command that creates a design doc and explain where the resulting file appears:
  - Command: `docmgr doc add --ticket MEN-3083 --doc-type design-doc --title "..."`
  - Appears under `ttmp/YYYY/MM/DD/MEN-3083-.../design-doc/NN-title.md` with an auto numeric prefix.
- Explain how to relate more than one file (with notes) in a single CLI invocation:
  - Use repeated `--file-note` flags:
    - `docmgr doc relate --ticket MEN-3083 --file-note "path1:note1" --file-note "path2:note2"`
- Describe the CLI verbs you reach for when tracking to-dos versus recording notable progress:
  - Tasks: `docmgr task add|list|check|uncheck|edit|remove`
  - Progress: `docmgr changelog update` (with optional `--file-note`)
- Summarize the warning produced by docmgr doctor during the sample run and outline the follow-up action:
  - Warning: `unknown_topics: [test]` on the ticket index. Follow-up: add `test` to vocabulary with `docmgr vocab add --category topics --slug test --description "..."`
    or change Topics to a known value.
- Outline the steps to append another changelog entry that includes file notes:
  - Run `docmgr changelog update --ticket MEN-3083 --entry "..." --file-note "path:note" [--file-note ...]`. Optionally also run `docmgr doc relate` for subdocuments and then validate with `docmgr doctor`.
- Describe how to relate files to a specific subdocument rather than the ticket index:
  - Use `--doc` targeting the exact file path:
    - `docmgr doc relate --doc ttmp/YYYY/MM/DD/MEN-3083-.../design-doc/NN-file.md --file-note "path:note"`
- Explain how to learn which topic/status values are acceptable when doctor reports an unknown value:
  - Use `docmgr vocab list` to inspect vocabulary; adjust via `docmgr vocab add` or change the field to a listed value.

## Repo state checks
- Verified repo exists: `/tmp/test-git-repo/ttmp`.
- Status summary showed 1 ticket and the expected docs.
- Manual run results matched the tutorial snippets, except for the `doc relate` anomaly mentioned.

## Logged findings (also recorded in the working note)
- Step 3: `doc relate` manual run showed “no changes specified” despite `--file-note` flags; the scripted run works. Suggest clarifying messaging or the tutorial’s example.
- Consistency of `docmgr doc relate` vs `docmgr relate` in examples.
- Numeric prefix behavior when repeating runs.

## Assessment of tutorial quality (gpt-5 low)
 
### Overall
- **Clarity**: Good foundational coverage; essentials are mostly  understandable without prior context.
- **Consistency**: Some command naming inconsistencies and path variations (e.g., `design` vs `design-doc`) create avoidable friction.
- **Reliability**: The scripted flow works end-to-end. The manual `doc relate` path produced a confusing “no changes specified” message.
- **Beginner resilience**: Generally strong, but a few small paper cuts likely to trip newcomers.

### Strengths
- **Actionable, progressive structure**: The tutorial flows from init → create-ticket → add-doc → relate → tasks → changelog → doctor.
- **Built-in guardrails**: Root discovery notes and doctor examples prepare users for common issues.
- **Automation parity**: The validation script mirrors human steps, useful for quick smoke tests and regression checks.
- **Output examples**: Human vs structured output guidance is practical and encourages CI/automation habits.

### Issues and confusions (ranked by severity)
1. **Command inconsistency in help text**: Examples show `docmgr relate` while actual usage is `docmgr doc relate`. Beginners will copy the wrong form.
2. **Relate command feedback**: Manual runs emitted “Error: no changes specified” despite providing `--file-note`. This likely indicates “no diff” (already related), but the message reads like a failure.
3. **Subdir naming drift**: The tutorial alternates references to `design/` and `design-doc/`. The generated path is `design-doc/`. This inconsistency causes path confusion when navigating and when using `--doc`.
4. **Root discovery friction**: Reminders exist, but beginners still easily run commands from the wrong working directory. This explains many “it didn’t work” reports.
5. **Vocabulary warnings without quick fix**: The expected `unknown_topics: [test]` warning is noted, but a “do-this-now” fix is not embedded where the warning first appears.
6. **Numeric prefixes surprise**: Re‑running steps yields `02-...` documents; this should be called out earlier with a rationale (ordering) to avoid confusion.

### Recommendations
- **Fix the help examples globally**: Standardize on `docmgr doc relate` in all examples and references.
- **Improve “no changes specified” wording**: Change to something like: “No changes to apply (all notes already present). Use --remove-files to drop entries or change notes to update.”
- **Normalize subdir naming**: Use `design-doc/` everywhere in the tutorial and help. If `design/` appears historically, add a short note explaining legacy naming.
- **Embed root safety in examples**: Prefer including `--root ttmp` in all tutorial commands or add a prominent callout box: “Run these from the repo root (or pass --root).”
- **Inline the vocabulary quick fix**: Right after demonstrating `doctor`, include a one‑liner: `docmgr vocab add --category topics --slug test --description "..."` and re‑run `doctor`.
- **Add an early callout about numeric prefixes**: “Files are auto‑prefixed to keep order. Re‑creates may increment to 02-, 03-, etc. This is expected.”

### Quick tutorial edits (high impact, low effort)
- In the “Relate files” section, add both successful and no‑op examples:
  - Success: `docmgr doc relate --ticket MEN-3083 --file-note "backend/api/register.go:Registers API routes" --file-note "web/src/store/api/chatApi.ts:Frontend integration"`
  - No‑op behavior: “If files are already present with the same notes, you’ll see a message indicating no changes to apply.”
- In each step, add a micro “verify” command:
  - After `doc add`: `docmgr list docs --ticket MEN-3083`
  - After `relate`: `docmgr doc search --file backend/api/register.go`
  - After `task add`: `docmgr task list --ticket MEN-3083`
  - After `changelog update`: `grep -n \"Initial tutorial validation pass\" ttmp/.../changelog.md`
- Add a small “Common mistakes” box at the end of Part 1 (Essentials): wrong working dir; forgetting `--root`; mixing `docmgr relate` vs `docmgr doc relate`; not quoting paths/notes with spaces/parentheses.

### Troubleshooting the relate anomaly
- Probable cause: attempting to relate files that are already present with identical notes (from the earlier scripted setup).
- Best practice guidance to include:
  - “If you see ‘no changes specified’, list current relationships by opening the document’s frontmatter or by using `docmgr doc search --file <path>` to confirm it’s already linked.”
  - “To update a note, repeat `--file-note` with the same path and a changed note; the tool will update in place.”
  - “To remove, use `--remove-files path1,path2`.”

### UX suggestions
- Add a `--verbose` flag suggestion in the tutorial when demonstrating `relate`, so beginners can see a summary of computed diffs.
- Provide a “First successful run checklist” card that learners can quickly compare against: repo exists, status summary counts, ticket path, one design doc created, at least one related file, one task, one changelog entry, doctor warning acknowledged/fixed.

### Repetition and inaccurate guidance
- **Duplicate explanations**: The “Record Changes in Changelog” section appears twice with nearly identical wording (Part 2 §8 and Part 2 §8 again near line 525). This wastes attention and makes it unclear which version is authoritative.
- **Outdated command references**: Sections still mention `docmgr relate` (without `doc`) and refer to `design/` directories. These references are simply wrong in current releases.
- **Conflicting instructions about `.ttmp.yaml`**: Early sections say “docmgr init creates .ttmp.yaml in repo root,” later sections imply `.ttmp.yaml` is optional and may live elsewhere without reinforcing how to configure `--root`. Needs a single, precise explanation.
- **Changelog/file-note mismatch**: The tutorial celebrates file-note enforcement yet continues to show `--files` examples (e.g., legacy text around line 408). Those instructions can’t be followed because the flag was removed; this is a correctness bug.

### Full assessment (critical take)
1. **Structure**: The tutorial is comprehensive but bloated. Parts 2 and 3 repeat subsections verbatim (e.g., “Record Changes in Changelog” shows up three times). Readers can’t tell when they’ve already learned a concept. Recommend collapsing duplicates into one canonical section with cross-links.
2. **Accuracy**: Several snippets no longer match the CLI:
   - `docmgr relate` vs `docmgr doc relate`
   - `--files` in `changelog update`
   - Directory names (`design` vs `design-doc`)
   - Mentioning `docmgr init` as optional in some places when it’s required for the workflows being described.
   These need immediate fixes to preserve trust.
3. **Workflow guidance**: The tutorial explains *what* commands do but often omits *how to verify success*. The new playbook tries to plug this gap with explicit `status`, `list`, and `grep` steps. The canonical doc should embed those verification steps instead of relying on external checklists.
4. **Error-handling realism**: Common failure modes (wrong working directory, missing vocabulary entries, re-running `relate` with identical notes) are acknowledged but buried. They should be surfaced as inline “If you see X, do Y” callouts right after each command intro.
5. **Beginner friction**: For novices, the tutorial’s density and occasional contradictions cause cognitive overload. Adding safety guardrails (no chained commands, always confirm `pwd`, etc.) directly into Part 1 would prevent many self-inflicted wounds.

Overall, the tutorial still provides a strong foundation, but quality suffers from drift: repeated sections, stale flags, and inconsistent terminology. Prioritizing accuracy fixes and trimming duplicate content will have outsized impact on reader trust and productivity.


