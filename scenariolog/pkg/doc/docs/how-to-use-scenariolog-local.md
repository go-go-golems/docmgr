---
Title: How to use scenariolog-local
Slug: how-to-use-scenariolog-local
Short: Build /tmp/scenariolog-local, record runs/steps into sqlite, and query logs with Glazed output (including optional FTS5 search).
Topics:
  - scenariolog
  - sqlite
  - testing
  - diagnostics
  - fts
Commands:
  - scenariolog
  - run
  - exec
  - search
  - summary
  - failures
  - timings
Flags:
  - --db
  - --run-id
  - --root-dir
  - --work-dir
  - --log-dir
  - --kv
  - --output
  - --query
  - --top
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

`scenariolog` is a small sqlite-backed “scenario flight recorder”: it records runs, steps, and captured stdout/stderr as artifacts so you can query failures and search logs after the fact, instead of grepping a giant mixed output stream.

## Build (recommended: enable FTS5)

Build the local binary with FTS5 enabled:

```bash
go -C scenariolog build -tags sqlite_fts5 -o /tmp/scenariolog-local ./cmd/scenariolog
```

## FTS5 degraded mode

If sqlite doesn’t have FTS5 support, `scenariolog init`/migrations still work (runs/steps/artifacts/kv), but:

- `scenariolog search` fails with “FTS not available”.

## Run lifecycle (manual)

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

## Query (Glazed structured output)

All of these commands support Glazed output flags like `--output json|yaml|csv|table`, plus `--fields` and `--sort-columns`.

```bash
/tmp/scenariolog-local summary  --db "$DB" --output json
/tmp/scenariolog-local timings  --db "$DB" --top 10 --output table
/tmp/scenariolog-local failures --db "$DB" --output table
/tmp/scenariolog-local search   --db "$DB" --run-id "$RUN_ID" --query "warning OR error" --limit 20 --output table
```

## Example: docmgr scenario suite integration

The docmgr scenario suite can build and use `scenariolog` automatically.

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


