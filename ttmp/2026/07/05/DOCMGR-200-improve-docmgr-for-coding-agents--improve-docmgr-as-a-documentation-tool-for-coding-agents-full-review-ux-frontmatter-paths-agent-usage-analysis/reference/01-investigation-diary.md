---
Title: Investigation diary
Ticket: DOCMGR-200-improve-docmgr-for-coding-agents
Status: active
Topics:
    - docmgr
    - ux
    - cli
    - documentation
    - tooling
DocType: reference
Intent: ticket-specific
Owners: []
RelatedFiles:
    - Path: ttmp/2026/07/05/DOCMGR-200-improve-docmgr-for-coding-agents--improve-docmgr-as-a-documentation-tool-for-coding-agents-full-review-ux-frontmatter-paths-agent-usage-analysis/scripts/03-stage-and-convert.sh
      Note: Session staging + minitrace conversion pipeline built in Step 3
    - Path: ttmp/2026/07/05/DOCMGR-200-improve-docmgr-for-coding-agents--improve-docmgr-as-a-documentation-tool-for-coding-agents-full-review-ux-frontmatter-paths-agent-usage-analysis/scripts/query-commands/docmgr/usage.js
      Note: JS analysis verbs (command-freq / error-summary / retry-chains / session-stats) from Step 3
    - Path: ttmp/2026/07/05/DOCMGR-200-improve-docmgr-for-coding-agents--improve-docmgr-as-a-documentation-tool-for-coding-agents-full-review-ux-frontmatter-paths-agent-usage-analysis/sources/minitrace-command-freq.json
      Note: Raw per-framework frequency data behind Step 4 findings
ExternalSources: []
Summary: Chronological diary of the DOCMGR-200 investigation - codebase review fan-out, go-minitrace JS analysis of agent sessions, and findings.
LastUpdated: 2026-07-05T18:50:00-04:00
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

Capture the full investigation journey for DOCMGR-200: reviewing docmgr end-to-end (CLI, internals, web UI, peripheral features), mining past pi/codex/claude coding sessions with go-minitrace for real docmgr usage patterns and failure modes, and turning both into an intern-ready analysis/design/implementation guide.

## Step 1: Ticket setup and review fan-out

Created the DOCMGR-200 ticket workspace and split the codebase review into four parallel review tracks (CLI surface, core internals/paths, web UI + HTTP API, peripheral subsystems), while keeping the agent-transcript analysis for the main thread. The four reviews came back with heavily file:line-anchored reports that now form the evidence base for the design doc.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to improve docmgr itself, in ./docmgr . First, just do a full review of the project. You can run experiments in the scripts/ folder of the docmgr ticket. Keep a detailed diary as you work, commit at appropriate intervals. We want to look at everything that could be worth improving, especially in terms of UI/UX both of the UI, and the CLI. There are also functionalities we barely use, and the paths handling of the frontmatter is tricky and all that. You can really imagine you could do anything to improve it as a tool for coding agents to maintain consistent documentation. You can also use `go-minitrace help --all` and use go-minitrace to write JS and SQL scripts to examine past coding sessions for pi and codex and claude and examine their use of docmgr in tool calls, what they struggle with, etc... Create a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Full evidence-based review of docmgr (CLI UX, web UI, rarely-used features, frontmatter path handling), plus empirical analysis of how coding agents actually use and struggle with docmgr, mined from historical transcripts via go-minitrace; deliver one intern-level analysis/design/implementation guide in the ticket and upload it to reMarkable.

**Inferred user intent:** Make docmgr a genuinely better tool for coding agents that maintain documentation, prioritizing improvements by what agents actually hit in practice rather than speculation.

### What I did
- `docmgr ticket create-ticket --ticket DOCMGR-200-improve-docmgr-for-coding-agents ...` with topics docmgr,ux,cli,documentation,tooling; added the design doc and this diary via `docmgr doc add`.
- Launched four parallel read-only review agents over the repo (`cmd/`+`pkg/commands`, `internal/paths`+frontmatter model, `ui/`+`internal/httpapi`, peripheral subsystems + tests).

