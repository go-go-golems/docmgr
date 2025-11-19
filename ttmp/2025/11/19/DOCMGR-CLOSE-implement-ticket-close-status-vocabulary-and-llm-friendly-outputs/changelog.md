# Changelog

## 2025-11-19

- Initial workspace created


## 2025-11-19

Implemented ticket close command with atomic status/intent/changelog updates, structured output support, status vocabulary with doctor warnings, and tasks check improvements

### Related Files

- cmd/docmgr/cmds/ticket/close.go — CLI wiring for ticket close
- pkg/commands/doctor.go — Added status vocabulary validation with warnings
- pkg/commands/tasks.go — Added all_tasks_done suggestion and structured output to tasks check
- pkg/commands/ticket_close.go — New ticket close command implementation
- pkg/commands/vocab_add.go — Added status category support
- pkg/commands/vocab_list.go — Added status category listing
- pkg/models/document.go — Added Status field to Vocabulary struct
- ttmp/vocabulary.yaml — Added status vocabulary entries

