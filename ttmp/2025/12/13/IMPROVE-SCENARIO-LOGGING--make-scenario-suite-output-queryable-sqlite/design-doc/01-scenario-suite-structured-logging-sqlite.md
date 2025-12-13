---
Title: Scenario suite structured logging (sqlite)
Ticket: IMPROVE-SCENARIO-LOGGING
Status: active
Topics:
    - testing
    - tooling
    - diagnostics
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-13T16:13:24.509538838-05:00
---

# Scenario suite structured logging (sqlite)

## Executive Summary

The `test-scenarios/testing-doc-manager` integration suite produces 40KB+ of verbose output mixing test execution, command output, and validation results. This design proposes capturing scenario execution in a queryable **sqlite database** with structured metadata (runs, steps, commands, exit codes, timings) and per-step `stdout`/`stderr` files, enabling fast diagnosis of failures without manual log scanning.

**Value proposition**: Transform a 1000+ line text dump into a database you can query in milliseconds to answer: "Which commands failed?", "What warnings appeared?", "How long did each step take?"

## Problem Statement

### Current State

The scenario harness (`run-all.sh`) executes 19+ bash scripts sequentially, each invoking multiple `docmgr` commands. Output is captured to a single text stream that mixes:

1. **Harness control flow**: `[ok]` / `[info]` / `[fail]` markers from the harness itself
2. **Command stdout**: Human-readable `docmgr` output (tables, summaries, reminders)
3. **Command stderr**: Errors, warnings, diagnostics (sometimes interleaved)
4. **Validation noise**: Doctor warnings that are intentionally induced then fixed

**Pain points** observed during Phase 1–3 migrations:

- **Missed regressions**: Critical failures are buried in 1000+ lines of output. Example: `Error: unknown flag: --out` (Phase 1) was visible but hard to spot; we initially didn't notice which binary was under test.
- **No structured search**: Finding "all warnings" or "commands that failed" requires manual `rg` invocations with regex patterns.
- **No timing data**: Can't measure which steps are slow or regressing in performance.
- **Lost context**: When a command fails, its surrounding context (env vars, working directory, prior commands) isn't easily available.
- **No diffing**: Can't compare two scenario runs to detect behavior changes.

### Motivating Example

After the Phase 3 run, we had to manually:

```bash
LOG="agent-tools/03b2ac14-e8a7-4f45-8d0c-06bcbc6f5481.txt"
rg -n -i "\b(error|fail)" "$LOG" | head -n 80
rg -n -i "\b(warn)" "$LOG" | head -n 80
tail -n 30 "$LOG"
```

to determine if failures were real or expected (they were expected).

With structured logging, this becomes:

```sql
-- Show all failed commands
SELECT step_name, command, exit_code, stderr_path 
FROM commands WHERE exit_code != 0;

-- Show all warnings
SELECT step_name, command, stdout_path 
FROM commands WHERE stdout LIKE '%warning%' OR stderr LIKE '%warning%';

-- Timing breakdown
SELECT step_name, duration_ms, exit_code FROM steps ORDER BY duration_ms DESC LIMIT 10;
```

## Proposed Solution

### Architecture Overview

1. **Structured logging during execution**: Each scenario step writes metadata to a **sqlite database** (`$ROOT_DIR/.scenario-run.db`) and captures `stdout`/`stderr` to separate files (`$ROOT_DIR/.logs/step-NN-stdout.txt`, etc.).

2. **Minimal harness changes**: Wrap each script invocation in a lightweight bash helper that:
   - Records step start/end times
   - Captures exit code
   - Writes a row to the `steps` table
   - For individual `docmgr` commands within steps, optionally instrument with a wrapper function to log to `commands` table

3. **Query interface**: Post-run, use `sqlite3` CLI or a small Go helper (`docmgr-scenario-query`) to analyze failures, timings, and warnings.

### SQLite Schema

