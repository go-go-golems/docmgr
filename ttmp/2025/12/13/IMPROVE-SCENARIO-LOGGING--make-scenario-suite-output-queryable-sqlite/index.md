---
Title: Make scenario suite output queryable (sqlite)
Ticket: IMPROVE-SCENARIO-LOGGING
Status: active
Topics:
    - testing
    - tooling
    - diagnostics
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: test-scenarios/testing-doc-manager/README.md
      Note: Scenario suite documentation
    - Path: test-scenarios/testing-doc-manager/run-all.sh
      Note: Current harness entrypoint (to be instrumented)
ExternalSources:
    - local:phase-3-scenario-log-2025-12-13.txt
Summary: "Make the integration scenario suite output queryable by logging runs/steps/commands to sqlite with captured stdout/stderr artifacts."
LastUpdated: 2025-12-13T16:15:22.087377798-05:00
---



# Make scenario suite output queryable (sqlite)

## Overview

Turn the `test-scenarios/testing-doc-manager` harness output into a **queryable sqlite run database** (plus per-step artifacts), so diagnosing regressions becomes “run a query” instead of “scroll 1000 lines and pray”.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Design-doc #1 (suite-specific)**: [Scenario suite structured logging (sqlite)](./design-doc/01-scenario-suite-structured-logging-sqlite.md)
- **Design-doc #2 (reusable tool)**: [Generic sqlite scenario logger (Go tool)](./design-doc/02-generic-sqlite-scenario-logger-go-tool.md)
- **Brainstorm / idea bank**: [Brainstorm: scenario logging ideas (wild + useful)](./reference/01-brainstorm-scenario-logging-ideas-wild-useful.md)

## Status

Current status: **active**

## Topics

- testing
- tooling
- diagnostics

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
