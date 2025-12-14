# Phase 2 Integration Suite Failure Analysis — `DOCMGR_PATH` points to a non-existent binary

## Summary

While running the integration scenario suite after completing Phase 2 (tasks `[5]–[7]`), the suite failed immediately because `DOCMGR_PATH` was set to a **path that does not exist**.

This is not a functional regression in `docmgr`; it’s a runner error: `DOCMGR_PATH` must point to a real executable binary.

## Symptoms

The scenario suite starts normally, then fails on the first script that invokes `docmgr`:

- `./02-init-ticket.sh: line 11: .../docmgr: No such file or directory`

## Root cause

I ran:

```bash
make build && DOCMGR_PATH="$(pwd)/docmgr" bash test-scenarios/testing-doc-manager/run-all.sh
```

But `make build` is:

- `go generate ./...`
- `go build ./...`

…which **does not produce a `./docmgr` binary** at the repo root.

So `DOCMGR_PATH="$(pwd)/docmgr"` was invalid, and every scenario script that tried to execute it failed with “No such file or directory”.

## Fix

Build an explicit binary for the CLI entrypoint and point `DOCMGR_PATH` at it:

```bash
go build -o /tmp/docmgr-scenario-local ./cmd/docmgr
DOCMGR_PATH=/tmp/docmgr-scenario-local bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario
```

This matches the documented workflow in:

- `test-scenarios/testing-doc-manager/README.md`
- `test-scenarios/testing-doc-manager/SCENARIO.md`

## What should be done in the future

- Consider adding a **Makefile target that produces a predictable CLI binary path** for local testing (for example, `make build-cli` -> `./dist/docmgr`) to reduce this class of runner error.
- When a scenario fails before executing `docmgr` logic, always sanity-check:
  - `ls -l "${DOCMGR_PATH}"`, and
  - `file "${DOCMGR_PATH}"` / `--version`.


