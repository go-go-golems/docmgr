---
Title: Investigation diary
Ticket: DOCMGR-FRICTION-001
Status: active
Topics:
    - docmgr
    - cli
    - usability
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://internal/documents/frontmatter.go
      Note: Lossless body framing, commit 91e0603
    - Path: repo://internal/documents/stability_test.go
      Note: Ten-cycle, CRLF and unknown metadata regressions
    - Path: repo://internal/documents/write.go
      Note: Atomic no-op writer and permission preservation
    - Path: repo://internal/httpapi/milestones.go
      Note: HTTP service parity
    - Path: repo://internal/operations/store.go
      Note: Prepared journals and cooperative recovery
    - Path: repo://pkg/commands/add.go
      Note: Explicit creation separator
    - Path: repo://pkg/commands/changelog_entries.go
      Note: Pure canonical entry builder
    - Path: repo://pkg/commands/changelog_stability_test.go
      Note: Boundary regression matrix
    - Path: repo://pkg/commands/close_service.go
      Note: Shared close planning
    - Path: repo://pkg/commands/create_ticket.go
      Note: Explicit index creation separator
    - Path: repo://pkg/commands/milestone_commands.go
      Note: CLI adapters
    - Path: repo://pkg/commands/milestone_service.go
      Note: Milestone and resume projections
    - Path: repo://pkg/doc/milestone-workflows.md
      Note: Public contracts and limitations
    - Path: repo://pkg/models/document.go
      Note: Unknown YAML node retention
    - Path: repo://ttmp/2026/09/10/DOCMGR-FRICTION-001--reduce-documentation-workflow-friction-with-deterministic-writes-and-coherent-milestones/design-doc/01-intern-guide-to-deterministic-document-mutation-and-milestone-workflows.md
      Note: Primary research handoff
    - Path: repo://ttmp/2026/09/10/DOCMGR-FRICTION-001--reduce-documentation-workflow-friction-with-deterministic-writes-and-coherent-milestones/scripts/03-render-guides.cjs
      Note: Reproducible diagram and print export
    - Path: repo://ttmp/2026/09/10/DOCMGR-FRICTION-001--reduce-documentation-workflow-friction-with-deterministic-writes-and-coherent-milestones/sources/implementation-validation.json
      Note: Full committed code inventory and validation evidence
    - Path: repo://ttmp/2026/09/10/DOCMGR-FRICTION-001--reduce-documentation-workflow-friction-with-deterministic-writes-and-coherent-milestones/sources/remarkable-upload.log
      Note: Verified delivery result
ExternalSources: []
Summary: Research chronology and review instructions for docmgr workflow friction.
LastUpdated: 2026-09-10T15:37:22Z
WhatFor: Preserve inspected evidence and explain the implementation handoff.
WhenToUse: Resuming this ticket or reviewing the reproductions.
---


# Investigation diary

## Goal

The original research explained session friction and delivered an intern guide without modifying product behavior. Steps 2 onward record the subsequently authorized implementation, commits, validation and limits.

## Step 1: Investigate persistence and lifecycle behavior

Created this ticket and the cross-linked skills ticket, inspected the relevant command/persistence APIs and converted a frozen copy of the motivating Pi session. Wrote the intern guide after reproducing the whitespace behavior rather than attributing every administrative failure to docmgr on the basis of memory.

### Prompt Context

**User prompt (verbatim):**

> ok, create a new docmgr ticket in ~/code/wesen/go-go-golems/docmgr itself about improving docmgr, and another one about improving the skills in ~/.pi/agent/skills (which is a git repo) itself.
>
> In both cases, Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.
>
> Cross reference the things if necessary. Do go into the details of this session in particular, reference which skills and other guidance was wrong specifically, etc...
>
> [REMINDER] Output a <summary>...</summary> block at the VERY END of your response. This is mandatory.

**Assistant interpretation:** Create two researched implementation handoffs and upload them; do not implement the proposed product/skill changes yet.

**Inferred user intent:** Make the session's avoidable work actionable for a new contributor while retaining useful correctness checks.

### What I did

- Read `AGENT.md`, document persistence, changelog, close, relation, task and HTTP boundaries.
- Built a local pinned CLI and ran the two ticket-contained reproducers in temporary directories.
- Ran existing focused document/changelog tests.
- Wrote the design guide with source maps, API sketches, decisions, phases and regression criteria.

