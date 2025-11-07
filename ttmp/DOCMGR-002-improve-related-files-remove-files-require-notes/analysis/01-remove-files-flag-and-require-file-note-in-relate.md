---
Title: Remove --files flag and require file-note in relate
Ticket: DOCMGR-002
Status: active
Topics:
    - docmgr
    - ux
    - cli
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: docmgr/pkg/commands/add.go
      Note: Seeds RelatedFiles via --related-files; add note-aware behavior or deprecate
    - Path: docmgr/pkg/commands/changelog.go
      Note: Supports --files and --file-note; consider consistency later
    - Path: docmgr/pkg/commands/relate.go
      Note: Implements relate CLI; remove --files
    - Path: docmgr/pkg/commands/search.go
      Note: Uses --files boolean for suggestions; unaffected by this change
    - Path: docmgr/pkg/doc/docmgr-cli-guide.md
      Note: CLI guide examples use --files; update to new syntax
    - Path: docmgr/pkg/doc/docmgr-how-to-use.md
      Note: Tutorial examples use --files; update to new syntax
    - Path: docmgr/pkg/models/document.go
      Note: RelatedFiles structure; note optional today; doctor considerations
    - Path: docmgr/test-scenarios/testing-doc-manager/SCENARIO.md
      Note: Scenario references --files; update accordingly
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-07T13:49:25.372914869-05:00
---



# Remove --files flag and require file-note in relate

## Purpose and scope

We will change the `docmgr relate` workflow to remove the `--files` flag and require users to always supply notes via repeated `--file-note "path:note"` arguments. This enforces better rationale capture at the moment of relating files.

Scope of this ticket:
- Update the `relate` command to accept additions only through `--file-note` mappings
- Disallow legacy `--files` usage with a helpful error
- Keep removals via `--remove-files` and suggestion flow intact; when applying suggestions, auto-generate a note from reasons as today
- Update documentation, examples, and test scenarios

Out of scope (follow-ups suggested):
- Enforcing notes globally in existing documents (doctor validation severity)
- Aligning `changelog update` to require notes (today it accepts `--files` + optional notes)

## Current behavior (summary)

- CLI:
  - `docmgr relate` supports `--files`, `--remove-files`, and repeated `--file-note "path:note"`.
  - `docmgr changelog update` supports `--files` and repeated `--file-note` to include a Related Files list in entries.
  - `docmgr add` supports `--related-files` to seed RelatedFiles on new docs; it does not accept notes.
- Implementation:
  - relate: additions are applied from `settings.Files`, with optional note lookup in a parsed `noteMap` from `settings.FileNotes`.
  - changelog: builds a file→note map from `--files` + `--file-note`; renders a section in the entry.
  - add: converts `--related-files` to `RelatedFiles` with empty notes.
- Suggestions: with `--suggest --apply-suggestions` (relate/changelog), a note is synthesized from reasons when no explicit note exists.
- Data model: `RelatedFiles` entries have `Path` and optional `Note` (omitempty) in YAML frontmatter.

Key implementation areas:
- Flag definitions and help text in `docmgr/pkg/commands/relate.go`.
- Addition/removal logic in `RunIntoGlazeProcessor` of `relate.go`.
- Frontmatter read/write helpers in `pkg/commands` (`readDocumentWithContent`, `writeDocumentWithFrontmatter`).
- Documentation and examples in `pkg/doc` and `test-scenarios`.

## Proposed behavior (by verb)

- relate
  - Remove support for `--files`.
  - Additions and updates must be expressed via repeated `--file-note "path:note"`.
  - Error on empty/missing note values; keep `--remove-files` unchanged.
  - Keep `--suggest --apply-suggestions`; ensure a note is always present (explicit or synthesized).
  - Fail fast on legacy `--files` with a clear migration message.

- changelog update
  - Deprecate and remove `--files` for consistency; require repeated `--file-note` for related files in entries.
  - When using `--suggest --apply-suggestions`, synthesize notes as today.
  - Provide a migration error with concrete examples if `--files` is provided.

- add
  - Replace `--related-files` with repeated `--file-note` to seed `RelatedFiles` with notes at creation time; error if notes are missing.
  - Alternative (softer) path: keep `--related-files` but warn and require `--file-note` when related files are specified; however, preferred is to remove `--related-files` for clarity.

## Changes by area

### CLI and command logic

File: `docmgr/pkg/commands/relate.go`

- Remove the `files` flag and `Files []string` from `RelateSettings`.
- Update long help/examples to show only `--file-note` for additions.
- Addition logic: iterate over the parsed `noteMap` keys to add or update entries; do not read from `settings.Files`.
- Validation: if no additions were specified (no `--file-note`), and neither `--remove-files` nor `--apply-suggestions` are used, return an error like: "No changes specified. Use --file-note path:note to add/update, --remove-files to remove, or --suggest --apply-suggestions."
- Validation: for each `--file-note`, require a non-empty note value; error if missing.
- Legacy guard: if the (now removed) `--files` parameter is encountered by older shells or integrations, return a targeted error: "--files has been removed. Use repeated --file-note 'path:note' instead. Example: docmgr relate --file-note 'a/b.go:reason' --file-note 'c/d.ts:reason'".
File: `docmgr/pkg/commands/changelog.go`

