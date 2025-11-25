---
Title: 'Fix how-to-use tutorial: new verb structure'
Ticket: DOCMGR-DOC-VERBS
Status: active
Topics:
    - docmgr
    - documentation
    - cli
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: cmd/docmgr/cmds/root.go
      Note: CLI command registration point (where verb changes happen)
    - Path: pkg/commands/relate.go
      Note: Source of no changes specified error (line 471)
    - Path: ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/02-debate-format-and-candidates.md
      Note: Debate personas (4 human devs
    - Path: ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/03-debate-questions.md
      Note: 10 debate questions for tutorial quality review
    - Path: ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/04-jamie-proposed-question-changes.md
      Note: Technical writer's proposed debate question changes (2 replacements
    - Path: ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/05-debate-round-1-go-no-go.md
      Note: Round 1 - Unanimous GO decision to fix tutorial
    - Path: ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/06-debate-round-2-patch-or-restructure.md
      Note: Round 2 - HYBRID approach (patch now
    - Path: ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/07-debate-round-3-severity-triage.md
      Note: Round 3 - Severity triage (2 CRITICAL
    - Path: ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/08-debate-round-8-error-messages.md
      Note: Round 8 - Error messages with codebase analysis (fix reset script + troubleshooting section)
    - Path: ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/09-debate-round-4-missing-functionality.md
      Note: Round 4 (REPLACEMENT) - Missing functionality analysis (add task editing + maintenance commands)
    - Path: ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/10-debate-round-12-regression-prevention.md
      Note: Round 12 - Three-layer prevention strategy (automation + process + ownership)
    - Path: ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/11-debate-synthesis-and-action-plan.md
      Note: Complete synthesis of all debate rounds with 4-phase action plan
    - Path: ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/12-complete-cli-command-inventory.md
      Note: Complete CLI command tree with aliasing structure (10 undocumented commands found)
    - Path: ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/13-jamie-commentary-on-glazed-style-guide.md
      Note: Technical writer analysis of Glazed docs standards (5 violations found)
    - Path: ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/14-final-comprehensive-action-plan.md
      Note: Final action plan with all 41 findings (validation + CLI + style guide)
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-19T13:58:24.683790459-05:00
---









---
Title: Fix how-to-use tutorial: new verb structure
Ticket: DOCMGR-DOC-VERBS
Status: draft
Topics:
  - docmgr
  - documentation
  - cli
DocType: index
Intent: short-term
Owners:
  - manuel
RelatedFiles: []
ExternalSources: []
Summary: >
  
LastUpdated: 2025-11-19
---

# Fix how-to-use tutorial: new verb structure

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- docmgr
- documentation
- cli

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
