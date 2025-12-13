# Tasks

## TODO

This task list is the actionable checklist that implements:

- design-doc #2: `design-doc/02-generic-sqlite-scenario-logger-go-tool.md`
- implementation plan: `design-doc/03-implementation-plan-scenariolog-mvp-kv-artifacts-fts-glazed-cli.md`

### Phase 0 — Repo scaffolding

- [x] Create self-contained tool directory `scenariolog/`
- [x] Add `scenariolog/go.mod` + `scenariolog/go.sum` (module-local deps)
- [x] Add `scenariolog/README.md` (how to build/run; how to integrate)
- [ ] Create `scenariolog/internal/scenariolog/` package skeleton (`db.go`, `migrate.go`, `logger.go`, `artifacts.go`, `fts.go`)
- [x] Create `scenariolog/cmd/scenariolog` Cobra root command (standalone first; docmgr wrapper can come later)

### Phase 1 — SQLite schema + migrations (KV + artifacts + FTS)

- [x] Implement DB open + pragmas for file-backed sqlite (foreign_keys, WAL, synchronous, busy timeout)
- [x] Implement schema versioning via `PRAGMA user_version`
- [ ] Add migrations that create:
- [x] `scenario_runs`
- [x] `steps`
- [x] `kv` (run/step/command-scoped tags)
- [x] `artifacts` (stdout/stderr + arbitrary outputs, with size/hash)
  - [ ] `log_lines_fts` (FTS5 virtual table for search; best-effort / degrade if missing)
- [x] Add indexes and constraints (unique scope keys, artifact uniqueness, etc.)
- [x] Unit test: migrate empty DB and assert tables exist (`sqlite_master`)

### Phase 2 — Exec wrapper + artifact capture

- [x] Implement process execution wrapper with:
- [x] stdout/stderr capture to files
- [x] timing + exit code capture
  - [ ] signal/ctx cancellation best-effort finalization
- [x] Insert stdout/stderr into `artifacts` (kind, path, sha256, size_bytes, is_text)
- [ ] Optional: also keep `steps.stdout_path`/`steps.stderr_path` columns in sync (if we keep those columns)
- [ ] Add KV tags (provenance):
  - [ ] suite name/version
  - [ ] hostname/user
  - [ ] `docmgr_path`/`docmgr_version` (where applicable)
  - [ ] best-effort git SHA + dirty flag (guarded)
- [ ] Unit test: run a command that writes to stdout+stderr and exits nonzero; verify DB rows + artifacts + hash/size

### Phase 3 — FTS ingestion + search

- [ ] Implement FTS ingestion for text artifacts (line-by-line, store line_num)
- [ ] Add bounded ingestion guardrails (max line length, max bytes per artifact; skip `is_text=0`)
- [ ] Decide indexing mode:
  - [ ] Default: auto-index stdout/stderr artifacts on `exec` completion
  - [ ] Optional: explicit `index fts` command (for re-indexing / toggling)
- [ ] Implement `search` query (`MATCH`) returning run_id, artifact_id, line_num, text/snippet
- [x] Ensure graceful fallback if FTS5 is unavailable (clear error or explicit no-op; core features still work)
- [ ] Unit test: index an artifact and query for a keyword; assert expected hits

### Phase 4 — CLI (Cobra + Glazed)

- [ ] Create `scenariolog/cmd/scenariolog` Cobra root command
- [ ] Add Glazed help system wiring (help system + richer help)
- [ ] Implement command settings parsing via `parsedLayers.InitializeStruct(...)` (don’t read Cobra flags directly)
- [ ] Implement commands:
- [x] `scenariolog init --db`
- [x] `scenariolog run start --db --root-dir --suite` (returns run_id)
- [x] `scenariolog run end --db --run-id --exit-code`
- [x] `scenariolog exec --db --run-id --kind step --step-num --name --log-dir -- <cmd...>`
  - [ ] `scenariolog search --db --query ...` (Glazed structured output)
  - [ ] `scenariolog summary --db [--run-id]` (Glazed structured output)
  - [ ] `scenariolog failures --db [--run-id]` (Glazed structured output)
  - [ ] `scenariolog timings --db [--run-id]` (Glazed structured output)
- [ ] Optional: dual-mode for `exec` (`--with-glaze-output`) so humans get nice text by default

### Phase 5 — Integrate into `test-scenarios/testing-doc-manager`

- [ ] Build `scenariolog` binary as part of the scenario run:
  - [ ] `go -C scenariolog build -o /tmp/scenariolog-local ./cmd/scenariolog`
  - [ ] pass the resulting path to the harness (e.g., `SCENARIOLOG_PATH=/tmp/scenariolog-local`)
- [ ] Update `run-all.sh` to:
  - [ ] create a run (`run start`)
  - [ ] wrap each step script invocation via `exec --kind step ...`
  - [ ] finalize run (`run end`)
- [ ] Update scenario README with:
  - [ ] how to run
  - [ ] where DB + artifacts live
  - [ ] how to query with `scenariolog search/failures/timings` and/or `sqlite3`

### Phase 6 — Hardening / polish

- [ ] Ensure ctrl-c produces a usable DB (run row + partial steps)
- [ ] Ensure sqlite locking is handled (busy timeout, single writer patterns)
- [ ] Add a “prune” or “cleanup” helper (optional)