### Why

The original EOF failures alone do not identify every root cause. A source-level reproducer separates formatting defects from excessive evidence-handling procedure.

### What worked

The frontmatter reproducer observed leading newline counts 1/2/3/4; changelog append produced two trailing newlines. The close probe observed repeated changes/history and an error after status had already become complete. Existing focused tests passed, showing the missing regression coverage. No live workspace was used for destructive fault injection.

### What didn't work

An initial discovery grep included an invalid literal newline regex and reported `rg: the literal "\n" is not allowed in a regex`; corrected the search to explicit symbols. Skills-repository status initially reported a missing `ttmp` root; initialized it as requested. The close probe's exit 1 is an intentionally injected failure, not a failed investigation. Final staging rejected trailing tab padding in the human `go version -m` capture; removed only that display whitespace. The first commit hook ran the full Go tests successfully, then errcheck rejected `01-reproduce-writes.go:25:20: Error return value of os.RemoveAll is not checked`. Replaced the unchecked deferred cleanup with a checked helper call; no product code or experiment policy changed.

### What I learned

The writer already uses per-file temporary replacement. The missing guarantee is stable framing and multi-file lifecycle consistency, not the complete absence of atomic file writes. `ApplyRelatedFilesUpdate` already has a no-op guard, so the report must not claim all mutations rewrite unchanged content.

### What was tricky to build

Distinguishing raw authored Markdown from serializer-owned separators, and choosing a deterministic close failure that does not depend on Unix permission assumptions. A directory at the changelog path reliably proves partial close behavior.

### What warrants a second pair of eyes

Review the proposed body-preservation contract, unknown metadata and permission behavior, operation-ID semantics and the difference between recoverability and atomic visibility. No new milestone API exists yet.

### What should be done in the future

Implement the open phases in tasks.md after the design is accepted. Keep this research ticket active rather than confusing guide delivery with implemented fixes.

### Code review instructions

Start with the design guide and `sources/source-map.txt`. Re-run scripts 01/02 from the docmgr root using a locally built binary. Compare JSON evidence and check that no product source changed.

### Technical details

Source pin: `710f31384d2d3242bfb84bf37cd06d2868a63bd3`. Session evidence is owned by SKILLS-FRICTION-001; both guides reference that frozen snapshot and the original project commit. Delivery/render receipts are stored under this ticket's sources directory. The guide and diary bundle was successfully uploaded to `/ai/2026/09/10/DOCMGR-FRICTION-001` and its listing verified. Both diagrams were inspected; the second was narrowed for print readability. Print copies add lowercase `title` metadata for the uploader while preserving canonical docmgr frontmatter. The uploaded diary is retained as `sources/print/diary.md`; this final delivery note was recorded afterward without requiring a recursive re-upload.

## Step 2: Establish byte-stable persistence contracts

Implementation began with the lowest-level reproduced defects, rather than teaching skills to repair generated files. The serializer now owns only delimiter termination; callers explicitly supply creation spacing. Changelog boundaries are canonical without trimming authored Markdown spaces.

The original research captures and uploaded snapshots remain historical evidence. New tests establish the changed behavior, including ten no-op cycles, CRLF body preservation, unknown metadata and file modes.

### Prompt Context

**User prompt (verbatim):**

> let's actually build it all. commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill).
>
> [REMINDER] Output a <summary>...</summary> block at the VERY END of your response. This is mandatory.

**Assistant interpretation:** Implement both researched tickets with focused commits and detailed investigation entries.

**Inferred user intent:** Replace the identified friction with tested tooling and coherent instructions, not another proposal.

**Commit (code):** `91e0603` — "fix: preserve document bodies and stabilize generated write boundaries"

### What I did
- Added a pure document serializer and an atomic, permission-preserving write-if-changed primitive.
- Retained unknown YAML fields through an inline node map; comments/style remain outside the fidelity guarantee.
- Audited creation versus metadata-update call sites; only create-ticket and doc-add need initial separators.
- Extracted a deterministic changelog builder, preserving internal Markdown and sorting file notes.
- Added regression tests using the standard testing library.

### Why
- Metadata-neutral cycles must not grow the body or replace unchanged files.
- Atomic single-file replacement is useful but must not imply multi-file consistency.

