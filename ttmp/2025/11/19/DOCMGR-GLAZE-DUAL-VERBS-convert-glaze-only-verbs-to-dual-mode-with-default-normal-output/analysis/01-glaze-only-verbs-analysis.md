---
Title: Analysis — Glaze-Only Verbs That Should Be Dual Mode
Ticket: DOCMGR-GLAZE-DUAL-VERBS
Status: active
Topics:
    - tooling
    - ux
    - cli
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Comprehensive analysis of docmgr verbs currently using glaze output only that should be converted to dual mode with default normal output
LastUpdated: 2025-11-19T00:00:00Z
---

# Analysis — Glaze-Only Verbs That Should Be Dual Mode

## Executive Summary

This analysis identifies all docmgr verbs that currently use **glaze output only** (GlazeCommand interface) but should be converted to **dual mode** (both BareCommand and GlazeCommand) with **default normal output** (human-friendly text). These commands are typically used interactively in terminals and LLM prompts, not in structured data contexts, making human-friendly output the better default.

## Background

### Current Output Modes in docmgr

docmgr supports three output modes:

1. **Glaze-only mode**: Commands implement only `GlazeCommand` interface, output structured data (table/JSON/YAML/CSV) by default
2. **Dual mode**: Commands implement both `BareCommand` (human-friendly) and `GlazeCommand` (structured), with human-friendly as default
3. **Bare-only mode**: Commands implement only `BareCommand`, output human-friendly text only

### Dual Mode Pattern

Dual mode commands:
- Default to human-friendly output (BareCommand implementation)
- Support structured output via `--with-glaze-output --output json|yaml|csv|table`
- Are registered with `cli.WithDualMode(true)` and `cli.WithGlazeToggleFlag("with-glaze-output")`

Example from `changelog update`:
```go
// In cmd/docmgr/cmds/changelog/update.go
return common.BuildCommand(
    cmd,
    cli.WithDualMode(true),
    cli.WithGlazeToggleFlag("with-glaze-output"),
)
```

## Commands Currently Glaze-Only (Should Be Dual Mode)

### 1. `doc add` / `docmgr add` (AddCommand)

**Current State**: GlazeCommand only
**Location**: `docmgr/pkg/commands/add.go`
**Registration**: `docmgr/cmd/docmgr/cmds/doc/add.go`

**Current Behavior**: Outputs structured table with ticket, doc_type, title, path, status

**Why Should Be Dual Mode**:
- Primarily used interactively when creating documents
- Users want immediate feedback about what was created (path, confirmation)
- Guidelines are printed after creation (lines 272-286), which is human-friendly output
- The structured output is rarely needed in scripts

**Recommended Human Output**:
```
Document created: design-doc/01-path-normalization-strategy.md
Ticket: MEN-4242
Title: Path Normalization Strategy

===== Guidelines for design-doc =====
[guideline text]
```

**Implementation Notes**:
- Already prints guidelines (human-friendly) in RunIntoGlazeProcessor
- Need to extract this to BareCommand Run() method
- Keep GlazeCommand for structured output when needed

---

### 2. `ticket create-ticket` (CreateTicketCommand)

**Current State**: GlazeCommand only
**Location**: `docmgr/pkg/commands/create_ticket.go`
**Registration**: `docmgr/cmd/docmgr/cmds/ticket/create.go`

**Current Behavior**: Outputs structured table with ticket, path, title, status

**Why Should Be Dual Mode**:
- Most common interactive command for starting new tickets
- Users want clear confirmation of what was created and where
- The path output is critical for users to know where to navigate
- Structured output rarely needed in automation (ticket creation is usually manual)

**Recommended Human Output**:
```
Ticket workspace created: DOCMGR-GLAZE-DUAL-VERBS
Path: ttmp/2025/11/19/DOCMGR-GLAZE-DUAL-VERBS-convert-glaze-only-verbs-to-dual-mode-with-default-normal-output

Created files:
  - index.md
  - tasks.md
  - changelog.md
  - README.md
```

**Implementation Notes**:
- Currently creates multiple files (index.md, tasks.md, changelog.md, README.md)
- Human output should list what was created
- Keep GlazeCommand for CI/automation scenarios

