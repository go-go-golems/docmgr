---
Title: CLI Regression Checklist
Ticket: DOCMGR-REFACTOR
Status: draft
Topics:
    - docmgr
    - testing
DocType: playbooks
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Validated CLI regression steps
LastUpdated: 2025-11-18T21:07:41.674145654-05:00
---


# CLI Regression Checklist

Use this quick run-book after refactors that touch Cobra wiring, document helpers, or metadata commands.

1. **Help docs render**
   ```bash
   go run ./cmd/docmgr help how-to-use | head -n 20
   ```
   Confirms the embedded tutorial still loads.
2. **Ticket list works**
   ```bash
   go run ./cmd/docmgr ticket tickets --root ttmp | head -n 10
   ```
   Expect a Markdown table; verify no cobra flag errors.
3. **Doc listing filters by ticket**
   ```bash
   go run ./cmd/docmgr doc docs --ticket DOCMGR-REFACTOR --root ttmp
   ```
   Should show the latest docs you created via `doc add`.
4. **Task list + mutations**
   ```bash
   go run ./cmd/docmgr tasks list --ticket DOCMGR-REFACTOR --root ttmp
   go run ./cmd/docmgr tasks check --ticket DOCMGR-REFACTOR --id 4 --root ttmp
   ```
   Ensure both listing and checkbox updates work.
5. **Metadata lifecycle**
   ```bash
   go run ./cmd/docmgr doc add --ticket DOCMGR-REFACTOR --doc-type design-doc --title "Smoke Test Doc" --root ttmp
   go run ./cmd/docmgr meta update --doc ttmp/.../design-doc/NN-smoke-test-doc.md --field Status --value review
   go run ./cmd/docmgr doc relate --doc ttmp/.../design-doc/NN-smoke-test-doc.md --file-note "/abs/path:file reason"
   go run ./cmd/docmgr changelog update --ticket DOCMGR-REFACTOR --entry "Smoke test doc created"
   ```
   Validates frontmatter read/write + changelog integration.
6. **`go test` sanity**
   ```bash
   go test ./...
   ```
   Keeps the package build green.