### What worked
- `go test ./internal/documents ./pkg/commands ./pkg/models` passes.
- Ten cycles retain identical bytes, sentinel mtime and mode; CRLF body and custom nested YAML survive metadata edits.

### What didn't work
- The first test command returned `go: updates to go.mod needed; to update it: go mod tidy` because the new tests imported testify, which this module does not directly use. Rewrote these tests with the standard library rather than adding a test-only dependency; the same focused command then passed.

### What I learned
- YAML formatting and Markdown fidelity need separate contracts. Canonical YAML can lose comments while retaining unknown values and exact body bytes.

### What was tricky to build
- Removing the extra serializer newline alone changes new-document presentation. An explicit CreationBody helper fixes that boundary without guessing whether existing leading blank lines were authored.
- Replacing an existing file via CreateTemp used to impose temporary-file permissions. The shared writer now preserves target permissions and rejects symlink/non-regular targets.

### What warrants a second pair of eyes
- Unknown YAML nodes with aliases and unusual scalar styles; first-write metadata normalization is intentionally not byte-identical YAML.
- File replacement remains subject to uncooperative filesystem races; no multi-file or adversarial filesystem guarantee is claimed.

### What should be done in the future
- Build cooperative operation locking, recoverable projections, shared close behavior, milestone and resume APIs in the next steps.

### Code review instructions
- Start at `internal/documents/frontmatter.go`, `write.go` and `stability_test.go`; then inspect changelog builder/tests and the two creation callers.
- Run the focused command above and the repository commit hooks. `format_file` is not exposed; the documented equivalent `gofmt` is used.

### Technical details
- Metadata framing uses LF; supplied body bytes, including CRLF and hard-break spaces, are untouched.
- New files default to 0644; existing permission bits survive replacement. Identical canonical bytes are not rewritten.

## Step 3: Recoverable operations, shared close and milestone services

The next layer replaces the misleading multi-file atomicity claim with a concrete recovery protocol. A prepared record stores intended projection bytes and before/after hashes; operations apply under a cooperative ticket-directory lock and return explicit recovery information on failure. Close writes history before status and both output modes call one service.

Milestones use the same store to complete stable task IDs and append history. Resume derives remaining tasks, evidence staleness and pending-operation warnings rather than creating another independently maintained current-state document.

### Prompt Context
**User prompt (verbatim):** (see Step 2)

**Assistant interpretation:** Build the accepted operation, milestone and resume phases with explicit consistency semantics and tests.

**Inferred user intent:** Eliminate manual reconciliation while retaining truthful error reporting and reviewable evidence.

**Commit (code):** `582171f` — "feat: recover ticket projections and share close and milestone operations"

### What I did
- Added a bounded operation journal, request identity checking, dry-run planning, cooperative locking and hash-checked recovery.
- Shared close planning between human and structured output; retained warnings for unfinished tasks.
- Added milestone services with stable-ID checks and ticket-contained evidence revisions, plus derived resume state.
- Started CLI/HTTP adapters against the repository's pinned Glazed v1.3.6 API rather than blindly applying the newer global skill's v1.4 output surface.
- Tested each journal/projection write boundary, replay identity, cancellation, concurrent duplicate requests, symlinks, human edits and status preflight failures.

### Why
- A sequence of renames is not atomic across paths. A recoverable plan provides inspectable intent without pretending external readers see a snapshot.
- Artifact SHA256 and stable task IDs prevent identity from depending on textual coincidence or mutable task positions.

### What worked
- `go test ./internal/operations ./pkg/commands` and `go test -race ./internal/operations` pass.
- Existing HTTP and command tests pass with the shared services and new registrations.
- Every selected interrupted stage recovers by retrying the same request; different payloads under an existing ID fail.

### What didn't work
- One cleanup edit was rejected before application: `Found 2 occurrences of edits[2]`. The repeated append line required surrounding context; a targeted unique replacement succeeded. No source changes were lost.
- The installed embedded verb-authoring help still contains removed layers/parameters examples. Current first-party command sources and the pinned module were used for actual API calls instead.

### What I learned
- Pretty-printing a journal reformats embedded raw JSON. Request identity must compact that JSON before hashing on replay, not hash incidental indentation.
- An interrupted default close must report its generated operation ID so the caller can explicitly recover it.