---

### 3. `meta update` (MetaUpdateCommand)

**Current State**: GlazeCommand only
**Location**: `docmgr/pkg/commands/meta_update.go`
**Registration**: `docmgr/cmd/docmgr/cmds/meta/update.go`

**Current Behavior**: Outputs structured table with doc, field, value, status

**Why Should Be Dual Mode**:
- Frequently used interactively to update metadata
- Users want confirmation of what changed
- Often used in quick workflows where structured output adds noise
- The confirmation message is more useful than a table

**Recommended Human Output**:
```
Updated field 'Owners' to 'manuel,alex' in ttmp/.../index.md
Updated field 'Status' to 'active' in ttmp/.../design/01-architecture.md
```

**Implementation Notes**:
- Can update multiple docs at once (via --ticket or --doc-type)
- Human output should show each update clearly
- Keep GlazeCommand for bulk operations in scripts

---

### 4. `doc relate` (RelateCommand)

**Current State**: GlazeCommand only
**Location**: `docmgr/pkg/commands/relate.go`
**Registration**: `docmgr/cmd/docmgr/cmds/doc/relate.go`

**Current Behavior**: Outputs structured table with doc, added, updated, removed, total, status

**Why Should Be Dual Mode**:
- Used interactively when relating files to docs
- Users want confirmation of what files were linked
- The counts (added/updated/removed) are useful but better as human-readable text
- Structured output rarely needed

**Recommended Human Output**:
```
Related files updated: ttmp/.../index.md
  Added: 2 files
  Updated: 1 file
  Removed: 0 files
  Total: 5 files

Files added:
  - backend/api/register.go: Registers API routes
  - backend/ws/manager.go: WebSocket lifecycle management
```

**Implementation Notes**:
- Currently tracks added/updated/removed counts
- Human output should show the actual files changed
- Keep GlazeCommand for bulk operations

---

### 5. `task edit` (TasksEditCommand)

**Current State**: GlazeCommand only
**Location**: `docmgr/pkg/commands/tasks.go` (line 482)
**Registration**: `docmgr/cmd/docmgr/cmds/tasks/edit.go`

**Current Behavior**: Outputs structured table (implementation not shown in grep results)

**Why Should Be Dual Mode**:
- Other task commands (add, check, uncheck, remove) are BareCommand only
- Consistency: task operations should have similar UX
- Used interactively when editing task text
- Users want simple confirmation

**Recommended Human Output**:
```
Task #2 updated in ttmp/.../tasks.md
  Old: "Update API docs"
  New: "Update API docs for /chat/v2"
```

**Implementation Notes**:
- Should match pattern of other task commands (BareCommand)
- Keep GlazeCommand for consistency with list command

---

### 6. `vocab add` (VocabAddCommand)

**Current State**: GlazeCommand only
**Location**: `docmgr/pkg/commands/vocab_add.go`
**Registration**: `docmgr/cmd/docmgr/cmds/vocab/add.go`

**Current Behavior**: Outputs structured table with category, slug, description, status

**Why Should Be Dual Mode**:
- `vocab list` is already dual mode
- Used interactively when adding vocabulary entries
- Users want confirmation of what was added
- Structured output rarely needed

**Recommended Human Output**:
```
Vocabulary entry added:
  Category: topics
  Slug: frontend
  Description: Frontend code and components
```

**Implementation Notes**:
- Should match pattern of vocab list (dual mode)
- Keep GlazeCommand for bulk vocabulary imports

---

### 7. `import file` (ImportFileCommand)

**Current State**: GlazeCommand only
**Location**: `docmgr/pkg/commands/import_file.go`
**Registration**: `docmgr/cmd/docmgr/cmds/importcmd/file.go`

**Current Behavior**: Outputs structured table (implementation not fully visible)

**Why Should Be Dual Mode**:
- Used interactively when importing external documents
- Users want confirmation of import location and status
- The import path is critical information
- Structured output rarely needed

**Recommended Human Output**:
```
File imported: external-doc.md
  Source: /path/to/source.md
  Destination: ttmp/.../sources/external-doc.md
  Status: imported
```

