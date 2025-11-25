# Changelog

## 2025-11-25

- Initial workspace created


## 2025-11-25

Created comprehensive analysis document of carapace's dynamic flag completion implementation. Analyzed storage system, flag registration, action system, context handling, and traversal logic. Documented key components with code references and usage examples.


## 2025-11-25

Wire carapace into root; add completions for 'doc add' (ticket/doc-type/topics/status/intent/related-files); add action helpers (tickets, vocab, files/dirs)

### Related Files

- docmgr/cmd/docmgr/cmds/doc/add.go — Register FlagCompletion for doc add
- docmgr/cmd/docmgr/cmds/root.go — Attach carapace (dynamic completion bridge)
- docmgr/pkg/completion/actions.go — Action providers for tickets and vocabulary
- docmgr/pkg/completion/carapace.go — Attach()


## 2025-11-25

tmux smoke test: dynamic completion works in bash and zsh for doc add flags (doc-type/status/intent/topics/ticket); will expand to fish and remaining verbs next

### Related Files

- docmgr/cmd/docmgr/cmds/doc/add.go — Registered FlagCompletion for doc add
- docmgr/pkg/completion/actions.go — Action providers used by completion
- docmgr/ttmp/2025/11/25/DOCMGR-DYNAMIC-COMPLETION-add-dynamic-flag-completion-to-docmgr-using-carapace/analysis/01-dev-diary-dynamic-carapace-completion.md — Diary entry with results


## 2025-11-25

Added tmux helper script to send actual TAB and capture pane outputs for bash/zsh; outputs saved under /tmp/dctest_*.txt; attach to diary.

### Related Files

- docmgr/ttmp/2025/11/25/DOCMGR-DYNAMIC-COMPLETION-add-dynamic-flag-completion-to-docmgr-using-carapace/analysis/01-dev-diary-dynamic-carapace-completion.md — Diary updated with tmux test notes
- docmgr/ttmp/2025/11/25/DOCMGR-DYNAMIC-COMPLETION-add-dynamic-flag-completion-to-docmgr-using-carapace/scripts/01-tmux-completion-test.sh — Tmux completion test helper


## 2025-11-25

Verified dynamic completion updates: added docType 'til' and topic 'ai' now appear; new ticket 'DEMO-100' appears under --ticket completion.

### Related Files

- docmgr/ttmp/2025/11/25/DOCMGR-DYNAMIC-COMPLETION-add-dynamic-flag-completion-to-docmgr-using-carapace/scripts/01-tmux-completion-test.sh — Helper used for testing


## 2025-11-25

Completed FlagCompletion wiring across verbs (doc list/search/relate, ticket create/list/close/rename, tasks list/add/check/uncheck/edit/remove, meta update, changelog update, vocab list/add, workspace doctor, template validate).

### Related Files

- docmgr/cmd/docmgr/cmds/** — Registered completions across commands
- docmgr/pkg/completion/actions.go — Added helpers (TaskIDs

