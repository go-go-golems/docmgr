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

// InitCommand initializes a documentation root (ttmp/) with vocabulary, templates, and guidelines
type InitCommand struct {
	*cmds.CommandDescription
}

// InitSettings holds the parameters for the root init command
type InitSettings struct {
	Root  string `glazed.parameter:"root"`
	Force bool   `glazed.parameter:"force"`
}

func NewInitCommand() (*InitCommand, error) {
	return &InitCommand{
		CommandDescription: cmds.NewCommandDescription(
			"init",
			cmds.WithShort("Initialize a docs root (ttmp/) with vocabulary, templates, and guidelines"),
			cmds.WithLong(`Sets up a documentation root directory with the standard scaffolding.

What this does:
- Creates the docs root if missing (defaults to 'ttmp' or .ttmp.yaml root)
- Creates an empty 'vocabulary.yaml' if missing
- Scaffolds '_templates/' and '_guidelines/' with default files (respecting existing ones unless --force)

Examples:
  # Initialize default root (ttmp) next to the nearest .ttmp.yaml or from CWD
  docmgr init

  # Initialize a specific root path
  docmgr init --root ttmp
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs (defaults to 'ttmp' or .ttmp.yaml root)"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"force",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Overwrite existing template/guideline files if present"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *InitCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &InitSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = ResolveRoot(settings.Root)

	// Create root directory
	if err := os.MkdirAll(settings.Root, 0755); err != nil {
		return fmt.Errorf("failed to create docs root %s: %w", settings.Root, err)
	}

	// Create .docmgrignore with sensible defaults
	ignorePath := filepath.Join(settings.Root, ".docmgrignore")
	ignoreContent := "# Default ignores for docmgr\n.git/\n_templates/\n_guidelines/\n"
	if err := writeFileIfNotExists(ignorePath, []byte(ignoreContent), settings.Force); err != nil {
		return fmt.Errorf("failed to write .docmgrignore: %w", err)
	}

	// Create vocabulary.yaml if missing (empty lists)
	vocabPath := filepath.Join(settings.Root, "vocabulary.yaml")
	if err := writeFileIfNotExists(vocabPath, []byte("topics: []\ndocTypes: []\nintent: []\n"), settings.Force); err != nil {
		return fmt.Errorf("failed to write vocabulary.yaml: %w", err)
	}

	// Scaffold _templates/ and _guidelines/
	if err := scaffoldTemplatesAndGuidelines(settings.Root, settings.Force); err != nil {
		return fmt.Errorf("failed to scaffold templates and guidelines: %w", err)
	}

	row := types.NewRow(
		types.MRP("root", settings.Root),
		types.MRP("vocabulary", vocabPath),
		types.MRP("docmgrignore", ignorePath),
		types.MRP("status", "initialized"),
	)
	return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &InitCommand{}
