# Test Notes — docmgr Scenario

## 2025-11-03

### Change: init semantics
- `init` now initializes the docs root (creates `ttmp/`, `vocabulary.yaml`, `_templates/`, `_guidelines/`).
- Use `create-ticket` to scaffold a ticket workspace: `docmgr create-ticket --ticket MEN-4242 --title ...`.

### Issue: search positional query — Too many arguments
- Command: `docmgr search "WebSocket" --ticket MEN-4242`
- Observed: `Error: Too many arguments`
- Cause: `search` expects `--query` for content search.
- Fix: Use `docmgr search --query "WebSocket" --ticket MEN-4242` (scripts and docs updated).

### Observation: doctor scans _templates/ and _guidelines
- Doctor output shows warnings/errors for `ttmp/_guidelines/index.md` (missing fields) and `ttmp/_templates/index.md` (invalid frontmatter parse).
- Fix: Updated doctor to ignore root directories starting with `_` (e.g., `_templates`, `_guidelines`).
- Verified: `docmgr doctor --root ttmp` now reports only the ticket workspace.

### Added: doctor vocabulary and related file checks
- Doctor now warns on unknown `Topics`, `DocType`, or `Intent` (validated against `ttmp/vocabulary.yaml` or the path configured via `.ttmp.yaml:vocabulary`).
- Doctor warns when `RelatedFiles` paths do not exist (resolved relative to repo root).
- Scenario uses valid vocab and existing files; doctor remains OK.

### Search validations (after fixes)
- `--query WebSocket` returns `index` and `reference` docs with snippets (headings shown).
- Metadata filter (`--topics websocket,backend --doc-type design-doc`) returns only the design doc as expected.
- Reverse lookup (`--file backend/chat/api/register.go`) returns ticket docs (works).
- Directory lookup (`--dir web/src/store/api/`) returns ticket docs (works).
- External source (`--external-source https://example.com/ws-lifecycle`) returns the ticket `index` (works).
- Date filters (`--updated-since "1 day ago"`, `--since "last month"`) include current docs as expected.
- File suggestions (`--files`) listed from `RelatedFiles` only; no git/rg-derived suggestions surfaced in this tiny repo (acceptable for now).

### Open follow-ups
- Expose doctor ignore patterns via flags (done) and document examples (partially in scenario).
- Validate `Topics`, `DocType`, `Intent` (done); consider `--fail-on warning` in CI if desired.
- Add help consistency audit and troubleshooting guide.
