package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// ListTicketsCommand lists ticket workspaces
type ListTicketsCommand struct {
	*cmds.CommandDescription
}

// ListTicketsSettings holds the parameters for the list tickets command
type ListTicketsSettings struct {
	Root   string `glazed.parameter:"root"`
	Ticket string `glazed.parameter:"ticket"`
	Status string `glazed.parameter:"status"`
}

func NewListTicketsCommand() (*ListTicketsCommand, error) {
	return &ListTicketsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"tickets",
			cmds.WithShort("List ticket workspaces"),
			cmds.WithLong(`Lists all ticket workspaces in the root directory.

Example:
  docmgr list tickets
  docmgr list tickets --ticket MEN-3475
  docmgr list tickets --status active
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
					"status",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by status"),
					parameters.WithDefault(""),
				),
			),
		),
	}, nil
}

func (c *ListTicketsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ListTicketsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

    // Apply config root if present
    settings.Root = ResolveRoot(settings.Root)

    if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	entries, err := os.ReadDir(settings.Root)
	if err != nil {
		return fmt.Errorf("failed to read root directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
        // Skip scaffolding directories
        if strings.HasPrefix(entry.Name(), "_") {
            continue
        }

		indexPath := filepath.Join(settings.Root, entry.Name(), "index.md")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			continue
		}

		doc, err := readDocumentFrontmatter(indexPath)
		if err != nil {
			// Skip documents with invalid frontmatter
			continue
		}

		// Apply filters
		if settings.Ticket != "" && !strings.Contains(doc.Ticket, settings.Ticket) {
			continue
		}
		if settings.Status != "" && doc.Status != settings.Status {
			continue
		}

		row := types.NewRow(
			types.MRP("ticket", doc.Ticket),
			types.MRP("title", doc.Title),
			types.MRP("status", doc.Status),
			types.MRP("topics", strings.Join(doc.Topics, ", ")),
			types.MRP("path", filepath.Join(settings.Root, entry.Name())),
			types.MRP("last_updated", doc.LastUpdated.Format("2006-01-02")),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &ListTicketsCommand{}

// Implement BareCommand for human-friendly output
func (c *ListTicketsCommand) Run(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
) error {
    settings := &ListTicketsSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to parse settings: %w", err)
    }

    // Apply config root if present
    settings.Root = ResolveRoot(settings.Root)

    if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
        return fmt.Errorf("root directory does not exist: %s", settings.Root)
    }

    entries, err := os.ReadDir(settings.Root)
    if err != nil {
        return fmt.Errorf("failed to read root directory: %w", err)
    }

    for _, entry := range entries {
        if !entry.IsDir() { continue }
        if strings.HasPrefix(entry.Name(), "_") { continue }
        indexPath := filepath.Join(settings.Root, entry.Name(), "index.md")
        if _, err := os.Stat(indexPath); os.IsNotExist(err) { continue }
        doc, err := readDocumentFrontmatter(indexPath)
        if err != nil { continue }
        if settings.Ticket != "" && !strings.Contains(doc.Ticket, settings.Ticket) { continue }
        if settings.Status != "" && doc.Status != settings.Status { continue }
        fmt.Printf("%s ‘%s’ status=%s topics=%s updated=%s path=%s\n",
            doc.Ticket,
            doc.Title,
            doc.Status,
            strings.Join(doc.Topics, ", "),
            doc.LastUpdated.Format("2006-01-02"),
            filepath.Join(settings.Root, entry.Name()),
        )
    }
    return nil
}

var _ cmds.BareCommand = &ListTicketsCommand{}

