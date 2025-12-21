---
Title: Add docmgr skill create
Ticket: 006-SKILL-CREATE
Status: active
Topics:
    - skills
    - cli
    - ux
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/docmgr/cmds/skill/list.go
      Note: Cobra wiring for skill list
    - Path: cmd/docmgr/cmds/skill/show.go
      Note: Cobra wiring for skill show
    - Path: cmd/docmgr/cmds/skill/skill.go
      Note: Cobra command group wiring for skill subcommands
    - Path: pkg/commands/skill_list.go
      Note: Current skill list UX + filtering logic (active tickets by default)
    - Path: pkg/commands/skill_show.go
      Note: Current skill show UX + ticket scoping + matching
    - Path: test-scenarios/testing-doc-manager/20-skills-smoke.sh
      Note: End-to-end skills scenario suite; extend for skill create
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-21T08:27:16.214712144-05:00
WhatFor: ""
WhenToUse: ""
---


# Add docmgr skill create

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- skills
- cli
- ux

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
