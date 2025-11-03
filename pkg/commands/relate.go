package commands

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"

    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/types"
)

// RelateCommand updates RelatedFiles metadata and can suggest files to relate
type RelateCommand struct {
    *cmds.CommandDescription
}

type RelateSettings struct {
    Ticket           string   `glazed.parameter:"ticket"`
    Doc              string   `glazed.parameter:"doc"`
    Files            []string `glazed.parameter:"files"`
    RemoveFiles      []string `glazed.parameter:"remove-files"`
    Suggest          bool     `glazed.parameter:"suggest"`
    ApplySuggestions bool     `glazed.parameter:"apply-suggestions"`
    Query            string   `glazed.parameter:"query"`
    Topics           []string `glazed.parameter:"topics"`
    Root             string   `glazed.parameter:"root"`
}

func NewRelateCommand() (*RelateCommand, error) {
    return &RelateCommand{
        CommandDescription: cmds.NewCommandDescription(
            "relate",
            cmds.WithShort("Relate code files to a document or ticket"),
            cmds.WithLong(`Update RelatedFiles in a document's frontmatter or the ticket index.

Examples:
  # Relate files to the ticket index
  docmgr relate --ticket MEN-4242 --files backend/chat/api/register.go,web/src/store/api/chatApi.ts

  # Relate files to a specific document
  docmgr relate --doc ttmp/MEN-4242-.../design/path-normalization-strategy.md --files backend/chat/ws/manager.go

  # Suggest files using heuristics (git + ripgrep + existing RelatedFiles)
  docmgr relate --ticket MEN-4242 --suggest --query WebSocket --topics chat,backend

  # Apply suggestions automatically to the ticket index
  docmgr relate --ticket MEN-4242 --suggest --apply-suggestions --query WebSocket
`),
            cmds.WithFlags(
                parameters.NewParameterDefinition(
                    "ticket",
                    parameters.ParameterTypeString,
                    parameters.WithHelp("Ticket identifier (updates ticket index when --doc not provided)"),
                    parameters.WithDefault(""),
                ),
                parameters.NewParameterDefinition(
                    "doc",
                    parameters.ParameterTypeString,
                    parameters.WithHelp("Path to a specific document to update"),
                    parameters.WithDefault(""),
                ),
                parameters.NewParameterDefinition(
                    "files",
                    parameters.ParameterTypeStringList,
                    parameters.WithHelp("Comma-separated list of files to add to RelatedFiles"),
                    parameters.WithDefault([]string{}),
                ),
                parameters.NewParameterDefinition(
                    "remove-files",
                    parameters.ParameterTypeStringList,
                    parameters.WithHelp("Comma-separated list of files to remove from RelatedFiles"),
                    parameters.WithDefault([]string{}),
                ),
                parameters.NewParameterDefinition(
                    "suggest",
                    parameters.ParameterTypeBool,
                    parameters.WithHelp("Suggest related files using heuristics (git + ripgrep + existing docs)"),
                    parameters.WithDefault(false),
                ),
                parameters.NewParameterDefinition(
                    "apply-suggestions",
                    parameters.ParameterTypeBool,
                    parameters.WithHelp("Apply suggested files to the target document (requires --suggest)"),
                    parameters.WithDefault(false),
                ),
                parameters.NewParameterDefinition(
                    "query",
                    parameters.ParameterTypeString,
                    parameters.WithHelp("Optional query to seed suggestions (e.g., a keyword)"),
                    parameters.WithDefault(""),
                ),
                parameters.NewParameterDefinition(
                    "topics",
                    parameters.ParameterTypeStringList,
                    parameters.WithHelp("Topics to seed suggestions (comma-separated)"),
                    parameters.WithDefault([]string{}),
                ),
                parameters.NewParameterDefinition(
                    "root",
                    parameters.ParameterTypeString,
                    parameters.WithHelp("Root directory for docs"),
                    parameters.WithDefault("ttmp"),
                ),
            ),
        ),
    }, nil
}

