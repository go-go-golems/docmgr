---
title: "Intern guide to deterministic document mutation and milestone workflows"
Title: Intern guide to deterministic document mutation and milestone workflows
Ticket: DOCMGR-FRICTION-001
Status: active
Topics:
    - docmgr
    - cli
    - usability
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: abs:///home/manuel/.pi/agent/skills/ttmp/2026/09/10/SKILLS-FRICTION-001--make-skills-phase-aware-proportional-and-consistent-across-long-sessions/design-doc/01-intern-guide-to-phase-aware-skills-and-evidence-proportional-workflows.md
      Note: Companion instruction and workflow guide
    - Path: repo://internal/documents/frontmatter.go
      Note: Reproduced separator growth and existing atomic replacement
    - Path: repo://pkg/commands/changelog_entries.go
      Note: Shared append formatting
    - Path: repo://pkg/commands/ticket_close.go
      Note: Partial close and duplicate entry point logic
    - Path: repo://ttmp/2026/09/10/DOCMGR-FRICTION-001--reduce-documentation-workflow-friction-with-deterministic-writes-and-coherent-milestones/scripts/01-reproduce-writes.go
      Note: Executed byte reproduction
    - Path: repo://ttmp/2026/09/10/DOCMGR-FRICTION-001--reduce-documentation-workflow-friction-with-deterministic-writes-and-coherent-milestones/scripts/02-reproduce-close.py
      Note: Executed lifecycle failure probe
ExternalSources: []
Summary: Evidence-backed architecture and implementation plan for stable writes, recoverable lifecycle updates, structured milestone receipts and resumable ticket state.
LastUpdated: 2026-09-10T15:37:22Z
WhatFor: Give an intern enough source context and executable evidence to improve docmgr without weakening validation or silently rewriting authored content.
WhenToUse: Before implementing documentation mutation or workflow improvements.
---


# Deterministic document mutation and coherent milestones

## 1. Executive summary and scope

Docmgr organizes engineering documentation into ticket workspaces. A ticket contains an index, tasks, changelog and typed documents with YAML frontmatter. Agents use its CLI to keep implementation evidence connected to those documents. The underlying information is useful; unnecessary file changes and repeated bookkeeping make maintaining it expensive.

This ticket comes from the Video Observatory engineering session `01a0837a-fb1c-7183-9bf5-26fce6d60551`, spanning September 9–10, 2026. That session implemented a synthetic media lab and published eight textbook chapters. Its final audit and commits are real deliverables, not evidence that every administrative step was necessary. The investigation separates defects in docmgr from excessive agent procedure and from user-requested rigor.

Three findings are now reproduced against this repository's source:

1. Frontmatter read/write cycles grow the body's leading newline count: **1, 2, 3, 4** over successive observations.
2. A newly appended changelog entry ends in **two newline bytes** and can fail Git's added-blank-EOF check.
3. `ticket close` can return an error while leaving `index.md` marked **complete**, because the index is written before a failing changelog append. The command's help nevertheless says it closes a ticket “Atomically.”

A repeated successful close also changes index bytes again and appends a second “Ticket closed” entry. Whether repeated close should append is a policy decision; it must not be conflated with an accidental serialization defect.

The first implementation phase should fix predictable persistence behavior and misleading guarantees. A later phase can introduce one structured milestone operation and a compact resume view. Do not start by building a large workflow engine, automatically editing prose, or removing meaningful tests.

**Companion ticket:** `SKILLS-FRICTION-001`, in `/home/manuel/.pi/agent/skills`, under `ttmp/2026/09/10/`. Run `docmgr ticket show SKILLS-FRICTION-001` there to locate its workspace; this document's RelatedFiles metadata links its exact guide. Its intern guide owns diary proportionality, conflicting upload instructions, skill loading and validation policy. This ticket owns the CLI, persistence and data contracts that make a better workflow possible.

## 2. Repository orientation

The inspected source revision is `710f31384d2d3242bfb84bf37cd06d2868a63bd3`. The working tree was clean before these ticket documents were created. Reproducers build or call this checkout; they do not assume the installed `docmgr` binary equals HEAD. Installed build information is retained in `sources/installed-binary.txt` and the locally built binary fingerprint in `sources/close-reproduction.json`.