### What was tricky to build
- Locking the existing ticket directory avoids creating files during dry-run/resume and releases the lock on process exit. Supported platforms are Linux/macOS/FreeBSD; other builds return an explicit capability error.
- Recovery preflights every projection against before/after hashes, then checks each write again. Conflicts stop rather than overwrite an unrelated edit. External editors still have a narrow check/rename race and do not participate in the lock.
- Committed replay returns historical evidence rather than claiming files are still unchanged; resume separately reports staleness.

### What warrants a second pair of eyes
- Cooperative consistency only: legacy mutation commands and external editors are not globally locked.
- Process-crash recovery is not a power-loss durability guarantee; journals retain bounded projection bytes and should not contain secrets.
- HTTP must preserve committed receipts if a later index refresh fails.

### What should be done in the future
- Complete entry-point parity, public help, composed skills and final scenario validation in subsequent steps.

### Code review instructions
- Start with `internal/operations/store.go` and its fault matrix, then `close_service.go`, `milestone_service.go` and their tests.
- Review request/receipt bounds and filesystem containment independently from command rendering.

### Technical details
- Projection allowlist: index.md, tasks.md, changelog.md. Record IDs: 1–64 alphanumeric/underscore/hyphen characters, beginning alphanumeric.
- Bounded files/journals: 4 MiB each; history: 1024 records and 64 MiB; requests: 64 KiB; evidence: 64 refs; completed tasks: 128.

## Step 4: Finish API integration, harden review findings and validate the implementation

The new services are now exposed through registered dual-mode CLI commands, strict bounded HTTP requests, typed frontend hooks and embedded help. A pinned-binary smoke script exercises real CLI parsing, dry-run, bare/JSON replay, resume and close no-op behavior; HTTP tests compare the same persisted receipt against the service.

Review also uncovered two persistence edge cases and a relevant security baseline issue. Valid YAML must be parsed before scalar repair to preserve aliases, generated scaffolding needs the same EOF discipline as changelog entries, and the new os.Root containment code must run on a patched Go version. Those fixes were committed separately from the API integration.

### Prompt Context
**User prompt (verbatim):** (see Step 2)

**Assistant interpretation:** Finish the implemented interfaces, test integration and record evidence without confusing future adoption measurement with completed code.

**Inferred user intent:** Obtain usable, tested changes with a reviewable implementation history.

**Commit (code):** `7de2ea4` — "security: require patched Go and update reachable vulnerable dependencies"

**Commit (code):** `b1f53fc` — "fix: preserve valid YAML aliases and canonicalize generated scaffolds"

**Commit (code):** `1516abc` — "feat: expose milestone and resume workflows through CLI and HTTP"

### What I did
- Added milestone record and ticket resume adapters, HTTP routes, typed RTK Query endpoints and embedded help.
- Added actual CLI smoke and HTTP/service parity tests, and monotonic operation sequence numbers so wall-clock rollback cannot select an older checkpoint as latest.
- Tested scaffold EOF and no-op mtime, and aliases referring to known metadata fields. Scalar repair now runs only after YAML parse failure; alias expansion is bounded and avoids dangling anchors.
- Required Go 1.26.6 and upgraded the two reachable vulnerable dependency families identified by govulncheck, including their selected transitive updates.
- Wired the pinned-version Glazed analyzer into aggregate lint, generated the new command package's logcopter file and corrected the unavailable-formatter instruction in AGENT.md.
- Completed the companion skill implementation at `112f8aa`, with a detailed diary and nine policy/dependency tests.

### Why
- Real entry-point tests catch parsing and projection integration that pure service tests cannot establish.
- The scanner's os.Root finding concerns the precise containment primitive introduced here, so leaving the old toolchain preference was not appropriate.
- A controlled future-session study cannot be fabricated from deterministic helper fixtures.

### What worked
- Full Go tests with and without sqlite_fts5, race tests for operations/commands/HTTP, go vet, go build, aggregate lint and logcopter checks pass.
- The pinned-binary scenario suite completes successfully, as does the new CLI smoke.
- Frontend `pnpm exec tsc -b` and targeted ESLint pass; no UI component or rendering behavior was changed.
- Post-update govulncheck reports zero reachable vulnerabilities. It still reports one imported-package and five module-level advisories without reachable calls; this is not a claim that every dependency has no advisory.
- The skills checker and nine tests pass, and both pre-existing go-minitrace file hashes remain unchanged.

