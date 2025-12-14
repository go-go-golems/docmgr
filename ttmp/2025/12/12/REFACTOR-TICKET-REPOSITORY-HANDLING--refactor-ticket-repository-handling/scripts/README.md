# Comparison Scripts

## `compare-docmgr-versions.sh`

Compares system docmgr vs locally built docmgr by running the full scenario suite with both versions and recording results in scenariolog SQLite databases.

### Purpose

This script helps validate that the refactored docmgr (local build) produces equivalent behavior to the system docmgr (old version). It's particularly useful for:

- **Regression detection**: Identify any behavior changes introduced by the refactor
- **Performance comparison**: Compare execution times between versions
- **Output validation**: Ensure outputs are equivalent (or identify intentional changes)
- **Failure analysis**: Compare failure modes and error messages

### Usage

```bash
# Basic usage (uses default paths)
./compare-docmgr-versions.sh

# Custom paths
./compare-docmgr-versions.sh /tmp/docmgr-system /tmp/docmgr-local
```

### What It Does

1. **Builds scenariolog** (if not already built) with FTS5 support for log searching
2. **Builds local docmgr** from the repo (`go build ./cmd/docmgr`)
3. **Finds system docmgr** from PATH (`command -v docmgr`)
4. **Runs scenario suite with system docmgr**:
   - Creates test repository at `${SYSTEM_ROOT}/acme-chat-app`
   - Records all steps/outputs in `${SYSTEM_ROOT}/.scenario-run.db`
5. **Runs scenario suite with local docmgr**:
   - Creates test repository at `${LOCAL_ROOT}/acme-chat-app`
   - Records all steps/outputs in `${LOCAL_ROOT}/.scenario-run.db`
6. **Prints comparison commands** for analyzing differences

### Output Locations

- **System docmgr run**: `${SYSTEM_ROOT}` (default: `/tmp/docmgr-system`)
  - Database: `${SYSTEM_ROOT}/.scenario-run.db`
  - Logs: `${SYSTEM_ROOT}/.logs/`
  - Test repo: `${SYSTEM_ROOT}/acme-chat-app/`

- **Local docmgr run**: `${LOCAL_ROOT}` (default: `/tmp/docmgr-local`)
  - Database: `${LOCAL_ROOT}/.scenario-run.db`
  - Logs: `${LOCAL_ROOT}/.logs/`
  - Test repo: `${LOCAL_ROOT}/acme-chat-app/`

### Querying Results

After running the comparison, you can use the helper script or scenariolog directly:

#### Quick Comparison (Recommended)

```bash
# Use the helper script for side-by-side comparison
./compare-results.sh /tmp/docmgr-system/.scenario-run.db /tmp/docmgr-local/.scenario-run.db
```

This will show:
- Step exit code comparison
- Step duration comparison (sorted by difference)
- Steps with different exit codes
- Total run duration for both versions

#### Manual Querying with scenariolog

For more detailed analysis, use scenariolog directly:

#### Quick Summary

```bash
# System docmgr summary
/tmp/scenariolog-local summary --db /tmp/docmgr-system/.scenario-run.db --output table

# Local docmgr summary
/tmp/scenariolog-local summary --db /tmp/docmgr-local/.scenario-run.db --output table
```

#### Find Failures

```bash
# System failures
/tmp/scenariolog-local failures --db /tmp/docmgr-system/.scenario-run.db --output table

# Local failures
/tmp/scenariolog-local failures --db /tmp/docmgr-local/.scenario-run.db --output table
```

#### Compare Step Exit Codes and Durations

Use the helper script (recommended):
```bash
./compare-results.sh /tmp/docmgr-system/.scenario-run.db /tmp/docmgr-local/.scenario-run.db
```

Or use SQLite directly (see helper script source for SQL queries).

#### Search Logs

```bash
# Search system logs for errors
/tmp/scenariolog-local search \
  --db /tmp/docmgr-system/.scenario-run.db \
  --run-id "${SYSTEM_RUN_ID}" \
  --query "error OR warning OR fail" \
  --limit 20 \
  --output table

# Search local logs for errors
/tmp/scenariolog-local search \
  --db /tmp/docmgr-local/.scenario-run.db \
  --run-id "${LOCAL_RUN_ID}" \
  --query "error OR warning OR fail" \
  --limit 20 \
  --output table
```

#### View Step Artifacts (stdout/stderr)

```bash
# View artifacts for step 5 (search-scenarios) from system run
/tmp/scenariolog-local artifacts \
  --db /tmp/docmgr-system/.scenario-run.db \
  --run-id "${SYSTEM_RUN_ID}" \
  --step-num 5 \
  --output table

# View artifacts for step 5 from local run
/tmp/scenariolog-local artifacts \
  --db /tmp/docmgr-local/.scenario-run.db \
  --run-id "${LOCAL_RUN_ID}" \
  --step-num 5 \
  --output table
```

### Manual Comparison

#### Compare Test Repositories

```bash
# Diff the entire test repositories
diff -r /tmp/docmgr-system/acme-chat-app /tmp/docmgr-local/acme-chat-app || true
```

#### Compare Specific Command Outputs

```bash
# System docmgr search output
DOCMGR_PATH="$(command -v docmgr)" docmgr doc search \
  --root /tmp/docmgr-system/acme-chat-app/ttmp \
  --query "chat" \
  --output json > /tmp/system-search.json

# Local docmgr search output
DOCMGR_PATH="/tmp/docmgr-local/docmgr-local" docmgr doc search \
  --root /tmp/docmgr-local/acme-chat-app/ttmp \
  --query "chat" \
  --output json > /tmp/local-search.json

# Diff the JSON outputs
diff /tmp/system-search.json /tmp/local-search.json || true
```

### Environment Variables

- `SCENARIOLOG_PATH`: Path to scenariolog binary (default: `/tmp/scenariolog-local`)
- `DOCMGR_PATH`: Not used by this script (it sets this internally for each run)

### Prerequisites

- `go` installed (for building docmgr and scenariolog)
- `bash` (POSIX shell compatible)
- `sqlite3` (for querying databases)
- System `docmgr` installed and in PATH
- `test-scenarios/testing-doc-manager/` directory exists in repo

### Troubleshooting

#### scenariolog build fails

If FTS5 build fails, the script will fall back to building without FTS5. Search functionality will be disabled, but other features work.

#### System docmgr not found

Ensure `docmgr` is installed and in PATH:
```bash
command -v docmgr
```

#### Scenario suite fails

Check the scenariolog databases for detailed error logs:
```bash
/tmp/scenariolog-local failures --db /tmp/docmgr-system/.scenario-run.db --output table
```

### Related Documentation

- `test-scenarios/testing-doc-manager/README.md` - Scenario suite documentation
- `scenariolog/pkg/doc/docs/how-to-use-scenariolog-local.md` - scenariolog usage guide

