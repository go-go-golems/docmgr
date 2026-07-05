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
RelatedFiles: []
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

## Step 2: go-minitrace corpus discovery and conversion pipeline

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

## Step 3: Findings from the session mining

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

## Step 4: Design doc, bookkeeping, upload

(Filled in as the step completes; see changelog.)

### Prompt Context

**User prompt (verbatim):** "you can also investigate say, adding llm functionality to docmgr itself, to maybe alleviate the need to do loops in the top-level LLM harness."

**Assistant interpretation:** Add a design investigation for embedding LLM calls in docmgr (auto-relate, summarization, changelog drafting, doc QA) so the expensive top-level agent doesn't burn turns on bookkeeping loops.

**Inferred user intent:** Reduce agent-context cost and retry loops by moving mechanical-but-fuzzy documentation chores into docmgr with a cheap model.

### What I did
- Wrote `design-doc/01-improving-docmgr-for-coding-agents-analysis-design-and-implementation-guide.md` (architecture walkthrough, evidence, gap analysis, decision records, phased implementation plan, LLM-in-docmgr design).
- Related key files, updated changelog and tasks, ran `docmgr doctor`, committed, uploaded the bundle to reMarkable.

### Code review instructions
- Start with the design doc's executive summary and the "life of a related file" walkthrough; validate claims against `internal/paths/resolver.go` and `pkg/commands/relate.go`.
- Reproduce any statistic with the query commands under `scripts/query-commands/` (exact invocations in the doc's appendix).