### 2.1 Entry points and responsibilities

| Layer | Concrete files and APIs | Responsibility |
|---|---|---|
| CLI assembly | `cmd/docmgr/cmds/root.go:27`, `NewRootCommand`; verb packages expose `Attach` | Register command groups and integrate help. |
| Commands | `pkg/commands/ticket_close.go`, `changelog.go`, `tasks.go`, `relate.go` | Decode settings, discover workspace, invoke operations and report outcomes. |
| Shared mutations | `pkg/commands/changelog_entries.go:71`, `AppendChangelogEntry`; `relate_apply.go:40`, `ApplyRelatedFilesUpdate` | Reusable operations already shared with HTTP handlers. |
| Document persistence | `internal/documents/frontmatter.go:23,79,136` | Read metadata/body, serialize through a temporary file, split frontmatter. |
| Metadata schema | `pkg/models` and `pkg/frontmatter` | Typed document metadata and YAML preprocessing. |
| Workspace/index | `internal/workspace` | Root/config discovery, ticket queries and an in-memory SQLite index. |
| Paths | `internal/paths` | Resolve anchored paths such as `repo://`, `docs://` and `abs://`. |
| Tasks | `internal/tasksmd/tasksmd.go:183–214` | Parse/write task Markdown, preserve stable task IDs, toggle state. |
| HTTP integration | `internal/httpapi/tickets_changelog.go:90` and `related_files.go` | Expose the same domain operations through validated request handlers. |
| Human diagnostics | `pkg/commands/doctor.go:1563–1606` | Render doctor findings as text; separate from document serialization. |
| Tests/help | `internal/documents/frontmatter_test.go`, `pkg/commands/*_test.go`, `pkg/doc`, `test-scenarios/testing-doc-manager` | Unit, scenario and discoverable CLI documentation. |

Read `AGENT.md` before changing code. Build the CLI with `go build -tags sqlite_fts5 ./cmd/docmgr`; full UI embedding is unnecessary for persistence tests. Use the repository's pinned-binary E2E convention rather than allowing a scenario to find an ambiguous executable on PATH.

![](/home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/09/10/DOCMGR-FRICTION-001--reduce-documentation-workflow-friction-with-deterministic-writes-and-coherent-milestones/sources/print/figure-1.png)

This is the desired shared flow, not a claim that every current command follows it. In particular, `ticket_close.go` duplicates substantial mutation logic in `Run` and `RunIntoGlazeProcessor`. Both paths must be covered when fixing close behavior.

## 3. What happened in the motivating session

The source project is `/home/manuel/code/wesen/2026-09-07--streaming-system`. Its ticket is `WEBUI-LAB-001`, under `ttmp/2026/09/09/`.

### 3.1 Verified incidents

The companion ticket contains a frozen-session conversion, selected normalized tool records and a correction ledger. Relevant normalized emitting-turn indexes are:

- **1228:** Git reported `changelog.md:66: new blank line at EOF.`
- **1497:** Git reported `changelog.md:94: new blank line at EOF.`
- **1499:** staged validation reported `sources/textbook-final-doctor.log:6: new blank line at EOF.` and exited with code 2.
- **1440:** the agent reloaded the textbook, vault-writing, docmgr and diary skills together before Chapter 8 work.

The original final diary, Step 22, records the corrections. Commit `ed265a5b049f882a15f97ddbabdefa2b94a413c4` contains the final audit, a normalized doctor log and `textbook-final-doctor.raw.log.gz`. The compressed success log is evidence of an over-elaborate preservation response, not a feature docmgr should require.

The original final publication was verified at vault commit `67ef451b09cd124c1371c6964260fe80bbd67ae9`. Its tests, real media evidence, remote verification and visual review should remain. The proposed improvement removes repeated manual coordination around those checks.

### 3.2 Attribution boundaries

Do not turn this evidence into an invented “minutes wasted” measurement. Repeated loads and file mutations are observable; their causal share of elapsed time is not isolated. The session also contained substantial implementation, browser investigation, publication and user-directed documentation.

