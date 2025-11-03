package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
					parameters.WithDefault("ttmp"),
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

		ticketPath := filepath.Join(settings.Root, entry.Name())
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

		// Track all issues found
		hasIssues := false

		// Check for unique index.md (should only be one per workspace)
		indexFiles := findIndexFiles(ticketPath)
		if len(indexFiles) > 1 {
			hasIssues = true
			row := types.NewRow(
				types.MRP("ticket", doc.Ticket),
				types.MRP("issue", "multiple_index"),
				types.MRP("severity", "warning"),
				types.MRP("message", fmt.Sprintf("Multiple index.md files found (%d), should be only one", len(indexFiles))),
				types.MRP("path", ticketPath),
				types.MRP("index_files", fmt.Sprintf("%v", indexFiles)),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}

		// Check for staleness (LastUpdated > 14 days)
		if !doc.LastUpdated.IsZero() {
			daysSinceUpdate := time.Since(doc.LastUpdated).Hours() / 24
			if daysSinceUpdate > 14 {
				hasIssues = true
				row := types.NewRow(
					types.MRP("ticket", doc.Ticket),
					types.MRP("issue", "stale"),
					types.MRP("severity", "warning"),
					types.MRP("message", fmt.Sprintf("LastUpdated is %.0f days old (threshold: 14 days)", daysSinceUpdate)),
					types.MRP("path", indexPath),
					types.MRP("last_updated", doc.LastUpdated.Format("2006-01-02")),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
			}
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
			hasIssues = true
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
		}

		// Only report "All checks passed" if there are truly no issues
		if !hasIssues {
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

// findIndexFiles recursively searches for all index.md files in a directory tree
func findIndexFiles(rootPath string) []string {
	var indexFiles []string
	
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}
		if !info.IsDir() && info.Name() == "index.md" {
			indexFiles = append(indexFiles, path)
		}
		return nil
	})
	
	if err != nil {
		// Return what we found even if there was an error
		return indexFiles
	}
	
	return indexFiles
}

var _ cmds.GlazeCommand = &DoctorCommand{}
