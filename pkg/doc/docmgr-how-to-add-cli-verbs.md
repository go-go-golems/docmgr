---
Title: How to Add CLI Verbs to docmgr
Slug: how-to-add-cli-verbs
Short: Step-by-step guide for implementing new CLI commands in docmgr, covering command structure, workspace integration, and output formatting.
Topics:
- docmgr
- implementation
- commands
- cli
- development
IsTemplate: false
IsTopLevel: true
ShowPerDefault: false
SectionType: GeneralTopic
---

# How to Add CLI Verbs to docmgr

## Overview

docmgr commands follow a consistent pattern that separates command definition from business logic, integrates with the workspace system for document discovery and querying, and supports both human-friendly and structured output formats. This design enables commands to focus on their specific functionality while leveraging shared infrastructure for workspace resolution, indexing, and output formatting.

**This guide covers:** Creating new command groups, implementing list and action commands, integrating with the workspace system, and supporting dual-mode output (human-friendly and structured).

**Intended audience:** Developers extending docmgr with new commands or command groups.

## Command Architecture

docmgr uses Cobra for CLI parsing and the Glazed framework for structured output. Commands are organized in `cmd/docmgr/cmds/` with subdirectories for each command group (doc, vocab, ticket, etc.). Each group follows a consistent structure with an `Attach()` function that registers subcommands and individual command files implementing specific operations.

**Command hierarchy:**

```
docmgr (root)
├── doc
│   ├── add
│   ├── list
│   ├── search
│   └── relate
├── vocab
│   ├── list
│   └── add
├── ticket
│   ├── create-ticket
│   ├── list
│   └── close
└── skill (new)
    ├── list
    └── show
```

**Dual-mode output:**

Commands implement two interfaces to support different use cases:
- **BareCommand**: Human-friendly text output (default mode)
- **GlazeCommand**: Structured output (JSON/YAML/CSV) for scripting

Users enable structured output with `--with-glaze-output`, then select format via `--output json|yaml|csv|table`.

## Step 1: Create Command Group Structure

Create a new subdirectory in `cmd/docmgr/cmds/` for your command group:

```bash
mkdir -p cmd/docmgr/cmds/skill
```

Create the main attachment file (`cmd/docmgr/cmds/skill/skill.go`):

```go
package skill

import (
    "github.com/spf13/cobra"
)

// Attach registers the skill command tree under the provided root command.
func Attach(root *cobra.Command) error {
    skillCmd := &cobra.Command{
        Use:   "skill",
        Short: "Manage skills documentation",
    }
    
    listCmd, err := newListCommand()
    if err != nil {
        return err
    }
    showCmd, err := newShowCommand()
    if err != nil {
        return err
    }
    
    skillCmd.AddCommand(listCmd, showCmd)
    root.AddCommand(skillCmd)
    return nil
}
```

**Register in root** (`cmd/docmgr/cmds/root.go`):

```go
import (
    // ... existing imports
    "github.com/go-go-golems/docmgr/cmd/docmgr/cmds/skill"
)

func NewRootCommand(helpSystem *help.HelpSystem) (*cobra.Command, error) {
    // ... existing code
    
    if err := skill.Attach(rootCmd); err != nil {
        return nil, err
    }
    
    return rootCmd, nil
}
```

## Step 2: Implement List Command

List commands follow a consistent pattern: discover workspace, query documents, output results. They implement both `BareCommand` and `GlazeCommand` interfaces.

**Create command file** (`cmd/docmgr/cmds/skill/list.go`):

```go
package skill

import (
    "github.com/carapace-sh/carapace"
    "github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
    "github.com/go-go-golems/docmgr/pkg/commands"
    "github.com/go-go-golems/docmgr/pkg/completion"
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/spf13/cobra"
)

func newListCommand() (*cobra.Command, error) {
    cmd, err := commands.NewSkillListCommand()
    if err != nil {
        return nil, err
    }
    cobraCmd, err := common.BuildCommand(
        cmd,
        cli.WithDualMode(true),
        cli.WithGlazeToggleFlag("with-glaze-output"),
    )
    if err != nil {
        return nil, err
    }
    
    carapace.Gen(cobraCmd).FlagCompletion(carapace.ActionMap{
        "root":   completion.ActionDirectories(),
        "ticket": completion.ActionTickets(),
        "topics": completion.ActionTopics(),
    })
    return cobraCmd, nil
}
```

