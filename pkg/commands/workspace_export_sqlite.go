package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
)

// ExportSQLiteCommand exports a workspace index to a persistent sqlite file for debugging/sharing.
type ExportSQLiteCommand struct {
	*cmds.CommandDescription
}

type ExportSQLiteSettings struct {
	Root        string `glazed:"root"`
	Out         string `glazed:"out"`
	Force       bool   `glazed:"force"`
	IncludeBody bool   `glazed:"include-body"`
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
  # Root-level command (recommended)
  docmgr export-sqlite --out /tmp/docmgr-index.sqlite
  docmgr export-sqlite --out /tmp/docmgr-index.sqlite --force
  docmgr export-sqlite --out /tmp/docmgr-index.sqlite --include-body

  # Equivalent namespaced form
  docmgr workspace export-sqlite --out /tmp/docmgr-index.sqlite
`),
			cmds.WithFlags(
				fields.New(
					"root",
					fields.TypeString,
					fields.WithHelp("Root directory for docs"),
					fields.WithDefault("ttmp"),
				),
				fields.New(
					"out",
					fields.TypeString,
					fields.WithHelp("Output sqlite file path"),
					fields.WithDefault(""),
				),
				fields.New(
					"force",
					fields.TypeBool,
					fields.WithHelp("Overwrite output file if it already exists"),
					fields.WithDefault(false),
				),
				fields.New(
					"include-body",
					fields.TypeBool,
					fields.WithHelp("Store markdown bodies in the exported sqlite (larger file)"),
					fields.WithDefault(false),
				),
			),
		),
	}, nil
}

// Run implements cmds.BareCommand (classic/human mode only).
func (c *ExportSQLiteCommand) Run(ctx context.Context, parsedValues *values.Values) error {
	settings := &ExportSQLiteSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
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
