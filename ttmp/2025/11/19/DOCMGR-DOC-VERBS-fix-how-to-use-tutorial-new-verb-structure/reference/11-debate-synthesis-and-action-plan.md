---
Title: Debate synthesis and action plan — Tutorial fixes
Ticket: DOCMGR-DOC-VERBS
Status: active
Topics:
    - docmgr
    - documentation
    - tutorial
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/pkg/doc/docmgr-how-to-use.md
      Note: Tutorial to be fixed based on debate decisions
    - Path: docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/05-debate-round-1-go-no-go.md
      Note: Round 1 decisions
    - Path: docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/06-debate-round-2-patch-or-restructure.md
      Note: Round 2 decisions
    - Path: docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/07-debate-round-3-severity-triage.md
      Note: Round 3 decisions
    - Path: docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/09-debate-round-4-missing-functionality.md
      Note: Round 4 decisions
    - Path: docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/08-debate-round-8-error-messages.md
      Note: Round 8 decisions
    - Path: docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/10-debate-round-12-regression-prevention.md
      Note: Round 12 decisions
ExternalSources: []
Summary: "Synthesis of all debate rounds with complete action plan for fixing docmgr tutorial."
LastUpdated: 2025-11-25
---

# Debate Synthesis and Action Plan

## Executive Summary

We conducted a presidential-style debate with 10 personas (4 human developers, 3 document entities, 3 wildcards) to review the docmgr tutorial quality. Based on validation data from 3 beginner testers, we identified 25 issues, prioritized fixes, and designed a comprehensive prevention strategy.

**Verdict:** Fix the tutorial using a hybrid approach (patch urgently, restructure later if needed) with three-layer regression prevention (automation, process, ownership).

---

## Debate Rounds Completed

✅ **Round 1:** Should We Fix This At All? → **UNANIMOUS GO**  
✅ **Round 2:** Patch or Restructure? → **HYBRID APPROACH**  
✅ **Round 3:** What's Actually Broken? → **2 CRITICAL, 2 HIGH, 9 MEDIUM, 12 LOW**  
✅ **Round 4:** Missing Functionality? → **Add task editing + maintenance (~110 lines)**  
✅ **Round 8:** Error Messages? → **Split reset script + troubleshooting section**  
✅ **Round 12:** Regression Prevention? → **Three-layer defense (automation + process + ownership)**  

---

## Key Decisions Summary

### Round 1: Fix It (Go/No-Go)

**Decision:** Fix the tutorial.

**Evidence:**
- Tutorial: 1,457 lines
- Issues: 25 distinct issues found
- Completion: 100% but took 2-3x advertised time
- Validator quote: "We succeeded DESPITE the tutorial, not because of it"

**Reasoning:**
- Wrong commands erode trust (Maya)
- Completion rate is vanity metric—look at time and errors (Jamie)
- Beginners blame themselves, not docs (Sam)
- With AI tools, fixes take ~70 minutes (negligible cost)

---

### Round 2: Hybrid Approach (Patch Then Restructure)

**Decision:** Two-phase fix.

**Phase 1 (This Week — 90 min):**
- Fix command syntax errors (docmgr relate → docmgr doc relate)
- Fix path inconsistencies (design/ → design-doc/)
- Fix removed flags (--files → --file-note)
- Remove obvious duplicate section

**Phase 2 (Next Sprint — if needed):**
- Consolidate remaining duplicates
- Evaluate Part 2 length (480 lines, 33% of tutorial)
- Improve navigation
- Based on Phase 1 feedback

**Reasoning:**
- Urgency favors fast fixes (Maya)
- Risk management favors phased approach (Jamie)
- Restructuring needs more validation data (all)
- Part 2 length not proven as problem yet (Maya)

---

### Round 3: Severity Triage

**Decision:** Prioritized all 25 issues.

**CRITICAL (Fix Immediately — Day 1):**
1. Issue #16: Command syntax errors (docmgr relate)
2. Issue #24: Removed flags (--files)

**HIGH (Fix This Week):**
3. Issue #15: Reset script pre-executes steps
4. Issue #25: Path inconsistencies (design/ vs design-doc/)

**MEDIUM (Next Sprint — 9 issues):**
- Duplicates, doctor warnings, --suggest flag, clarity issues