### Why
- The repo is ~35k LOC Go plus a React SPA; four scoped reviewers with file:line requirements produce better coverage than one linear pass, and it freed the main thread for the minitrace work that needed interactive iteration.

### What worked
- All four reviews returned with concrete, verifiable anchors. Highlights: three exit-0 silent failures in the CLI (`relate` malformed `--file-note`, `meta update` field errors, empty `changelog update`); the fts5 build-tag trap (`Error: fts5 not available (docs_fts missing)` on a default `go build`); the internals review reduced the paths problem to one sentence - the frontmatter stores a bare string whose meaning is reconstructed at read time by a 5-anchor, existence-dependent, CWD-sensitive guesser, and the write side (`relate`), index side (`NormalizeNoFS`), doctor side, and UI side interpret it with four different algorithms.

### What didn't work
- N/A for this step (agents completed cleanly).

### What I learned
- `doc relate` itself manufactures `../../../../glazed/...` entries for cross-repo files that doctor then flags as missing (`internal/paths/resolver.go:110-114` repo-containment guard vs `pkg/commands/relate.go:653-672` canonicalization) - the tool disagrees with itself.
- The UI is a read-mostly browser: exactly two write endpoints (task check/add); mermaid renders only in the graph tab, not in doc bodies.

### Code review instructions
- The four review reports are summarized and cited throughout the design doc (`design-doc/01-...md`); spot-check any claim via its file:line anchor in the docmgr repo.

### Technical details
- Exact setup commands:
  - `docmgr ticket create-ticket --ticket DOCMGR-200-improve-docmgr-for-coding-agents --title "Improve docmgr as a documentation tool for coding agents: full review, UX, frontmatter paths, agent-usage analysis" --topics docmgr,ux,cli,documentation,tooling`
  - `docmgr doc add --ticket DOCMGR-200-improve-docmgr-for-coding-agents --doc-type design-doc --title "Improving docmgr for coding agents: analysis, design, and implementation guide"`
  - `docmgr doc add --ticket DOCMGR-200-improve-docmgr-for-coding-agents --doc-type reference --title "Investigation diary"`
- Repo size at review time: 258 Go files / ~35,007 LOC (main module), React SPA under `ui/`, separate `scenariolog` module, 53 existing tickets / 232 docs in `ttmp/`.
- Vocabulary already contained suitable topics (docmgr, ux, cli, documentation, tooling), so no `vocab add` was needed at setup time.

## Step 2: Codebase review findings (four parallel reviews)

The four scoped reviews returned dense, file:line-anchored reports. This step records their headline findings so the diary is self-contained; the design doc integrates and cites them in full.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

### What I did
- Collected and cross-checked the four reports (CLI surface; core internals/paths; web UI + HTTP API; peripheral subsystems + tests + docs accuracy).

### Key findings per track

**CLI surface** (cmd/, pkg/commands):
- Three exit-0 silent failures: malformed `--file-note` silently skipped (`pkg/commands/relate.go:460-462`); `meta update` prints per-file errors but returns nil (`pkg/commands/meta_update.go:297-305`); `changelog update` without `--entry` writes an empty dated heading (`pkg/commands/changelog.go:65`).
- fts5 build-tag trap: default `go build` → `Error: fts5 not available (docs_fts missing)` on any text query (`internal/workspace/query_docs.go:36-37`); only the Makefile builds `-tags sqlite_fts5`.
- `ignore explain` human mode is broken: `Error: no non-Glaze run method implemented for *commands.IgnoreExplainCommand` (only `RunIntoGlazeProcessor` implemented, wired as dual-mode).
- Doctor issue numbering stuck at `1)` (counter resets per taxonomy: `pkg/diagnostics/render/render.go:19` vs `adapter.go:99-119`); doctor's suggested `vocab add --category doctype` command is invalid twice over (bad category, missing required `--description`).
- Fresh `init` + first ticket immediately fails doctor (vocabulary not seeded by default but create-ticket writes vocab-checked values).
- Naming: `ticket create-ticket`/`rename-ticket` stutter; `ticket tickets` canonical vs `doc list`; same op spelled 2-4 ways (root + workspace group, `list` group, `search` alias).
- Dead code: `pkg/commands/list.go` (`NewListCommand`) never registered.
- Output noise: 3-line root/config/vocabulary banner on every command; unconditional reminder nags on task/changelog mutations; `doctor --all` = 112 KB on this repo.