### What didn't work
- Alias regression initially failed: `stability_test.go:65: lost anchored metadata value`. The preprocessor had quoted valid `&heading`/`*heading` syntax into literal text. Parsing valid YAML first and expanding unknown-field aliases fixed the regression; the focused suite then passed.
- A source-discovery command guessed a nonexistent helper: `rg: pkg/commands/init_helpers.go: No such file or directory (os error 2)`. Located the actual `scaffold.go` before editing.
- The first `make govulncheck` failed with `Your code is affected by 9 vulnerabilities from 2 modules and the Go standard library.` The retained before-log includes GO-2026-4970 (os.Root), x/text and Excelize findings. After the listed fixed versions, the scan reports zero reachable findings.
- UI node_modules was absent (`ls: cannot access 'ui/node_modules/.bin/tsc': No such file or directory`); installed the unchanged lockfile with `pnpm install --frozen-lockfile --ignore-scripts`, then typechecked successfully.

### What I learned
- A content hash alone cannot order checkpoints if wall time moves backward; sequence numbers under the cooperative lock provide that ordering.
- Unknown YAML preservation requires retaining values, not simply storing a raw alias node whose anchor may be discarded by typed serialization.
- Routine docs-only commits now skip Go hooks, while code commits run full tests and both analyzers; there is no need to manufacture extra validation receipts for every prose edit.

### What was tricky to build
- HTTP mutation success must survive a later index-refresh error: return the committed receipt with a warning, not an apparent failed mutation that invites duplicate work.
- The original design assumed future command names; skill guidance now checks actual binary capabilities and uses the pinned Glazed output surface. The global binary was intentionally not overwritten with a CLI-only build that could discard its embedded UI.
- Research delivery manifests describe the earlier uploaded snapshots; they are not rolling hashes of documents subsequently updated with implementation notes.

### What warrants a second pair of eyes
- The documented cooperative-lock/external-editor boundary and journal archival/idempotency lifetime.
- The intentional close output-contract change and no-op policy for already matching status/intent.
- Lower-administration workflow claims require real post-adoption evidence, not just passing policy fixtures.

### What should be done in the future
- SKILLS-FRICTION-001 task vewh remains open for equivalent real sessions. Non-reachable dependency advisories remain a dependency-maintenance concern, not an ignored reachable scan failure.

### Code review instructions
- Review the five code commits separately, then run the commands recorded in `sources/implementation-validation.json`.
- Use `/tmp/docmgr-friction-local` for the verified CLI. `scripts/04-milestone-smoke.py` and the repository scenario suite write only isolated temporary workspaces.
- Inspect both ticket diaries for evidence and limits; unrelated skills files and original research uploads were preserved.

### Technical details
- Final toolchain used: `go version go1.26.6 linux/amd64`.
- HTTP: POST /api/v1/tickets/milestone, GET /api/v1/tickets/resume; CLI: milestone record, ticket resume, shared ticket close.
- New journal directories use 0700. File hashes are evidence revisions, not authorization or automatic proof of a claim.

## Step 5: Reconcile recorded closure without false stale warnings

Before using the new commands on their own tickets, review showed that resume compared projections only with the last milestone. A subsequent recorded close legitimately updates history and status, so it should become the latest known owner of those projection hashes without replacing the milestone's narrative checkpoint.

Resume now derives each projection's expected hash from the latest committed operation, while retaining the latest milestone's phase, next action and evidence. Unrecorded edits still trigger warnings.

### Prompt Context
**User prompt (verbatim):** (see Step 2)

**Assistant interpretation:** Validate the combined milestone-then-close lifecycle before final ticket bookkeeping.

**Inferred user intent:** Make the new workflow coherent in actual use, not merely correct as isolated commands.

**Commit (code):** `1d0710d` — "fix: reconcile resume against the latest recorded projections"

### What I did
- Derived current projection expectations from all committed records.
- Added a regression covering milestone, close, clean resume, then an unrecorded edit.

### Why
- A known, recorded state transition should not create a false stale-evidence alarm.

### What worked
- Focused command/HTTP tests and full commit-hook tests/lint pass. Recorded closure is clean; subsequent unrecorded history changes are still detected.

### What didn't work
- No failed test run in this step; the edge case was identified during lifecycle review.

