---
Title: Reduce documentation workflow friction with deterministic writes and coherent milestones
Ticket: DOCMGR-FRICTION-001
Status: active
Topics:
    - docmgr
    - cli
    - usability
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://internal/documents/write.go
      Note: Permission-preserving no-op writer, commit 91e0603
ExternalSources: []
Summary: Research delivered; implementation remains open for stable persistence, recoverable close, explicit milestone identities and resume views.
LastUpdated: 2026-09-10T15:37:22Z
WhatFor: Turn verified Video Observatory workflow friction into implementable docmgr improvements.
WhenToUse: Starting the implementation phases or reviewing session evidence.
---

# Docmgr workflow friction

The research reproduces accumulating frontmatter whitespace, noncanonical changelog EOF and partial ticket closure after a changelog write error. The guide maps existing CLI, persistence, workspace, task and HTTP APIs before proposing changes.

## Documents

- [Intern analysis, design and implementation guide](design-doc/01-intern-guide-to-deterministic-document-mutation-and-milestone-workflows.md)
- [Investigation diary](reference/01-investigation-diary.md)
- [Tasks](tasks.md) and [changelog](changelog.md)
- [Write reproduction](sources/write-reproduction.json) and [close reproduction](sources/close-reproduction.json)

## Scope and status

The requested research documents are complete. Product changes are not implemented; the ticket remains active with explicit implementation phases. Reproduction scripts only mutate temporary fixtures. No production Go code was changed.

Companion: **SKILLS-FRICTION-001**, rooted at `/home/manuel/.pi/agent/skills/ttmp/2026/09/10/SKILLS-FRICTION-001--make-skills-phase-aware-proportional-and-consistent-across-long-sessions`. It owns instruction conflicts, diary proportionality and resume/validation conventions.

## Delivery

The reMarkable bundle destination is `/ai/2026/09/10/DOCMGR-FRICTION-001`. Dry-run, upload and verification receipts are retained under `sources/`; the upload result is authoritative for delivery status. The bundle contains this guide and its investigation diary, with rendered diagrams.