**Core internals / paths** (pkg/models, internal/paths, internal/workspace):
- `models.Document` has no `omitempty`; rewrites materialize all 13 keys incl. `LastUpdated: 0001-01-01T00:00:00Z`. RelatedFiles unmarshal silently drops malformed entries (`pkg/models/document.go:341`).
- `NormalizedPath` carries six representations; `Normalize` picks the first anchor where the file exists among repo/doc/config/docsRoot/docsParent (`internal/paths/resolver.go:90-131`); `NormalizeNoFS` (used by the index) picks the first valid base - the two can disagree for the same string.
- Absolute inputs short-circuit with `Exists=true` and are never stat'd (resolver.go:86-88).
- Cross-repo relate writes `../../../../glazed/...` doc-relative entries that doctor then can't resolve (doc anchor must stay inside repo, resolver.go:110-114) - the tool contradicts itself; UI 403s the same entries (`internal/httpapi/path_safety.go:48-53`, `RelatedFilesList.tsx:37` assumes repo-relative).
- `PreprocessYAML` has a dead nested-structure guard (`pkg/frontmatter/frontmatter.go:56` checks prefix after TrimLeft) so indented block-scalar lines can be rewritten at read time.
- tasks.md IDs are positional and recomputed each parse (`internal/tasksmd/tasksmd.go:84-97`).
- Doctor validates RelatedFiles/vocab/staleness only on index.md (doctor.go:411-630); `is_sources_path` is a dead column (computed, stored, never filtered).
- `go.work` never consulted; RepoRoot has two different definitions (`config.go:297-323` incl. legacy `doc/` marker vs `resolver.go:385-401`).

**Web UI / HTTP API** (ui/, internal/httpapi, internal/web):
- 20 REST routes under `/api/v1`; reads from an in-memory snapshot, no file watching, manual `POST /index/refresh`.
- UI is read-mostly: exactly two write endpoints (task check, task add - section hardcoded to TODO). No doc create/edit, no meta update, no relate, no doctor, no changelog.
- Mermaid renders only in the graph tab; doc bodies use a bare `MarkdownBlock` (no component overrides) - diagrams in docs are invisible; relative doc links and images break.
- Careful path-traversal defenses in `path_safety.go` (normalization + EvalSymlinks + NUL checks) with tests.
- Hygiene: `StatusBadge` copy-pasted 5x, duplicate legacy CSS selectors, dev-port 3001 hardcoded in error hints while embedded serve defaults to 8787, changelog tab is a link stub.

**Peripheral subsystems**:
- Search is SQLite FTS5, not bleve; sibling repos goja-bleve/goja-text/go-go-goja have zero integration with docmgr.
- Skills subsystem (1,447 LOC) active and tested; templates stable; verb `.templ` output templates niche but working (9 verbs, `--print-template-schema`).
- scenariolog: separate module, needs `GOWORK=off`, glazed pinned v1.0.5 vs main v1.3.6.
- Tests: all green in 3 configs, but CI runs untagged tests only - the FTS5 production path and the 22-scenario E2E suite never run in CI.
- Docs drift: README broken code fence + missing features (API/UI/skills); AGENT.md describes a different repo (ttmp/YYYY-MM-DD/, doc/, bun+templ); CONTRIBUTING points at nonexistent `pkg/commands/config.go` and obsolete glazed APIs; `test-api.sh` is dead (pre-`/api/v1` routes).

### What worked
- Requiring file:line anchors in each report made cross-checking cheap; several claims (silent failures, fts5 error, doctor numbering) were verified by the reviewers actually running the commands in sandboxes.

### What didn't work
- N/A - all four reviews completed.