**LOW (Backlog — 12 issues):**
- Polish, minor improvements, nice-to-haves

**Reasoning:**
- All candidates agreed on CRITICAL tier (factually wrong)
- 3/4 agreed on HIGH tier (high user impact)
- Severity = user pain, not just correctness (Checklist)

---

### Round 4: Missing Functionality

**Decision:** Add ~110 lines to existing sections.

**Add to Part 2, Section 10 (Tasks):**
- `task edit/remove/uncheck` commands (+30 lines)
- Natural completion of task workflow

**Add to Part 4, Section 17 (Maintenance):**
- `doc layout-fix` — Reorganize structure (+40 lines)
- `config show` — Debugging tool (+20 lines)
- Expand `renumber` — Resequencing (+20 lines)

**Skip:**
- rename-ticket (too rare)
- import file (defer)
- template validate (separate guide)

**Reasoning:**
- Task editing is everyday workflow (Jamie)
- Maintenance commands users will encounter (all)
- Validation showed current coverage sufficient for beginners (Maya)
- Don't bloat unnecessarily (Maya)

---

### Round 8: Error Messages & Troubleshooting

**Decision:** Fix reset script + add troubleshooting section.

**For This Ticket:**
1. **Split reset script:**
   - `setup-practice-repo.sh` — Skeleton (init + create only)
   - `validate-tutorial.sh` — Full workflow (for testing)

2. **Add troubleshooting section** with top 5 errors:
   - "no changes specified" → Not always an error
   - "Unknown topic: [X]" → How to add to vocab
   - "Must specify --doc or --ticket" → Which to use
   - File not found → Common causes
   - Doctor warnings → Resolution steps

**For Separate CLI Ticket (optional):**
- Improve "no changes specified" context detection
- Add doctor warning resolution guidance
- Defer until Phase 1 data proves need

**Reasoning:**
- Reset script is root cause of "already related" errors (Reset Script)
- Error messages need empathy (Sam)
- Fix CLI at source, then document edge cases (Jamie)
- Tutorial troubleshooting helps all users (all)

---

### Round 12: Regression Prevention

**Decision:** Three-layer defense.

**Layer 1: Automation (CI):**
1. Pattern linting (catches deprecated syntax) — 5 sec
2. Command coverage (finds undocumented commands) — 10 sec
3. Link validation (finds broken links) — 30 sec
**Total: 45 seconds per PR**

**Layer 2: Process:**
1. CODEOWNERS (doc maintainer reviews CLI changes)
2. PR template (doc impact assessment)
3. CLI changelog (explicit tracking)
4. Tiered review (not everything needs maintainer)

**Layer 3: Ownership:**
1. Assign documentation maintainer (role)
2. Create style guide
3. Quarterly validation schedule
4. Tutorial health dashboard

**Reasoning:**
- Automation catches 80% of drift (CI Robot)
- Process prevents drift at source (Git History)
- Ownership ensures quality (Jamie)
- Metrics make quality visible (Tutorial)
- All layers needed for 100% prevention

---

## Complete Action Plan

### Phase 1 (This Week — DOCMGR-DOC-VERBS Ticket)

**Day 1: CRITICAL Fixes (2 hours)**

1. **Fix command syntax errors:**
   ```bash
   # Find and replace in docmgr-how-to-use.md
   docmgr relate      → docmgr doc relate      (6 instances)
   docmgr add         → docmgr doc add         (4 instances)
   docmgr search      → docmgr doc search      (5 instances)
   docmgr guidelines  → docmgr doc guidelines  (2 instances)
   ```

2. **Fix removed flags:**
   ```bash
   # Replace all instances
   --files → --file-note "path:note" (with explanation)
   ```

3. **Test changes:**
   - Run grep to verify no old patterns remain
   - Manually verify examples make sense

---

**Day 2-3: HIGH Fixes (3 hours)**

4. **Split reset script:**
   - Create `setup-practice-repo.sh` (skeleton only: init + create ticket)
   - Rename existing to `validate-tutorial.sh` (full workflow)
   - Update checklist to reference both scripts
   - Test both scripts work

5. **Fix path inconsistencies:**
   ```bash
   # Standardize in tutorial
   design/ → design-doc/ (all instances)
   playbooks/ → playbook/ (check consistency)
   ```

