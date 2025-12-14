---
Title: Comparison Suite Report — System docmgr vs Local (Refactor) docmgr
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/scripts/compare-docmgr-versions.sh
      Note: Runs the common scenario subset twice (system vs local) and records runs in scenariolog
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/scripts/compare-results.sh
      Note: Cross-run comparison helper (SQLite ATTACH) for step exit codes and durations
    - Path: test-scenarios/testing-doc-manager/run-all.sh
      Note: Scenario suite used as the behavioral baseline
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-13T20:40:00.000000000-05:00
---

# Comparison Suite Report — System docmgr vs Local (Refactor) docmgr

## Executive Summary

We ran a **common scenario subset** (steps 1–14) of `test-scenarios/testing-doc-manager/` against:
- **System docmgr**: `/home/manuel/.local/bin/docmgr`
- **Local docmgr (repo build)**: `/tmp/docmgr-compare-local/docmgr-local`

Both runs were recorded with **scenariolog** into separate SQLite DBs and then compared step-by-step.

**Result**: ✅ **Behavior matches at the scenario level for the common subset**.
- All 14 steps exit with code **0** for both binaries.
- No step exit codes differ.
- Runtime is slightly faster for the local binary in this run (local: **7.46s**, system: **8.94s**).

## Why the “common subset” exists

The upstream scenario suite includes steps that require **newer docmgr features** (notably `workspace export-sqlite`), which the system docmgr does not support. For cross-version comparison, we therefore run a subset that both binaries can execute, while still covering the core behaviors (init, doc creation, relate, doctor, search, status, configuration, path normalization).

## Run Metadata

### System run
- **Root**: `/tmp/docmgr-system`
- **DB**: `/tmp/docmgr-system/.scenario-run.db`
- **Run ID**: `2025-12-14T01:36:45.053745795Z-pid-2592017-439dc61eaf955b6a`
- **Suite**: `testing-doc-manager-common-system`
- **Steps**: 14
- **Exit**: 0
- **Duration**: 8936ms

### Local run
- **Root**: `/tmp/docmgr-local`
- **DB**: `/tmp/docmgr-local/.scenario-run.db`
- **Run ID**: `2025-12-14T01:36:54.054533601Z-pid-2593497-9bd89fe76d3f1190`
- **Suite**: `testing-doc-manager-common-local`
- **Steps**: 14
- **Exit**: 0
- **Duration**: 7462ms

## Step-by-step Comparison

### Exit codes

All steps matched:

| Step | Name | System | Local | Status |
|------|------|--------|-------|--------|
| 1 | 01-create-mock-codebase | 0 | 0 | match |
| 2 | 02-init-ticket | 0 | 0 | match |
| 3 | 03-create-docs-and-meta | 0 | 0 | match |
| 4 | 04-relate-and-doctor | 0 | 0 | match |
| 5 | 05-search-scenarios | 0 | 0 | match |
| 6 | 06-doctor-advanced | 0 | 0 | match |
| 7 | 07-status | 0 | 0 | match |
| 8 | 08-configure | 0 | 0 | match |
| 9 | 09-relate-from-git | 0 | 0 | match |
| 10 | 10-status-warnings | 0 | 0 | match |
| 11 | 11-changelog-file-notes | 0 | 0 | match |
| 12 | 12-vocab-add-output | 0 | 0 | match |
| 13 | 13-template-schema-output | 0 | 0 | match |
| 14 | 14-path-normalization | 0 | 0 | match |

### Durations (largest deltas first)

The local binary was faster on most steps in this run; one step (`14-path-normalization`) was slightly slower.

| Step | Name | System (s) | Local (s) | Diff (s) |
|------|------|------------|-----------|----------|
| 13 | 13-template-schema-output | 1.15 | 0.75 | -0.40 |
| 3 | 03-create-docs-and-meta | 1.82 | 1.52 | -0.29 |
| 5 | 05-search-scenarios | 1.73 | 1.51 | -0.22 |
| 6 | 06-doctor-advanced | 0.74 | 0.60 | -0.14 |
| 12 | 12-vocab-add-output | 0.21 | 0.09 | -0.11 |
| 14 | 14-path-normalization | 0.66 | 0.71 | +0.06 |

**Total runtime**:
- System: **8.94s**
- Local: **7.46s**

## Diagnostics / Warning Parity

Both runs surface the same expected warning patterns in the logs during the doctor-advanced step (`06-doctor-advanced`), notably warnings about:
- unknown vocabulary values for `Topics`
- missing related file entry / missing related file note

