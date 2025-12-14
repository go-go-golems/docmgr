---
Title: 'Test Playbook: Contestant #2 Ingrid the Indexer (Index Builder / InitIndex)'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/workspace/index_builder.go
      Note: InitIndex + ingestion implementation under test
    - Path: internal/workspace/index_builder_test.go
      Note: End-to-end ingestion unit test
    - Path: internal/workspace/sqlite_schema.go
      Note: Schema creation used by ingestion
    - Path: internal/workspace/sqlite_schema_test.go
      Note: Schema smoke test
    - Path: test-scenarios/testing-doc-manager/19-export-sqlite.sh
      Note: Scenario-level export-sqlite smoke test
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T20:08:04.520831536-05:00
---


# Test Playbook: Contestant #2 Ingrid the Indexer (Index Builder / InitIndex)

## Purpose

This playbook validates Workspace index ingestion (“Ingrid the Indexer”):

- `Workspace.InitIndex` sets up an in-memory SQLite DB
- schema is created
- document walk runs once
- docs are inserted with correct parse state
- topics and RelatedFiles are ingested
- skip policy is applied (cross-checks contestant #1)

This is the closest thing we have today to an end-to-end proof of the refactor’s core invariants, without yet porting “normal verbs” like `doc search`.

## Environment Assumptions

- You are in the `docmgr` repo working tree.
- You have Go installed.
- You can run tests that use the sqlite3 driver.

This playbook uses an isolated temporary workspace layout created by the unit test (no external fixtures required).

## Commands

```bash
cd /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr && \
go test ./internal/workspace -run TestCreateWorkspaceSchema_InMemory -count=1 -v && \
go test ./internal/workspace -run TestWorkspaceInitIndex_IngestsDocsTopicsAndRelatedFiles -count=1 -v
```

## Exit Criteria

- Both tests exit 0.

### What `TestCreateWorkspaceSchema_InMemory` proves

- `openInMemorySQLite` works
- `createWorkspaceSchema` creates the key tables:
  - `docs`
  - `doc_topics`
  - `related_files`

### What `TestWorkspaceInitIndex_IngestsDocsTopicsAndRelatedFiles` proves (high signal)

The test constructs a minimal but realistic ticket layout under a temporary `ttmp/` and asserts:

- **Skip policy actually applied**:
  - a doc under `ticket/.meta/` is not indexed
  - a doc under `ttmp/_guidelines/` is not indexed
- **Parse error docs are indexed (repair-friendly)**:
  - malformed frontmatter doc is present in `docs`
  - `parse_ok=0`
  - `parse_err` is non-empty
- **Control-doc tagging works**:
  - `tasks.md` at ticket root is `is_control_doc=1` because `index.md` exists
- **RelatedFiles normalization envelope exists**:
  - row exists in `related_files` for `raw_path='backend/main.go'`
  - normalization keys are populated as expected:
    - `norm_canonical` and `norm_repo_rel` are `backend/main.go`
    - `norm_docs_rel` is empty (because file is not under docs root)
    - `norm_doc_rel` is non-empty and ends with `backend/main.go`
    - `norm_abs` ends with `/backend/main.go`
    - `norm_clean` is `backend/main.go`
    - `anchor='repo'`
- **Topic lowercasing invariant**:
  - topics `[a, B]` in frontmatter produce `topic_lower='b'` exactly once

## Notes

### Why this is the best “current state” test for Ingrid

As of the current integration status (per the review guide / diary), the new Workspace backend is not yet wired into most user-facing commands. That means the most reliable way to judge ingestion quality today is:

- this unit test (fast, deterministic, isolated), and
- the `workspace export-sqlite` scenario smoke (next section).

### Optional integration smoke: export the DB and inspect it

The scenario suite includes:

- `test-scenarios/testing-doc-manager/19-export-sqlite.sh`

You can run the full suite against a local binary:

```bash
cd /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr && \
go build -o /tmp/docmgr-local ./cmd/docmgr && \
DOCMGR_PATH=/tmp/docmgr-local bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario-local
```

This specifically validates that:

- index can be built for a realistic mock workspace
- export uses `VACUUM INTO` and produces a non-empty sqlite file
- `README` table is present and contains embedded docs
