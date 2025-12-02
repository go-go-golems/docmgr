# Tasks

## TODO

- [x] Add tasks here

- [x] Design ticket-move verb: inputs (ticket/id or path), new path templating (date-based), validation (source/dest), and safety checks.
- [x] Implement pkg/commands/ticket_move.go + cobra wiring; move ticket directory to new path template; update Ticket paths in .ttmp config if needed.
- [x] Move/rename ticket directory; rewrite internal paths (related files? relative links?) where necessary; ensure index/timestamps preserved.
- [x] Add scenario test `test-scenarios/testing-doc-manager/17-ticket-move.sh` (runs with DOCMGR_PATH=/tmp/docmgr-bin); follow up with unit coverage.
- [ ] Add unit tests for path computation/move to complement the scenario.
- [x] Update help/docs to include ticket move verb and any caveats (overwrites, ignores, config).
