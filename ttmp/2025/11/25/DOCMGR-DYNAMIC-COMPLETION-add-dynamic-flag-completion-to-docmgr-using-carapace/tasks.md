# Tasks

## TODO

- [ ] Add tasks here

- [x] Wire carapace into root: add AttachCarapace(rootCmd) and hidden _carapace
- [x] Implement action helpers: tickets, vocab (docTypes/status/intent/topics), files/dirs, task IDs
- [x] Register FlagCompletion for doc add (ticket/doc-type/topics/status/intent)
- [x] Register FlagCompletion for doc list/search (ticket/doc-type/topics/status,file,dir)
- [x] Register FlagCompletion for doc relate (ticket, doc, file-note MultiParts, remove-files, topics)
- [x] Register FlagCompletion for ticket create (topics), list (status), close (ticket/status/intent), rename (ticket)
- [x] Register FlagCompletion for tasks (ticket, tasks-file, id from tasks.md) across list/add/check/uncheck/edit/remove
- [x] Register FlagCompletion for meta update (doc, ticket, doc-type, field enum; value dyn for status/intent/topics)
- [x] Register FlagCompletion for changelog update (ticket, changelog-file, file-note MultiParts, topics)
- [x] Register FlagCompletion for vocab list/add (category enum, root)
- [x] Register FlagCompletion for workspace (root; doctor: ticket/ignore-dir/ignore-glob/fail-on)
- [x] Register FlagCompletion for template validate (root, path *.templ)
- [ ] tmux test matrix: build, PATH setup, source _carapace for bash/zsh/fish, validate all verbs
- [x] Add tmux helper script to send TAB and capture outputs; write to /tmp and attach to diary