These warnings are part of the scenario’s intentional “doctor warnings” coverage; importantly, they appear in **both** runs (parity).

## Earlier attempt failures (and what we fixed)

Before switching to the common subset + correcting the comparison tooling, we observed two failure modes that were **harness issues**, not docmgr behavior differences:

1. **System docmgr failed** the full suite on `19-export-sqlite` with:
   - `Error: unknown flag: --out`
   This indicates the system docmgr binary lacks the newer `workspace export-sqlite` command/flags.

2. **Local docmgr run failed early** because the local binary was built under the scenario root and got deleted by the suite’s reset step (`00-reset.sh`).
   - `./02-init-ticket.sh: line 11: /tmp/docmgr-local/docmgr-local: No such file or directory`

We fixed the harness by:
- Building the local docmgr binary outside the scenario root (`/tmp/docmgr-compare-local/docmgr-local`)
- Running a common subset that excludes the incompatible `19-export-sqlite` step
- Updating the comparison helper to use the correct scenariolog schema (`steps.step_name`, `scenario_runs.completed_at`) and to compare two DBs via SQLite `ATTACH DATABASE`

## How to reproduce

Run comparison:

```bash
cd /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr
bash ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/scripts/compare-docmgr-versions.sh /tmp/docmgr-system /tmp/docmgr-local
```

Compare DBs:

```bash
cd /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr
bash ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/scripts/compare-results.sh /tmp/docmgr-system/.scenario-run.db /tmp/docmgr-local/.scenario-run.db
```

Inspect details with scenariolog:

```bash
/tmp/scenariolog-local summary  --db /tmp/docmgr-system/.scenario-run.db --output table
/tmp/scenariolog-local summary  --db /tmp/docmgr-local/.scenario-run.db --output table
/tmp/scenariolog-local failures --db /tmp/docmgr-system/.scenario-run.db --output table
/tmp/scenariolog-local failures --db /tmp/docmgr-local/.scenario-run.db --output table
```

## Full command + SQL transcript (exactly what we ran)

This section captures the **exact CLI commands and SQL** used while producing this report, so we can reproduce it later.

### 1) Verify scenariolog DB schema (so our SQL targets the right tables/columns)

```bash
cd /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr
sqlite3 /tmp/docmgr-system/.scenario-run.db "PRAGMA table_info(steps);"
sqlite3 /tmp/docmgr-system/.scenario-run.db "PRAGMA table_info(scenario_runs);"
sqlite3 /tmp/docmgr-system/.scenario-run.db ".tables"
```

### 2) Run the comparison suite (fresh run for both binaries)

We ran the harness to create fresh run DBs:

```bash
cd /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr
rm -rf /tmp/docmgr-system /tmp/docmgr-local
timeout 900 bash ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/scripts/compare-docmgr-versions.sh /tmp/docmgr-system /tmp/docmgr-local 2>&1 | tee /tmp/docmgr-compare-run.log
```

The harness automatically:
- builds `scenariolog` (default: `/tmp/scenariolog-local`) if missing
- builds the local docmgr binary at `/tmp/docmgr-compare-local/docmgr-local`
- runs a common subset of scenario steps (1–14) for system and local docmgr

### 3) Summaries and failures (scenariolog commands)

```bash
/tmp/scenariolog-local summary --db /tmp/docmgr-system/.scenario-run.db --output table
/tmp/scenariolog-local summary --db /tmp/docmgr-local/.scenario-run.db --output table

/tmp/scenariolog-local failures --db /tmp/docmgr-system/.scenario-run.db --output table
/tmp/scenariolog-local failures --db /tmp/docmgr-local/.scenario-run.db --output table
```

### 4) Search for warnings/errors (scenariolog search)

System run:

```bash
/tmp/scenariolog-local search \
  --db /tmp/docmgr-system/.scenario-run.db \
  --run-id 2025-12-14T01:36:45.053745795Z-pid-2592017-439dc61eaf955b6a \
  --query "error OR warning OR fail" \
  --limit 50 \
  --output table
```

Local run:

```bash
/tmp/scenariolog-local search \
  --db /tmp/docmgr-local/.scenario-run.db \
  --run-id 2025-12-14T01:36:54.054533601Z-pid-2593497-9bd89fe76d3f1190 \
  --query "error OR warning OR fail" \
  --limit 50 \
  --output table
```

### 5) Step-by-step diff across runs (our helper script)

