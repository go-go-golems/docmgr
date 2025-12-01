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
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// DocMoveCommand moves a markdown document from one ticket to another and rewrites its Ticket field.
type DocMoveCommand struct {
	*cmds.CommandDescription
}

type DocMoveSettings struct {
	Root       string `glazed.parameter:"root"`
	Doc        string `glazed.parameter:"doc"`
	DestTicket string `glazed.parameter:"dest-ticket"`
	DestDir    string `glazed.parameter:"dest-dir"`
	Overwrite  bool   `glazed.parameter:"overwrite"`
}

type DocMoveResult struct {
	SourceTicket string
	DestTicket   string
	SourcePath   string
	DestPath     string
	CompletedAt  time.Time
}

func NewDocMoveCommand() (*DocMoveCommand, error) {
	return &DocMoveCommand{
		CommandDescription: cmds.NewCommandDescription(
			"move",
			cmds.WithShort("Move a document to another ticket"),
			cmds.WithLong(`Moves a markdown document between tickets and updates the Ticket field in frontmatter.

Typical uses:
  - Consolidate related docs under a single ticket
  - Recover stray docs that were created under the wrong ticket

Behavior:
  - Reads frontmatter to detect the source ticket
  - Resolves the destination ticket directory
  - Rewrites the Ticket field and writes the doc at the destination path
  - Removes the source file after a successful write

Use --dest-dir to override the relative subdirectory under the destination ticket.
By default the original relative path is preserved. Use --overwrite to replace an
existing file at the destination.`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"doc",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to the markdown document to move"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"dest-ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Destination ticket ID"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"dest-dir",
					parameters.ParameterTypeString,
					parameters.WithHelp("Relative directory inside the destination ticket (defaults to original subpath)"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Docs root (ttmp)"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"overwrite",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Overwrite destination file if it exists"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *DocMoveCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &DocMoveSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.applyMove(settings)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("source_ticket", result.SourceTicket),
		types.MRP("dest_ticket", result.DestTicket),
		types.MRP("source_path", result.SourcePath),
		types.MRP("dest_path", result.DestPath),
		types.MRP("status", "moved"),
		types.MRP("time", result.CompletedAt.Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func (c *DocMoveCommand) applyMove(settings *DocMoveSettings) (*DocMoveResult, error) {
	settings.Root = workspace.ResolveRoot(settings.Root)

	srcPath, err := resolveDocPath(settings.Root, settings.Doc)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(srcPath); err != nil {
		return nil, fmt.Errorf("source document not found: %s", srcPath)
	}

	doc, body, err := documents.ReadDocumentWithFrontmatter(srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read document frontmatter: %w", err)
	}
	srcTicket := strings.TrimSpace(doc.Ticket)
	if srcTicket == "" {
		return nil, fmt.Errorf("source document missing Ticket in frontmatter: %s", srcPath)
	}

	srcTicketDir, err := findTicketDirectory(settings.Root, srcTicket)
	if err != nil {
		return nil, fmt.Errorf("failed to find source ticket directory: %w", err)
	}
	destTicketDir, err := findTicketDirectory(settings.Root, settings.DestTicket)
	if err != nil {
		return nil, fmt.Errorf("failed to find destination ticket directory: %w", err)
	}

	relFromSrc, err := filepath.Rel(srcTicketDir, srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compute relative path: %w", err)
	}
	if strings.HasPrefix(relFromSrc, "..") {
		return nil, fmt.Errorf("document is not inside its ticket directory: %s", srcPath)
	}

	destRel := relFromSrc
	if settings.DestDir != "" {
		destRel = filepath.Join(filepath.Clean(settings.DestDir), filepath.Base(relFromSrc))
	}
	if strings.HasPrefix(destRel, "..") {
		return nil, fmt.Errorf("dest-dir must stay within the ticket: %s", settings.DestDir)
	}
	destPath := filepath.Join(destTicketDir, destRel)

	if !settings.Overwrite {
		if _, err := os.Stat(destPath); err == nil {
			return nil, fmt.Errorf("destination file already exists (use --overwrite to replace): %s", destPath)
		}
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Rewrite Ticket in frontmatter for the destination copy.
	doc.Ticket = settings.DestTicket
	if err := documents.WriteDocumentWithFrontmatter(destPath, doc, body, true); err != nil {
		return nil, fmt.Errorf("failed to write destination document: %w", err)
	}

	if err := os.Remove(srcPath); err != nil {
		return nil, fmt.Errorf("failed to remove source document: %w", err)
	}

	return &DocMoveResult{
		SourceTicket: srcTicket,
		DestTicket:   settings.DestTicket,
		SourcePath:   srcPath,
		DestPath:     destPath,
		CompletedAt:  time.Now(),
	}, nil
}

func resolveDocPath(root, doc string) (string, error) {
	if doc == "" {
		return "", fmt.Errorf("doc path is required")
	}
	if filepath.IsAbs(doc) {
		return doc, nil
	}
	candidate := filepath.Join(root, doc)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	abs, err := filepath.Abs(doc)
	if err != nil {
		return "", err
	}
	return abs, nil
}

var _ cmds.GlazeCommand = &DocMoveCommand{}
