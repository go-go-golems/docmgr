---
Title: 'CLI verbs mapping: old vs new'
Ticket: DOCMGR-DOC-VERBS
Status: active
Topics:
    - docmgr
    - documentation
    - cli
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/doc/docmgr-how-to-use.md
      Note: Source tutorial to update (verbs outdated)
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-19T13:58:29.134137645-05:00
---


---
# CLI verbs mapping: old vs new

## Purpose
Align the “how-to-use” tutorial with the new grouped verb structure. This is the authoritative mapping to use when updating examples.

## Mapping (Old → New)

| Area               | Old                         | New                          | Notes |
|--------------------|-----------------------------|------------------------------|-------|
| Add document       | `docmgr add`                | `docmgr doc add`             | Doc operations live under `doc` |
| Search             | `docmgr search`             | `docmgr doc search`          | Full-text/metadata search moved under `doc` |
| Guidelines         | `docmgr guidelines`         | `docmgr doc guidelines`      | Guidelines grouped with docs |
| Relate files       | `docmgr relate`             | `docmgr doc relate`          | File relations are doc-scoped |
| Create ticket      | `docmgr create-ticket`      | `docmgr ticket create-ticket`| Ticket operations under `ticket` |
| List docs          | `docmgr list docs`          | `docmgr list docs`           | Preferred (also available: `docmgr doc docs`) |
| List tickets       | `docmgr list tickets`       | `docmgr list tickets`        | Preferred (also available: `docmgr ticket tickets`) |
| Status             | `docmgr status`             | `docmgr status`              | Unchanged |
| Tasks              | `docmgr tasks …`            | `docmgr tasks …`             | Unchanged |
| Metadata           | `docmgr meta update …`      | `docmgr meta update …`       | Unchanged |
| Changelog          | `docmgr changelog update …` | `docmgr changelog update …`  | Unchanged |
| Doctor             | `docmgr doctor …`           | `docmgr doctor …`            | Unchanged |
| Init               | `docmgr init …`             | `docmgr init …`              | Unchanged |
| Configure/Config   | `docmgr configure|config`   | `docmgr configure|config`    | Unchanged |

## Updated examples to use in the tutorial

### Create a ticket
```bash
docmgr ticket create-ticket --ticket MEN-4242 \
  --title "Normalize chat API paths and WebSocket lifecycle" \
  --topics chat,backend,websocket
```

### Add documents
```bash
docmgr doc add --ticket MEN-4242 --doc-type design-doc --title "Path Normalization Strategy"
docmgr doc add --ticket MEN-4242 --doc-type reference --title "Chat WebSocket Lifecycle"
docmgr doc add --ticket MEN-4242 --doc-type playbook  --title "Smoke Tests for Chat"
```

### Search
```bash
# Full-text search
docmgr doc search --query "WebSocket"

# Filter by metadata
docmgr doc search --query "API" --topics backend --doc-type design-doc

# Reverse lookup by file
docmgr doc search --file backend/api/register.go
```

### Relate files to a document or ticket
```bash
# Relate to ticket index (repeat --file-note)
docmgr doc relate --ticket MEN-4242 \
  --file-note "backend/api/register.go:Registers API routes (normalization logic)" \
  --file-note "backend/ws/manager.go:WebSocket lifecycle management"
```

### Guidelines
```bash
docmgr doc guidelines --doc-type design-doc
```

## Next steps (TODO)
- [ ] Update `docmgr/pkg/doc/docmgr-how-to-use.md` to reflect the new verbs:
  - Section 3 (Create Your First Ticket): use `docmgr ticket create-ticket`
  - Section 4 (Add Documents): use `docmgr doc add`
  - Section 5 (Search): use `docmgr doc search`
  - Section 7 (Relating Files): use `docmgr doc relate`
  - Section 12 (Guidelines mention): use `docmgr doc guidelines`
- [ ] Open PR, link this reference, and announce in release notes
