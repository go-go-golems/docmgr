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

// GuidelinesCommand displays guidelines for document types
type GuidelinesCommand struct {
	*cmds.CommandDescription
}

// GuidelinesSettings holds the parameters for the guidelines command
type GuidelinesSettings struct {
	DocType string `glazed.parameter:"doc-type"`
	Root    string `glazed.parameter:"root"`
	List    bool   `glazed.parameter:"list"`
}

func NewGuidelinesCommand() (*GuidelinesCommand, error) {
	return &GuidelinesCommand{
		CommandDescription: cmds.NewCommandDescription(
			"guidelines",
			cmds.WithShort("Show guidelines for document types"),
			cmds.WithLong(`Displays guidelines for writing different types of documents.

Example:
  docmgr guidelines --doc-type design-doc
  docmgr guidelines --list
  docmgr guidelines --doc-type reference --root ttmp
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"doc-type",
					parameters.ParameterTypeString,
					parameters.WithHelp("Document type to show guidelines for"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"list",
					parameters.ParameterTypeBool,
					parameters.WithHelp("List all available document types"),
					parameters.WithDefault(false),
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

func (c *GuidelinesCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
    settings := &GuidelinesSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

    // Apply config root if present
    settings.Root = ResolveRoot(settings.Root)

	// List all available types
	if settings.List {
		docTypes := ListGuidelineTypes()
		for _, docType := range docTypes {
			row := types.NewRow(
				types.MRP("doc_type", docType),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
		return nil
	}

	// Show specific guideline
	if settings.DocType == "" {
		return fmt.Errorf("must specify --doc-type or use --list to see available types")
	}

	// Try to load from file system first
	guidelinePath := filepath.Join(settings.Root, "_guidelines", fmt.Sprintf("%s.md", settings.DocType))
	if _, err := os.Stat(guidelinePath); err == nil {
		// File exists, read and output it
		content, err := os.ReadFile(guidelinePath)
		if err != nil {
			return fmt.Errorf("failed to read guideline file: %w", err)
		}

		// Output as a single row with the content
		row := types.NewRow(
			types.MRP("doc_type", settings.DocType),
			types.MRP("content", string(content)),
		)
		return gp.AddRow(ctx, row)
	}

	// Fall back to embedded guidelines
	guideline, ok := GetGuideline(settings.DocType)
	if !ok {
		return fmt.Errorf("unknown document type: %s (use --list to see available types)", settings.DocType)
	}

	row := types.NewRow(
		types.MRP("doc_type", settings.DocType),
		types.MRP("content", guideline),
	)
	return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &GuidelinesCommand{}