```sql
-- scenario_runs: one row per full run-all.sh invocation
CREATE TABLE scenario_runs (
    run_id TEXT PRIMARY KEY,           -- UUID or timestamp-based (e.g., "2025-12-13T16:09:42-run-abc123")
    root_dir TEXT NOT NULL,            -- /tmp/docmgr-scenario
    docmgr_path TEXT NOT NULL,         -- /tmp/docmgr-local or $(which docmgr)
    docmgr_version TEXT,               -- from "docmgr --version" if available
    started_at TEXT NOT NULL,          -- ISO8601 timestamp
    completed_at TEXT,                 -- ISO8601 or NULL if incomplete
    exit_code INTEGER,                 -- 0=success, nonzero=failure
    duration_ms INTEGER,               -- total run duration
    notes TEXT                         -- optional freeform notes (git hash, branch, etc.)
);

-- steps: one row per scenario script (01-create-mock-codebase.sh, 02-init-ticket.sh, etc.)
CREATE TABLE steps (
    step_id TEXT PRIMARY KEY,          -- "<run_id>-step-<NN>" (e.g., "run-abc123-step-02")
    run_id TEXT NOT NULL,              -- FK to scenario_runs
    step_num INTEGER NOT NULL,         -- 1, 2, 3, ...
    step_name TEXT NOT NULL,           -- "01-create-mock-codebase", "02-init-ticket", etc.
    script_path TEXT NOT NULL,         -- ./01-create-mock-codebase.sh
    started_at TEXT NOT NULL,
    completed_at TEXT,
    exit_code INTEGER,
    duration_ms INTEGER,
    stdout_path TEXT,                  -- relative path: .logs/step-02-stdout.txt
    stderr_path TEXT,                  -- relative path: .logs/step-02-stderr.txt
    FOREIGN KEY (run_id) REFERENCES scenario_runs(run_id)
);

-- commands: individual docmgr invocations within a step (optional fine-grain tracking)
CREATE TABLE commands (
    command_id TEXT PRIMARY KEY,       -- "<step_id>-cmd-<NN>"
    step_id TEXT NOT NULL,             -- FK to steps
    command_num INTEGER NOT NULL,      -- 1, 2, 3 within the step
    command TEXT NOT NULL,             -- "docmgr init --seed-vocabulary"
    started_at TEXT NOT NULL,
    completed_at TEXT,
    exit_code INTEGER,
    duration_ms INTEGER,
    stdout_path TEXT,                  -- .logs/step-02-cmd-01-stdout.txt (or empty if captured by step)
    stderr_path TEXT,
    FOREIGN KEY (step_id) REFERENCES steps(step_id)
);

-- diagnostics: structured findings extracted from stdout/stderr (optional)
CREATE TABLE diagnostics (
    diag_id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL,
    step_id TEXT,                      -- NULL if from run metadata
    command_id TEXT,                   -- NULL if from step/run metadata
    level TEXT NOT NULL,               -- "error", "warning", "info"
    category TEXT,                     -- "doctor", "parse", "unknown_flag", etc.
    message TEXT NOT NULL,
    file_path TEXT,                    -- associated doc/code file if applicable
    details TEXT,                      -- JSON blob or multiline text
    FOREIGN KEY (run_id) REFERENCES scenario_runs(run_id),
    FOREIGN KEY (step_id) REFERENCES steps(step_id),
    FOREIGN KEY (command_id) REFERENCES commands(command_id)
);

CREATE INDEX idx_steps_run ON steps(run_id);
CREATE INDEX idx_commands_step ON commands(step_id);
CREATE INDEX idx_diagnostics_run ON diagnostics(run_id);
CREATE INDEX idx_diagnostics_level ON diagnostics(level);
```

### Output File Structure

```
$ROOT_DIR/
  .scenario-run.db           ← sqlite database
  .logs/
    step-01-stdout.txt       ← per-step stdout
    step-01-stderr.txt       ← per-step stderr
    step-02-cmd-01-stdout.txt  ← per-command (optional)
    ...
  acme-chat-app/             ← test workspace (unchanged)
```

### Harness Integration

**Minimal changes to `run-all.sh`**:

```bash
#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-/tmp/docmgr-scenario}"
RUN_ID="$(date +%Y-%m-%dT%H:%M:%S)-run-$$"
DB="${ROOT_DIR}/.scenario-run.db"
LOG_DIR="${ROOT_DIR}/.logs"

source "$(dirname "$0")/scenario-harness.sh"  # provides: run_step(), init_db(), finalize_run()

init_db "$DB" "$RUN_ID" "$ROOT_DIR" "${DOCMGR_PATH}"

run_step 1 "00-reset" "./00-reset.sh" "${ROOT_DIR}"
run_step 2 "01-create-mock-codebase" "./01-create-mock-codebase.sh" "${ROOT_DIR}"
run_step 3 "02-init-ticket" "./02-init-ticket.sh" "${ROOT_DIR}"
# ... etc.

finalize_run "$DB" "$RUN_ID" "$?"
```

