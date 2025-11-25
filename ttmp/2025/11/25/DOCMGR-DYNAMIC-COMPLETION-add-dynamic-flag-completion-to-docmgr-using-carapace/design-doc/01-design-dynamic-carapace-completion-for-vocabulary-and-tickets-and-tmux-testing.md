---
Title: Design — Dynamic Carapace Completion for Vocabulary and Tickets, and tmux Testing
Ticket: DOCMGR-DYNAMIC-COMPLETION
Status: active
Topics:
    - backend
    - cli
    - completion
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Add carapace-based dynamic completions for vocabulary-driven fields and tickets; document tmux-based testing setup and workflow.
LastUpdated: 2025-11-25T16:30:34.120979956-05:00
---

# Design — Dynamic Carapace Completion for Vocabulary and Tickets, and tmux Testing

## Executive Summary

We will integrate carapace dynamic completion into `docmgr` to provide runtime completions for:
- `--doc-type`, `--status`, `--topics`, `--intent` (from vocabulary)
- `--ticket` (discovered from existing ticket workspaces)

This uses carapace’s hidden `_carapace` subcommand and per-flag Actions to compute completion values based on the current workspace. We also define a reproducible tmux-based testing flow that builds `docmgr`, sources the carapace snippet for the active shell, and validates that completions resolve correctly.

## Problem Statement

- Current completion is static and limited; it doesn’t reflect vocabulary updates or existing tickets at runtime.
- New users (like interns) struggle to discover valid values for `--doc-type`, `--status`, `--topics`, `--intent`, and available tickets.
- We need first-class dynamic completions, plus a simple testing workflow to validate them across shells.

## Proposed Solution

1) Add carapace to `docmgr` and generate the hidden `_carapace` subcommand on the root command.
2) For key flags, define `carapace.Action` providers:
   - `doc-types`, `status`, `topics`, `intent`: read from `vocabulary.yaml` using existing code.
   - `ticket` (and other flags referring to tickets): discover current tickets from the docs root.
3) Register per-command `FlagCompletion` maps for the above flags on relevant commands (e.g., `doc add`, `ticket`, `list tickets`, `meta update`, etc.).
4) Document a tmux testing workflow to build, source, and verify completions in bash/zsh/fish.

## Key Code Reference

Root command construction:

```21:36:docmgr/cmd/docmgr/cmds/root.go
func NewRootCommand(helpSystem *help.HelpSystem) (*cobra.Command, error) {
    rootCmd := &cobra.Command{
        Use:   "docmgr",
        Short: "Document Manager for LLM Workflows",
        Long:  `docmgr is a structured document manager ...`,
    }
    help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)
    // ... attaches subcommands
    return rootCmd, nil
}
```

Entrypoint:

```12:20:docmgr/cmd/docmgr/main.go
func main() {
    helpSystem := help.NewHelpSystem()
    _ = doc.AddDocToHelpSystem(helpSystem)

    rootCmd, err := appcmds.NewRootCommand(helpSystem)
    if err != nil { /* ... */ }

    if err := rootCmd.Execute(); err != nil { /* ... */ }
}
```

Vocabulary loader:

```13:33:docmgr/pkg/commands/vocabulary.go
func LoadVocabulary() (*models.Vocabulary, error) {
    if path, err := workspace.ResolveVocabularyPath(); err == nil {
        if _, err2 := os.Stat(path); err2 == nil {
            return loadVocabularyFromFile(path)
        }
    }
    return &models.Vocabulary{ Topics: [], DocTypes: [], Intent: [], Status: [] }, nil
}
```

Ticket discovery:

```21:31:docmgr/internal/workspace/discovery.go
func CollectTicketWorkspaces(root string, skipDir func(relPath, baseName string) bool) ([]TicketWorkspace, error) {
    workspaces := []TicketWorkspace{}
    err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        // looks for index.md and parses frontmatter
    })
    // returns discovered workspaces (sorted)
}
```

`doc add` command (flags we’ll complete):

```71:119:docmgr/pkg/commands/add.go
cmds.WithFlags(
    parameters.NewParameterDefinition("ticket",   parameters.ParameterTypeString,     ...),
    parameters.NewParameterDefinition("doc-type", parameters.ParameterTypeString,     ...),
    parameters.NewParameterDefinition("title",    parameters.ParameterTypeString,     ...),
    parameters.NewParameterDefinition("topics",   parameters.ParameterTypeStringList, ...),
    parameters.NewParameterDefinition("owners",   parameters.ParameterTypeStringList, ...),
    parameters.NewParameterDefinition("status",   parameters.ParameterTypeString,     ...),
    parameters.NewParameterDefinition("intent",   parameters.ParameterTypeString,     ...),
)
```

## Design Details

### A. Wire carapace into docmgr