**Implement command** (`pkg/commands/skill_list.go`):

```go
package commands

import (
    "context"
    "fmt"
    
    "github.com/go-go-golems/docmgr/internal/workspace"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/types"
)

// SkillListCommand lists all skills
type SkillListCommand struct {
    *cmds.CommandDescription
}

// SkillListSettings holds command parameters
type SkillListSettings struct {
    Root   string   `glazed.parameter:"root"`
    Ticket string   `glazed.parameter:"ticket"`
    Topics []string `glazed.parameter:"topics"`
}

func NewSkillListCommand() (*SkillListCommand, error) {
    return &SkillListCommand{
        CommandDescription: cmds.NewCommandDescription(
            "list",
            cmds.WithShort("List all skills"),
            cmds.WithLong(`Lists all skills found in the workspace.

Skills can be located in:
  - Workspace root: /skills directory
  - Ticket-specific: <ticket>/skills directory

Columns:
  skill,what_for,when_to_use,topics,related_paths,path

Examples:
  # Human output
  docmgr skill list
  docmgr skill list --ticket MEN-3475
  docmgr skill list --topics api,backend

  # Structured output
  docmgr skill list --with-glaze-output --output json
`),
            cmds.WithFlags(
                parameters.NewParameterDefinition(
                    "root",
                    parameters.ParameterTypeString,
                    parameters.WithHelp("Root directory for docs"),
                    parameters.WithDefault("ttmp"),
                ),
                parameters.NewParameterDefinition(
                    "ticket",
                    parameters.ParameterTypeString,
                    parameters.WithHelp("Filter by ticket identifier"),
                    parameters.WithDefault(""),
                ),
                parameters.NewParameterDefinition(
                    "topics",
                    parameters.ParameterTypeStringList,
                    parameters.WithHelp("Filter by topics"),
                ),
            ),
        ),
    }, nil
}

// RunIntoGlazeProcessor implements GlazeCommand for structured output
func (c *SkillListCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    settings := &SkillListSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to parse settings: %w", err)
    }
    
    // Discover workspace
    ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{
        RootOverride: settings.Root,
    })
    if err != nil {
        return fmt.Errorf("failed to discover workspace: %w", err)
    }
    
    // Initialize index
    if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{
        IncludeBody: false,
    }); err != nil {
        return fmt.Errorf("failed to initialize workspace index: %w", err)
    }
    
    // Query skills (DocType == "skill")
    scope := workspace.Scope{Kind: workspace.ScopeRepo}
    if settings.Ticket != "" {
        scope = workspace.Scope{
            Kind:     workspace.ScopeTicket,
            TicketID: settings.Ticket,
        }
    }
    
    res, err := ws.QueryDocs(ctx, workspace.DocQuery{
        Scope: scope,
        Filters: workspace.DocFilters{
            DocType:   "skill",
            TopicsAny: settings.Topics,
        },
        Options: workspace.DocQueryOptions{
            IncludeErrors:       false,
            IncludeArchivedPath: true,
            IncludeScriptsPath:  true,
            IncludeControlDocs:  true,
        },
    })
    if err != nil {
        return fmt.Errorf("failed to query skills: %w", err)
    }
    
    // Output results
    for _, h := range res.Docs {
        if h.Doc == nil {
            continue
        }
        
        // Extract related paths
        relatedPaths := make([]string, 0, len(h.Doc.RelatedFiles))
        for _, rf := range h.Doc.RelatedFiles {
            relatedPaths = append(relatedPaths, rf.Path)
        }
        
        row := types.NewRow(
            types.MRP("skill", h.Doc.Title),
            types.MRP("what_for", h.Doc.WhatFor),      // Custom field
            types.MRP("when_to_use", h.Doc.WhenToUse), // Custom field
            types.MRP("topics", h.Doc.Topics),
            types.MRP("related_paths", relatedPaths),
            types.MRP("path", h.Path),
        )
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}

// Run implements BareCommand for human-friendly output
func (c *SkillListCommand) Run(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
) error {
    settings := &SkillListSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to parse settings: %w", err)
    }
    
    // Apply config root if present
    settings.Root = workspace.ResolveRoot(settings.Root)
    
    // Discover workspace and query (same as GlazeCommand)
    ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{
        RootOverride: settings.Root,
    })
    if err != nil {
        return fmt.Errorf("failed to discover workspace: %w", err)
    }
    
    if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{
        IncludeBody: false,
    }); err != nil {
        return fmt.Errorf("failed to initialize workspace index: %w", err)
    }
    
    scope := workspace.Scope{Kind: workspace.ScopeRepo}
    if settings.Ticket != "" {
        scope = workspace.Scope{
            Kind:     workspace.ScopeTicket,
            TicketID: settings.Ticket,
        }
    }
    
    res, err := ws.QueryDocs(ctx, workspace.DocQuery{
        Scope: scope,
        Filters: workspace.DocFilters{
            DocType:   "skill",
            TopicsAny: settings.Topics,
        },
        Options: workspace.DocQueryOptions{
            IncludeErrors:       false,
            IncludeArchivedPath: true,
            IncludeScriptsPath:  true,
            IncludeControlDocs:  true,
        },
    })
    if err != nil {
        return fmt.Errorf("failed to query skills: %w", err)
    }
    
    // Human-friendly output
    for _, h := range res.Docs {
        if h.Doc == nil {
            continue
        }
        fmt.Printf("Skill: %s\n", h.Doc.Title)
        if h.Doc.WhatFor != "" {
            fmt.Printf("  What for: %s\n", h.Doc.WhatFor)
        }
        if h.Doc.WhenToUse != "" {
            fmt.Printf("  When to use: %s\n", h.Doc.WhenToUse)
        }
        if len(h.Doc.Topics) > 0 {
            fmt.Printf("  Topics: %s\n", strings.Join(h.Doc.Topics, ", "))
        }
        fmt.Printf("  Path: %s\n\n", h.Path)
    }
    
    return nil
}

var _ cmds.GlazeCommand = &SkillListCommand{}
var _ cmds.BareCommand = &SkillListCommand{}
```

