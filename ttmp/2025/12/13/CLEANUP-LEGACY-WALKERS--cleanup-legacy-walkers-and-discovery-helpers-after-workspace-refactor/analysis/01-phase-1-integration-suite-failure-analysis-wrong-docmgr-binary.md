---
Title: Phase 1 integration suite failure analysis (wrong docmgr binary)
Ticket: CLEANUP-LEGACY-WALKERS
Status: active
Topics:
    - refactor
    - tickets
    - docmgr-internals
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: test-scenarios/testing-doc-manager/19-export-sqlite.sh
      Note: Flag mismatch surfaced when running old system docmgr
    - Path: test-scenarios/testing-doc-manager/README.md
      Note: Docs now require pinned DOCMGR_PATH
    - Path: test-scenarios/testing-doc-manager/SCENARIO.md
      Note: Docs now explain how to run with pinned DOCMGR_PATH
    - Path: test-scenarios/testing-doc-manager/run-all.sh
      Note: Hardened harness; requires pinned DOCMGR_PATH
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-13T11:27:20.167996952-05:00
---


# Phase 1 integration suite failure analysis (wrong docmgr binary)

## Purpose

Capture a reproducible explanation for a Phase 1 scenario-suite failure that was initially observed in the Cursor/agent environment, so future work can distinguish **real regressions** from **harness/binary selection issues**.

## Context

- Ticket: `CLEANUP-LEGACY-WALKERS`
- Suite: `test-scenarios/testing-doc-manager/run-all.sh`
- The suite runs inside a small mock **git repo** created by `01-create-mock-codebase.sh` (it does `git init`, commits, etc.), so git-based features are exercised.

## What happened

### Observed failure

The scenario suite failed with:

```text
==> Exporting workspace index sqlite to /tmp/docmgr-scenario-cleanup-phase1-2025-12-13/workspace-index.sqlite
Error: unknown flag: --out
...
Error: unknown flag: --out
```

This came from the export-sqlite scenario step.

### Why this was surprising

The repo-under-test supports `workspace export-sqlite` with the flags the scenario uses. Phase 1 cleanup changes were unrelated to export-sqlite behavior, so a true regression here was unlikely.

## Root cause

The failing run executed the **system** `docmgr` binary from `PATH` (older install) rather than a freshly built binary from this repo. That system binary did not recognize the `--out` flag, so the suite failed even though the repo code supports it.

In other words: this was **not** a cleanup regression; it was a **test harness / binary selection mismatch**.

## Reproduction details

### Failing invocation pattern (ambiguous binary)

```bash
# Uses whatever `docmgr` is on PATH (may be older)
bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario-cleanup-phase1-2025-12-13
```

### Passing invocation pattern (pin the binary under test)

```bash
BIN=/tmp/docmgr-cleanup-phase1
ROOT=/tmp/docmgr-scenario-cleanup-phase1-2025-12-13
LOG=/tmp/docmgr-scenario-cleanup-phase1-2025-12-13.log

go build -o "$BIN" ./cmd/docmgr && \
DOCMGR_PATH="$BIN" bash test-scenarios/testing-doc-manager/run-all.sh "$ROOT" >"$LOG" 2>&1
```

This also keeps console output small while preserving the full run log.

## What we did next

- Re-ran the suite with `DOCMGR_PATH` pinned to a freshly built binary.
- Result: the suite completed successfully. No Phase 1 regression.

## What should be done in the future

- When you see “unknown flag” in scenarios, verify **which `docmgr` binary** is being executed first (PATH vs `DOCMGR_PATH`).
- Tighten the harness contract so it prints the resolved `DOCMGR_PATH` and fails when unset, to prevent silent usage of an older system binary.

## Related files (jump points)

- `test-scenarios/testing-doc-manager/run-all.sh`
- `test-scenarios/testing-doc-manager/01-create-mock-codebase.sh`
- `test-scenarios/testing-doc-manager/19-export-sqlite.sh`