**New helper: `scenario-harness.sh`**:

```bash
#!/usr/bin/env bash

init_db() {
    local db="$1" run_id="$2" root_dir="$3" docmgr_path="$4"
    sqlite3 "$db" < scenario-schema.sql
    sqlite3 "$db" "INSERT INTO scenario_runs (run_id, root_dir, docmgr_path, started_at) VALUES ('$run_id', '$root_dir', '$docmgr_path', datetime('now'));"
}

run_step() {
    local num="$1" name="$2" script="$3" root="$4"
    local step_id="${RUN_ID}-step-$(printf %02d $num)"
    local start_time="$(date -Iseconds)"
    local start_ms="$(date +%s%3N)"
    
    mkdir -p "$LOG_DIR"
    local stdout_log="${LOG_DIR}/step-$(printf %02d $num)-stdout.txt"
    local stderr_log="${LOG_DIR}/step-$(printf %02d $num)-stderr.txt"
    
    sqlite3 "$DB" "INSERT INTO steps (step_id, run_id, step_num, step_name, script_path, started_at) VALUES ('$step_id', '$RUN_ID', $num, '$name', '$script', '$start_time');"
    
    set +e
    bash "$script" "$root" >"$stdout_log" 2>"$stderr_log"
    local exit_code="$?"
    set -e
    
    local end_ms="$(date +%s%3N)"
    local duration=$(( end_ms - start_ms ))
    
    sqlite3 "$DB" "UPDATE steps SET completed_at=datetime('now'), exit_code=$exit_code, duration_ms=$duration, stdout_path='$stdout_log', stderr_path='$stderr_log' WHERE step_id='$step_id';"
    
    if [[ $exit_code -ne 0 ]]; then
        echo "[fail] Step $num ($name) exited with code $exit_code" >&2
        return $exit_code
    fi
    echo "[ok] Step $num ($name) completed" >&2
}

finalize_run() {
    local db="$1" run_id="$2" exit_code="$3"
    sqlite3 "$db" "UPDATE scenario_runs SET completed_at=datetime('now'), exit_code=$exit_code WHERE run_id='$run_id';"
}
```

## Design Decisions

### 1. **Why sqlite, not JSON/YAML?**

- **Queryable**: Standard SQL for filtering, aggregation, and joins without custom parsers.
- **Relational**: Captures run → steps → commands hierarchy naturally.
- **Portable**: Single file, no external dependencies, works everywhere Go/bash exist.
- **Indexable**: Fast lookups even with 100s of runs.

### 2. **Per-step file capture, not inline storage**

- **Rationale**: Storing 40KB+ of output in a `TEXT` column bloats the DB and slows queries. Files keep the DB lean and make `less`/`cat` workflows still viable.
- **Trade-off**: Requires `.logs/` directory management (cleanup policy TBD).

### 3. **Optional `commands` table granularity**

- **Phase 1**: Only track `steps` (script-level).
- **Phase 2** (optional): Instrument individual `docmgr` commands within scripts for fine-grain analysis (useful if a step has 10+ commands and we need to identify which failed).
- **Trade-off**: More instrumentation overhead vs more granular data.

### 4. **Diagnostics extraction is optional/future work**

- The `diagnostics` table is designed but **not required for MVP**. We can seed it later by post-processing logs (e.g., parsing `Error:` lines or `doctor` JSON output).
- **Rationale**: Avoids coupling the harness to `docmgr`-specific output formats initially.

### 5. **`run_id` includes timestamp + PID for uniqueness**

- Format: `YYYY-MM-DDTHH:MM:SS-run-<pid>` (e.g., `2025-12-13T16:09:42-run-12345`)
- **Rationale**: Human-readable, sortable, unique even if runs overlap.

## Query Examples

### Find Failed Steps

```sql
SELECT run_id, step_num, step_name, exit_code, stderr_path
FROM steps
WHERE exit_code != 0
ORDER BY run_id DESC, step_num;
```

### Find Steps Matching a Pattern (e.g., "doctor")

```sql
SELECT run_id, step_num, step_name, duration_ms, exit_code
FROM steps
WHERE step_name LIKE '%doctor%'
ORDER BY run_id DESC, step_num;
```

### Show Run Summary

