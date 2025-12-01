# Changelog

## 2025-11-18

- Initial workspace created


## 2025-11-18

Added Document.Validate() method to check required fields (Title, Ticket, DocType). Returns error listing all missing fields.

### Related Files

- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/models/document.go — Added Validate() method for Document type


## 2025-11-18

Integrated Document.Validate() method into doctor command. Doctor now uses the centralized validation method for required fields (Title, Ticket, DocType) and reports errors with clear messages.

### Related Files

- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/doctor.go — Updated to use Document.Validate() method


## 2025-11-18

Consolidated frontmatter parsing on adrg/frontmatter library. Removed manual splitFrontmatter() function from import_file.go and replaced extractFrontmatterAndBody() in templates.go with library-based parsing. This eliminates fragile edge case handling and provides robust YAML parsing.

### Related Files

- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/import_file.go — Replaced manual frontmatter splitting with adrg/frontmatter library
- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/templates.go — Replaced manual frontmatter extraction with adrg/frontmatter library


## 2025-11-18

Implemented 'docmgr config show' command to display configuration resolution process and active settings. Shows all configuration sources checked in precedence order with indicators for which source was used, and displays active configuration values.

### Related Files

- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/cmd/docmgr/main.go — Registered config show command
- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/config_show.go — New config show command


## 2025-11-18

Updated 'docmgr init' command to automatically create .ttmp.yaml config file template at repository root when initializing a new workspace. The config file includes root and vocabulary paths relative to the repo root.

### Related Files

- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/pkg/commands/init.go — Added automatic .ttmp.yaml config file creation


## 2025-12-01

Auto-closed: ticket was active but not created today

