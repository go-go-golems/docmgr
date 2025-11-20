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