```sql
SELECT 
    run_id,
    docmgr_path,
    started_at,
    exit_code,
    duration_ms,
    (SELECT COUNT(*) FROM steps WHERE steps.run_id = scenario_runs.run_id AND exit_code != 0) AS failed_steps
FROM scenario_runs
ORDER BY started_at DESC
LIMIT 10;
```

### Extract Warnings from a Specific Run

```bash
# Get stdout paths for steps in run X
sqlite3 .scenario-run.db "SELECT stdout_path FROM steps WHERE run_id='2025-12-13T16:09:42-run-12345';" | \
    while read -r log; do
        echo "==> $log"
        rg -n -i "\bwarning\b" "$log" || true
    done
```

### Compare Two Runs (duration regression)

```sql
SELECT 
    a.step_name,
    a.duration_ms AS run_old_ms,
    b.duration_ms AS run_new_ms,
    (b.duration_ms - a.duration_ms) AS delta_ms
FROM steps a
JOIN steps b ON a.step_num = b.step_num AND a.step_name = b.step_name
WHERE a.run_id = '2025-12-12T10:00:00-run-old'
  AND b.run_id = '2025-12-13T16:09:42-run-new'
ORDER BY ABS(b.duration_ms - a.duration_ms) DESC;
```

## Implementation Plan

### Phase 1: Core Harness (MVP)

- [ ] **Schema definition**: Write `test-scenarios/testing-doc-manager/scenario-schema.sql` with `scenario_runs` and `steps` tables.
- [ ] **Bash helpers**: Implement `scenario-harness.sh` with `init_db()`, `run_step()`, `finalize_run()`.
- [ ] **Update `run-all.sh`**: Source `scenario-harness.sh`, wrap each script with `run_step()`, and populate DB.
- [ ] **Manual validation**: Run a scenario, query the DB with `sqlite3`, verify structure + timings + exit codes.

### Phase 2: Query Tooling

- [ ] **Summary reporter**: Add `test-scenarios/testing-doc-manager/summarize-run.sh` that queries the DB and prints:
  - Run metadata (duration, exit code, docmgr version)
  - Failed steps (if any)
  - Top 5 slowest steps
  - Warning/error counts (via `rg` on captured logs)

- [ ] **Interactive query helper**: Optional: `docmgr-scenario-query --db .scenario-run.db --show-failures` (Go binary or bash script wrapping `sqlite3`).

### Phase 3: Fine-Grain Commands Table (Optional)

- [ ] Instrument individual `docmgr` commands within scripts (either via a bash wrapper or by modifying scripts to call a logging function before/after each command).
- [ ] Populate `commands` table with per-command metadata.
- [ ] Add queries for "which specific command failed in step X?"

### Phase 4: Diagnostics Extraction (Future)

- [ ] Parse `doctor` JSON output (`--diagnostics-json`) and insert into `diagnostics` table.
- [ ] Parse known error patterns (regex-based) from logs and classify them.
- [ ] Add queries for "show all doctor warnings across runs".

## Alternatives Considered

### 1. **Structured JSON logs (one JSON object per line)**

**Pros**: Easy to parse with `jq`, standard format.  
**Cons**: No relational queries (joins, aggregations require external tooling), harder to index/search across runs, JSON serialization overhead.  
**Verdict**: Rejected; SQL is more powerful for analytics.

### 2. **Keep text logs, add regex/awk post-processing scripts**

**Pros**: Minimal harness changes, no schema to maintain.  
**Cons**: Brittle (regex patterns break when output format changes), slow for large logs, no structured metadata (timings, exit codes).  
**Verdict**: This is the current state; inadequate for detecting regressions.

### 3. **Use TAP (Test Anything Protocol) output**

**Pros**: Standardized test format, tooling exists (`prove`).  
**Cons**: TAP is line-oriented (no hierarchy), doesn't capture timings/metadata/logs naturally, requires TAP-compliant test harness (not bash scripts).  
**Verdict**: Over-engineering; sqlite is simpler and more flexible.

### 4. **Emit GitHub Actions workflow logs / JUnit XML**

**Pros**: CI-native formats.  
**Cons**: Designed for CI environments, not local dev; we'd still need local queryability; JUnit XML is verbose and doesn't capture logs naturally.  
**Verdict**: sqlite is better for local dev; we can emit JUnit from sqlite later if CI needs it.

## Schema Evolution and Compatibility

