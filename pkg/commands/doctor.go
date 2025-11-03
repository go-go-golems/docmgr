package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// DoctorCommand validates document workspaces
type DoctorCommand struct {
	*cmds.CommandDescription
}

// DoctorSettings holds the parameters for the doctor command
type DoctorSettings struct {
	Root   string `glazed.parameter:"root"`
	Ticket string `glazed.parameter:"ticket"`
	All    bool   `glazed.parameter:"all"`
}

func NewDoctorCommand() (*DoctorCommand, error) {
	return &DoctorCommand{
		CommandDescription: cmds.NewCommandDescription(
			"doctor",
			cmds.WithShort("Validate document workspaces"),
			cmds.WithLong(`Checks document workspaces for issues like missing frontmatter,
invalid metadata, or broken structure.

Example:
  docmgr doctor --ticket MEN-3475
  docmgr doctor --all
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs"),
					parameters.WithDefault("docs"),
				),
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Check specific ticket"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"all",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Check all tickets"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *DoctorCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &DoctorSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	activePath := filepath.Join(settings.Root, "active")
	if _, err := os.Stat(activePath); os.IsNotExist(err) {
		return fmt.Errorf("active directory does not exist: %s", activePath)
	}

	entries, err := os.ReadDir(activePath)
	if err != nil {
		return fmt.Errorf("failed to read active directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		ticketPath := filepath.Join(activePath, entry.Name())
		indexPath := filepath.Join(ticketPath, "index.md")

		// Check if index.md exists
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			row := types.NewRow(
				types.MRP("ticket", entry.Name()),
				types.MRP("issue", "missing_index"),
				types.MRP("severity", "error"),
				types.MRP("message", "index.md not found"),
				types.MRP("path", ticketPath),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
			continue
		}

		// Try to parse frontmatter
		doc, err := readDocumentFrontmatter(indexPath)
		if err != nil {
			row := types.NewRow(
				types.MRP("ticket", entry.Name()),
				types.MRP("issue", "invalid_frontmatter"),
				types.MRP("severity", "error"),
				types.MRP("message", fmt.Sprintf("Failed to parse frontmatter: %v", err)),
				types.MRP("path", indexPath),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
			continue
		}

		// Filter by ticket if specified
		if settings.Ticket != "" && doc.Ticket != settings.Ticket {
			continue
		}

		// Validate required fields
		issues := []string{}
		if doc.Title == "" {
			issues = append(issues, "missing Title")
		}
		if doc.Ticket == "" {
			issues = append(issues, "missing Ticket")
		}
		if doc.Status == "" {
			issues = append(issues, "missing Status")
		}
		if len(doc.Topics) == 0 {
			issues = append(issues, "missing Topics")
		}

		if len(issues) > 0 {
			for _, issue := range issues {
				row := types.NewRow(
					types.MRP("ticket", doc.Ticket),
					types.MRP("issue", "missing_field"),
					types.MRP("severity", "warning"),
					types.MRP("message", issue),
					types.MRP("path", indexPath),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
			}
		} else {
			// No issues found
			row := types.NewRow(
				types.MRP("ticket", doc.Ticket),
				types.MRP("issue", "none"),
				types.MRP("severity", "ok"),
				types.MRP("message", "All checks passed"),
				types.MRP("path", ticketPath),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &DoctorCommand{}
