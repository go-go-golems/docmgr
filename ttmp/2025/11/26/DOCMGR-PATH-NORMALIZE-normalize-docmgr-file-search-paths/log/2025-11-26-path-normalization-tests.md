# Path Normalization Test Runs — 2025-11-26

## Commands Executed

1. `docmgr/test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-path-normalize`
   - Purpose: ensure the entire regression suite (init → relate/search → doctor → template schema → new path-normalization playbook) still completes end-to-end.
   - Result: ✅ success after updating scripts to the current CLI (`docmgr ticket create-ticket`, `docmgr doc add`, `docmgr doc relate`, `docmgr doc search`). Outputs confirm
     - baseline tickets/docs created under `/tmp/docmgr-path-normalize/acme-chat-app/ttmp`
     - doctor/status/vocab commands clean except for the intentionally introduced warnings in the “advanced doctor” scenario.
     - Added `reference/99-wonky-paths-fixture.md` containing doc-relative, ttmp-relative, and absolute `RelatedFiles` entries. `05-search-scenarios.sh` now issues searches using each variant, plus suffix-only (`register.go`), all of which succeed.

2. `docmgr/test-scenarios/testing-doc-manager/14-path-normalization.sh /tmp/docmgr-path-normalize`
   - Purpose: targeted verification that `docmgr doc relate` canonicalizes doc-relative, ttmp-relative, and absolute paths, and that `docmgr doc search --file` finds the
     same doc using any of those inputs (plus bare suffix).
   - Result: ✅ relates were already canonicalized (reported as no-op) and each search variant produced the expected doc listing.

## Notes

- All fixture runs use a clean `/tmp/docmgr-path-normalize` sandbox, so reruns do not pollute the repo.
- Scenario scripts now match the instructions documented in `docmgr-help-how-to-use` (namespaced `doc` / `ticket` verbs).
- Search output snippets show the canonicalized representations (`../../../../../backend/chat/api/register.go`, `backend/chat/api/register.go`, absolute path) validating the resolver’s suffix/prefix logic.

