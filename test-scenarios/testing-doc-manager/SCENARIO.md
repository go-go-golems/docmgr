# Scenario: MEN-4242 — Normalize chat API paths and WebSocket lifecycle

This scenario creates a mock codebase (`acme-chat-app`) and a ticket workspace under `ttmp/` to exercise all implemented `docmgr` features.

## Overview

- Mock repository: `acme-chat-app`
- Ticket: `MEN-4242` — "Normalize chat API paths and WebSocket lifecycle"
- Code: Go backend + TypeScript frontend fragments for search and relate heuristics
- Docs root: `ttmp/` (RFC-aligned)
- Repository ignores: `.docmgrignore` excludes `.git/`, `node_modules/`, `dist/`

## Steps (scripts)

1) 00-reset.sh — reset working directory
2) 01-create-mock-codebase.sh — bootstrap mock repository with chat-related files
3) 02-init-ticket.sh — initialize docs root, seed vocabulary, and create ticket workspace
4) 03-create-docs-and-meta.sh — add docs and enrich frontmatter metadata
5) 04-relate-and-doctor.sh — relate code files and run doctor
6) 05-search-scenarios.sh — exercise search (content, metadata, reverse lookup, external sources, dates, file suggestions)
7) 06-doctor-advanced.sh — induce warnings (unknown vocab, missing file, duplicate index), use ignore-glob, then fix and pass
8) 07-status.sh — summarize workspace and staleness

Optional manual steps:
- Use `docmgr changelog update` to append dated entries and include related files with notes
- Use `docmgr tasks` to manage the ticket checklist in `tasks.md`

## Expected Results

- `ttmp/YYYY/MM/DD/MEN-4242-normalize-chat-api-paths-and-websocket-lifecycle/` exists with RFC-aligned scaffolding
- `ttmp/vocabulary.yaml` contains topics/docTypes/intent used by this scenario
- `index.md` updated with Owners, Summary, ExternalSources, and RelatedFiles
- `doctor` reports OK unless staleness or multiple index.md introduced
- `search` returns expected rows and snippets for queries and filters
-. `changelog.md` shows appended dated entries where used
-. `tasks.md` can be listed/updated via CLI

## Commands Summary

### Vocabulary (seed)
- `docmgr vocab add --category topics --slug chat`
- `docmgr vocab add --category topics --slug backend`
- `docmgr vocab add --category topics --slug websocket`
- `docmgr vocab add --category docTypes --slug design-doc`
- `docmgr vocab add --category docTypes --slug reference`
- `docmgr vocab add --category docTypes --slug playbook`
- `docmgr vocab add --category docTypes --slug index`
- `docmgr vocab add --category intent --slug long-term`

### Init (docs root) and create ticket
- `docmgr init`
- `docmgr create-ticket --ticket MEN-4242 --title "Normalize chat API paths and WebSocket lifecycle" --topics chat,backend,websocket`

### Add documents
- `docmgr add --ticket MEN-4242 --doc-type design-doc --title "Path Normalization Strategy"`
- `docmgr add --ticket MEN-4242 --doc-type reference --title "Chat WebSocket Lifecycle"`
- `docmgr add --ticket MEN-4242 --doc-type playbook --title "Smoke Tests for Chat"`

### Guidelines
- `docmgr guidelines --doc-type design-doc --output markdown`

### Metadata updates
- `docmgr meta update --doc <index.md> --field Owners --value "manuel,alex"`
- `docmgr meta update --doc <index.md> --field Summary --value "Unify chat HTTP paths and stabilize WebSocket flows."`
- `docmgr meta update --doc <index.md> --field ExternalSources --value "https://example.com/rfc/chat-api,https://example.com/ws-lifecycle"`
- `docmgr meta update --doc <index.md> --field RelatedFiles --value "backend/chat/api/register.go,backend/chat/ws/manager.go,web/src/store/api/chatApi.ts"`

### List and doctor
- `docmgr list tickets --ticket MEN-4242`
- `docmgr list docs --ticket MEN-4242`
- `docmgr doctor --ignore-dir _templates --ignore-dir _guidelines --stale-after 30 --fail-on error`

### Search
### Relate
- Add explicit related files to the ticket index:
  - `docmgr relate --ticket MEN-4242 --files backend/chat/api/register.go,backend/chat/ws/manager.go,web/src/store/api/chatApi.ts`
- See suggested files with explanations (no changes applied):
  - `docmgr relate --ticket MEN-4242 --suggest --query WebSocket --topics chat`
  - Suggestions include a `source` and a `reason` (e.g., "recent commit activity", "working tree modified", "referenced by documents").

- Content: `docmgr search --query "WebSocket" --ticket MEN-4242`
- Metadata: `docmgr search --ticket MEN-4242 --topics websocket,backend --doc-type design-doc`
- Reverse lookup: `docmgr search --file backend/chat/api/register.go`
- Directory lookup: `docmgr search --dir web/src/store/api/`
- External sources: `docmgr search --external-source "https://example.com/ws-lifecycle"`
- Date filters: `docmgr search --updated-since "1 day ago" --ticket MEN-4242`
- Since last month: `docmgr search --since "last month" --ticket MEN-4242`
- File suggestions (heuristics): `docmgr search --ticket MEN-4242 --topics chat --files`

### Doctor (advanced)
### Changelog
- Minimal entry:
  - `docmgr changelog update --ticket MEN-4242 --entry "Normalize chat API paths"`
- With related files and notes:
  - `docmgr changelog update --ticket MEN-4242 \
     --files backend/chat/api/register.go,web/src/store/api/chatApi.ts \
     --file-note "backend/chat/api/register.go:Source of path normalization" \
     --file-note "web/src/store/api/chatApi.ts=Frontend integration"`
- Suggestions only / apply suggestions:
  - `docmgr changelog update --ticket MEN-4242 --suggest --query WebSocket`
  - `docmgr changelog update --ticket MEN-4242 --suggest --apply-suggestions --query WebSocket`

### Tasks
- List tasks: `docmgr tasks list --ticket MEN-4242`
- Add task: `docmgr tasks add --ticket MEN-4242 --text "Run smoke test suite"`
- Check/uncheck: `docmgr tasks check|uncheck --ticket MEN-4242 --id 1`
- Edit/remove: `docmgr tasks edit|remove --ticket MEN-4242 --id 1`
### Status
- Full status: `docmgr status`
- Summary only: `docmgr status --summary-only`
- Custom threshold: `docmgr status --stale-after 30`
- From nested dir (verifies `.ttmp.yaml` resolution):
  - `cd web && docmgr status --summary-only`

- Induce warnings: add unknown topic, add non-existing RelatedFiles, create `design/index.md`
- Show `--fail-on warning` returns nonzero
- Use `--ignore-glob "<ticket-dir>/design/index.md"` (for example, `ttmp/YYYY/MM/DD/MEN-4242-.../design/index.md`) to suppress duplicate index
- Fix metadata and re-run to pass

## Notes

- If `rg` is not available, file suggestions fall back to `grep`.
- To simulate staleness warnings, edit `LastUpdated` in `index.md` to an older date and re-run `doctor`.
