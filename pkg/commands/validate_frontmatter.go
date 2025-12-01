package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/models"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// ValidateFrontmatterCommand validates YAML frontmatter for a document.
type ValidateFrontmatterCommand struct {
	*cmds.CommandDescription
}

// ValidateFrontmatterSettings holds parameters for validation.
type ValidateFrontmatterSettings struct {
	Doc  string `glazed.parameter:"doc"`
	Root string `glazed.parameter:"root"`
}

func NewValidateFrontmatterCommand() (*ValidateFrontmatterCommand, error) {
	return &ValidateFrontmatterCommand{
		CommandDescription: cmds.NewCommandDescription(
			"frontmatter",
			cmds.WithShort("Validate YAML frontmatter for a document"),
			cmds.WithLong(`Validates YAML frontmatter for a single markdown file.

If parsing fails, the command surfaces a diagnostics taxonomy (line/column/snippet when available).
Use this before running doctor when iterating on frontmatter edits.

Examples:
  docmgr validate frontmatter --doc ttmp/2025/11/29/DOC-1234/index.md
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"doc",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to the markdown document to validate"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Docs root (used when doc is relative)"),
					parameters.WithDefault("ttmp"),
				),
			),
		),
	}, nil
}

func (c *ValidateFrontmatterCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ValidateFrontmatterSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	docPath := settings.Doc
	if !filepath.IsAbs(docPath) {
		root := workspace.ResolveRoot(settings.Root)
		docPath = filepath.Join(root, docPath)
	}

	doc, err := validateFrontmatterFile(docPath)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("doc", docPath),
		types.MRP("title", doc.Title),
		types.MRP("ticket", doc.Ticket),
		types.MRP("docType", doc.DocType),
		types.MRP("status", "ok"),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return fmt.Errorf("failed to emit validation result: %w", err)
	}
	return nil
}

func (c *ValidateFrontmatterCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &ValidateFrontmatterSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	docPath := settings.Doc
	if !filepath.IsAbs(docPath) {
		root := workspace.ResolveRoot(settings.Root)
		docPath = filepath.Join(root, docPath)
	}

	doc, err := validateFrontmatterFile(docPath)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Frontmatter OK: %s (Ticket=%s DocType=%s)\n", docPath, doc.Ticket, doc.DocType)
	return nil
}

func validateFrontmatterFile(path string) (*models.Document, error) {
	doc, _, err := documents.ReadDocumentWithFrontmatter(path)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

var _ cmds.GlazeCommand = &ValidateFrontmatterCommand{}
var _ cmds.BareCommand = &ValidateFrontmatterCommand{}
