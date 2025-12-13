# scenariolog

`scenariolog` is a small, self-contained CLI tool that logs scenario-style execution (runs/steps/commands) into a **sqlite** database with:

- **KV tags** (run/step/command-scoped metadata)
- **Artifacts** (stdout/stderr files + metadata such as hashes/sizes)
- **FTS** search over text artifacts

This tool is intentionally generic and reusable outside `docmgr`.

## Build

From the repo root:

```bash
go -C scenariolog build -tags sqlite_fts5 -o /tmp/scenariolog-local ./cmd/scenariolog
```

## Quick start

```bash
/tmp/scenariolog-local init --db /tmp/scenario/.scenario-run.db
```

## FTS5 support (and fallback)

By default we try to create the FTS5 table (`log_lines_fts`) during migrations, but **we degrade gracefully** if the sqlite library does not have FTS5 enabled:

- Without FTS5: migrations succeed; FTS-backed search/indexing features are unavailable (or will be no-ops).
- With FTS5: build with `-tags sqlite_fts5` (CGO + bundled sqlite compile options) to enable full text search.


