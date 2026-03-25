package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/utils"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// TicketMoveCommand moves an existing ticket directory to a new path based on the current path template.
type TicketMoveCommand struct {
	*cmds.CommandDescription
}

type TicketMoveSettings struct {
	Root         string `glazed:"root"`
	Ticket       string `glazed:"ticket"`
	PathTemplate string `glazed:"path-template"`
	Overwrite    bool   `glazed:"overwrite"`
}

type TicketMoveResult struct {
	Ticket      string
	SourcePath  string
	DestPath    string
	CompletedAt time.Time
}

func NewTicketMoveCommand() (*TicketMoveCommand, error) {
	return &TicketMoveCommand{
		CommandDescription: cmds.NewCommandDescription(
			"move",
			cmds.WithShort("Move a ticket directory to the current path template"),
			cmds.WithLong(`Moves an existing ticket workspace to the path derived from the current path template.

Useful for migrating legacy tickets created under an older layout (e.g., flat ttmp/)
into the date-based template. The command preserves ticket contents and updates LastUpdated.

Behavior:
  - Resolves source ticket directory
  - Renders destination path using the provided or configured path template
  - Moves the directory (rename) unless destination exists
  - Updates LastUpdated in index.md to now (best effort)

Examples:
  # Migrate a legacy ticket to the current date-based path template
  docmgr ticket move --ticket MEN-4242

  # Explicitly override the destination template
  docmgr ticket move --ticket MEN-4242 --path-template "legacy/{{TICKET}}--{{SLUG}}"

  # Overwrite destination if it already exists (use with care)
  docmgr ticket move --ticket MEN-4242 --overwrite
`),
			cmds.WithFlags(
				fields.New(
					"ticket",
					fields.TypeString,
					fields.WithHelp("Ticket identifier to move"),
					fields.WithRequired(true),
				),
				fields.New(
					"root",
					fields.TypeString,
					fields.WithHelp("Docs root (ttmp)"),
					fields.WithDefault("ttmp"),
				),
				fields.New(
					"path-template",
					fields.TypeString,
					fields.WithHelp("Path template to render destination (overrides config/default)"),
					fields.WithDefault(""),
				),
				fields.New(
					"overwrite",
					fields.TypeBool,
					fields.WithHelp("Overwrite destination if it exists (use with care)"),
					fields.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *TicketMoveCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	pl *values.Values,
	gp middlewares.Processor,
) error {
	settings := &TicketMoveSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.applyMove(ctx, settings)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("ticket", result.Ticket),
		types.MRP("source_path", result.SourcePath),
		types.MRP("dest_path", result.DestPath),
		types.MRP("status", "moved"),
		types.MRP("time", result.CompletedAt.Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func (c *TicketMoveCommand) applyMove(ctx context.Context, settings *TicketMoveSettings) (*TicketMoveResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil context")
	}
	settings.Root = workspace.ResolveRoot(settings.Root)

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return nil, fmt.Errorf("failed to discover workspace: %w", err)
	}
	settings.Root = ws.Context().Root
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return nil, fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	srcDir, err := resolveTicketDirViaWorkspace(ctx, ws, settings.Ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to find source ticket: %w", err)
	}

	pathTemplate := settings.PathTemplate
	if pathTemplate == "" {
		pathTemplate = DefaultTicketPathTemplate
	}

	indexPath := filepath.Join(srcDir, "index.md")
	srcDoc, _, _ := readDocumentWithContent(indexPath)
	title := strings.TrimSpace(settings.Ticket)
	if srcDoc != nil && strings.TrimSpace(srcDoc.Title) != "" {
		title = strings.TrimSpace(srcDoc.Title)
	}
	slug := utils.SlugifyTitleForTicket(settings.Ticket, title)

	// Use current time for new path template rendering.
	now := time.Now()
	destDir, err := renderTicketPath(settings.Root, pathTemplate, settings.Ticket, slug, title, now)
	if err != nil {
		return nil, fmt.Errorf("failed to render destination path: %w", err)
	}

	if strings.HasPrefix(destDir, srcDir) && destDir != srcDir {
		return nil, fmt.Errorf("destination cannot be nested inside source")
	}
	if !settings.Overwrite {
		if _, err := os.Stat(destDir); err == nil {
			return nil, fmt.Errorf("destination already exists (use --overwrite to replace): %s", destDir)
		}
	}

	if err := os.MkdirAll(filepath.Dir(destDir), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create destination parent: %w", err)
	}

	if err := os.Rename(srcDir, destDir); err != nil {
		return nil, fmt.Errorf("failed to move ticket directory: %w", err)
	}

	// Best effort: touch LastUpdated in index.md if present.
	destIndexPath := filepath.Join(destDir, "index.md")
	if doc, body, err := readDocumentWithContent(destIndexPath); err == nil && doc != nil {
		doc.LastUpdated = now
		_ = documents.WriteDocumentWithFrontmatter(destIndexPath, doc, body, true)
	}

	return &TicketMoveResult{
		Ticket:      settings.Ticket,
		SourcePath:  srcDir,
		DestPath:    destDir,
		CompletedAt: time.Now(),
	}, nil
}

var _ cmds.GlazeCommand = &TicketMoveCommand{}