- Add dependency: `github.com/carapace-sh/carapace`
- In `NewRootCommand`, after building the tree, initialize carapace and add the hidden `_carapace` command.

Sketch:

```go
// cmd/docmgr/cmds/completion.go (new file)
package cmds

import (
    "github.com/carapace-sh/carapace"
    "github.com/spf13/cobra"
)

func AttachCarapace(root *cobra.Command) {
    carapace.Gen(root) // adds hidden `_carapace` and registers bridge
}
```

Call from `NewRootCommand`:

```go
rootCmd := /* ... */
// after all Attach(...)
AttachCarapace(rootCmd)
return rootCmd, nil
```

### B. Define Action providers

Create vocabulary and ticket Actions (in `cmd/docmgr/cmds/completion_actions.go`):

```go
func actionVocabDocTypes() carapace.Action {
    return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
        v, _ := commands.LoadVocabulary()
        vals := make([]string, 0, len(v.DocTypes))
        for _, dt := range v.DocTypes { vals = append(vals, dt.Slug) }
        return carapace.ActionValues(vals...)
    })
}

func actionVocabStatus() carapace.Action { /* same with v.Status */ }
func actionVocabIntent() carapace.Action { /* same with v.Intent */ }
func actionVocabTopicsList() carapace.Action {
    return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
        v, _ := commands.LoadVocabulary()
        vals := make([]string, 0, len(v.Topics))
        for _, t := range v.Topics { vals = append(vals, t.Slug) }
        // topics is a list: support comma-separated unique lists
        return carapace.ActionValues(vals...).UniqueList(",")
    })
}

func actionTickets() carapace.Action {
    return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
        root := "ttmp" // default; consider reading from config similar to list tickets
        ws, _ := workspace.CollectTicketWorkspaces(workspace.ResolveRoot(root), nil)
        vals := make([]string, 0, len(ws))
        for _, w := range ws {
            if w.Doc != nil { vals = append(vals, w.Doc.Ticket) }
        }
        return carapace.ActionValues(vals...)
    })
}
```

### C. Register per-command flag completions

For example, in `cmd/docmgr/cmds/doc/add.go`, after building the command with `common.BuildCommand(...)`:

```15:20:docmgr/cmd/docmgr/cmds/doc/add.go
return common.BuildCommand(
    cmd,
    cli.WithDualMode(true),
    cli.WithGlazeToggleFlag("with-glaze-output"),
)
```

Update to register completions on the returned `*cobra.Command`:

```go
co, err := common.BuildCommand(
    cmd,
    cli.WithDualMode(true),
    cli.WithGlazeToggleFlag("with-glaze-output"),
)
if err != nil { return nil, err }
carapace.Gen(co).FlagCompletion(carapace.ActionMap{
    "ticket":   actionTickets(),
    "doc-type": actionVocabDocTypes(),
    "status":   actionVocabStatus(),
    "intent":   actionVocabIntent(),
    "topics":   actionVocabTopicsList(),
})
return co, nil
```

Repeat analogous registrations for other commands that expose these flags (e.g., `list tickets --status`, `ticket close --status`, etc.).

### D. Shell integration (dynamic)

Carapace requires sourcing a shell snippet that calls back into the binary on each completion.

- For bash:
  ```bash
  source <(docmgr _carapace bash)
  ```
- For zsh:
  ```bash
  source <(docmgr _carapace zsh)
  ```
- For fish:
  ```bash
  docmgr _carapace fish | source
  ```

This differs from cobra’s static `docmgr completion <shell>`; use `_carapace` for dynamic values.

## Implementation Plan

1) Dependencies
   - Add `github.com/carapace-sh/carapace` to `go.mod`.

2) Core wiring
   - Create `cmd/docmgr/cmds/completion.go` and call `AttachCarapace(rootCmd)` from `NewRootCommand`.

3) Actions
   - Create `cmd/docmgr/cmds/completion_actions.go` with `actionVocabDocTypes`, `actionVocabStatus`, `actionVocabTopicsList`, `actionVocabIntent`, `actionTickets`.
   - Reuse `commands.LoadVocabulary()` and `workspace.CollectTicketWorkspaces()`.

4) Per-command registrations
   - `doc add`: complete `ticket`, `doc-type`, `topics`, `status`, `intent`.
   - `list tickets`: complete `status`.
   - `ticket close`: complete `status`.
   - Any other commands that expose these flags.

5) Docs & examples
   - Update `how-to-use` (already done) to explain dynamic completion.
   - Add examples to the reference/analysis doc.