- Remove `files` flag and `Files []string` from settings; require `--file-note` for any file references in entries.
- Keep suggestions behavior; ensure notes are present for applied suggestions.
- Update help/examples accordingly; add migration error message when `--files` is used.

File: `docmgr/pkg/commands/add.go`

- Add support for repeated `--file-note` to seed `RelatedFiles` with notes on new docs.
- Remove `--related-files` (or accept but error with a migration message pointing to `--file-note`).
- Update help/examples.

- Suggestions path: unchanged; when applying, continue generating notes from reasons where none provided explicitly.

Notes on code touchpoints:
- Flag removal/update in `NewRelateCommand()`.
- Addition/removal application block currently around the additions loop (see ~`for _, af := range settings.Files`): replace with iteration over `noteMap`.
- Maintain stable ordering and existing frontmatter preservation behavior.

### Data model

File: `docmgr/pkg/models/document.go`

- No structural change required. `Note` remains optional in the schema to maintain backward compatibility for existing documents. Enforcement occurs at the CLI layer and (optionally) doctor.

### Validation (doctor)

File: `docmgr/pkg/commands/doctor.go`

- Add a check that emits a warning for `RelatedFiles` entries with an empty `Note`. Consider a `--enforce-notes` flag (or configuration) to upgrade this to an error in CI, but for now keep a warning to avoid breaking existing workspaces.

### Documentation updates

Update all examples to remove `--files` usage and demonstrate repeated `--file-note` entries:

- `docmgr/pkg/doc/docmgr-how-to-use.md`
  - Sections showing `docmgr relate --files` must change to only `--file-note` entries.
  - Keep suggestion examples; clarify that applying suggestions always produces a note.

- `docmgr/pkg/doc/docmgr-cli-guide.md`
  - Replace relate examples to use repeated `--file-note`.
  - Update any automation snippets that relied on `--files` to instead produce `--file-note "$f:$NOTE"`.

- Add command sections
  - Replace `--related-files` with repeated `--file-note` in examples and templates.

- `docmgr/test-scenarios/testing-doc-manager/SCENARIO.md`
  - Update the Relate and Changelog examples to match the new syntax for relate.
  - Update Add examples to seed `RelatedFiles` using `--file-note`.

### Other commands (review only)

- `docmgr search` uses a boolean `--files` to trigger suggestion output. This is unrelated to list-style `--files` and remains unchanged. Disambiguate in docs.

## Migration and compatibility

- relate: legacy `--files` errors with a migration hint to repeated `--file-note`.
- changelog: legacy `--files` errors with a migration hint to repeated `--file-note`.
- add: legacy `--related-files` errors with a migration hint to repeated `--file-note` (or warn-only if a softer rollout is preferred).
- Existing documents without notes remain valid. Doctor will surface warnings; teams can resolve incrementally.

## Test and scenario updates

- Update `test-scenarios/testing-doc-manager/SCENARIO.md` relate examples.
- Update changelog examples to use only `--file-note`.
- Update add examples to seed with `--file-note`.
- Grep and replace `docmgr relate --files`, `docmgr changelog update --files`, and `docmgr add --related-files` across repository docs.
- Add small scenarios: legacy flags → expect clear migration errors; note-only flows → success; suggestions → success with synthesized notes.

## Implementation checklist

- [ ] Relate: remove `--files`; require `--file-note`; validate non-empty notes
- [ ] Relate: fail fast on legacy `--files`; keep suggestions and removals
- [ ] Changelog: remove `--files`; require `--file-note`; suggestions synthesize notes; legacy error
- [ ] Add: support repeated `--file-note`; remove `--related-files` (or error with migration)
- [ ] Doctor: warn for blank `Note` entries in `RelatedFiles`
- [ ] Docs: update tutorial/CLI guide/add & changelog examples
- [ ] Scenarios and CI snippets updated
- [ ] Changelog entry documenting breaking changes

## Key files impacted

- `docmgr/pkg/commands/relate.go` — CLI flags, addition/removal logic, help text
- `docmgr/pkg/commands/changelog.go` — CLI flags, file-note enforcement, suggestions
- `docmgr/pkg/commands/add.go` — add note-aware seeding, remove `--related-files`
- `docmgr/pkg/commands/doctor.go` — new warning for missing notes (optional in this ticket)
- `docmgr/pkg/doc/docmgr-how-to-use.md` — tutorial updates
- `docmgr/pkg/doc/docmgr-cli-guide.md` — reference/automation updates
- `docmgr/test-scenarios/testing-doc-manager/SCENARIO.md` — scenario commands
- (follow-up) `docmgr/pkg/commands/changelog.go` — decide on consistency

## Example new usage

Add two files with notes to the ticket index:

```bash
docmgr relate --ticket DOCMGR-002 \
  --file-note "docmgr/pkg/commands/relate.go:Implements relate CLI changes" \
  --file-note "docmgr/pkg/doc/docmgr-how-to-use.md:Update examples to note-only"
```

Remove a file:

```bash
docmgr relate --ticket DOCMGR-002 --remove-files docmgr/pkg/commands/relate.go
```

Apply suggestions (notes auto-generated from reasons; can be overridden with explicit `--file-note`):

```bash
docmgr relate --ticket DOCMGR-002 --suggest --apply-suggestions --query "relate command"
```