## Step 3: Implement Show Command

Show commands display detailed information about a single item. They typically use human-friendly output by default but can support structured output for scripting.

**Create command file** (`cmd/docmgr/cmds/skill/show.go`):

```go
package skill

import (
    "github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
    "github.com/go-go-golems/docmgr/pkg/commands"
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/spf13/cobra"
)

func newShowCommand() (*cobra.Command, error) {
    cmd, err := commands.NewSkillShowCommand()
    if err != nil {
        return nil, err
    }
    cobraCmd, err := common.BuildCommand(
        cmd,
        cli.WithDualMode(true),
        cli.WithGlazeToggleFlag("with-glaze-output"),
    )
    if err != nil {
        return nil, err
    }
    return cobraCmd, nil
}
```

**Implement command** (`pkg/commands/skill_show.go`):

```go
package commands

import (
    "context"
    "fmt"
    "strings"
    
    "github.com/go-go-golems/docmgr/internal/documents"
    "github.com/go-go-golems/docmgr/internal/workspace"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// SkillShowCommand shows detailed information about a skill
type SkillShowCommand struct {
    *cmds.CommandDescription
}

// SkillShowSettings holds command parameters
type SkillShowSettings struct {
    Root string `glazed.parameter:"root"`
    Skill string `glazed.parameter:"skill"`
}

func NewSkillShowCommand() (*SkillShowCommand, error) {
    return &SkillShowCommand{
        CommandDescription: cmds.NewCommandDescription(
            "show",
            cmds.WithShort("Show detailed information about a skill"),
            cmds.WithLong(`Shows detailed information about a specific skill.