- **V1 schema** (MVP): `scenario_runs` + `steps` only.
- **V2**: Add `commands` table when fine-grain tracking is needed.
- **V3**: Add `diagnostics` table for extracted findings.

**Migration strategy**: Use `PRAGMA user_version` to track schema version. Add a `scenario-migrate-db.sh` script that applies incremental DDL if version mismatches.

## Example Workflow

### 1. Run the scenario

```bash
cd test-scenarios/testing-doc-manager/
go build -o /tmp/docmgr-local ../../cmd/docmgr
DOCMGR_PATH=/tmp/docmgr-local bash run-all.sh /tmp/docmgr-scenario
```

Produces:

```
/tmp/docmgr-scenario/
  .scenario-run.db
  .logs/
    step-01-stdout.txt
    step-01-stderr.txt
    ...
  acme-chat-app/
```

### 2. Query the run

```bash
cd /tmp/docmgr-scenario
sqlite3 .scenario-run.db

sqlite> SELECT * FROM scenario_runs;
2025-12-13T16:09:42-run-12345|/tmp/docmgr-scenario|/tmp/docmgr-local|v0.1.14|2025-12-13T16:09:42|2025-12-13T16:10:15|0|33000|

sqlite> SELECT step_num, step_name, duration_ms, exit_code FROM steps ORDER BY step_num;
1|00-reset|50|0
2|01-create-mock-codebase|120|0
3|02-init-ticket|450|0
...

sqlite> SELECT step_name, exit_code, stderr_path FROM steps WHERE exit_code != 0;
(no rows = all passed)
```

### 3. Analyze warnings

```bash
# Show all steps that produced warnings in output
sqlite3 .scenario-run.db "SELECT step_num, step_name, stdout_path FROM steps;" | while IFS='|' read num name log; do
    if rg -q -i "\bwarning\b" "$log" 2>/dev/null; then
        echo "Step $num ($name) has warnings: $log"
    fi
done
```

## Open Questions

### 1. **Should we capture full command lines for `docmgr` invocations?**

**Options**:
- A: Capture every `docmgr` command individually in the `commands` table (fine-grain).
- B: Only capture script-level steps; individual commands stay in the step's stdout file (coarse-grain, MVP).

**Recommendation**: Start with **B** (script-level) for MVP. Add **A** (command-level) later if needed for debugging complex multi-command steps.

### 2. **How long should we retain `.logs/` and `.scenario-run.db`?**

**Options**:
- Keep forever (grows unbounded).
- Keep last N runs (e.g., 10).
- Keep runs from last 30 days.
- User-managed cleanup (document in README).

**Recommendation**: **User-managed** for MVP. Add a `scenario-cleanup.sh` helper later that prunes old runs from DB and deletes stale log files.

### 3. **Should we extract diagnostics automatically?**

**Options**:
- A: Parse `doctor` JSON (`--diagnostics-json`) and known error patterns during the run.
- B: Post-process logs after the run (separate tool).
- C: Manual (query stdout/stderr files via sqlite + shell scripts).

**Recommendation**: Start with **C** (manual) for MVP. The DB gives us structure; parsing heuristics can be added later without changing the harness.

### 4. **Integration with CI?**

If we run scenarios in CI, we may want:
- Artifact upload (`.scenario-run.db` + `.logs/`)
- JUnit XML export (`sqlite → XML` converter)
- GitHub Actions annotations for failures

**Recommendation**: Defer until we validate the design locally. The sqlite DB is self-contained and portable, so CI integration should be straightforward later.

## Security and Privacy Considerations

- **No sensitive data**: Scenario uses mock data (`MEN-4242`, fake files). Logs may contain absolute paths (e.g., `/home/manuel/...`), but no credentials or secrets.
- **File permissions**: `.scenario-run.db` and `.logs/` inherit `$ROOT_DIR` permissions (typically `0755`/`0644`). No special handling needed.

## References

- **Current scenario harness**: `test-scenarios/testing-doc-manager/run-all.sh` (non-instrumented)
- **Example verbose log**: `ttmp/2025/12/13/IMPROVE-SCENARIO-LOGGING--make-scenario-suite-output-queryable-sqlite/sources/local/phase-3-scenario-log-2025-12-13.txt` (40KB, 1033 lines)
- **Sqlite CLI docs**: https://www.sqlite.org/cli.html
- **TAP protocol** (rejected alternative): https://testanything.org/