### What warrants a second pair of eyes
- The `PreprocessYAML` dead-guard claim (frontmatter.go:56) and the `Exists=true` short-circuit (resolver.go:86-88) are the two most consequential internals claims; re-verify before building on them.

## Step 3: go-minitrace corpus discovery and conversion pipeline

Started with raw `rg` greps over the three native transcript stores to size the corpus, then - after user feedback - pivoted to a proper go-minitrace pipeline: stage a top-usage sample of sessions into store-shaped trees, convert to minitrace archives, and do all real analysis as reusable JS query commands stored in this ticket's `scripts/query-commands/`.

### Prompt Context

**User prompt (verbatim):** "why are you using grep and not go-minitrace? can's you write go-minitrace JS scripts?" - and later: "go-minitrace help --all should give you information about JS API and all" and "the source is here btw if you need go-minitrace information. I really wantyou to explore what is possible with it. ~/code/wesen/go-go-golems/go-minitrace and also suggest improvements and what was tricky."

**Assistant interpretation:** Do the transcript analysis with go-minitrace's structured JS/SQL query commands, not ad-hoc grep; also treat go-minitrace itself as an object of study and report friction/improvement ideas.

**Inferred user intent:** Exercise and dogfood go-minitrace's JS query-command surface on a real analysis task; get transferable, reusable analysis scripts instead of one-off shell pipelines.

### What I did
- Corpus sizing: 1,017 codex, 825 pi, 145 claude session files contain real docmgr subcommand invocations (`scripts/01-docmgr-command-frequency.sh` writes per-store command/flag frequency tables; grep stayed only as the narrowing pass since ~2,000 sessions / many GB was too much to convert wholesale).
- `scripts/02-docmgr-error-patterns.sh`: error-signature counts across the raw stores (patterns seeded from `errors.Errorf` strings in docmgr source).
- `scripts/03-stage-and-convert.sh`: selects top-N sessions by docmgr-hit count per store plus every STRIDE-th of the remainder, symlink-stages them into `.codex`/`.pi`/`.claude`-shaped trees, and runs `go-minitrace convert codex|pi|claude-code`. Full run: 98 codex + 88 pi + 54 claude sessions -> ~1.1 GB of minitrace archives in a few minutes.
- JS query commands in `scripts/query-commands/docmgr/`: `probe.js` (schema dump), `usage.js` (command-freq, error-summary, retry-chains, session-stats), `paths.js` (path-shapes for `--doc`/`--file-note`), `volume.js` (output bytes per verb). Results saved as JSON under `sources/`.

### Why
- Structured tool_calls tables (command, success, exit_code, error, result, position_in_session) let us measure actual executions, not prompt-echo. The raw grep counts were inflated ~7,000x for `doc search` (169k occurrences vs 22 real calls in the converted sample) because system prompts and AGENT.md instructions echo docmgr commands in every request payload.

### What worked
- The builder JS API (`mt.db().RuntimeArchives().QueryCommandDefaults().MaxRows(500000).Build()`) with SQL + JS post-processing was a good fit; `__section__`/`__verb__` markers turned each analysis into a typed CLI command with `--framework`/`--limit`/`--flag` flags for free.
- Symlink staging works: converters follow symlinked JSONL files, so no copying of multi-GB stores was needed.

### What didn't work
- `bash -euo pipefail` + `rg -c` exits 1 on zero matches and killed script 02 mid-run; fixed with `(rg ... || true)`.
- Pi's slugged session dirs start with `--`, so `dirname $rel` parsed them as flags: `dirname: unrecognized option '--home-manuel-workspaces-...'`. Fixed with `dirname -- "$rel"`.
- First probe attempt: `SELECT ... FROM sqlite_master` -> `GoError: query references disallowed table/view "sqlite_master"`. The JS query validator allowlists only the normalized tables; used `db.schema()` instead.
- `zsh` ate a bare `echo ===` separator (`(eval):1: == not found`) - quote separators.

### What I learned
- Normalized schema is rich: `tool_calls` carries `command`, `success`, `exit_code`, `error`, `result`, `position_in_session`, `tools_before_json` - enough for retry-chain reconstruction without touching raw JSONL.
- Codex converter takes a *home*-shaped dir (`--source-dir X` where X contains `sessions/`), pi/claude take the sessions/projects dir directly - easy to get wrong.

