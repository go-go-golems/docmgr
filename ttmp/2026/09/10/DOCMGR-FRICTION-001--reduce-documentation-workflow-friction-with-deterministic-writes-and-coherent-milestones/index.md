---
Title: Reduce documentation workflow friction with deterministic writes and coherent milestones
Ticket: DOCMGR-FRICTION-001
Status: complete
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
Summary: Implemented stable persistence, recoverable close and milestone projections, CLI/HTTP resume, regression coverage and patched dependencies.
LastUpdated: 2026-09-10T15:58:52.605009986-04:00
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

The requested implementation is complete and validated: stable framing and generated EOFs, recoverable close/milestones, derived resume, CLI/HTTP parity and typed frontend API hooks. Code checkpoints: `91e0603`, `582171f`, `7de2ea4`, `b1f53fc`, `1516abc`, `1d0710d`. See the detailed diary and `sources/implementation-validation.json` for evidence and limits. Research captures remain historical; new tests establish current behavior. The locally built CLI is `/tmp/docmgr-friction-local`; the global installed binary was not replaced.

Companion: **SKILLS-FRICTION-001**, rooted at `/home/manuel/.pi/agent/skills/ttmp/2026/09/10/SKILLS-FRICTION-001--make-skills-phase-aware-proportional-and-consistent-across-long-sessions`. It owns instruction conflicts, diary proportionality and resume/validation conventions.

## Delivery

The original research delivery (not a new implementation upload) used the reMarkable bundle destination `/ai/2026/09/10/DOCMGR-FRICTION-001`. Dry-run, upload and verification receipts are retained under `sources/`; the upload result is authoritative for delivery status. The bundle contains this guide and its investigation diary, with rendered diagrams.
