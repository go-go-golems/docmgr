---
Title: External postfix templates for verb outputs
Ticket: DOCMGR-OUTPUT-TEMPLATES
Status: complete
Topics:
    - cli
    - templates
    - glaze
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/cmd/docmgr/cmds/template/
      Note: Template command structure
    - Path: docmgr/pkg/commands/guidelines_cmd.go
      Note: Added schema flags and postfix template rendering for guidelines verb
    - Path: docmgr/pkg/commands/search.go
      Note: Added schema flags and postfix template rendering for search verb
    - Path: docmgr/pkg/commands/status.go
      Note: Added schema flags and postfix template rendering for status verb
    - Path: docmgr/pkg/commands/tasks.go
      Note: Added schema flags and postfix template rendering for tasks list verb
    - Path: docmgr/pkg/commands/template_validate.go
      Note: Template validation command - checks syntax and undefined functions
    - Path: docmgr/pkg/commands/vocab_list.go
      Note: Added schema flags and postfix template rendering for vocab list verb
    - Path: docmgr/pkg/doc/docmgr-how-to-use.md
      Note: Added documentation for --print-template-schema flag in Template Schema Discovery section
    - Path: docmgr/test-scenarios/testing-doc-manager/13-template-schema-output.sh
      Note: Integration test verifying --print-template-schema outputs only schema (no human output) for all templated verbs
    - Path: docmgr/test-scenarios/testing-doc-manager/run-all.sh
      Note: Added template schema output test to test suite
    - Path: docmgr/ttmp/2025/11/19/DOCMGR-OUTPUT-TEMPLATES-external-postfix-templates-for-verb-outputs/playbooks/01-intern-playbook-continuing-external-templates.md
      Note: Updated playbook with lessons learned and improved guidance
    - Path: docmgr/ttmp/templates/doc/guidelines.templ
      Note: Example template file for guidelines command
    - Path: docmgr/ttmp/templates/doc/search.templ
      Note: Example template file for search command
    - Path: docmgr/ttmp/templates/status.templ
      Note: Example template file for status command
    - Path: docmgr/ttmp/templates/tasks/list.templ
      Note: Example template file for tasks list command
    - Path: docmgr/ttmp/templates/vocab/list.templ
      Note: Example template file for vocab list command
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-01T16:01:37.91252771-05:00
---










# External postfix templates for verb outputs

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- cli
- templates
- glaze

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