The skill name can be matched by:
  - Exact title match
  - Filename slug (e.g., "01-api-design" matches "API Design")

Examples:
  docmgr skill show "API Design"
  docmgr skill show 01-api-design
`),
            cmds.WithFlags(
                parameters.NewParameterDefinition(
                    "root",
                    parameters.ParameterTypeString,
                    parameters.WithHelp("Root directory for docs"),
                    parameters.WithDefault("ttmp"),
                ),
                parameters.NewParameterDefinition(
                    "skill",
                    parameters.ParameterTypeString,
                    parameters.WithHelp("Skill name or slug to show"),
                    parameters.WithRequired(true),
                ),
            ),
        ),
    }, nil
}

// Run implements BareCommand (show commands typically don't need structured output)
func (c *SkillShowCommand) Run(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
) error {
    settings := &SkillShowSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to parse settings: %w", err)
    }
    
    settings.Root = workspace.ResolveRoot(settings.Root)
    
    // Discover workspace and query for matching skills
    ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{
        RootOverride: settings.Root,
    })
    if err != nil {
        return fmt.Errorf("failed to discover workspace: %w", err)
    }
    
    if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{
        IncludeBody: true,  // Need body for full display
    }); err != nil {
        return fmt.Errorf("failed to initialize workspace index: %w", err)
    }
    
    // Query all skills and find matches
    res, err := ws.QueryDocs(ctx, workspace.DocQuery{
        Scope: workspace.Scope{Kind: workspace.ScopeRepo},
        Filters: workspace.DocFilters{
            DocType: "skill",
        },
        Options: workspace.DocQueryOptions{
            IncludeBody:         true,
            IncludeErrors:       false,
            IncludeArchivedPath: true,
            IncludeScriptsPath:  true,
            IncludeControlDocs:  true,
        },
    })
    if err != nil {
        return fmt.Errorf("failed to query skills: %w", err)
    }
    
    // Find matching skill(s)
    var matches []workspace.DocHandle
    skillQuery := strings.ToLower(strings.TrimSpace(settings.Skill))
    
    for _, h := range res.Docs {
        if h.Doc == nil {
            continue
        }
        // Match by title or slug
        titleLower := strings.ToLower(h.Doc.Title)
        if titleLower == skillQuery || strings.Contains(titleLower, skillQuery) {
            matches = append(matches, h)
        }
    }
    
    if len(matches) == 0 {
        return fmt.Errorf("skill not found: %s", settings.Skill)
    }
    if len(matches) > 1 {
        fmt.Fprintf(os.Stderr, "Multiple skills found, showing first match:\n")
        for _, m := range matches {
            fmt.Fprintf(os.Stderr, "  - %s (%s)\n", m.Doc.Title, m.Path)
        }
    }
    
    // Display skill details
    h := matches[0]
    doc := h.Doc
    
    fmt.Printf("Title: %s\n", doc.Title)
    if doc.WhatFor != "" {
        fmt.Printf("\nWhat this skill is for:\n%s\n", doc.WhatFor)
    }
    if doc.WhenToUse != "" {
        fmt.Printf("\nWhen to use this skill:\n%s\n", doc.WhenToUse)
    }
    if len(doc.Topics) > 0 {
        fmt.Printf("\nTopics: %s\n", strings.Join(doc.Topics, ", "))
    }
    if len(doc.RelatedFiles) > 0 {
        fmt.Printf("\nRelated Files:\n")
        for _, rf := range doc.RelatedFiles {
            fmt.Printf("  - %s", rf.Path)
            if rf.Note != "" {
                fmt.Printf(": %s", rf.Note)
            }
            fmt.Printf("\n")
        }
    }
    if h.Body != "" {
        fmt.Printf("\n%s\n", h.Body)
    }
    
    return nil
}

var _ cmds.BareCommand = &SkillShowCommand{}
```

