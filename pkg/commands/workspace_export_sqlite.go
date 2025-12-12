package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
)

// ExportSQLiteCommand exports a workspace index to a persistent sqlite file for debugging/sharing.
type ExportSQLiteCommand struct {
	*cmds.CommandDescription
}

type ExportSQLiteSettings struct {
	Root        string `glazed.parameter:"root"`
	Out         string `glazed.parameter:"out"`
	Force       bool   `glazed.parameter:"force"`
	IncludeBody bool   `glazed.parameter:"include-body"`
}

func NewExportSQLiteCommand() (*ExportSQLiteCommand, error) {
	return &ExportSQLiteCommand{
		CommandDescription: cmds.NewCommandDescription(
			"export-sqlite",
			cmds.WithShort("Export workspace index to a SQLite file (for debugging/sharing)"),
			cmds.WithLong(`Exports the in-memory Workspace index to a persistent SQLite database file.

The exported DB includes:
  - Workspace index tables (docs, doc_topics, related_files, ...)
  - README table containing docmgr embedded documentation (pkg/doc/*.md)

Examples:
  docmgr workspace export-sqlite --out /tmp/docmgr-index.sqlite
  docmgr workspace export-sqlite --out /tmp/docmgr-index.sqlite --force
  docmgr workspace export-sqlite --out /tmp/docmgr-index.sqlite --include-body
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"out",
					parameters.ParameterTypeString,
					parameters.WithHelp("Output sqlite file path"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"force",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Overwrite output file if it already exists"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"include-body",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Store markdown bodies in the exported sqlite (larger file)"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

// Run implements cmds.BareCommand (classic/human mode only).
func (c *ExportSQLiteCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &ExportSQLiteSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	if settings.Out == "" {
		return errors.New("--out is required")
	}

	// Apply config root if present.
	settings.Root = workspace.ResolveRoot(settings.Root)

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{
		RootOverride: settings.Root,
	})
	if err != nil {
		return err
	}

	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: settings.IncludeBody}); err != nil {
		return err
	}

	if err := ws.ExportIndexToSQLiteFile(ctx, workspace.ExportSQLiteOptions{
		OutPath: settings.Out,
		Force:   settings.Force,
	}); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "Exported workspace index to %s\n", settings.Out)
	return nil
}

var _ cmds.BareCommand = &ExportSQLiteCommand{}


