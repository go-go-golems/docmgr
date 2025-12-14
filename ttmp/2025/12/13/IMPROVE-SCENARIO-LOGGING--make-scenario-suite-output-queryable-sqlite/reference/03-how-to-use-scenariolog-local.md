---
Title: How to use scenariolog-local
Ticket: IMPROVE-SCENARIO-LOGGING
Status: active
Topics:
    - testing
    - tooling
    - diagnostics
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: scenariolog/README.md
      Note: Build instructions + FTS5 enablement/fallback notes
    - Path: scenariolog/cmd/scenariolog/main.go
      Note: Cobra entrypoint + exec/run commands
    - Path: scenariolog/cmd/scenariolog/glazed_cmds.go
      Note: Glazed query/report commands (search/summary/failures/timings)
    - Path: test-scenarios/testing-doc-manager/run-all.sh
      Note: Scenario harness integration (builds scenariolog + wraps steps)
    - Path: test-scenarios/testing-doc-manager/README.md
      Note: Scenario suite docs (mentions scenariolog usage)
ExternalSources:
    - local:glaze-help-how-to-write-good-documentation-pages-2025-12-14.txt
    - local:glaze-help-writing-help-entries-2025-12-14.txt
Summary: "Copy/paste guide for building and using /tmp/scenariolog-local: run lifecycle, step wrapping, Glazed query output, and FTS5/degraded-mode behavior."
LastUpdated: 2025-12-13T19:08:22.030701209-05:00
---

# How to use scenariolog-local

## Goal

Give a copy/paste-ready workflow for using `scenariolog` as a local “scenario flight recorder” binary (`/tmp/scenariolog-local`): build it, run scenarios, and query the resulting sqlite DB quickly.

## Context

This content is now also available from the `scenariolog` binary itself via the Glazed help system:

```bash
scenariolog help how-to-use-scenariolog-local
```

`scenariolog` turns “a big mixed stdout/stderr dump” into a sqlite database you can query. It records:

- `scenario_runs` (one row per run)
- `steps` (one row per step/script)
- `artifacts` (stdout/stderr files with hashes and sizes)
- `kv` (metadata tags)
- optional `log_lines_fts` for fast log search (FTS5)

The most common consumer is the docmgr integration scenario suite:

- `test-scenarios/testing-doc-manager/run-all.sh`

## Quick Reference

### Build (recommended: enable FTS5)

From the repo root:

```bash
go -C scenariolog build -tags sqlite_fts5 -o /tmp/scenariolog-local ./cmd/scenariolog
```

### FTS5 degraded mode

If sqlite doesn’t have FTS5 support, `scenariolog init`/migrations still work (runs/steps/artifacts/kv), but:

- `scenariolog search` will fail with a clear “FTS not available” error.

### Run lifecycle (manual)

```bash
DB=/tmp/scenario/.scenario-run.db
ROOT=/tmp/scenario
mkdir -p "$ROOT" "$ROOT/.logs"

/tmp/scenariolog-local init --db "$DB"
RUN_ID=$(/tmp/scenariolog-local run start --db "$DB" --root-dir "$ROOT" --suite demo --kv env:local --kv build_id:123)

# Wrap a step
/tmp/scenariolog-local exec \
  --db "$DB" \
  --run-id "$RUN_ID" \
  --root-dir "$ROOT" \
  --work-dir "$(pwd)" \
  --log-dir ".logs" \
  --step-num 1 \
  --name "demo-step" \
  --script-path "./some-script.sh" \
  --kv step_kind:demo \
  -- bash --noprofile --norc -c 'echo out; echo err 1>&2; exit 0'

/tmp/scenariolog-local run end --db "$DB" --run-id "$RUN_ID" --exit-code 0
```

### Query (Glazed structured output)

```bash
/tmp/scenariolog-local summary  --db "$DB" --output json
/tmp/scenariolog-local timings  --db "$DB" --top 10 --output table
/tmp/scenariolog-local failures --db "$DB" --output table
/tmp/scenariolog-local search   --db "$DB" --run-id "$RUN_ID" --query "warning OR error" --limit 20 --output table
```

## Usage Examples

### Example: run the docmgr scenario suite with automatic logging

```bash
cd test-scenarios/testing-doc-manager/
go build -o /tmp/docmgr-scenario-local ../../cmd/docmgr
export DOCMGR_PATH=/tmp/docmgr-scenario-local

# Optional: pin scenariolog; if unset, run-all.sh builds it to /tmp/scenariolog-local
# go -C ../../scenariolog build -tags sqlite_fts5 -o /tmp/scenariolog-local ./cmd/scenariolog
# export SCENARIOLOG_PATH=/tmp/scenariolog-local

./run-all.sh /tmp/docmgr-scenario
```

Then query:

```bash
DB=/tmp/docmgr-scenario/.scenario-run.db
RUN_ID=$(sqlite3 "$DB" "SELECT run_id FROM scenario_runs ORDER BY started_at DESC LIMIT 1;")

/tmp/scenariolog-local summary --db "$DB" --output table
/tmp/scenariolog-local timings --db "$DB" --top 5 --output table
/tmp/scenariolog-local search --db "$DB" --run-id "$RUN_ID" --query "warning" --limit 10 --output table
```

### Example: find the first failing step quickly

```bash
DB=/tmp/docmgr-scenario/.scenario-run.db
/tmp/scenariolog-local failures --db "$DB" --output table
```

## Related

- `design-doc/02-generic-sqlite-scenario-logger-go-tool.md`
- `design-doc/03-implementation-plan-scenariolog-mvp-kv-artifacts-fts-glazed-cli.md`