6. **Remove duplicate section:**
   - Delete line 528 (accidental duplicate of Section 8)
   - Verify other duplicates have distinct purposes
   - Add cross-references if needed

---

**Day 4: Documentation Additions (3 hours)**

7. **Expand Part 2, Section 10 (Manage Tasks):**
   Add +30 lines:
   ```markdown
   # Edit task text
   docmgr task edit --ticket T --id 3 --text "Updated description"
   
   # Remove tasks
   docmgr task remove --ticket T --id 5
   
   # Uncheck if needed
   docmgr task uncheck --ticket T --id 2
   ```

8. **Add Part 4, Section 17 (Maintenance Commands):**
   Add +80 lines:
   ```markdown
   ## 17. Maintenance Commands [ADVANCED]
   
   ### layout-fix — Reorganize Document Structure
   [Purpose, usage, when needed, examples]
   
   ### config show — Display Configuration
   [Purpose, usage, when needed, examples]
   
   ### renumber — Resequence Numeric Prefixes
   [Expand existing brief mention]
   ```

9. **Add Troubleshooting Section (Appendix):**
   Add +100 lines:
   ```markdown
   ## Appendix: Troubleshooting Common Errors
   
   ### "Error: no changes specified"
   [What it means, causes, how to fix, examples]
   
   ### "Unknown topic: [X]"
   [What it means, causes, how to fix, examples]
   
   ### [3 more common errors]
   ```

---

**Day 5: Validation (2 hours)**

10. **Test all changes:**
    - Read through entire tutorial (flow still makes sense?)
    - Run commands from examples (all work?)
    - Check cross-references (links valid?)

11. **Run validation checklist:**
    - Fresh environment (use new `setup-practice-repo.sh`)
    - Follow tutorial manually
    - Time Part 1 completion
    - Log any issues

12. **Update changelog:**
    ```bash
    docmgr changelog update --ticket DOCMGR-DOC-VERBS \
      --entry "Phase 1 complete: Fixed 2 CRITICAL + 2 HIGH issues, added 110 lines (task editing + maintenance + troubleshooting)"
    ```

---

### Phase 2 (Week 2 — Automation Setup)

**Automation Layer (3 hours):**

1. **Create `scripts/lint-doc-commands.sh`:**
   ```bash
   #!/bin/bash
   # Pattern linting for deprecated commands
   
   ERRORS=0
   
   # Check for old verb structure
   if grep -rn "docmgr relate[^d]" pkg/doc/*.md; then
     echo "ERROR: Found 'docmgr relate' (should be 'docmgr doc relate')"
     ERRORS=$((ERRORS + 1))
   fi
   
   if grep -rn "docmgr add[^-]" pkg/doc/*.md; then
     echo "ERROR: Found 'docmgr add' (should be 'docmgr doc add' or 'task add')"
     ERRORS=$((ERRORS + 1))
   fi
   
   if grep -rn "\--files[^-]" pkg/doc/*.md; then
     echo "ERROR: Found '--files' (removed flag)"
     ERRORS=$((ERRORS + 1))
   fi
   
   exit $ERRORS
   ```

2. **Create `scripts/check-command-coverage.sh`:**
   ```bash
   #!/bin/bash
   # Verify command coverage
   
   docmgr help | grep -E "^  [a-z]+" | awk '{print $1}' | \
     while read cmd; do
       if ! grep -q "docmgr $cmd" pkg/doc/docmgr-how-to-use.md; then
         echo "WARN: Command 'docmgr $cmd' not documented"
       fi
     done
   ```

3. **Create `.github/workflows/lint-docs.yml`:**
   ```yaml
   name: Documentation Validation
   
   on: [pull_request, push]
   
   jobs:
     lint-docs:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v3
         
         - name: Check command patterns
           run: ./scripts/lint-doc-commands.sh
         
         - name: Check command coverage
           run: |
             make build
             ./scripts/check-command-coverage.sh
         
         - name: Check markdown links
           uses: gaurav-nelson/github-action-markdown-link-check@v1
           with:
             use-quiet-mode: 'yes'
             config-file: '.markdown-link-check.json'
   ```

4. **Test automation:**
   - Run locally before committing
   - Verify catches known bad patterns
   - Ensure doesn't false-fail on valid content