The diary skill required a full narrative template and broad file relations. The research workflow required tasks, relations, changelog, diary, doctor and upload receipts. The writing contract required final comprehensive review and preservation. Those instructions encouraged administrative multiplication, but neither docmgr nor the user required copying the same current-state fact into every possible file after every small change.

The companion guide distinguishes instructions that were wrong from safeguards that were merely applied too broadly. Docmgr should provide consistent primitives, not silently decide that the user's evidence requirements are unnecessary.

## 4. Root cause: separator ownership is undefined

`ReadDocumentWithFrontmatter(path)` returns `(*models.Document, string, error)`. `extractFrontmatter` splits the raw bytes on newline, finds the closing delimiter and joins every line after it into the body. The empty line after the delimiter is therefore part of the returned body.

`WriteDocumentWithFrontmatter(path, doc, body, force)` writes a fresh closing delimiter plus **two** newlines, then appends that body unchanged. The first newline terminates the delimiter; the second creates a separator. On the next cycle, the reader includes that separator in the body, and the writer adds another.

```text
first write:       ---\n\n# Body\n
reader body:             \n# Body\n
second write:      ---\n\n\n# Body\n
```

This is not a YAML decoder defect and should not be fixed by globally trimming Markdown. `strings.TrimSpace(body)` could remove authored whitespace, including trailing spaces used for Markdown line breaks. The contract must define which bytes belong to the framing and which belong to authored content.

### Decision: lossless body persistence before optional normalization

- **Context:** Metadata-only edits should not progressively alter the document body.
- **Options:** Trim all body whitespace; canonicalize leading blank lines automatically; preserve body bytes and make creation framing explicit.
- **Proposed decision:** The low-level serializer owns delimiter termination only. It preserves the supplied body bytes. A creation helper supplies the preferred initial separator once. Optional formatting is a separate explicit operation.
- **Rationale:** Lossless round trips are easier to test and safer than guessing which whitespace was authored.
- **Consequences:** Audit creation call sites and templates. Test bodies with no leading newline as well as existing files. Do not bulk-reformat historical documents as part of the fix.
- **Status:** proposed.

A proposed implementation sketch:

```go
func SerializeDocument(doc *models.Document, body []byte) ([]byte, error) {
    yamlBytes := encodeMetadata(doc)
    // Closing delimiter gets one terminator; body is otherwise untouched.
    return concat("---\n", yamlBytes, "---\n", body), nil
}

func WriteIfChanged(path string, expectedHash Digest, next []byte) (WriteResult, error) {
    old := read(path)
    if hash(old) != expectedHash { return conflict() }
    if bytes.Equal(old, next) { return unchanged(hash(old)) }
    return replaceUsingExistingTempFileStrategy(path, next)
}
```

These APIs are proposed, not present today. Preserve permissions deliberately: the existing writer uses `os.CreateTemp`, whose new-file mode needs review when replacing an existing file. Also test unknown YAML keys before assuming typed metadata decode/re-encode is lossless. Body stability and metadata fidelity are separate requirements.

## 5. Changelog and output hygiene are separate problems

`AppendChangelogEntry` is already a shared CLI/HTTP primitive. It creates `# Changelog\n\n`, then appends a string beginning with another newline. Entries without file notes end with `\n\n`; entries with notes add another blank separator after the list. The result parses, but generates avoidable diffs and EOF cleanup.

For a docmgr-owned changelog, define a canonical boundary: one blank line between sections and one final LF. Preserve the entry's internal Markdown. Do not deduplicate entries by text: two legitimate milestones can have identical descriptions. Retry idempotency needs an explicit operation identity.

Doctor output is different. `doctor.go` writes a blank line after each ticket group. That is human-oriented formatting; treating a captured success report as a source-code-quality fixture created part of the friction. Prefer structured output for retained machine evidence, or a documented text normalization rule. Do not force users to compress every successful diagnostic report merely to preserve decorative whitespace.

Existing tests validate YAML safety and changelog parse/append behavior. They do not currently establish all of the proposed byte-stability contracts. The focused existing tests pass while the reproducers still demonstrate the defects; passing them is not proof that the bugs are absent.

## 6. Close is not a multi-file transaction

At `ticket_close.go:157–199` and the corresponding human path around `292–328`, the command writes `index.md` before opening/appending `changelog.md`. If the second write fails, it returns an error but leaves the index changed.

