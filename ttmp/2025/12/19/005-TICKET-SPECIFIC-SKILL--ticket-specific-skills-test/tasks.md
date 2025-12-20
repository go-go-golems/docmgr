# Tasks

## Done

- [x] Create ticket-specific skill fixtures
- [x] Print ticket id + ticket title for ticket-scoped skills (list + show)
- [x] Ensure `skill show --ticket ...` works with the installed PATH binary
- [x] Filter out skills from non-active tickets by default unless `--ticket` is provided

- [x] Read 005 diary + 004 analysis so you know what changed and why
- [x] Run smoke: DOCMGR_PATH=/home/manuel/.local/bin/docmgr bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario (confirm skills smoke passes)
- [x] Update docs: document --ticket for skill show, ticket printing in list/show, and active-ticket-only default behavior
- [x] Add a guideline/template for DocType=skill so doc add --doc-type skill has good scaffolding (and avoids 'No guidelines found')
- [x] Decide/confirm convention: ticket skills live under /skill/ (doc-type folder) vs /skills/ (docs mention both); update docs accordingly
- [ ] Review 'active tickets only' semantics: should Status=review count as active? If yes, adjust filter + tests
- [ ] Perf cleanup (optional): avoid re-discovering/re-indexing when fetching ticket titles/status inside skill list/show