**Implementation Notes**:
- Import operations benefit from clear human feedback
- Keep GlazeCommand for bulk imports

---

### 8. `doc layout-fix` (LayoutFixCommand)

**Current State**: GlazeCommand only
**Location**: `docmgr/pkg/commands/layout_fix.go`
**Registration**: `docmgr/cmd/docmgr/cmds/doc/layout_fix.go`

**Current Behavior**: Outputs structured table (implementation not fully visible)

**Why Should Be Dual Mode**:
- Used interactively to fix file layout issues
- Users want to see what was fixed
- Diagnostic/repair commands benefit from human-readable output
- Structured output rarely needed

**Recommended Human Output**:
```
Layout fixes applied:
  Fixed: 3 files
  Moved: 1 file
  Created: 0 directories
```

**Implementation Notes**:
- Diagnostic commands should be human-friendly by default
- Keep GlazeCommand for CI validation scenarios

---

### 9. `doc renumber` (RenumberCommand)

**Current State**: GlazeCommand only
**Location**: `docmgr/pkg/commands/renumber.go`
**Registration**: `docmgr/cmd/docmgr/cmds/doc/renumber.go`

**Current Behavior**: Outputs structured table (implementation not fully visible)

**Why Should Be Dual Mode**:
- Used interactively to fix numeric prefixes
- Users want to see what was renumbered
- Diagnostic/repair commands benefit from human-readable output
- Structured output rarely needed

**Recommended Human Output**:
```
Renumbered documents in ttmp/.../design/
  01-old-name.md -> 01-new-name.md
  02-another.md -> 02-another.md (no change)
  03-final.md -> 03-final.md (no change)
```

**Implementation Notes**:
- Should show what changed vs what stayed the same
- Keep GlazeCommand for bulk operations

---

### 10. `workspace configure` (ConfigureCommand)

**Current State**: GlazeCommand only
**Location**: `docmgr/pkg/commands/configure.go`
**Registration**: `docmgr/cmd/docmgr/cmds/workspace/configure.go`

**Current Behavior**: Outputs structured table with config_file, status

**Why Should Be Dual Mode**:
- Used interactively to set up workspace configuration
- Users want confirmation of configuration file location
- Setup commands should be human-friendly
- Structured output rarely needed

**Recommended Human Output**:
```
Configuration file created: .ttmp.yaml
Location: /path/to/repo/.ttmp.yaml

Configuration:
  root: ttmp
  vocabulary: ttmp/vocabulary.yaml
```

**Implementation Notes**:
- Should show what was configured
- Keep GlazeCommand for CI setup scenarios

---

### 11. `workspace init` (InitCommand)

**Current State**: GlazeCommand only
**Location**: `docmgr/pkg/commands/init.go`
**Registration**: `docmgr/cmd/docmgr/cmds/workspace/init.go`

**Current Behavior**: Outputs structured table with root, vocabulary, templates, guidelines, docmgrignore, status

**Why Should Be Dual Mode**:
- Used interactively to initialize workspace
- Users want clear confirmation of what was created
- Setup commands should be human-friendly
- Structured output rarely needed

**Recommended Human Output**:
```
Workspace initialized: ttmp/

Created:
  - vocabulary.yaml
  - _templates/ (with default templates)
  - _guidelines/ (with default guidelines)
  - .docmgrignore

Ready to create tickets!
```

**Implementation Notes**:
- Should list what was created
- Keep GlazeCommand for CI setup scenarios

---

## Commands Already Dual Mode (Reference)

These commands already implement dual mode correctly:

1. **`doc list`** (ListDocsCommand) - Dual mode ✓
2. **`ticket list`** (ListTicketsCommand) - Dual mode ✓
3. **`doc search`** (SearchCommand) - Dual mode ✓
4. **`workspace status`** (StatusCommand) - Dual mode ✓
5. **`task list`** (TasksListCommand) - Dual mode ✓
6. **`changelog update`** (ChangelogUpdateCommand) - Dual mode ✓
7. **`doc guidelines`** (GuidelinesCommand) - Dual mode ✓
8. **`vocab list`** (VocabListCommand) - Dual mode ✓
9. **`config show`** (ConfigShowCommand) - Dual mode ✓
10. **`ticket rename`** (RenameTicketCommand) - Dual mode ✓