6) Testing (tmux)
   - Build: `go build -o ./dist/docmgr ./cmd/docmgr`
   - Ensure PATH:
     ```bash
     export PATH="$PWD/dist:$PATH"
     which docmgr   # should point to ./dist/docmgr
     ```
   - Start tmux: `tmux new -s docmgr-test`
   - In pane A (bash):
     ```bash
     source <(docmgr _carapace bash)
     # Test
     docmgr doc add --doc-type <TAB>
     docmgr doc add --status <TAB>
     docmgr doc add --intent <TAB>
     docmgr doc add --topics <TAB>
     docmgr doc add --ticket <TAB>
     ```
   - In pane B (zsh):
     ```bash
     source <(docmgr _carapace zsh)
     # Same tests
     ```
   - Prepare data if needed:
     ```bash
     # Ensure vocabulary exists (docTypes, topics, status, intent)
     docmgr vocab list
     # Create a sample ticket for ticket completion
     docmgr ticket create-ticket --ticket DEMO-1 --title "Demo" --topics demo
     ```

7) Validation
   - Verify vocabulary changes are reflected immediately (edit `ttmp/vocabulary.yaml` and re-run `<TAB>`).
   - Verify new tickets show up instantly under `--ticket <TAB>`.

## Open Questions

- Should we complete `--owners` as well (from recent Owners used in workspace)?
- Should we add per-command context (e.g., filter `--status` to valid transitions for `ticket close`)?

## References

- Carapace wiring and snippet generation:

```114:121:carapace/carapace.go
// Snippet creates completion script for given shell.
func (c Carapace) Snippet(name string) (string, error) { /* ... */ }
```

- Hidden subcommand and completion broker:

```21:47:carapace/command.go
carapaceCmd := &cobra.Command{ Use: "_carapace", Hidden: true, Run: func(...) { /* calls complete() */ } }
```

- Cobra bridge:

```52:67:carapace/storage.go
func (s _storage) bridge(cmd *cobra.Command) { /* registers valid args + flag completion */ }
```

## Verb-by-Verb Dynamic Completion Plan (Do Not Omit Any)

Below is an exhaustive map of all verbs (command groups) and subcommands, with the flags that benefit from dynamic completion and the source of values. Use this as the canonical checklist when wiring `carapace.Gen(cmd).FlagCompletion(...)`.

- workspace
  - init (pkg: `docmgr/pkg/commands/init.go`)
    - --root: directories (ActionDirectories)
  - status (pkg: `docmgr/pkg/commands/status.go`)
    - --root: directories
  - doctor (pkg: `docmgr/pkg/commands/doctor.go`)
    - --root: directories
    - --ticket: tickets (CollectTicketWorkspaces)
    - --ignore-dir: directories (ActionDirectories)
    - --ignore-glob: file globs (suggest common patterns)
    - --stale-after: numeric (no dynamic set)
    - --fail-on: enum [none,warning,error]
  - configure (pkg: `docmgr/pkg/commands/configure.go`)
    - --root: directories
    - --intent: vocabulary.intent
    - --vocabulary: files (ActionFiles, prefer vocabulary.yaml)

- ticket
  - create-ticket (pkg: `docmgr/pkg/commands/create_ticket.go`)
    - --ticket: freeform
    - --title: freeform
    - --topics: vocabulary.topics (UniqueList(","))
    - --root: directories
    - --path-template: freeform
  - rename (pkg: `docmgr/pkg/commands/rename_ticket.go`)
    - --ticket: tickets
    - --new-ticket: freeform
    - --root: directories
  - close (pkg: `docmgr/pkg/commands/ticket_close.go`)
    - --ticket: tickets
    - --status: vocabulary.status
    - --intent: vocabulary.intent
    - --changelog-entry: freeform
    - --root: directories
  - list (pkg: `docmgr/pkg/commands/list_tickets.go`)
    - --root: directories
    - --ticket: tickets
    - --status: vocabulary.status

- doc
  - add (pkg: `docmgr/pkg/commands/add.go`)
    - --ticket: tickets
    - --doc-type: vocabulary.docTypes
    - --topics: vocabulary.topics (UniqueList(","))
    - --owners: freeform (could later suggest known Owners from config)
    - --status: vocabulary.status
    - --intent: vocabulary.intent
    - --root: directories
  - list (pkg: `docmgr/pkg/commands/list_docs.go`)
    - --root: directories
    - --ticket: tickets
    - --status: vocabulary.status
    - --doc-type: vocabulary.docTypes
    - --topics: vocabulary.topics (UniqueList(","))
  - search (pkg: `docmgr/pkg/commands/search.go`)
    - --root: directories
    - --ticket: tickets
    - --doc-type: vocabulary.docTypes
    - --status: vocabulary.status
    - --topics: vocabulary.topics (UniqueList(","))
    - --file: files (ActionFiles)
    - --dir: directories (ActionDirectories)
    - --external-source: freeform (URL)
    - --since/--until/--created-since/--updated-since: common presets [today,yesterday,last week,last month,...]
  - relate (pkg: `docmgr/pkg/commands/relate.go`)
    - --ticket: tickets
    - --doc: files (ActionFiles, markdown)
    - --file-note: MultiParts(":") with left part ActionFiles and right freeform note
    - --remove-files: files (ActionFiles)
    - --topics: vocabulary.topics (UniqueList(","))
    - --root: directories
  - renumber (pkg: `docmgr/pkg/commands/renumber.go`)
    - --ticket: tickets
    - --doc-type: vocabulary.docTypes
    - --root: directories
  - layout-fix (pkg: `docmgr/pkg/commands/layout_fix.go`)
    - --ticket: tickets
    - --root: directories
  - guidelines (pkg: `docmgr/pkg/commands/guidelines_cmd.go`)
    - --doc-type: vocabulary.docTypes