### What was tricky to build
- The invocation extractor: shell commands in transcripts include `cd x && docmgr ... && docmgr ...`, heredocs, and prose echoes. A conservative regex over `tool_calls.command` plus per-call aggregation got clean numbers; prose noise like "docmgr and re" only appears in the raw-grep pass, not in tool_calls.

### What warrants a second pair of eyes
- The `success`/`exit_code` semantics differ per adapter (pi rows have NULL duration_ms; the claude sample is small at 17 sessions with docmgr calls). Failure *rates* are comparable within a framework, less so across frameworks.

### What should be done in the future
- go-minitrace improvement notes (also in the final report): allow `sqlite_master` or point at `db.schema()` in the validator error text; codex/claude converters accept only `--source-dir` (no `--source-session`), forcing the staging dance; a `discover --with-cwd` filter would remove the grep-narrowing step entirely; pi adapter doesn't map `duration_ms`.

### Code review instructions
- Re-run everything: `scripts/03-stage-and-convert.sh <hits-dir> <work-dir> 50 20`, then any `go-minitrace query commands docmgr <group> <verb> --query-repository scripts/query-commands --archive-glob '<work>/archive/*/active/*/*.minitrace.json'`.
- Saved outputs in `sources/minitrace-*.json`.

## Step 4: Findings from the session mining

The numbers align remarkably well with the static review. Sample (139 sessions with docmgr activity, 14,166 docmgr tool calls): pi 8,049 calls / 74 sessions, codex 5,819 / 48, claude-code 298 / 17.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

### What I did
- Ran command-freq, error-summary, retry-chains, path-shapes, output-volume across all three frameworks; saved JSON to `sources/`.

