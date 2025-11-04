package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/docmgr/pkg/models"
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
    SeedVocabulary bool `glazed.parameter:"seed-vocabulary"`
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
                parameters.NewParameterDefinition(
                    "seed-vocabulary",
                    parameters.ParameterTypeBool,
                    parameters.WithHelp("Seed a default vocabulary.yaml with common topics/docTypes/intent"),
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
    // Echo resolved context prior to write
    cfgPath, _ := FindTTMPConfigPath()
    vocabPath, _ := ResolveVocabularyPath()
    absRoot := settings.Root
    if !filepath.IsAbs(absRoot) {
        if cwd, err := os.Getwd(); err == nil {
            absRoot = filepath.Join(cwd, absRoot)
        }
    }
    fmt.Printf("root=%s config=%s vocabulary=%s\n", absRoot, cfgPath, vocabPath)

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
    vocabFilePath := filepath.Join(settings.Root, "vocabulary.yaml")
    if err := writeFileIfNotExists(vocabFilePath, []byte("topics: []\ndocTypes: []\nintent: []\n"), settings.Force); err != nil {
		return fmt.Errorf("failed to write vocabulary.yaml: %w", err)
	}

    // Optionally seed vocabulary with defaults
    if settings.SeedVocabulary {
        if err := seedDefaultVocabulary(); err != nil {
            return fmt.Errorf("failed to seed vocabulary: %w", err)
        }
    }

	// Scaffold _templates/ and _guidelines/
	if err := scaffoldTemplatesAndGuidelines(settings.Root, settings.Force); err != nil {
		return fmt.Errorf("failed to scaffold templates and guidelines: %w", err)
	}

    row := types.NewRow(
		types.MRP("root", settings.Root),
        types.MRP("vocabulary", vocabFilePath),
		types.MRP("docmgrignore", ignorePath),
		types.MRP("status", "initialized"),
	)
	return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &InitCommand{}

// seedDefaultVocabulary populates vocabulary.yaml with a minimal default set if entries are missing.
func seedDefaultVocabulary() error {
    vocab, err := LoadVocabulary()
    if err != nil {
        return err
    }

    // Helpers to add if missing
    addItem := func(items *[]models.VocabItem, slug, desc string) {
        for _, it := range *items {
            if it.Slug == slug {
                return
            }
        }
        *items = append(*items, models.VocabItem{Slug: slug, Description: desc})
    }

    // Topics
    addItem(&vocab.Topics, "chat", "Chat backend and frontend surfaces")
    addItem(&vocab.Topics, "backend", "Backend services")
    addItem(&vocab.Topics, "websocket", "WebSocket lifecycle & events")

    // DocTypes
    addItem(&vocab.DocTypes, "design-doc", "Structured rationale and architecture notes")
    addItem(&vocab.DocTypes, "reference", "Reference docs and API contracts")
    addItem(&vocab.DocTypes, "playbook", "Operational procedures and QA/Smoke steps")
    addItem(&vocab.DocTypes, "index", "Ticket landing page")

    // Intent
    addItem(&vocab.Intent, "long-term", "Likely to persist")

    // Persist
    repoRoot, err := findRepoRoot()
    if err != nil {
        return err
    }
    if err := SaveVocabulary(vocab, repoRoot); err != nil {
        return err
    }
    return nil
}