### What I learned
- Checkpoint identity and projection ownership are related but distinct views of the same journal.

### What was tricky to build
- Updating expected projection hashes must not discard the milestone's evidence revisions or mask genuinely unrecorded changes. A stable projection-name iteration also keeps warning order deterministic.

### What warrants a second pair of eyes
- Review combined operations, not only each endpoint separately.

### What should be done in the future
- No additional implementation follow-up from this step; the independent real-session skills comparison remains open.

### Code review instructions
- Read `milestone_service.go` and `resume_projection_test.go`; run `go test ./pkg/commands ./internal/httpapi`.

### Technical details
- Latest milestone remains the source of phase/next/evidence. Latest committed projection owner supplies current expected hashes.
- Used the new milestone service to complete this ticket's remaining implementation tasks, then closed it with explicit operation ID `implementation-closed`. The real resume view reports complete, zero remaining tasks and zero conflicts; doctor passes. Receipts and the bounded journals are retained in this ticket. The companion skills ticket reports only vewh remaining and also passes doctor. Both real resume outputs pass the new helper's shape check.

## Step 6: Install the full local binary and assess PR readiness

The implementation checkpoint deliberately left the global executable unchanged. The user now requested installation. I built the embedded-UI/FTS5 executable from clean revision e607822, backed up the previous executable and atomically replaced the PATH-resolved local binary. This supersedes the earlier not-installed checkpoint; the skills already reside in the live shared skill directory.

The installed CLI smoke and embedded-package tests pass. Fresh remote comparisons show both branches ahead without divergence. Exact CI-pinned lint and the CI GoSec command also pass locally, so the changes are ready to open as two repository-specific PRs, subject to actual remote checks.

### Prompt Context
**User prompt (verbatim):**
> install locally, and are we ready for a PR?
>
> [REMINDER] Output a <summary>...</summary> block at the VERY END of your response. This is mandatory.

**Assistant interpretation:** Install the complete local executable and assess readiness without pushing or opening PRs.

**Inferred user intent:** Start using the changes locally and decide whether review can begin.

### What I did
- Ran `make build-embed`, backed up the old binary, and installed `/home/manuel/.local/bin/docmgr` via a temporary file and rename.
- Verified Go 1.26.6, sqlite_fts5/embed tags, clean build revision and installed-binary smoke behavior.
- Fetched both origins, checked ancestry and committed diff whitespace, and inspected CI configuration/current base runs.
- Ran golangci-lint v2.12.2 and GoSec with the repository's CI exclusions using temporary tool installations.

### Why
- Installing the earlier CLI-only test build could discard the existing embedded web UI.
- Local aggregate lint alone is not proof that the exact GitHub lint configuration passes.

### What worked
- Full build, installed smoke, tagged web/HTTP tests, pinned lint and GoSec pass.
- Docmgr was nine commits ahead/zero behind; skills four ahead/zero behind before this installation note. No push or PR creation occurred.

### What didn't work
- `git -C /home/manuel/.pi/agent/skills symbolic-ref refs/remotes/origin/HEAD` returned `fatal: ref refs/remotes/origin/HEAD is not a symbolic ref`, interrupting the first combined preflight command. `git ls-remote --symref origin HEAD` confirmed main; explicit origin/main comparison succeeded. No repository configuration change was necessary.

### What I learned
- The skills repository lacks a local remote-HEAD symbolic ref despite having a valid remote default branch.

### What was tricky to build
- Preserve the full local installation while keeping unrelated skills edits and existing processes untouched.

### What warrants a second pair of eyes
- Review the operation recovery contract and API changes; local success does not replace remote CI or human review. Scheduled dependency checks on the unchanged remote base are failing; the previously recorded local vulnerability check passes after our dependency updates.

### What should be done in the future
- Create/push topic branches and open one PR per repository when requested. Real post-adoption efficiency measurement remains a nonblocking follow-up.

### Code review instructions
- Review origin/main...HEAD in each repository and the focused implementation commits. Installation identity and preflight results are in `sources/local-installation.json`.

### Technical details
- Installed SHA256: `449999482b7fbe9f1bc265a341b7e956bf4f699da50ede6a2aea73b76f462766`.
- Backup: `/tmp/docmgr-before-friction-install`. Build/install generated no tracked product changes.
