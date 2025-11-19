# Tasks

## TODO

- [x] Add 'ticket close' command under ticket/ namespace (skeleton)
- [x] Implement atomic close: update Status=complete (override via --status), optional Intent, changelog entry, LastUpdated
- [x] Add structured output to 'ticket close' (--with-glaze-output --output json) with operations + state
- [x] Extend vocabulary.yaml with status values; update doctor to warn on unknown Status
- [ ] Document suggested status transitions (not enforced) in help/docs
- [x] On all tasks done, print actionable suggestion in 'tasks check' to run 'ticket close' (no auto exec)
- [x] (Optional) Add '--with-glaze-output' to 'tasks check' exposing all_tasks_done and task counts
- [ ] Update docs: how-to-use + cli-guide to include 'ticket close' + structured output guidance
- [x] After implementation, relate modified files with notes; add changelog entries