The research script makes the changelog path a directory in a temporary workspace. This avoids permission assumptions. The result is exit 1 with “is a directory,” while the persisted status is `complete`. This reproduces partial application; it does not simulate a power failure.

The current document writer already uses a temporary file and rename **for one file**. Do not propose that as if it were missing, and do not equate it with an atomic transaction across several files.

### Decision: distinguish preflight, recoverability and atomic visibility

- **Context:** A milestone may update index, tasks, changelog and evidence metadata.
- **Options:** Sequential writes with honest partial results; one canonical record with derived views; a recoverable write journal; a database as the new source of truth.
- **Proposed decision:** First remove misleading atomicity language and unify close logic. For milestone operations, prefer one canonical operation record with recoverable projections and explicit consistency semantics.
- **Rationale:** Markdown remains the user-facing storage model; filesystem rename cannot make multiple paths simultaneously visible.
- **Consequences:** Cooperative readers need a ticket lock or committed-generation rule if they require consistent snapshots. External editors may still observe intermediate states. Crash recovery and concurrent edits need tests.
- **Status:** proposed.

A bounded operation flow:

```text
resolve ticket and acquire per-ticket cooperative lock
read current bytes and revisions
validate operation, referenced task IDs and required evidence
compute all intended outputs in memory
if operation ID already committed with same digest: return original receipt
if operation ID exists with different digest: reject
preflight paths; record prepared journal with old/new hashes
replace each changed projection; record progress
mark operation committed; refresh/invalidate workspace index
return receipt with changed paths and hashes
```

Preflight reduces predictable failures; it does not prove future writes cannot fail. Recovery must compare hashes before retrying, never overwrite an unrelated human edit, and report a conflict instead of guessing. The initial implementation can explicitly report partial application before a full recovery journal is introduced.

## 7. Proposed milestone and resume interfaces

### 7.1 One domain result, multiple renderers

The current command ecosystem uses Glazed command descriptions and decoded settings. Preserve that integration, but have CLI human/JSON paths and HTTP call the same operation service. `ApplyRelatedFilesUpdate` is a useful existing example: it returns `Changed=false` without writing when no effective relation change occurs. Do not characterize every current mutation as non-idempotent.

Proposed types:

```go
type MilestoneRequest struct {
    TicketID    string
    OperationID string
    Summary     string
    TaskIDs     []string
    Evidence    []EvidenceRef
    Expected    map[string]Digest
}
type EvidenceRef struct {
    Kind, Path, Revision, Claim string
}
type MutationReceipt struct {
    OperationID, Status string
    ChangedPaths []string
    Before, After map[string]Digest
    CompletedTasks []string
    Warnings []string
}
```

A proposed CLI is `docmgr milestone record --ticket ... --operation-id ... --evidence-file ...`. A future HTTP endpoint can expose the same request/result after the domain operation is tested. These are design sketches, not instructions to call nonexistent commands.

Keep the first schema small: one ticket, stable task IDs, summary, references and change receipt. The agent—not docmgr—judges whether evidence supports a technical claim. A passing test is not automatically sufficient proof for every requirement.

### 7.2 Resume is a derived view, not another diary

A proposed `docmgr ticket resume --ticket ID --output json` should return current status, latest committed milestone, remaining tasks, relevant document paths, evidence revisions and conflicts/staleness. It should reference historical narrative rather than rewriting it.

![](/home/manuel/code/wesen/go-go-golems/docmgr/ttmp/2026/09/10/DOCMGR-FRICTION-001--reduce-documentation-workflow-friction-with-deterministic-writes-and-coherent-milestones/sources/print/figure-2.png)

The skills ticket owns how much of that view an agent loads. Docmgr should expose enough information for targeted resumption without instructing agents to reread every historical document.

## 8. Implementation phases for a new intern

### Phase 0 — reproduce and learn

Read the source-map files above, `AGENT.md`, and `docmgr help how-to-add-cli-verbs`. Build an explicit local binary. Run the supplied scripts in the repository root:

```sh
T=$(find ttmp -type d -name 'DOCMGR-FRICTION-001--*' -print -quit)
go run -tags sqlite_fts5 "./$T/scripts/01-reproduce-writes.go"
go build -tags sqlite_fts5 -o /tmp/docmgr-friction-local ./cmd/docmgr
python3 "$T/scripts/02-reproduce-close.py" /tmp/docmgr-friction-local
```

Both scripts isolate mutations in temporary directories. They are research artifacts, not production fixes. Understand why existing tests pass before adding regressions.

### Phase 1 — byte-level contracts

Add table-driven tests to `internal/documents/frontmatter_test.go` for repeated read/write stability, body-byte preservation, creation framing, CRLF policy, empty bodies, permissions and no-op replacement. Add changelog fixtures with/without title and file notes, authored internal blank lines and deterministic clocks. Fix only the behavior specified by those tests.

Acceptance: repeating a metadata-neutral operation does not change bytes or mtime after any explicitly documented first normalization; substantive Markdown whitespace is unchanged; generated changelogs satisfy their declared EOF contract. Do not claim a global canonical serializer if only one command is fixed.

### Phase 2 — unify close and report failures honestly

Extract shared close planning/application into one callable service. Reuse changelog construction instead of duplicated string formatting. Keep human and structured output as renderers of the same result. Inject clock and filesystem failures so tests cover index-write failure, changelog-write failure, close retry and conflicting human edits.

Acceptance: help text describes the actual guarantee; errors identify which effects occurred; repeated operations with explicit idempotency identity do not add duplicate history. Preserve existing task-warning semantics unless a separate explicit strict-close option is accepted.

### Phase 3 — structured milestone primitive

Implement the smallest request/receipt contract, stable operation IDs, bounded references and a dry-run that computes changes without writing. Reuse task parsing and anchored path resolution. Add journaling/recovery only with a written consistency contract and fault-injection tests; do not label a sequence of renames atomic.

Acceptance: same ID/same request returns the same receipt; same ID/different request fails; no arbitrary path escape; partial application is recoverable or explicitly reported. A command failure cannot quietly masquerade as full success.

### Phase 4 — resume, documentation and HTTP parity

Generate a compact resume view from authoritative data. Add embedded help and examples. Exercise CLI human/JSON parity and the HTTP service boundary. Update the companion skill only after real CLI commands exist and are verified; until then, its proposal must remain labeled future API.

Run focused tests first, then the full Go suite with/without relevant build tags and the pinned-binary scenario suite. A persistence-only change does not require rebuilding the embedded frontend on every edit.

## 9. Test matrix and review gates

| Property | Test | Failure to prevent |
|---|---|---|
| Round-trip framing | Ten read/write cycles | Accumulating blank lines |
| Body fidelity | Code fences, hard-break spaces, blank lines | Destructive global trimming |
| No-op semantics | Compare bytes and mtime | Rewrites for unchanged relations/status |
| Explicit retry identity | Repeat operation and mismatch its payload | Duplicate history or accidental text deduplication |
| Failure boundaries | Inject each write failure | Complete status with unreported missing history |
| Concurrency | Modify a file after planning | Overwrite of a human edit |
| Projection recovery | Interrupt each journal stage | Permanent mixed milestone state |
| Entry-point parity | Same request via human/JSON/service | Divergent mutation logic |
| Diagnostics | Structured result and normalized text fixture | Decorative output becoming an audit burden |

Review source changes separately from generated fixture updates. Require an explanation for any changed Markdown byte that is not part of the requested mutation. Keep operation receipts bounded and avoid embedding entire transcripts or secrets.

## 10. Evidence, open decisions and handoff

The source/reproduction evidence is in `sources/write-reproduction.json`, `close-reproduction.json`, `existing-tests.log`, `installed-binary.txt` and `source-map.txt`. The session's precise normalized records and instruction conflict analysis live in the companion skills ticket; its session snapshot fingerprint makes the analysis cutoff explicit.

Open design decisions are intentional implementation choices, not missing research results: whether close gains an explicit strict mode; the scope of body normalization; how a journal coordinates external readers; and the minimum evidence schema. An intern should resolve these with maintainers before expanding the public API.

Success means fewer unnecessary changes and reconciliation steps **without** weaker technical validation. This ticket delivers a researched implementation guide and reproducers. The proposed product changes remain unimplemented and their tasks remain open.