```bash
cd /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr
bash ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/scripts/compare-results.sh \
  /tmp/docmgr-system/.scenario-run.db \
  /tmp/docmgr-local/.scenario-run.db
```

### 6) Raw SQL used for cross-run comparison (from `compare-results.sh`)

**Exit-code comparison** (SQLite cross-file join via `ATTACH DATABASE`):

```sql
ATTACH DATABASE '/tmp/docmgr-local/.scenario-run.db' AS local_db;
SELECT
  s1.step_num,
  s1.step_name,
  s1.exit_code AS system_exit,
  s2.exit_code AS local_exit,
  CASE
    WHEN s1.exit_code = s2.exit_code THEN 'match'
    ELSE 'DIFFERENT'
  END AS status
FROM steps s1
JOIN local_db.steps s2 ON s1.step_num = s2.step_num AND s1.step_name = s2.step_name
WHERE s1.run_id = '2025-12-14T01:36:45.053745795Z-pid-2592017-439dc61eaf955b6a'
  AND s2.run_id = '2025-12-14T01:36:54.054533601Z-pid-2593497-9bd89fe76d3f1190'
ORDER BY s1.step_num;
DETACH DATABASE local_db;
```

**Duration comparison**:

```sql
ATTACH DATABASE '/tmp/docmgr-local/.scenario-run.db' AS local_db;
SELECT
  s1.step_num,
  s1.step_name,
  ROUND((julianday(s1.completed_at) - julianday(s1.started_at)) * 86400, 2) AS system_sec,
  ROUND((julianday(s2.completed_at) - julianday(s2.started_at)) * 86400, 2) AS local_sec,
  ROUND(
    ((julianday(s2.completed_at) - julianday(s2.started_at)) -
     (julianday(s1.completed_at) - julianday(s1.started_at))) * 86400, 2
  ) AS diff_sec
FROM steps s1
JOIN local_db.steps s2 ON s1.step_num = s2.step_num AND s1.step_name = s2.step_name
WHERE s1.run_id = '2025-12-14T01:36:45.053745795Z-pid-2592017-439dc61eaf955b6a'
  AND s2.run_id = '2025-12-14T01:36:54.054533601Z-pid-2593497-9bd89fe76d3f1190'
ORDER BY ABS(diff_sec) DESC;
DETACH DATABASE local_db;
```

**Total run duration**:

```sql
ATTACH DATABASE '/tmp/docmgr-local/.scenario-run.db' AS local_db;
SELECT
  'system' AS version,
  ROUND((julianday(completed_at) - julianday(started_at)) * 86400, 2) AS duration_sec
FROM scenario_runs
WHERE run_id = '2025-12-14T01:36:45.053745795Z-pid-2592017-439dc61eaf955b6a'
UNION ALL
SELECT
  'local' AS version,
  ROUND((julianday(completed_at) - julianday(started_at)) * 86400, 2) AS duration_sec
FROM local_db.scenario_runs
WHERE run_id = '2025-12-14T01:36:54.054533601Z-pid-2593497-9bd89fe76d3f1190';
DETACH DATABASE local_db;
```

### 7) Raw SQL we ran to print step rows (per-run)

System:

```bash
sqlite3 /tmp/docmgr-system/.scenario-run.db \
  "SELECT step_num, step_name, exit_code, duration_ms FROM steps WHERE run_id='2025-12-14T01:36:45.053745795Z-pid-2592017-439dc61eaf955b6a' ORDER BY step_num;"
```

Local:

```bash
sqlite3 /tmp/docmgr-local/.scenario-run.db \
  "SELECT step_num, step_name, exit_code, duration_ms FROM steps WHERE run_id='2025-12-14T01:36:54.054533601Z-pid-2593497-9bd89fe76d3f1190' ORDER BY step_num;"
```

### 8) Capturing the run log snippet

```bash
cat /tmp/docmgr-compare-run.log | tail -20
```

## Conclusions

- **Behavioral parity** (for shared feature set) is confirmed by scenario pass/fail: **14/14 steps match**.
- **Compatibility gap** is confirmed for `workspace export-sqlite` (system docmgr lacks it). This is expected until the system binary is upgraded.
- **Performance**: local was modestly faster in this run; treat as indicative only (single run, non-isolated environment).

## Next steps (recommended)

- Once system docmgr is upgraded (or a pinned “baseline old binary” is preserved), re-run the **full suite** including step 19 for both, to expand the comparison surface.
- Consider adding a third mode: run the full suite on local only, but keep cross-version comparison on the common subset.