---

### Phase 3 (Week 3 — Process Setup)

**Process Layer (2 hours):**

1. **Create CODEOWNERS:**
   ```
   # Documentation
   /pkg/doc/              @manuel
   
   # CLI commands (doc maintainer reviews for doc impact)
   /cmd/docmgr/cmds/      @manuel
   /pkg/commands/         @manuel
   ```

2. **Update PR template (`.github/pull_request_template.md`):**
   ```markdown
   ## Documentation Impact
   
   Does this PR change CLI behavior? [ ] Yes [ ] No
   
   If yes, which docs need updating?
   - [ ] docmgr-how-to-use.md (tutorial)
   - [ ] docmgr-how-to-setup.md (setup guide)
   - [ ] Command --help text
   - [ ] Troubleshooting section
   - [ ] None (internal refactor only)
   
   If docs updated: Commit hash: _______
   ```

3. **Create `docs/cli-changelog.md`:**
   ```markdown
   # CLI Changelog (User-Facing Changes)
   
   Track CLI changes that affect documentation.
   
   ## Unreleased
   - (none)
   
   ## v0.1.14 (2025-11)
   - BREAKING: Commands moved under verb groups (doc, ticket, task)
   - BREAKING: --files removed from relate (use --file-note)
   - Changed: Paths standardized to DocType (design-doc/ not design/)
   
   ## v0.1.13 (2025-10)
   - Added: ticket close command
   - Added: --fail-on flag to doctor
   ```

4. **Document tiered review process:**
   ```markdown
   # Documentation Review Process
   
   ## Tier 1 (No Review): Auto-approve
   - Internal refactors
   - Test changes
   - Non-CLI code
   
   ## Tier 2 (Peer Review): Any team member
   - Typo fixes
   - Broken link fixes
   - Example updates
   
   ## Tier 3 (Maintainer Review): Doc maintainer
   - CLI flag changes
   - Command behavior changes
   - Error messages
   
   ## Tier 4 (Full Validation): Maintainer + checklist
   - New commands
   - Command structure changes
   - Breaking changes
   ```

---

### Phase 4 (Week 4 — Ownership & Quality)

**Ownership Layer (4 hours):**

1. **Create style guide (`docs/docmgr-style-guide.md`):**
   ```markdown
   # docmgr Documentation Style Guide
   
   ## Voice and Tone
   - Active voice ("Run docmgr init" not "docmgr init should be run")
   - Second person ("you" not "the user")
   - Conversational but precise
   
   ## Command Examples
   - Always show full commands with flags
   - Use realistic ticket IDs (MEN-4242, not FOO-123)
   - Add comments for multi-line commands
   - Show expected output
   
   ## Terminology (Consistent Usage)
   - "ticket" (lowercase in prose)
   - "Ticket" (capitalized in field names)
   - "docs root" not "documentation root"
   - "frontmatter" not "front-matter"
   - "RelatedFiles" (capitalized, one word)
   
   ## Structure Conventions
   - Part 1: Essentials (init → create → add → search only)
   - Part 2: Everyday Workflows (common tasks)
   - Part 3: Power User (automation, advanced)
   - Part 4: Reference (complete command listing)
   - Appendices: Troubleshooting, glossary, advanced topics
   ```

2. **Create tutorial health dashboard (`docs/tutorial-health.md`):**
   ```markdown
   # Tutorial Health Dashboard
   
   ## Last Validation
   - Date: 2025-11-25
   - Validators: 3 (gpt-5-low, gpt-5-full, dumdum)
   - Completion rate: 100%
   - Average time (Part 1): 18 minutes (target: 15)
   - Issues found: 0 critical, 0 high, 2 medium
   
   ## Command Accuracy
   - Total commands in tutorial: 45
   - Commands verified: 45 (100%)
   - Deprecated patterns: 0
   - Unknown commands: 0
   - Last verified: 2025-11-25
   
   ## Content Health
   - Last review: 2025-11-25
   - Next review due: 2026-02-25
   - DOCDEBT markers: 0
   - Duplicate sections: 0
   - Status: ✓ Current
   
   ## Metrics Trends
   | Quarter | Validation Time | Issues | Command Accuracy |
   |---------|----------------|--------|------------------|
   | Q4 2025 | 18 min | 0 | 100% |
   ```

