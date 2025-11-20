# Changelog

## 2025-11-19

- Initial workspace created


## 2025-11-19

Created ticket and initial analysis for external postfix template rendering


## 2025-11-19

Implemented external postfix templates for verb outputs. Added template rendering infrastructure in internal/templates/verb_output.go with path resolution, FuncMap helpers, and rendering logic. Updated list_docs, list_tickets, and doctor commands to build template data structs and render postfix templates. Created example templates under ttmp/templates/ for doc/list, list/tickets, and doctor.

### Related Files

- docmgr/internal/templates/verb_output.go — Template rendering infrastructure with path resolution and FuncMap helpers
- docmgr/pkg/commands/doctor.go — Added template data struct building and postfix template rendering
- docmgr/pkg/commands/list_docs.go — Added template data struct building and postfix template rendering
- docmgr/pkg/commands/list_tickets.go — Added template data struct building and postfix template rendering
- docmgr/ttmp/templates/doc/list.templ — Example template for doc list command
- docmgr/ttmp/templates/doctor.templ — Example template for doctor command
- docmgr/ttmp/templates/list/tickets.templ — Example template for list tickets command


## 2025-11-19

Added newline separator before template output for better readability. Added missing topics (cli, templates, glaze) to vocabulary to resolve doctor warnings.

### Related Files

- docmgr/internal/templates/verb_output.go — Added newline separator before template rendering


## 2025-11-19

Added comprehensive unit tests for template FuncMap helpers. Created verb_output_test.go with 40+ test cases covering all helpers (slice, dict, set, get, add1, countBy) including edge cases and reflection-based struct field access. All tests pass.

### Related Files

- docmgr/internal/templates/verb_output_test.go — Comprehensive unit tests for all template FuncMap helpers


## 2025-11-19

Created comprehensive template data contracts reference documentation. Documented all data structures for doc list, list tickets, and doctor verbs, including common envelope, verb-specific fields, available template functions, and usage examples with best practices.

### Related Files

- docmgr/ttmp/2025/11/19/DOCMGR-OUTPUT-TEMPLATES-external-postfix-templates-for-verb-outputs/reference/01-template-data-contracts-reference.md — Complete reference documentation for template data contracts


## 2025-11-19

Created advanced template examples demonstrating complex patterns: nested loops, conditional filtering, map aggregations, slice operations, and data transformations. Examples show status breakdowns, topic frequency analysis, recent items filtering, and detailed issue reporting.

### Related Files

- docmgr/ttmp/templates/examples/ — Advanced template examples demonstrating complex patterns


## 2025-11-19

Session summary: Completed unit tests (40+ cases), comprehensive template data contracts documentation, and advanced template examples. 18/22 tasks complete. Core implementation is production-ready with full test coverage and documentation.


## 2025-11-19

Created reference document listing all verbs needing template support. Identified 5 verbs: status (high), search (high), tasks list (high), vocab list (medium), guidelines (low). Added tasks for high-priority verbs. Documented proposed data structures and implementation notes for each.

### Related Files

- docmgr/ttmp/2025/11/19/DOCMGR-OUTPUT-TEMPLATES-external-postfix-templates-for-verb-outputs/reference/02-verbs-needing-template-support.md — Reference document listing verbs needing templates with priority and data structures


## 2025-11-19

Prepared handoff tasks: Created individual tasks for each remaining verb (status, search, tasks list, vocab list, guidelines). Added enhancement tasks: migrate to glazed/sprig helpers, add --print-template-schema flag, create user tutorial. All tasks documented and ready for next developer.


## 2025-11-19

Final handoff preparation: Created individual tasks for all remaining verbs (5 tasks), added enhancement tasks (glazed helpers migration, --print-template-schema flag, user tutorial). Total: 18 completed, 12 pending. All tasks documented with implementation notes. Diary updated with handoff summary and next developer guidance.


## 2025-11-19

Added analysis: Template Schema Printing Design (flags, reflection approach, per-verb contracts). Surveyed glazed templating helpers (sprig + TemplateFuncs) and linked docs. Updated diary with findings.

### Related Files

- docmgr/ttmp/2025/11/19/DOCMGR-OUTPUT-TEMPLATES-external-postfix-templates-for-verb-outputs/analysis/02-template-schema-printing-design.md — Design for --print-template-schema with implementation outline


## 2025-11-19

Implemented template schema printing: added --print-template-schema and --schema-format flags to list docs, list tickets, and doctor; created templates.PrintSchema with reflection to JSON/YAML; added unit tests.

### Related Files

- docmgr/internal/templates/schema.go — Schema builder and printer
- docmgr/internal/templates/schema_test.go — Unit tests for schema printer
- docmgr/pkg/commands/doctor.go — Flags and wiring in Run
- docmgr/pkg/commands/list_docs.go — Flags and wiring in Run
- docmgr/pkg/commands/list_tickets.go — Flags and wiring in Run


## 2025-11-20

Added intern playbook for continuing external postfix templates: step-by-step workflow, schema printing, wiring, examples, and verification checklist.

### Related Files

- docmgr/ttmp/2025/11/19/DOCMGR-OUTPUT-TEMPLATES-external-postfix-templates-for-verb-outputs/playbooks/01-intern-playbook-continuing-external-templates.md — Intern playbook


## 2025-11-20

Added template support for status command: schema flags, template data building, and example template

### Related Files

- docmgr/pkg/commands/status.go — Added schema flags and postfix template rendering
- docmgr/ttmp/templates/status.templ — Example template for status command output


## 2025-11-20

Completed status command template support implementation


## 2025-11-20

Added template support for tasks list command: schema flags, template data building, and example template

### Related Files

- docmgr/pkg/commands/tasks.go — Added schema flags and postfix template rendering for tasks list verb
- docmgr/ttmp/templates/tasks/list.templ — Example template for tasks list command output

