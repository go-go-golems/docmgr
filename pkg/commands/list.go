package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/docmgr/docmgr/pkg/models"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// ListCommand lists document workspaces
type ListCommand struct {
	*cmds.CommandDescription
}

// ListSettings holds the parameters for the list command
type ListSettings struct {
	Root   string `glazed.parameter:"root"`
	Ticket string `glazed.parameter:"ticket"`
	Status string `glazed.parameter:"status"`
}

func NewListCommand() (*ListCommand, error) {
	return &ListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List document workspaces"),
			cmds.WithLong(`Lists all document workspaces in the root directory.

Example:
  docmgr list
  docmgr list --ticket MEN-3475
  docmgr list --status active
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

func (c *ListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

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

func readDocumentFrontmatter(path string) (*models.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var doc models.Document
	_, err = frontmatter.Parse(f, &doc)
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

var _ cmds.GlazeCommand = &ListCommand{}