func (c *RelateCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    settings := &RelateSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to parse settings: %w", err)
    }

    // Resolve target document path
    var targetDocPath string
    var ticketDir string
    var err error

    if settings.Doc != "" {
        targetDocPath = settings.Doc
    } else {
        if settings.Ticket == "" {
            return fmt.Errorf("must specify either --doc or --ticket")
        }
        ticketDir, err = findTicketDirectory(settings.Root, settings.Ticket)
        if err != nil {
            return fmt.Errorf("failed to find ticket directory: %w", err)
        }
        targetDocPath = filepath.Join(ticketDir, "index.md")
    }

    // Optional: collect suggestions
    type reasonSet map[string]bool
    suggestions := map[string]reasonSet{}
    if settings.Suggest {
        // Determine search root: ticket dir if available else repo root
        searchRoot := settings.Root
        if ticketDir == "" && settings.Ticket != "" {
            ticketDir, _ = findTicketDirectory(settings.Root, settings.Ticket)
        }
        if ticketDir != "" {
            searchRoot = ticketDir
        }

        // From existing docs' RelatedFiles within the search root
        existing := map[string]bool{}
        _ = filepath.Walk(searchRoot, func(path string, info os.FileInfo, err error) error {
            if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
                return nil
            }
            doc, err := readDocumentFrontmatter(path)
            if err != nil {
                return nil
            }
            // If topics provided, filter
            if len(settings.Topics) > 0 {
                match := false
                for _, ft := range settings.Topics {
                    for _, dt := range doc.Topics {
                        if strings.EqualFold(strings.TrimSpace(ft), strings.TrimSpace(dt)) {
                            match = true
                            break
                        }
                    }
                    if match {
                        break
                    }
                }
                if !match {
                    return nil
                }
            }
            for _, rf := range doc.RelatedFiles {
                if rf != "" {
                    existing[rf] = true
                }
            }
            return nil
        })

        for f := range existing {
            if _, ok := suggestions[f]; !ok {
                suggestions[f] = reasonSet{}
            }
            suggestions[f]["referenced by documents"] = true
        }

        // From git history
        terms := []string{}
        if settings.Query != "" {
            terms = append(terms, settings.Query)
        }
        terms = append(terms, settings.Topics...)
        if files, err := suggestFilesFromGit(searchRoot, terms); err == nil {
            for _, f := range files {
                if _, ok := suggestions[f]; !ok {
                    suggestions[f] = reasonSet{}
                }
                suggestions[f]["recent commit activity"] = true
            }
        }

        // From ripgrep/grep
        if files, err := suggestFilesFromRipgrep(searchRoot, terms); err == nil {
            for _, f := range files {
                if _, ok := suggestions[f]; !ok {
                    suggestions[f] = reasonSet{}
                }
                suggestions[f][fmt.Sprintf("content match: %s", firstTerm(terms))] = true
            }
        }

        // From git status (modified, staged, untracked)
        if modified, staged, untracked, err := suggestFilesFromGitStatus(searchRoot); err == nil {
            for _, f := range modified {
                if _, ok := suggestions[f]; !ok {
                    suggestions[f] = reasonSet{}
                }
                suggestions[f]["working tree modified"] = true
            }
            for _, f := range staged {
                if _, ok := suggestions[f]; !ok {
                    suggestions[f] = reasonSet{}
                }
                suggestions[f]["staged for commit"] = true
            }
            for _, f := range untracked {
                if _, ok := suggestions[f]; !ok {
                    suggestions[f] = reasonSet{}
                }
                suggestions[f]["untracked new file"] = true
            }
        }

        // Deduplicate suggestions
        var dedup []string
        for f := range suggestions {
            if f != "" {
                dedup = append(dedup, f)
            }
        }
        sort.Strings(dedup)

        // If we are not applying suggestions, just output them
        if !settings.ApplySuggestions {
            for _, f := range dedup {
                // join reasons
                reasons := make([]string, 0, len(suggestions[f]))
                for r := range suggestions[f] {
                    reasons = append(reasons, r)
                }
                sort.Strings(reasons)
                row := types.NewRow(
                    types.MRP("file", f),
                    types.MRP("source", "suggested"),
                    types.MRP("reason", strings.Join(reasons, "; ")),
                )
                if err := gp.AddRow(ctx, row); err != nil {
                    return err
                }
            }
            return nil
        }
    }

    // Read the target document
    doc, err := readDocumentFrontmatter(targetDocPath)
    if err != nil {
        return fmt.Errorf("failed to read document frontmatter: %w", err)
    }

    // Build sets for add/remove
    current := map[string]bool{}
    for _, f := range doc.RelatedFiles {
        if f != "" {
            current[f] = true
        }
    }

    // Apply removals
    removedCount := 0
    for _, rf := range settings.RemoveFiles {
        rf = strings.TrimSpace(rf)
        if rf == "" {
            continue
        }
        if current[rf] {
            delete(current, rf)
            removedCount++
        }
    }

    // Apply additions
    addedCount := 0
    for _, af := range settings.Files {
        af = strings.TrimSpace(af)
        if af == "" {
            continue
        }
        if !current[af] {
            current[af] = true
            addedCount++
        }
    }

    // Apply suggestions if requested
    if settings.Suggest && settings.ApplySuggestions {
        for f := range suggestions {
            sf := f
            if !current[sf] {
                current[sf] = true
                addedCount++
            }
        }
    }

    // Serialize back to slice
    out := make([]string, 0, len(current))
    for f := range current {
        out = append(out, f)
    }
    sort.Strings(out)
    doc.RelatedFiles = out

    // Preserve existing content: rewrite file with updated frontmatter
    // We need the content; use readDocumentWithContent
    _, content, err := readDocumentWithContent(targetDocPath)
    if err != nil {
        return fmt.Errorf("failed to read document content: %w", err)
    }
    if err := writeDocumentWithFrontmatter(targetDocPath, doc, content, true); err != nil {
        return fmt.Errorf("failed to write document: %w", err)
    }

    row := types.NewRow(
        types.MRP("doc", targetDocPath),
        types.MRP("added", addedCount),
        types.MRP("removed", removedCount),
        types.MRP("total", len(doc.RelatedFiles)),
        types.MRP("status", "updated"),
    )
    return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &RelateCommand{}


