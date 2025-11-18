package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/go-go-golems/docmgr/pkg/models"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// RenameTicketCommand renames a ticket identifier and moves the workspace directory.
type RenameTicketCommand struct {
	*cmds.CommandDescription
}

type RenameTicketSettings struct {
	Root      string `glazed.parameter:"root"`
	Ticket    string `glazed.parameter:"ticket"`
	NewTicket string `glazed.parameter:"new-ticket"`
	DryRun    bool   `glazed.parameter:"dry-run"`
}

func NewRenameTicketCommand() (*RenameTicketCommand, error) {
	return &RenameTicketCommand{
		CommandDescription: cmds.NewCommandDescription(
			"rename-ticket",
			cmds.WithShort("Rename a ticket identifier and move its workspace directory"),
			cmds.WithLong(`Renames the ticket ID across all frontmatter files in the workspace and
moves the ticket directory from <oldTicket>-<slug> to <newTicket>-<slug>.

Examples:
  docmgr rename-ticket --ticket MEN-1234 --new-ticket MEN-5678
  docmgr rename-ticket --ticket DOCMGR-1 --new-ticket DOCMGR-101 --dry-run
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Current ticket identifier (e.g., MEN-1234)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"new-ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("New ticket identifier (e.g., MEN-5678)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"dry-run",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Show planned changes without modifying files"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *RenameTicketCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &RenameTicketSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Resolve workspace root from config/ENV/git
	settings.Root = ResolveRoot(settings.Root)

	if settings.Ticket == settings.NewTicket {
		return fmt.Errorf("new ticket is identical to current ticket")
	}

	// Locate current ticket directory
	oldDir, err := findTicketDirectory(settings.Root, settings.Ticket)
	if err != nil {
		return fmt.Errorf("failed to find ticket directory: %w", err)
	}

	// Compute new directory name: replace leading ticket prefix, preserve slug suffix if present
	base := filepath.Base(oldDir)
	remainder := ""
	if strings.HasPrefix(base, settings.Ticket) {
		remainder = strings.TrimPrefix(base, settings.Ticket) // includes leading '-' if present
	}
	newBase := settings.NewTicket + remainder
	newDir := filepath.Join(filepath.Dir(oldDir), newBase)

	verboseLog("rename-ticket: oldDir=%s newDir=%s", oldDir, newDir)

	if settings.DryRun {
		row := types.NewRow(
			types.MRP("ticket_old", settings.Ticket),
			types.MRP("ticket_new", settings.NewTicket),
			types.MRP("from", oldDir),
			types.MRP("to", newDir),
			types.MRP("status", "dry-run"),
		)
		return gp.AddRow(ctx, row)
	}

	// Update frontmatter Ticket fields across all markdown files that contain frontmatter
	updated, err := updateTicketFrontmatter(oldDir, settings.NewTicket)
	if err != nil {
		return fmt.Errorf("failed to update ticket in frontmatter: %w", err)
	}

	// Ensure target doesn't exist
	if _, err := os.Stat(newDir); err == nil {
		return fmt.Errorf("target directory already exists: %s", newDir)
	}

	// Perform directory rename
	if err := os.Rename(oldDir, newDir); err != nil {
		return fmt.Errorf("failed to rename directory %s -> %s: %w", oldDir, newDir, err)
	}

	row := types.NewRow(
		types.MRP("ticket_old", settings.Ticket),
		types.MRP("ticket_new", settings.NewTicket),
		types.MRP("from", oldDir),
		types.MRP("to", newDir),
		types.MRP("updated_docs", updated),
		types.MRP("status", "renamed"),
		types.MRP("time", time.Now().Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

// updateTicketFrontmatter walks a directory and updates the Ticket field in frontmatter-capable markdown files.
func updateTicketFrontmatter(root string, newTicket string) (int, error) {
	updated := 0
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		// Attempt to parse frontmatter; skip files without valid frontmatter
		f, openErr := os.Open(path)
		if openErr != nil {
			return nil
		}
		defer func() { _ = f.Close() }()

		var doc models.Document
		body, parseErr := frontmatter.Parse(f, &doc)
		if parseErr != nil {
			return nil
		}

		// Update ticket and last-updated
		doc.Ticket = newTicket
		doc.LastUpdated = time.Now()

		if err := writeDocumentWithFrontmatter(path, &doc, string(body), true); err != nil {
			return fmt.Errorf("writing updated frontmatter failed for %s: %w", path, err)
		}
		updated++
		return nil
	})
	return updated, err
}

var _ cmds.GlazeCommand = &RenameTicketCommand{}

// Implement BareCommand for human-friendly output
func (c *RenameTicketCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &RenameTicketSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	settings.Root = ResolveRoot(settings.Root)

	if settings.Ticket == settings.NewTicket {
		return fmt.Errorf("new ticket is identical to current ticket")
	}

	oldDir, err := findTicketDirectory(settings.Root, settings.Ticket)
	if err != nil {
		return fmt.Errorf("failed to find ticket directory: %w", err)
	}
	base := filepath.Base(oldDir)
	remainder := ""
	if strings.HasPrefix(base, settings.Ticket) {
		remainder = strings.TrimPrefix(base, settings.Ticket)
	}
	newBase := settings.NewTicket + remainder
	newDir := filepath.Join(filepath.Dir(oldDir), newBase)

	if settings.DryRun {
		fmt.Printf("Would rename ticket %s -> %s: %s -> %s\n", settings.Ticket, settings.NewTicket, oldDir, newDir)
		return nil
	}

	updated, err := updateTicketFrontmatter(oldDir, settings.NewTicket)
	if err != nil {
		return err
	}

	if _, err := os.Stat(newDir); err == nil {
		return fmt.Errorf("target directory already exists: %s", newDir)
	}
	if err := os.Rename(oldDir, newDir); err != nil {
		return fmt.Errorf("failed to rename directory %s -> %s: %w", oldDir, newDir, err)
	}

	fmt.Printf("Renamed ticket %s -> %s, updated %d docs\nfrom: %s\nto:   %s\n",
		settings.Ticket, settings.NewTicket, updated, oldDir, newDir)
	return nil
}

var _ cmds.BareCommand = &RenameTicketCommand{}