## Commands That Should Stay Glaze-Only

### `doctor` (DoctorCommand)

**Current State**: GlazeCommand only
**Reason to Keep Glaze-Only**:
- Primarily used in CI/CD pipelines where structured output is needed
- Validation results are better consumed as JSON/YAML for reporting
- Human output can be enabled via `--output table` if needed
- The structured output is the primary use case

**Note**: Could potentially be dual mode, but glaze-only is acceptable given CI usage.

## Commands That Are Bare-Only (Should Stay As-Is)

These commands are intentionally BareCommand only and should not be changed:

1. **`task add`** (TasksAddCommand) - BareCommand only ✓
2. **`task check`** (TasksCheckCommand) - BareCommand only ✓
3. **`task uncheck`** (TasksUncheckCommand) - BareCommand only ✓
4. **`task remove`** (TasksRemoveCommand) - BareCommand only ✓

These are simple mutation commands that don't need structured output.

## Implementation Pattern

### Step 1: Add BareCommand Implementation

For each command, add a `Run()` method that implements `BareCommand`:

```go
// Implement BareCommand for human-friendly output
func (c *AddCommand) Run(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
) error {
    settings := &AddSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to parse settings: %w", err)
    }
    
    // ... perform the operation ...
    
    // Human-friendly output
    fmt.Printf("Document created: %s\n", docPath)
    fmt.Printf("Ticket: %s\n", settings.Ticket)
    fmt.Printf("Title: %s\n", settings.Title)
    
    // Print guidelines if applicable
    // ...
    
    return nil
}

var _ cmds.BareCommand = &AddCommand{}
```

### Step 2: Update Registration

Update the command registration to enable dual mode:

```go
// In cmd/docmgr/cmds/doc/add.go
func newAddCommand() (*cobra.Command, error) {
    cmd, err := commands.NewAddCommand()
    if err != nil {
        return nil, err
    }
    return common.BuildCommand(
        cmd,
        cli.WithDualMode(true),
        cli.WithGlazeToggleFlag("with-glaze-output"),
    )
}
```

### Step 3: Keep GlazeCommand Implementation

The existing `RunIntoGlazeProcessor()` method remains unchanged, providing structured output when `--with-glaze-output` is used.

## Summary Table

| Command | Current Mode | Should Be | Priority | Notes |
|---------|-------------|-----------|----------|-------|
| `doc add` | Glaze-only | Dual | High | Most common interactive command |
| `ticket create-ticket` | Glaze-only | Dual | High | Most common interactive command |
| `meta update` | Glaze-only | Dual | High | Frequently used interactively |
| `doc relate` | Glaze-only | Dual | Medium | Used interactively |
| `task edit` | Glaze-only | Dual | Medium | Consistency with other task commands |
| `vocab add` | Glaze-only | Dual | Medium | Consistency with vocab list |
| `import file` | Glaze-only | Dual | Low | Less frequently used |
| `doc layout-fix` | Glaze-only | Dual | Low | Diagnostic command |
| `doc renumber` | Glaze-only | Dual | Low | Diagnostic command |
| `workspace configure` | Glaze-only | Dual | Low | Setup command |
| `workspace init` | Glaze-only | Dual | Low | Setup command |
| `doctor` | Glaze-only | Keep | - | CI-focused, structured output needed |

## Testing Considerations

For each converted command, verify:

1. **Default behavior**: Human-friendly output when run without flags
2. **Structured output**: `--with-glaze-output --output json` produces valid JSON
3. **Backward compatibility**: Existing scripts using glaze output still work
4. **Help text**: Update help text to mention `--with-glaze-output` flag

## References

- Dual mode pattern: `docmgr/cmd/docmgr/cmds/changelog/update.go`
- Human output example: `docmgr/pkg/commands/changelog.go` (Run method, lines 360-525)
- Glaze output example: `docmgr/pkg/commands/changelog.go` (RunIntoGlazeProcessor method, lines 360-360)
- Registration pattern: `docmgr/cmd/docmgr/cmds/doc/list.go`

