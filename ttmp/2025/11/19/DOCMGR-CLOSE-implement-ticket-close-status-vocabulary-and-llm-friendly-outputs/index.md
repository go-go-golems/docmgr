---
Title: 'IMPL: Implement ticket close, status vocabulary, and LLM-friendly outputs'
Ticket: DOCMGR-CLOSE
Status: active
Topics:
    - docmgr
    - workflow
    - ux
    - automation
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/cmd/docmgr/cmds/tasks/check.go
      Note: Enabled dual-mode output
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/cmd/docmgr/cmds/ticket/close.go
      Note: CLI wiring for ticket close
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/cmd/docmgr/cmds/ticket/ticket.go
      Note: Registered ticket close command
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/doctor.go
      Note: Added status vocabulary validation
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/init.go
      Note: Added status vocabulary seeding
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/tasks.go
      Note: Added all_tasks_done suggestion and structured output
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/ticket_close.go
      Note: New ticket close command implementation
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/vocab_add.go
      Note: Added status category support
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/vocab_list.go
      Note: Added status category listing
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/vocabulary.go
      Note: Updated LoadVocabulary for Status
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/models/document.go
      Note: Added Status field to Vocabulary struct
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/ttmp/vocabulary.yaml
      Note: Added status vocabulary entries
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-19T14:58:36.016944946-05:00
---


# Implement ticket close, status vocabulary, and LLM-friendly outputs

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- docmgr
- workflow
- ux
- automation

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

## References

- Debate Synthesis — Closing Workflow, Status/Intent, and LLM UX  
  docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/reference/06-debate-synthesis-closing-workflow-status-intent-and-llm-ux.md
- Debate Round 1 — Workflow Friction  
  docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/12-debate-round-1-workflow-friction.md
- Debate Round 2 — New Verbs and Command Patterns  
  docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/13-debate-round-2-new-verbs-and-command-patterns.md
- Debate Round 3 — Status and Intent Lifecycle Transitions  
  docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/14-debate-round-3-status-and-intent-lifecycle-transitions.md
- Debate Round 4 — Automation vs Manual  
  docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/15-debate-round-4-automation-vs-manual.md
- Debate Round 5 — LLM Usage Patterns  
  docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/analysis/16-debate-round-5-llm-usage-patterns.md
- Debate Format and Candidates — Workflow Improvements  
  docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/reference/04-debate-format-and-candidates-workflow-improvements.md
- Debate Questions — Workflow Improvements  
  docmgr/ttmp/2025/11/18/DOCMGR-CODE-REVIEW-code-review-docmgr-codebase/reference/05-debate-questions-workflow-improvements.md
- CLI verbs mapping (old vs new)  
  docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/reference/01-cli-verbs-mapping-old-vs-new.md
