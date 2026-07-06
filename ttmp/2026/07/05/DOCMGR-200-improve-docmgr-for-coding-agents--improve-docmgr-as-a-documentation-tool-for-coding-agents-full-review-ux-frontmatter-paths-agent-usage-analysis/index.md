---
Title: 'Improve docmgr as a documentation tool for coding agents: full review, UX, frontmatter paths, agent-usage analysis'
Ticket: DOCMGR-200-improve-docmgr-for-coding-agents
Status: active
Topics:
    - docmgr
    - ux
    - cli
    - documentation
    - tooling
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Full evidence-based review of docmgr plus empirical mining of 240 coding-agent sessions (go-minitrace) to design improvements for agent-facing UX, frontmatter path handling, doctor, UI parity, and an LLM-in-docmgr subsystem. Companion deliverable - a go-minitrace field report.
LastUpdated: 2026-07-05T18:28:00.646175498-04:00
WhatFor: ""
WhenToUse: ""
---

# Improve docmgr as a documentation tool for coding agents: full review, UX, frontmatter paths, agent-usage analysis

## Overview

Investigation ticket for making docmgr a first-class documentation tool for coding agents. Two deliverables:

1. `design-doc/01-improving-docmgr-for-coding-agents-analysis-design-and-implementation-guide.md` — intern-ready guide: architecture tour, evidence from four code reviews plus go-minitrace mining of 139 sessions / 14,166 docmgr tool calls, gap analysis, design (path anchors, agent CLI contract, doctor overhaul, UI parity, `docmgr ai`), phased plan.
2. `analysis/01-go-minitrace-field-report-...md` — field report and assessment of go-minitrace itself from this project's usage (CLI ergonomics, JS API, schema, adapter fidelity).

Working evidence lives in `reference/01-investigation-diary.md` (chronological, includes two live bug reproductions), `scripts/` (corpus tooling + JS query commands), and `sources/` (saved analysis JSON).

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- docmgr
- ux
- cli
- documentation
- tooling

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
