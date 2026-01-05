---
Title: Modularize Web UI widgets (SearchPage extraction)
Ticket: 007-MODULARIZE-UI-WIDGETS
Status: active
Topics:
    - ui
    - web
    - ux
    - docmgr
    - refactor
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ui/src/App.tsx
      Note: CSS import split; route entrypoint using shared+search styles.
    - Path: ui/src/components/ApiErrorAlert.tsx
      Note: Shared error alert primitive
    - Path: ui/src/components/PathHeader.tsx
      Note: Shared primitive for path+actions; used by Search preview
    - Path: ui/src/components/RelatedFilesList.tsx
      Note: Shared related-files list widget
    - Path: ui/src/features/search/SearchPage.tsx
      Note: |-
        Refactor target; orchestrates Search UI and now consumes extracted leaf widgets.
        Search is now composed from ui/src/features/search/widgets/*
    - Path: ui/src/features/search/components/DiagnosticList.tsx
      Note: Extracted diagnostics renderer widget (cards + details).
    - Path: ui/src/features/search/components/MarkdownSnippet.tsx
      Note: Extracted snippet markdown renderer with query-term highlighting.
    - Path: ui/src/features/search/components/TopicMultiSelect.tsx
      Note: Extracted topics token input widget.
    - Path: ui/src/features/search/hooks/useIsMobile.ts
      Note: Extracted responsive hook from SearchPage (mobile breakpoint logic).
    - Path: ui/src/features/search/hooks/useSearchSelection.ts
      Note: Extracted selection model (selected/index/apply desired selection) from SearchPage.
    - Path: ui/src/features/search/hooks/useSearchUrlSync.ts
      Note: Extracted URL read/write sync (mode/q/filters/sel/preview) from SearchPage.
    - Path: ui/src/features/search/widgets/SearchDocsResults.tsx
      Note: Docs results list widget
    - Path: ui/src/features/search/widgets/SearchFiltersDesktop.tsx
      Note: Desktop filter panel widget
    - Path: ui/src/features/search/widgets/SearchPreviewModal.tsx
      Note: Mobile preview modal widget
    - Path: ui/src/lib/apiError.ts
      Note: Shared helper for docmgr API error envelope parsing (used by SearchPage error banner).
    - Path: ui/src/lib/clipboard.ts
      Note: Extracted clipboard helper; SearchPage uses for copy-path.
    - Path: ui/src/lib/time.ts
      Note: Extracted SearchPage timeAgo helper into shared lib (no behavior change for Search).
    - Path: ui/src/styles/design-system.css
      Note: Shared UI styling (DocCard/result cards
    - Path: ui/src/styles/search.css
      Note: Search-page-specific layout styling (container
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-05T08:49:54.386444656-05:00
WhatFor: ""
WhenToUse: ""
---








# Modularize Web UI widgets (SearchPage extraction)

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- ui
- web
- ux
- docmgr
- refactor

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