3. **Schedule quarterly validation:**
   - Add to project roadmap/calendar
   - Create issue template for validation runs
   - Document process in `docs/validation-process.md`

4. **Update tutorial frontmatter:**
   ```yaml
   ---
   # ... existing fields ...
   LastReviewedBy: manuel
   LastReviewDate: 2025-11-25
   ReviewCadence: quarterly
   NextReviewDue: 2026-02-25
   ValidationVersion: v2.0
   ---
   ```

---

## Implementation Checklist

### Phase 1: Tutorial Fixes (This Week)

**CRITICAL Issues:**
- [ ] Fix command syntax: `docmgr relate` → `docmgr doc relate` (6 instances)
- [ ] Fix command syntax: `docmgr add` → `docmgr doc add` (4 instances)
- [ ] Fix command syntax: `docmgr search` → `docmgr doc search` (5 instances)
- [ ] Fix removed flag: `--files` → `--file-note` (all examples)

**HIGH Issues:**
- [ ] Split reset script into learning vs testing versions
- [ ] Fix path inconsistencies: `design/` → `design-doc/`
- [ ] Remove duplicate section (line 528 changelog)

**Additions:**
- [ ] Add shell completion to Part 1 or Part 4 (+20 lines)
- [ ] Add task editing to Part 2, Section 10 (+30 lines)
- [ ] Add import workflow to Part 3 (+40 lines)
- [ ] Add maintenance commands to Part 4, Section 17 (+60 lines)
- [ ] Add command aliasing explanation to Part 4 (+30 lines)
- [ ] Add troubleshooting appendix (+100 lines)

**Validation:**
- [ ] Grep for remaining old patterns
- [ ] Run commands from examples
- [ ] Fresh validation run with new setup script
- [ ] Time Part 1 completion (target: <15 min)

---

### Phase 2: Automation (Week 2)

**CI Setup:**
- [ ] Create `scripts/lint-doc-commands.sh`
- [ ] Create `scripts/check-command-coverage.sh`
- [ ] Create `.github/workflows/lint-docs.yml`
- [ ] Create `.markdown-link-check.json` config
- [ ] Test locally
- [ ] Commit and verify CI runs

---

### Phase 3: Process (Week 3)

**Process Setup:**
- [ ] Create/update `CODEOWNERS`
- [ ] Update PR template (add doc impact section)
- [ ] Create `docs/cli-changelog.md` (backfill recent changes)
- [ ] Document tiered review process
- [ ] Communicate new process to team

---

### Phase 4: Ownership (Week 4)

**Ownership Setup:**
- [ ] Assign documentation maintainer (or recruit)
- [ ] Create `docs/docmgr-style-guide.md`
- [ ] Create `docs/tutorial-health.md`
- [ ] Schedule quarterly validation (calendar)
- [ ] Document validation process
- [ ] Update tutorial frontmatter with review metadata

---

## Success Criteria

### Phase 1 Success (Tutorial Fixes):
- ✅ Zero deprecated command patterns in tutorial
- ✅ Zero removed flags in examples
- ✅ Zero path inconsistencies
- ✅ Part 1 completes in <15 minutes (not 30)
- ✅ Validation run: 0 critical, 0 high issues

### Phase 2 Success (Automation):
- ✅ CI runs on every PR (100%)
- ✅ Pattern linting catches known issues
- ✅ <5% false positive rate
- ✅ Average CI time <1 minute

### Phase 3 Success (Process):
- ✅ 90%+ PRs with CLI changes have doc assessment
- ✅ CLI changelog updated within 1 week of changes
- ✅ Doc maintainer reviews 100% of Tier 3/4 PRs

### Phase 4 Success (Ownership):
- ✅ Documentation maintainer assigned
- ✅ Style guide published and used
- ✅ Quarterly validation scheduled and run
- ✅ Tutorial health dashboard updated monthly

---

## Estimated Effort

### Initial (4 weeks):
- Week 1 (Phase 1): 10 hours (tutorial fixes)
- Week 2 (Phase 2): 3 hours (automation)
- Week 3 (Phase 3): 2 hours (process)
- Week 4 (Phase 4): 4 hours (ownership)
**Total: 19 hours**