### What worked / key findings
- **Usage is concentrated on bookkeeping verbs**: doctor 3,330; doc relate 2,487; changelog update 1,861; task add 1,519; doc add 1,474; task check 1,049. Search is almost unused as a tool call (22 real `doc search` calls) despite being heavily promoted in prompts.
- **Failure-rate outliers**: `validate frontmatter` 22.9% (494 calls), `ticket show` 46.7% (command doesn't exist - agents guess it), `ticket create` 28.6% (they mean `create-ticket`), `vocab add` 5.4%, `doc relate` 3.7% (91 failures).
- **Retry tax**: 380 failure->next-docmgr-call chains; 58% retried the same verb immediately. Biggest retry sinks: `validate frontmatter` (78), `doc relate` (52).
- **Path-shape confusion is measurable**: agents pass `--doc` in 7+ shapes; most common is `ttmp/`-prefixed (1,107 uses, 7.6-14% failure), docs-root-relative date paths fail 11-18%, absolute paths are most reliable (2.5-5.6%). The `file-not-found` error class (96) includes the telltale `ttmp/ttmp/` double-join. `--file-note` is dominated by absolute paths (9,353 uses, 3.2% failure) with 31 `../../..`-chain cross-repo uses.
- **Error taxonomy** (error-summary): file-not-found 96; frontmatter-parse 57; unknown-flag 54 (`--ticket` on the `ticket` group, `--plain`); "expected exactly 1 doc for --doc" ~37; ticket-not-found 27 (agents pass the directory slug instead of the ID); ticket-ambiguous 18; task `--id` non-integer 5 (agents invent IDs like `F2.5.1`).
- **Context cost**: `ticket create-ticket` output averages 3.1 KB/call (the created-files banner); doctor 758 B avg x 1,999 calls = 1.4 MB of context across the sample; outputs are truncated at ~10 KB by agent harnesses, so anything bigger is wasted.

### What didn't work
- A quick save-loop with string surgery over command names wrote empty JSON files; re-ran each save explicitly.

### What warrants a second pair of eyes
- Failure classification regexes (`ERROR_CLASSES` in `usage.js`) are first-match-wins; "other" is still 80 items and contains e.g. `--doc` resolution failures wrapped in doctor report text.

### What should be done in the future
- The design doc turns these findings into prioritized improvements (silent-failure fixes, path-anchor scheme, verb aliases for what agents guess, output slimming) plus a design for LLM functionality inside docmgr itself (per user follow-up request).

## Step 5: Design doc, bookkeeping, and two live bug reproductions

Wrote the full design doc (14 sections: mental model, subsystem tour with mermaid diagrams, evidence tables, gap analysis G1-G12, design D1-D8 with four decision records, phased plan P0-P5, testing strategy, go-minitrace appendix, reproduction commands, key-file reference). While doing routine docmgr bookkeeping for this very ticket, docmgr reproduced two of the documented bugs on its own maintainers' workflow - the best possible validation of the analysis.

### Prompt Context

**User prompt (verbatim):** "you can also investigate say, adding llm functionality to docmgr itself, to maybe alleviate the need to do loops in the top-level LLM harness." - and: "write the diary as you go, not just at the end. So fill it out now, in detail, for all the investigation you already did"

**Assistant interpretation:** Add a design investigation for embedding LLM calls in docmgr (auto-relate, summarization, changelog drafting, frontmatter repair, doc QA) so the expensive top-level agent doesn't burn turns on bookkeeping loops; and keep the diary current during the work.

**Inferred user intent:** Reduce agent-context cost and retry loops by moving mechanical-but-fuzzy documentation chores into docmgr with a cheap model; keep the investigation auditable in real time.

### What I did
- Wrote `design-doc/01-improving-docmgr-for-coding-agents-analysis-design-and-implementation-guide.md` in four passes (sections 1-5 orientation/architecture; 6 evidence; 7-8 gaps/design incl. `docmgr ai` as D7 with a decision record on where LLM calls live; 9-14 plan/testing/risks/appendices).
- Bookkeeping: `docmgr doc relate` (7 file-notes on the design doc, 3 on the diary), `task add` x5 + `task check`, `changelog update`, `docmgr doctor --ticket DOCMGR-200-...`.

### What worked
- The design doc integrates all five evidence streams (4 reviews + minitrace) with every claim anchored; the `docmgr ai` design reuses the existing suggest/apply pattern from `doc relate --suggest/--apply-suggestions` as its guard-rail model.

### What didn't work (live bug reproductions!)
- **Comma-split file-notes**: `docmgr doc relate --file-note ".../resolver.go:...trickiness (sections 4.4, 8.1)"` stored the note truncated to "...(sections 4.4" - glazed parses the repeatable flag as a comma-separated stringList, and the post-comma fragment " 8.1)..." (no colon) was **silently dropped with exit 0** - exactly the G1 silent-failure bug documented hours earlier. Fixed the notes by editing frontmatter directly; added the reproduction to the design doc's G1 row.
- **Positional task-ID trap**: `docmgr task check --id 1,2,3,4` checked the scaffolded placeholder "Add tasks here" as ID 1 instead of my first real task; after `task remove --id 1` all IDs shifted again - exactly the D5/G8 instability. The scaffold placeholder also counts as an open task in `list tickets` until manually removed.

### What I learned
- Dogfooding the bookkeeping loop on the tool's own repo is the fastest way to hit the agent-facing bugs; both reproductions are now cited as first-party evidence in the design doc.

### What warrants a second pair of eyes
- The comma-split behavior means every historical `--file-note` containing a comma has a truncated note in frontmatter across all 53 tickets in this repo (and other repos) - worth a one-off audit query before/with the D1 migration.

### What should be done in the future
- P0 fix list in design doc section 9 (the comma-split fix joins it: make `--file-note` a repeatable plain string, not a comma list).

### Code review instructions
- Start with the design doc's executive summary and the "life of a related file" walkthrough (section 4.4); validate claims against `internal/paths/resolver.go` and `pkg/commands/relate.go`.
- Reproduce any statistic with the query commands under `scripts/query-commands/` (exact invocations in design doc Appendix B).
- Reproduce the comma-split bug: `docmgr doc relate --doc <any> --file-note "/abs/path:note with, a comma"` and inspect the stored Note.
