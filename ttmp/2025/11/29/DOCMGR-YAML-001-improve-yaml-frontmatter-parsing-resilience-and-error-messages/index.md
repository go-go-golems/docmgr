---
Title: Improve YAML frontmatter parsing resilience and error messages
Ticket: DOCMGR-YAML-001
Status: complete
Topics:
    - yaml
    - ux
    - errors
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/commands/doctor.go
      Note: doctor entrypoint
    - Path: pkg/commands/validate_frontmatter.go
      Note: validation verb
    - Path: pkg/doc/docmgr-doctor-validation-workflow.md
      Note: |-
        doctor/validate workflow walkthrough
        workflow doc
    - Path: test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh
      Note: diagnostics smoke
    - Path: test-scenarios/testing-doc-manager/18-validate-frontmatter-smoke.sh
      Note: validation/auto-fix smoke
    - Path: ttmp/2025/11/29/DOCMGR-YAML-001-improve-yaml-frontmatter-parsing-resilience-and-error-messages/design/03-validation-smoke-plan.md
      Note: validation smoke inventory/plan
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-02T13:16:42.289711322-05:00
---





# Improve YAML frontmatter parsing resilience and error messages

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- yaml
- ux
- errors

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