### Ongoing:
- Per PR: 45 seconds (automated) + 2.6 minutes average (review)
- Per month: 2 hours (micro-reviews, updates)
- Per quarter: 3 hours (full validation)
**Total: ~2-3 hours/month**

### ROI:
- Prevention cost: 2-3 hours/month
- Reactive fixing cost: 20+ hours (this debate + implementation)
- **Break-even: After preventing 1st drift incident**
- **Long-term savings: 10:1+ ROI**

---

## Risk Mitigation

### Risk 1: Automation false positives

**Mitigation:**
- Start with conservative patterns (obvious errors only)
- Iterate based on false positive rate
- Allow override mechanism for valid exceptions

### Risk 2: Process becomes bottleneck

**Mitigation:**
- Tiered review (only 5% of PRs need maintainer)
- Clear escalation criteria
- Async review (doesn't block PR)

### Risk 3: No one wants to be doc maintainer

**Mitigation:**
- Make role attractive (visible impact, recognized work)
- Rotate quarterly or annually
- Provide clear scope (2 hours/month, not full-time)

### Risk 4: Metrics become vanity metrics

**Mitigation:**
- Track user-relevant metrics (time, errors, satisfaction)
- Quarterly validation ensures metrics reflect reality
- Don't optimize metrics, optimize user experience

---

## Dependencies and Blockers

### Blockers:
- None. All decisions made, ready to implement.

### Dependencies:
- Phase 2-4 should wait for Phase 1 completion
- Can parallelize within phases (automation + process can happen together)

---

## Next Actions

**Immediate:**
1. Review this synthesis with team
2. Assign documentation maintainer (if not already decided)
3. Start Phase 1 implementation

**This Week:**
- Complete Phase 1 (tutorial fixes)
- Ship updated tutorial

**Next 3 Weeks:**
- Implement automation (Phase 2)
- Set up process (Phase 3)
- Establish ownership (Phase 4)

---

## Appendix: Command Reference (From Codebase Analysis)

### All docmgr Commands

**Documented in Tutorial:**
- `workspace init/configure/doctor/status`
- `ticket create-ticket/close/list`
- `doc add/search/relate/guidelines/list`
- `task add/check/list`
- `meta update`
- `vocab add/list`
- `changelog update`
- `list tickets/docs`

**Mentioned Briefly:**
- `doc renumber` (Part 4, Section 16)

**Not Documented:**
- `doc layout-fix` (maintenance)
- `doc add` (mentioned but could expand)
- `ticket rename-ticket` (rare)
- `import file` (advanced)
- `template validate` (advanced)
- `config show` (debugging)
- `task edit/remove/uncheck` (everyday but not documented)

**Decision:** Add task editing + maintenance. Skip rare/advanced for now.

---

## Files Created in Debate

1. `reference/02-debate-format-and-candidates.md` — 10 personas
2. `reference/03-debate-questions-REVISED.md` — 12 questions
3. `reference/04-jamie-proposed-question-changes.md` — Technical writer rationale
4. `reference/05-debate-round-1-go-no-go.md` — GO decision
5. `reference/06-debate-round-2-patch-or-restructure.md` — HYBRID approach
6. `reference/07-debate-round-3-severity-triage.md` — Priority ranking
7. `reference/09-debate-round-4-missing-functionality.md` — Feature coverage
8. `reference/08-debate-round-8-error-messages.md` — UX analysis
9. `reference/10-debate-round-12-regression-prevention.md` — Three-layer defense
10. `reference/11-debate-synthesis-and-action-plan.md` — This document

---

## Status

**Debates Complete:** 6 of 12 rounds (Rounds 1-4, 8, 12)  
**Decisions Made:** All critical decisions for Phase 1 implementation  
**Ready to Implement:** Yes

**Remaining rounds (optional):**
- Round 5: Priority (can derive from Round 3 triage)
- Round 6: Duplicates (decided: remove line 528, others have purpose)
- Round 7: Terminology (can be Phase 2 work)
- Round 9: Error UX (covered in Round 8)
- Round 10: Command accuracy scope (decided: tutorial only for now)
- Round 11: Reset script (decided: split into learning/testing)

**Recommendation:** Proceed with implementation. Remaining rounds are refinements, not blockers.

---

**Ready to implement Phase 1. All decisions made, action plan complete, acceptance criteria defined.**

