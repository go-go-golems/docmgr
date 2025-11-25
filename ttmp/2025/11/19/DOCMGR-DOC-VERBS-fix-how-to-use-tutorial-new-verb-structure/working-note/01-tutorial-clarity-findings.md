---
Title: Tutorial clarity findings
Ticket: DOCMGR-DOC-VERBS
Status: active
Topics:
    - docmgr
    - documentation
    - cli
DocType: working-note
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-24T16:25:38.208706183-05:00
---

# Tutorial clarity findings

## Summary

Log every confusing tutorial moment here so we can trace which section or workflow step needs edits.

## Notes

- Use `YYYY-MM-DD – Step <n> – <short description>` format when adding new entries.
- Reference the exact command or paragraph that caused confusion.

## Decisions

- All tester feedback for DOCMGR-DOC-VERBS lives in this working note (not in the sample ticket).

## Next Steps
- Review this log after each test run and convert recurring issues into concrete documentation changes or ticket tasks.

## Findings

- 2025-11-24 – Step 3 – Re-running `docmgr doc relate` with identical notes prints `Error: no changes specified`. This actually means “no diff to apply.” Consider clarifying the message in the tutorial so beginners aren’t alarmed by the word “Error.”
- 2025-11-24 – Step 3 – The reset script already seeds the related files during its run, so running `docmgr doc relate` against the ticket index immediately afterward repeats the exact same notes and hits the “no changes specified” error. Suggest calling this out (“script already applied Step 3; rerun with new notes or use `--doc` to relate to a subdocument”) so newcomers know their CLI request had no effect even though the output looks like an error.
- 2025-11-24 – Step 3 – Root vs CWD: commands run from the wrong directory will target the wrong docs root. Suggest explicitly passing `--root /tmp/test-git-repo/ttmp` in examples or reminding readers to run from `/tmp/test-git-repo`.
- 2025-11-24 – Help inconsistency: `docmgr doc relate --help` shows examples using `docmgr relate`, but the binary rejects `docmgr relate`. Update examples to consistently use `docmgr doc relate`.
- 2025-11-24 – Numeric prefixes: Creating a second design doc increments to `02-...`. Call this out earlier so users aren’t surprised when filenames change across runs.
