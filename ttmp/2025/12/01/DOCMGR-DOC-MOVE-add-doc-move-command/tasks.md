# Tasks

## TODO

- [x] Add tasks here

- [x] Design docmgr doc move command: UX (flags), validation (source/dest tickets), and file moves.
- [x] Implement pkg/commands/doc_move.go + cobra wiring under cmd/docmgr/cmds/doc to move a markdown doc between tickets.
- [x] Handle frontmatter rewrite (Ticket field) and optional renumber/prefix updates when moving; preserve body/content.
- [x] Add tests: add smoke scenario `test-scenarios/testing-doc-manager/16-doc-move.sh` (run) and keep a follow-up for unit coverage.
- [ ] Add unit tests for path resolution/frontmatter rewrite to complement the smoke scenario.
- [x] Update help/docs (how-to-use, cli-guide) with doc move usage examples and safeguards.
