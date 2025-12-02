---
Title: Add dynamic flag completion to docmgr using carapace
Ticket: DOCMGR-DYNAMIC-COMPLETION
Status: complete
Topics:
    - backend
    - cli
    - completion
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: carapace/action.go
      Note: Action type definition and composition methods
    - Path: carapace/carapace.go
      Note: Main API for carapace - Gen() and FlagCompletion() methods
    - Path: carapace/compat.go
      Note: Bridge between carapace Actions and cobra's RegisterFlagCompletionFunc
    - Path: carapace/context.go
      Note: Context type providing runtime information during completion
    - Path: carapace/example/cmd/flag.go
      Note: Example usage of FlagCompletion with various flag types
    - Path: carapace/storage.go
      Note: Storage system that maintains per-command registry of completion actions
    - Path: carapace/traverse.go
      Note: Command-line traversal logic that determines what to complete
ExternalSources: []
Summary: Add dynamic flag completion to docmgr using carapace library, enabling context-aware completion for flags like --ticket, --doc-type, --file, etc.
LastUpdated: 2025-12-01T16:01:37.817138709-05:00
---




# Add dynamic flag completion to docmgr using carapace

## Overview

This ticket tracks the implementation of dynamic flag completion for docmgr using the [carapace](https://github.com/carapace-sh/carapace) library. Currently, docmgr uses cobra's built-in completion generation, which produces static completion scripts. By integrating carapace, we can provide dynamic, context-aware completion that:

- Completes `--ticket` with actual ticket IDs from the workspace
- Completes `--doc-type` with values from vocabulary.yaml
- Completes `--topics` and `--status` with vocabulary values
- Completes `--file` paths filtered by ticket context
- Provides intelligent completion based on previously parsed flags

The analysis document in [`reference/01-carapace-dynamic-flag-completion-analysis.md`](./reference/01-carapace-dynamic-flag-completion-analysis.md) provides a comprehensive breakdown of how carapace implements this functionality, including architecture, key components, and integration patterns.

## Key Links

- **Analysis Document**: [`reference/01-carapace-dynamic-flag-completion-analysis.md`](./reference/01-carapace-dynamic-flag-completion-analysis.md) - Comprehensive analysis of carapace's flag completion implementation
- **Related Files**: See frontmatter RelatedFiles field (carapace source files)
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- backend
- cli
- completion

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
