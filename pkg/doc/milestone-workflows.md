---
Title: Recoverable milestones and ticket resumption
Slug: milestone-workflows
Short: Record stable milestone identities, recover interrupted projections, and resume from current evidence.
Topics: [docmgr, workflow, tickets]
Commands: [milestone, ticket]
Flags: [operation-id, dry-run, evidence-file, expected-file]
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

## Record one milestone

A milestone stores one canonical request and receipt and updates the requested task checkboxes and changelog. It does not write a diary, infer completion from a test result, change ticket status, or copy narrative into several documents. Use a stable operation ID for one logical action; retain the same arguments when retrying it.

```bash
docmgr milestone record --ticket EXAMPLE --operation-id tests-reviewed \
  --summary 'Reviewed persistence tests' --phase validate \
  --task-id ab12 --next 'Review HTTP parity' --dry-run
# Remove --dry-run to apply the same request.
```

Phases are `start`, `implement`, `validate`, `checkpoint`, `resume`, and `close`. They are advisory labels, not automatic scheduling or proof of goal completion. Task references must be stable IDs, not their mutable list positions. A matching ID/request returns the original receipt without adding history; an existing ID with a different request fails.

`--with-glaze-output --output json` selects structured output in docmgr's pinned Glazed v1.3.6. The milestone row contains `receipt`; resume rows contain `resume`. Do not substitute flags from a newer Glazed release without checking this binary's help.

## Attach evidence and expected revisions

Evidence is a bounded JSON array supplied with `--evidence-file`. Each record contains kind, path, SHA256 revision and claim. For example, compute the revision with `sha256sum` and use its actual 64-character digest:

```json
[
  {"kind":"test", "path":"doc://sources/test.log", "revision":"<SHA256 of test.log>", "claim":"The retained regression run passed"}
]
```

Paths must resolve inside the ticket. Bare paths are ticket-relative; anchored paths use the shared resolver but are still subject to ticket containment. Symlink escapes, missing files and revision mismatches fail. Store a ticket-local report referencing an external repository/commit rather than authorizing arbitrary file reads through an HTTP request. Revision means file bytes, not a Git commit. The claim remains the author's responsibility.

`--expected-file` optionally accepts a JSON map from `index.md`, `tasks.md`, or `changelog.md` to their expected SHA256, or `missing`. Obtain current revisions from structured resume output. These are optimistic guards, not an exclusive lock on external editors. A committed retry returns its historical receipt even if evidence has subsequently changed; resume reports that staleness separately.

## Consistency and recovery

Build with Go 1.26.6 or newer, as required by the module, for patched filesystem containment. Close and milestone operations lock the existing ticket directory, read/validate the plan, save a prepared journal under `.docmgr-operations/`, replace each changed projection, then mark the record committed. The lock is cooperative and available on Linux, macOS and FreeBSD. Other builds return a capability error for these operations.

This provides process-crash recovery, not atomic multi-file visibility or a power-loss durability guarantee. Legacy mutation commands and external editors do not acquire this lock. They may observe intermediate files; a narrow check/rename race with an uncooperative writer remains outside the contract. Do not run competing mutation tools against a ticket during application or recovery.

After interruption, retry the **same ID and request**. The saved plan, including its original dates and output bytes, is reused. Each projection must match its original before hash or intended after hash. Any other bytes produce a conflict, never an automatic overwrite. A pending operation blocks unrelated new operations. Inspect its JSON record and reconcile conflicting edits deliberately before retrying; there is no force-overwrite recovery flag.

Close uses the same service in human and structured modes. Structured output now carries a receipt instead of the old duplicated update-boolean map; this is an intentional output-contract change. History is written before status, so a predictable changelog failure cannot mark the ticket complete. An already matching status/intent is normally a no-op, even with a different message: use changelog update to append a standalone note. Supply `--operation-id` for explicit close retry identity. An interrupted default close reports its generated ID; retry using that ID and the original options. Open tasks still warn rather than preventing closure.

Limits protect local storage and requests: 4 MiB per file/journal, 64 KiB per request, 1024 journal records, 64 MiB total scanned history, 64 evidence refs, 128 task IDs, and 512 resume document pointers. Archive old committed history deliberately before reaching these bounds; never discard a pending record to evade a conflict. Retry identity applies to retained records: do not reuse archived operation IDs. New journal directories are owner-only; protect repository access too when journals are committed to Git. Journals contain projection bytes, so avoid secrets in managed metadata and notes.

## Resume without another status document

```bash
docmgr ticket resume --ticket EXAMPLE
docmgr ticket resume --ticket EXAMPLE --with-glaze-output --output json
```

Resume reads authoritative files under the same cooperative lock and does not write. It returns status, the latest committed milestone, phase/next action, remaining tasks, projection revisions, document pointers, evidence and conflict warnings. A newer standalone edit may legitimately trigger a staleness warning; investigate rather than treating the checkpoint as newer than the files or the current user request. Narrative history stays in the diary.

## HTTP parity

`POST /api/v1/tickets/milestone` accepts `ticket`, `operation_id`, `summary`, `phase`, `next`, `task_ids`, `evidence`, and `expected`. Add `?dry_run=true` to preview. JSON is strict and size-bounded. `GET /api/v1/tickets/resume?ticket=EXAMPLE` returns the same resume structure as the CLI. Both call the shared services.

Malformed JSON returns 400; unresolved tickets retain existing 404/ambiguity errors; service conflicts return 409 with a receipt and observed applied paths where available. A successful mutation followed by index-refresh failure returns the committed receipt with a warning, not a misleading mutation failure. The client can explicitly refresh the index afterward.

## Persistence contracts

Metadata serialization owns delimiter termination only. Creation helpers supply initial spacing; subsequent edits preserve body bytes, including hard-break spaces, leading blank lines and CRLF. YAML framing is LF and typed metadata formatting is canonical, not a promise to retain comments or scalar styles. Unknown metadata values are retained; aliases are expanded with depth 64 / 10,000-node limits so known-field re-encoding cannot leave dangling anchors. Valid YAML is parsed before scalar-repair heuristics, preserving real aliases and inline-comment semantics. Identical canonical bytes preserve mtime; replacing a regular file preserves permissions. Symlink/non-regular write targets are rejected.

Generated changelogs have one blank line between sections and one final LF. Internal entry Markdown is not globally trimmed. Text-equal entries are not deduplicated; retry identity belongs to milestone operations.

## Troubleshooting

| Problem | Cause | Action |
|---|---|---|
| ID reused for different request | Arguments changed on retry | Recover with original arguments, or choose a new ID for a genuinely new operation after recovery |
| Pending operation | Interrupted projection or commit marker | Inspect record, retry same ID/request |
| Revision conflict | Files changed since planning | Reconcile intentionally; do not force overwrite |
| Evidence mismatch | Wrong SHA256, missing or changed file | Review the file and record its actual digest |
| Unknown task ID | Positional ID or stale marker | Inspect current stable IDs with task list |
| Stale resume warning | Current files differ from checkpoint | Investigate and record a new reviewed milestone if warranted |

## See Also

- `docmgr help how-to-use`
- `docmgr ticket close --help`
- `docmgr milestone record --help`
- `docmgr ticket resume --help`