- meta
  - update (pkg: `docmgr/pkg/commands/meta_update.go`)
    - --doc: files (ActionFiles, markdown)
    - --ticket: tickets
    - --doc-type: vocabulary.docTypes (filters which docs under ticket)
    - --field: enum [Title,Ticket,Status,Topics,DocType,Intent,Owners,RelatedFiles,ExternalSources,Summary]
    - --value: contextual (if field in {Status,Intent} → vocabulary; if Topics → UniqueList(",") with vocabulary.topics)
    - --root: directories

- tasks
  - list (pkg: `docmgr/pkg/commands/tasks.go`)
    - --ticket: tickets
    - --tasks-file: files (ActionFiles → tasks.md)
    - --root: directories
  - add (pkg: `docmgr/pkg/commands/tasks.go`)
    - --ticket: tickets
    - --tasks-file: files (tasks.md)
    - --after: dynamic integers from current task IDs
    - --root: directories
  - check (pkg: `docmgr/pkg/commands/tasks.go`)
    - --ticket: tickets
    - --tasks-file: files (tasks.md)
    - --id: dynamic integers from current task IDs (UniqueList(","))
    - --match: freeform
    - --root: directories
  - uncheck (pkg: `docmgr/pkg/commands/tasks.go`)
    - --ticket: tickets
    - --tasks-file: files (tasks.md)
    - --id: dynamic task IDs
    - --match: freeform
    - --root: directories
  - edit (pkg: `docmgr/pkg/commands/tasks.go`)
    - --ticket: tickets
    - --tasks-file: files (tasks.md)
    - --id: dynamic task IDs
    - --root: directories
  - remove (pkg: `docmgr/pkg/commands/tasks.go`)
    - --ticket: tickets
    - --tasks-file: files (tasks.md)
    - --id: dynamic task IDs (UniqueList(","))
    - --root: directories

- changelog
  - update (pkg: `docmgr/pkg/commands/changelog.go`)
    - --ticket: tickets
    - --changelog-file: files (ActionFiles → changelog.md)
    - --file-note: MultiParts(":") with left part ActionFiles and right freeform note
    - --suggest/--apply-suggestions: bools
    - --query: freeform
    - --topics: vocabulary.topics (UniqueList(","))
    - --root: directories

- vocab
  - list (pkg: `docmgr/pkg/commands/vocab_list.go`)
    - --category: enum [topics,docTypes,intent,status]
    - --root: directories
  - add (pkg: `docmgr/pkg/commands/vocab_add.go`)
    - --category: enum [topics,docTypes,intent,status]
    - --slug: freeform
    - --root: directories

- list (aggregate)
  - list docs (pkg: `docmgr/pkg/commands/list_docs.go`)
    - same as doc list
  - list tickets (pkg: `docmgr/pkg/commands/list_tickets.go`)
    - same as ticket list

- template
  - validate (pkg: `docmgr/pkg/commands/template_validate.go`)
    - --root: directories
    - --path: files (ActionFiles, prefer *.templ)

- import
  - import file (pkg: `docmgr/pkg/commands/import_file.go`)
    - --file: files (ActionFiles)
    - --ticket: tickets
    - --doc-type: vocabulary.docTypes
    - --root: directories

- config
  - show (pkg: `docmgr/pkg/commands/config_show.go`)
    - no dynamic flags needed (reads resolved config)

### Notes
- For all `--root` flags, prefer `ActionDirectories()`.
- For all `--ticket` flags, use discovered tickets via `workspace.CollectTicketWorkspaces`.
- For vocabulary-backed values, call `commands.LoadVocabulary()` and map to slugs.
- For file-path flags, use `ActionFiles()`; for colon mappings like `--file-note`, use `MultiParts(":")` with left part `ActionFiles()`.
- For task IDs, dynamically parse `tasks.md` and supply numeric values (and consider adding descriptions from the task text).
