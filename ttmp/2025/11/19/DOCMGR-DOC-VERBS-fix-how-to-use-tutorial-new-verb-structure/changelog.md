# Changelog

## 2025-11-19

- Initial workspace created


## 2025-11-19

Audit CLI verbs and added reference mapping; added task to update tutorial

### Related Files

- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/doc/docmgr-how-to-use.md — Source tutorial to update (verbs outdated)


## 2025-11-25

Completed Rounds 1-4, 8, 12 of presidential debate on tutorial quality. Decisions: GO (fix it), HYBRID (patch then restructure), TRIAGE (2 CRITICAL, 2 HIGH, 9 MEDIUM), FUNCTIONALITY (add task editing + maintenance), ERROR UX (split script + troubleshooting), PREVENTION (three-layer defense).


## 2025-11-25

Discovered shell completion command and command aliasing structure (docmgr init ↔ workspace init). Analyzed Glazed documentation style guide—found 5 violations in our tutorial (missing topic intros, jargon before definition, wordiness, inconsistent terms, complex examples). Added to Phase 1: topic-focused intros + move glossary earlier.


## 2025-11-25

Phase 1: tutorial fixes, helper scripts split, and troubleshooting appendix

### Related Files

- docmgr/pkg/doc/docmgr-how-to-use.md — Implemented all Phase 1 edits (new glossary placement
- docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/playbook/01-beginner-tutorial-validation-checklist.md — Checklist now references setup + validation scripts.
- docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/14-final-comprehensive-action-plan.md — Marked Phase 1 tasks complete and aligned plan with latest work.
- docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/script/01-docmgr-tutorial-validation-run.md — Documented new helper scripts.
- docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/script/setup-practice-repo.sh — Added lightweight setup script for hands-on practice.
- docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/script/validate-tutorial.sh — Renamed reset helper and verified iterative workflow.


## 2025-11-25

Phase 2: documented layout-fix/config-show/renumber maintenance workflows

### Related Files

- docmgr/pkg/doc/docmgr-how-to-use.md — Added maintenance sections covering layout-fix


## 2025-11-25

Phase 2: documented import-file workflows and data flow

### Related Files

- docmgr/pkg/doc/docmgr-how-to-use.md — Added Section 14 with import file use cases


## 2025-11-25

Phase 2: audited and removed duplicate content (Appendix B fragment)

### Related Files

- docmgr/pkg/doc/docmgr-how-to-use.md — Removed doctor output fragment from Appendix B; verified no other substantive duplicates remain.


## 2025-11-25

Phase 2: extracted advanced content to separate guides (247 lines removed)

### Related Files

- docmgr/pkg/doc/docmgr-advanced-workflows.md — Created new guide covering import/root-config/layout-fix/multi-repo.
- docmgr/pkg/doc/docmgr-ci-automation.md — Added Glaze field contracts and automation patterns from tutorial.
- docmgr/pkg/doc/docmgr-how-to-use.md — Moved Glaze/CI/import/root-config to other docs; tutorial now 999 lines (was 1246).


## 2025-12-01

Auto-closed: ticket was active but not created today

