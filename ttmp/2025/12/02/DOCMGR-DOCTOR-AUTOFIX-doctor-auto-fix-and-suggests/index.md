---
Title: Add auto-fix and suggest-fixes support to doctor
Ticket: DOCMGR-DOCTOR-AUTOFIX
Status: active
Topics:
    - yaml
    - diagnostics
    - ux
DocType: index
Intent: short-term
Owners: []
RelatedFiles:
    - Path: internal/documents/frontmatter.go
      Note: parser/snippets/problem classification
    - Path: pkg/commands/doctor.go
      Note: doctor entry point; add suggest/auto-fix flags and logic
    - Path: pkg/commands/validate_frontmatter.go
      Note: fix generator + auto-fix orchestration to reuse
    - Path: pkg/diagnostics/docmgrctx/frontmatter.go
      Note: taxonomy context with Fixes
    - Path: pkg/diagnostics/docmgrrules/frontmatter_rules.go
      Note: rule rendering with fixes/action links
    - Path: pkg/doc/docmgr-doctor-validation-workflow.md
      Note: doctor/validate workflow walkthrough
    - Path: pkg/frontmatter/frontmatter.go
      Note: quoting/preprocess helpers
    - Path: ttmp/2025/11/29/DOCMGR-YAML-001-improve-yaml-frontmatter-parsing-resilience-and-error-messages/reference/02-frontmatter-healing-and-validation-guide.md
      Note: detailed frontmatter healing guide
    - Path: ttmp/2025/12/02/DOCMGR-DOCTOR-AUTOFIX-doctor-auto-fix-and-suggests/analysis/01-doctor-autofix-context.md
      Note: onboarding context for this ticket
ExternalSources: []
Summary: Add suggest/auto-fix support to doctor, reusing validate-frontmatter heuristics (fix generator, backups) and exposing optional flags for ticket/doctor runs.
LastUpdated: 2025-12-02T12:16:39.788036254-05:00
---




# Add auto-fix and suggest-fixes support to doctor

## Overview
Doctor currently reports frontmatter/schema issues but doesn’t suggest or apply fixes. The `validate frontmatter` verb already supports `--suggest-fixes` and `--auto-fix` (with backups). This ticket tracks adding similar capabilities to doctor, so workspace scans can offer or perform repairs (probably opt-in and scoped).

## Status
Active

## Tasks
See tasks.md

## Key Links
- tasks.md
- changelog.md
- reference/01-context.md

## Changelog
See changelog.md
