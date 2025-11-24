---
Title: 'Implementation Diary: Ticket Close, Status Vocabulary, and Frontmatter Fixes'
Ticket: DOCMGR-CLOSE
Status: active
Topics:
    - docmgr
    - workflow
    - ux
    - automation
DocType: log
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-19T15:32:43.95549767-05:00
---

# Implementation Diary: Ticket Close, Status Vocabulary, and Frontmatter Fixes

Implementation diary for DOCMGR-CLOSE tracking the journey of implementing `ticket close`, status vocabulary, structured output, and fixing frontmatter parsing issues.

---

## 2025-11-19 - Implementation Complete

### What We Implemented

**Core Feature: `ticket close` Command**
- Created `pkg/commands/ticket_close.go` with atomic close operation
- Wired into CLI via `cmd/docmgr/cmds/ticket/close.go`
- Implements both `Run` (human-friendly) and `RunIntoGlazeProcessor` (structured output)
- Behavior:
  - Updates Status (default: `complete`, override with `--status`)
  - Optionally updates Intent (via `--intent`)
  - Appends changelog entry
  - Updates LastUpdated timestamp
  - Checks if all tasks are done, warns if not (doesn't fail)

**Status Vocabulary System**
- Extended `models.Vocabulary` struct to include `Status []VocabItem`
- Updated `ttmp/vocabulary.yaml` with initial status values:
  - draft, active, review, complete, archived
- Modified `doctor` to validate Status field (warnings for unknown values)
- Updated `vocab add` to support `--category status`
- Updated `vocab list` to display status vocabulary
- Seeded default status values in `init` command

**Tasks Check Enhancements**
- Added suggestion when all tasks complete: "💡 All tasks complete! Consider closing the ticket: docmgr ticket close --ticket <ID>"
- Implemented `RunIntoGlazeProcessor` for structured output with:
  - `all_tasks_done` boolean
  - `open_tasks`, `done_tasks`, `total_tasks` counts
  - `checked_ids` array
- Enabled dual-mode output via `--with-glaze-output`

**Bonus: Frontmatter Parsing Fix**
- Fixed duplicate frontmatter issue in template rendering
- Problem: Templates with placeholders (`{{TITLE}}`) aren't valid YAML
- Solution: Hybrid approach in `ExtractFrontmatterAndBody`:
  1. Try library parsing (works for real documents)
  2. Fall back to manual delimiter detection when parsing fails
  3. Strip frontmatter by finding closing `---`
- Enhanced doctor to detect invalid frontmatter in ALL markdown files, not just index.md

### What Worked Well

1. **Architecture Clarity**: Following existing patterns from `ticket create-ticket` and `task check` made wiring straightforward
2. **Glazed Framework**: Dual-mode output (Run + RunIntoGlazeProcessor) pattern was easy to follow
3. **Vocabulary Model**: Extending the existing vocabulary system was clean and consistent
4. **Atomic Operations**: `ticket close` bundles multiple updates into a single, verifiable operation
5. **Testing with docmgr**: Using docmgr commands themselves to manage the implementation ticket validated UX

### What Didn't Work / Challenges

1. **Duplicate Frontmatter Bug**: The original DOCMGR-CLOSE index.md had duplicate frontmatter blocks
   - Root cause: Template frontmatter wasn't stripped because placeholders made YAML invalid
   - Parsing failed, so entire template (including frontmatter) was returned as body
   - Fixed by adding manual parsing fallback

2. **Ticket Discovery Issue**: `ticket close` couldn't find DOCMGR-CLOSE initially
   - Root cause: Invalid YAML (unquoted colon in title: `IMPL: Implement...`)
   - Discovery relies on parsing index.md frontmatter
   - If parsing fails, `ws.Doc` is nil, ticket not found
   - Fixed by quoting the Title field

3. **Function Signature Mismatch**: Initial doctor enhancement used wrong function signature
   - Called `readDocumentFrontmatter` expecting 3 returns, only has 2
   - Quick fix after checking `document_utils.go`

4. **Remembering to use --tasks-file vs --ticket**: When ticket discovery fails, need to use `--tasks-file` with absolute path

### Lessons Learned

**About YAML and Frontmatter**
- The YAML encoder automatically quotes values with colons, so title handling is robust
- Templates with placeholders can't be parsed by YAML libraries (expected behavior)
- Hybrid approach (library + manual fallback) is the right balance
- Invalid frontmatter silently breaks ticket discovery, but doctor now catches it

**About Command Implementation**
- Glazed dual-mode pattern is powerful: same business logic, two output formats
- Use `cli.WithDualMode(true)` and `cli.WithGlazeToggleFlag("with-glaze-output")` for consistency
- Human output should be concise and actionable
- Structured output should include state info (e.g., `all_tasks_done`) for orchestration

**About Testing**
- Using docmgr to manage docmgr tickets is dogfooding at its best
- Edge cases appear during real usage: colons, invalid YAML, duplicate blocks
- Testing workflows end-to-end reveals integration issues early

**About Vocabulary Systems**
- Extending vocabulary is straightforward when model is well-designed
- Status as vocabulary-guided (warnings only) strikes the right balance
- Intent stays vocabulary-controlled for stricter governance
- Teams can customize status values via `vocab add`

### What Should Be Done Next Time

**Pre-Implementation**
1. Check if existing index.md has parsing issues before starting work
2. Run `doctor` on ticket workspace at the beginning
3. Quote any titles with colons immediately

**During Implementation**
1. Test each component individually before integration:
   - Test ticket close command in isolation
   - Test vocabulary additions separately
   - Test structured output modes
2. Use test tickets liberally to validate UX
3. Clean up test tickets immediately after validation

**Code Quality**
1. Always check function signatures before calling utility functions
2. Verify both Run and RunIntoGlazeProcessor implementations compile
3. Add comments explaining why fallback logic exists (e.g., template placeholders)

**Documentation Discipline**
1. Use docmgr commands for changelog/relate as work progresses (not at the end)
2. Check off tasks immediately after completing them
3. Keep ticket metadata synchronized with actual state

### Remaining Work (from original tasks)

✅ Completed:
1. ticket close command (skeleton + implementation)
2. Atomic close operation (status, intent, changelog, LastUpdated)
3. Structured output with operations tracking
4. Status vocabulary with doctor warnings
5. Tasks check suggestion for all tasks done
6. Tasks check structured output
7. Related files and changelog entries

⏸️ Deferred (documentation):
- Document suggested status transitions in help/docs
- Update how-to-use + cli-guide with ticket close examples

These can be done in a follow-up since the implementation is complete and tested.

### Next Steps

1. Complete remaining documentation tasks (tasks #5 and #8)
2. Consider adding `ticket reopen` and `ticket archive` commands (follow-up tickets)
3. Explore opt-in `--auto-close` with `--yes` for CI/agents (after gathering feedback)
4. Monitor real-world usage for additional edge cases

### Key Takeaways

**Hybrid parsing is pragmatic**: Using a library for 95% of cases and manual fallback for edge cases (templates) is better than implementing full YAML parsing from scratch.

**Atomic operations reduce cognitive load**: `ticket close` bundles what used to be 3-5 commands into one, with clear output and proper error handling.

**Structured output enables automation**: LLMs and scripts can now reliably parse operation results without regex-ing human output.

**Vocabulary-guided vs controlled**: The distinction between Status (extensible, warnings) and Intent (controlled, errors) gives teams flexibility while maintaining data quality.

### Files Modified

See RelatedFiles in index.md frontmatter for the complete list with explanatory notes.
