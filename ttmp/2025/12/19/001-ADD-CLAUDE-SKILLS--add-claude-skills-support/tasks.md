# Tasks

## TODO

<!-- Tasks are managed via `docmgr task add/check/edit/remove`. -->
- [x] Add WhatFor/WhenToUse fields to pkg/models/document.go (YAML+JSON tags)
- [x] Extend internal/workspace/sqlite_schema.go docs table with what_for + when_to_use columns (and update schema tests)
- [x] Populate what_for/when_to_use during ingest in internal/workspace/index_builder.go
- [x] Hydrate what_for/when_to_use in QueryDocs (internal/workspace/query_docs_sql.go + query_docs.go)
- [x] Add Cobra command group cmd/docmgr/cmds/skill and register in cmd/docmgr/cmds/root.go
- [ ] Implement pkg/commands/skill_list.go (filters: --ticket, --topics, --file, --dir; outputs what_for/when_to_use/topics/related_paths)
- [ ] Implement pkg/commands/skill_show.go (show preamble + related files + body; decide ambiguity behavior)
- [ ] Add 'skill' to ttmp/vocabulary.yaml docTypes and add optional _templates skill scaffold
- [ ] Add tests: schema/query hydration + scenario test for skill list --file/--dir
- [ ] Update docs: pkg/doc/docmgr-cli-guide.md to include docmgr skill list/show + filtering examples
