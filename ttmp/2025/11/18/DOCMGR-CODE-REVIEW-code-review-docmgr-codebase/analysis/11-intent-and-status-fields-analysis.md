---
Title: Intent and Status Fields Analysis
Ticket: DOCMGR-CODE-REVIEW
Status: active
Topics:
    - docmgr
    - code-review
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Analysis of Intent and Status fields: how they're defined, validated, set, and modified in docmgr. Intent uses controlled vocabulary; Status is free-form.
LastUpdated: 2025-11-19T14:14:54.016246947-05:00
---

# Intent and Status Fields Analysis

## Overview

This document analyzes how `Intent` and `Status` fields are defined, validated, and managed for tickets and documents in docmgr. These fields serve different purposes and have different validation mechanisms.

## Intent Field

### Definition and Purpose

The `Intent` field indicates the longevity expectation for documentation:
- **Purpose**: Distinguishes between long-term documentation (should be maintained) and temporary documentation (for active work)
- **Type**: Controlled vocabulary (managed via `vocabulary.yaml`)
- **Field Location**: Both tickets (index.md) and individual documents

### Available Values

Currently defined in `vocabulary.yaml`:
- `long-term`: Long-term documentation that should be maintained

### Default Values

1. **When creating tickets** (`create-ticket` command):
   - Checks workspace config defaults first (`cfg.Defaults.Intent`)
   - Falls back to `"long-term"` if not configured
   - See: ```142:147:docmgr/pkg/commands/create_ticket.go```

2. **When creating documents** (`doc add` command):
   - Inherits from ticket's index.md `Intent` field
   - If ticket intent is empty, defaults to `"long-term"`
   - Can be overridden via `--intent` flag
   - See: ```215:221:docmgr/pkg/commands/add.go```

### How Intent is Set/Modified

1. **During creation**:
   ```bash
   # Via config defaults (for tickets)
   # Set in .ttmp.yaml:
   defaults:
     intent: long-term
   
   # Via command flag (for documents)
   docmgr doc add --ticket MEN-XXXX --doc-type design-doc --title "..." --intent long-term
   ```

2. **After creation**:
   ```bash
   # Update via meta update command
   docmgr meta update --ticket MEN-XXXX --field Intent --value long-term
   docmgr meta update --doc path/to/doc.md --field Intent --value long-term
   ```

3. **Via vocabulary management**:
   ```bash
   # Add new intent values to vocabulary
   docmgr vocab add --category intent --slug short-term --description "Temporary documentation for active work"
   ```

### Validation

- **Vocabulary validation**: The `doctor` command checks if intent values exist in vocabulary.yaml
- **Warning level**: Unknown intent values generate warnings (not errors)
- **See**: ```318:334:docmgr/pkg/commands/doctor.go```

### Code References

- Model definition: ```75:75:docmgr/pkg/models/document.go```
- Vocabulary structure: ```149:153:docmgr/pkg/models/document.go```
- Default handling in create_ticket: ```142:147:docmgr/pkg/commands/create_ticket.go```
- Default handling in add: ```215:221:docmgr/pkg/commands/add.go```
- Update handling: ```208:209:docmgr/pkg/commands/meta_update.go```

## Status Field

### Definition and Purpose

The `Status` field indicates the current state of a ticket or document:
- **Purpose**: Tracks workflow state (active, draft, review, complete, etc.)
- **Type**: Free-form string (NOT managed via vocabulary)
- **Field Location**: Both tickets (index.md) and individual documents

### Available Values

Status is **not constrained** by vocabulary. Common values found in the codebase:
- `active`: Ticket/document is actively being worked on
- `draft`: Document is in draft state
- `review`: Document is ready for review
- `complete`: Work is finished
- `needs-review`: Document needs review
- `archived`: Document/ticket is archived

**Note**: These are examples only - any string value is accepted.

### Default Values

1. **When creating tickets** (`create-ticket` command):
   - Hardcoded to `"active"` (no config override available)
   - See: ```139:139:docmgr/pkg/commands/create_ticket.go```

2. **When creating documents** (`doc add` command):
   - Inherits from ticket's index.md `Status` field
   - Can be overridden via `--status` flag
   - See: ```210:213:docmgr/pkg/commands/add.go```

### How Status is Set/Modified

1. **During creation**:
   ```bash
   # Via command flag (for documents)
   docmgr doc add --ticket MEN-XXXX --doc-type design-doc --title "..." --status draft
   ```

2. **After creation**:
   ```bash
   # Update via meta update command
   docmgr meta update --ticket MEN-XXXX --field Status --value active
   docmgr meta update --doc path/to/doc.md --field Status --value review
   
   # Update all documents of a specific type under a ticket
   docmgr meta update --ticket MEN-XXXX --doc-type design-doc --field Status --value review
   ```

3. **Manual editing**: Status can be edited directly in YAML frontmatter

### Validation

- **Presence check**: The `doctor` command warns if Status is empty (but doesn't fail validation)
- **No vocabulary validation**: Status values are NOT validated against vocabulary
- **No enum constraints**: Any string value is accepted
- **See**: ```268:275:docmgr/pkg/commands/doctor.go```

### Code References

- Model definition: ```72:72:docmgr/pkg/models/document.go```
- Default in create_ticket: ```139:139:docmgr/pkg/commands/create_ticket.go```
- Default handling in add: ```210:213:docmgr/pkg/commands/add.go```
- Update handling: ```194:195:docmgr/pkg/commands/meta_update.go```
- Status display: ```312:314:docmgr/pkg/commands/status.go```

## Key Differences

| Aspect | Intent | Status |
|--------|--------|--------|
| **Vocabulary managed** | ✅ Yes | ❌ No |
| **Default value** | `"long-term"` | `"active"` (tickets only) |
| **Configurable default** | ✅ Yes (via .ttmp.yaml) | ❌ No |
| **Validation** | ✅ Warns on unknown values | ⚠️ Only checks for presence |
| **Enum constraints** | ✅ Yes (via vocabulary) | ❌ No (free-form) |
| **Can be empty** | ✅ Yes (defaults applied) | ⚠️ Warns if empty |

## Recommendations

### For Intent

1. **Add common intent values to vocabulary**:
   - Consider adding `short-term` or `temporary` for experimental docs
   - Document intent values in team guidelines

2. **Use config defaults**:
   - Set default intent in `.ttmp.yaml` for consistency:
     ```yaml
     defaults:
       intent: long-term
     ```

### For Status

1. **Consider vocabulary management**:
   - Status is currently free-form, which allows flexibility but may lead to inconsistency
   - Consider adding status to vocabulary if standardization is desired
   - Alternatively, document common status values in team guidelines

2. **Document common workflows**:
   - Document typical status transitions (draft → review → active → complete)
   - Provide examples in documentation

## Implementation Notes

### Intent Implementation

- Intent values are stored in `vocabulary.yaml` under the `intent` key
- Vocabulary is loaded and validated by the `doctor` command
- Unknown intent values generate warnings but don't prevent document creation
- Intent can be managed via `docmgr vocab` commands

### Status Implementation

- Status is a simple string field with no vocabulary constraints
- Status defaults are hardcoded in `create_ticket.go` (always `"active"`)
- Status can be filtered in list/search commands (`--status` flag)
- Status is displayed in status command output

## Related Commands

- `docmgr doc add`: Create documents with intent/status flags
- `docmgr meta update`: Update intent/status fields
- `docmgr vocab add`: Add new intent values to vocabulary
- `docmgr vocab list`: List available intent values
- `docmgr doctor`: Validate intent values against vocabulary
- `docmgr list --status`: Filter by status value
- `docmgr search --status`: Search by status value