## Step 4: Common Patterns and Best Practices

**Workspace integration:**

Always use `workspace.DiscoverWorkspace()` to resolve the docs root. This ensures commands work consistently regardless of where they're invoked:

```go
ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{
    RootOverride: settings.Root,
})
if err != nil {
    return fmt.Errorf("failed to discover workspace: %w", err)
}
```

**Index initialization:**

Initialize the index only when you need to query documents. For simple file operations, you can use `documents.WalkDocuments()` instead:

```go
// For queries (filtering, searching)
if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{
    IncludeBody: false,  // Set true only if you need body content
}); err != nil {
    return fmt.Errorf("failed to initialize workspace index: %w", err)
}

// For simple traversal (no filtering needed)
err := documents.WalkDocuments(root, func(path string, doc *models.Document, body string, readErr error) error {
    // Process each document
    return nil
})
```

**Error handling:**

Handle parse errors gracefully. Documents with parse errors are still indexed (with error metadata), allowing diagnostics and repair workflows:

```go
for _, h := range res.Docs {
    if h.Doc == nil {
        // Document failed to parse, skip or report
        continue
    }
    // Process valid document
}
```

**Output formatting:**

Use `types.MRP()` (Make Row Pair) for structured output to ensure type-safe key-value pairs:

```go
row := types.NewRow(
    types.MRP("field1", value1),
    types.MRP("field2", value2),
    // Use appropriate types (string, int, []string, etc.)
)
gp.AddRow(ctx, row)
```

**Parameter definition:**

Define parameters with appropriate types and defaults:

```go
parameters.NewParameterDefinition(
    "ticket",
    parameters.ParameterTypeString,
    parameters.WithHelp("Filter by ticket identifier"),
    parameters.WithDefault(""),  // Empty string for optional filters
),
parameters.NewParameterDefinition(
    "topics",
    parameters.ParameterTypeStringList,  // For multiple values
    parameters.WithHelp("Filter by topics"),
),
parameters.NewParameterDefinition(
    "skill",
    parameters.ParameterTypeString,
    parameters.WithHelp("Skill name to show"),
    parameters.WithRequired(true),  // Required parameter
),
```

## Step 5: Testing Your Command

**Manual testing:**

```bash
# Build docmgr
cd docmgr
go build -o docmgr ./cmd/docmgr

# Test human output
./docmgr skill list
./docmgr skill show "API Design"

# Test structured output
./docmgr skill list --with-glaze-output --output json
./docmgr skill list --with-glaze-output --output yaml
```

**Integration testing:**

Create test documents in `ttmp/skills/`:

```yaml
---
Title: API Design Skill
DocType: skill
Topics: [api, design]
WhatFor: Designing RESTful APIs
WhenToUse: When starting a new API endpoint
---
# API Design Skill

This skill covers...
```

Run your command and verify output matches expectations.

## Common Pitfalls

**Forgetting to initialize index:**
- Always call `ws.InitIndex()` before `ws.QueryDocs()`
- Set `IncludeBody: true` only if you need body content (increases memory usage)

**Incorrect scope:**
- Use `ScopeRepo` for repository-wide queries
- Use `ScopeTicket` with `TicketID` for ticket-specific queries

**Missing error handling:**
- Check `h.Doc == nil` before accessing document fields
- Handle parse errors gracefully (skip or report)

**Wrong output format:**
- Use `types.MRP()` for structured output (not plain strings)
- Ensure field names match expected schema

## Next Steps

After implementing your command:

1. **Add completion support**: Register actions in `pkg/completion/actions.go`
2. **Add to help system**: Commands are automatically included via Cobra
3. **Write tests**: Add unit tests in `pkg/commands/` and integration tests in `test-scenarios/`
4. **Update documentation**: Add usage examples to relevant help docs

For more details on workspace integration, see `docmgr help codebase-architecture`.

