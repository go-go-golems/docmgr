package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/models"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"gopkg.in/yaml.v3"
)

// InitCommand initializes a documentation root (ttmp/) with vocabulary, templates, and guidelines
type InitCommand struct {
	*cmds.CommandDescription
}

// InitSettings holds the parameters for the root init command
type InitSettings struct {
	Root           string `glazed:"root"`
	Force          bool   `glazed:"force"`
	SeedVocabulary bool   `glazed:"seed-vocabulary"`
}

type InitResult struct {
	Root         string
	Vocabulary   string
	DocmgrIgnore string
	ConfigPath   string
	Status       string
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

  # Seed the vocabulary with common defaults
  docmgr init --root ttmp --seed-vocabulary
`),
			cmds.WithFlags(
				fields.New(
					"root",
					fields.TypeString,
					fields.WithHelp("Root directory for docs (defaults to 'ttmp' or .ttmp.yaml root)"),
					fields.WithDefault("ttmp"),
				),
				fields.New(
					"force",
					fields.TypeBool,
					fields.WithHelp("Overwrite existing template/guideline files if present"),
					fields.WithDefault(false),
				),
				fields.New(
					"seed-vocabulary",
					fields.TypeBool,
					fields.WithHelp("Seed a default vocabulary.yaml with common topics/docTypes/intent"),
					fields.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *InitCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp middlewares.Processor,
) error {
	settings := &InitSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.initializeWorkspace(settings)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("root", result.Root),
		types.MRP("vocabulary", result.Vocabulary),
		types.MRP("docmgrignore", result.DocmgrIgnore),
		types.MRP("status", result.Status),
	)
	if result.ConfigPath != "" {
		row.Set("config", result.ConfigPath)
	}
	return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &InitCommand{}
var _ cmds.BareCommand = &InitCommand{}

func (c *InitCommand) initializeWorkspace(settings *InitSettings) (*InitResult, error) {
	settings.Root = workspace.ResolveRoot(settings.Root)
	cfgPath, _ := workspace.FindTTMPConfigPath()
	absRoot := settings.Root
	if !filepath.IsAbs(absRoot) {
		if cwd, err := os.Getwd(); err == nil {
			absRoot = filepath.Join(cwd, absRoot)
		}
	}

	if err := os.MkdirAll(settings.Root, 0755); err != nil {
		return nil, fmt.Errorf("failed to create docs root %s: %w", settings.Root, err)
	}

	ignorePath := filepath.Join(settings.Root, ".docmgrignore")
	ignoreContent := "# Default ignores for docmgr\n.git/\n_templates/\n_guidelines/\n"
	if err := writeFileIfNotExists(ignorePath, []byte(ignoreContent), settings.Force); err != nil {
		return nil, fmt.Errorf("failed to write .docmgrignore: %w", err)
	}

	vocabFilePath := filepath.Join(settings.Root, "vocabulary.yaml")
	if err := writeFileIfNotExists(vocabFilePath, []byte("topics: []\ndocTypes: []\nintent: []\n"), settings.Force); err != nil {
		return nil, fmt.Errorf("failed to write vocabulary.yaml: %w", err)
	}

	if settings.SeedVocabulary {
		if err := seedDefaultVocabulary(); err != nil {
			return nil, fmt.Errorf("failed to seed vocabulary: %w", err)
		}
	}

	if err := scaffoldTemplatesAndGuidelines(settings.Root, settings.Force); err != nil {
		return nil, fmt.Errorf("failed to scaffold templates and guidelines: %w", err)
	}

	status := "initialized"
	configPath := cfgPath
	repoRoot, err := workspace.FindRepositoryRoot()
	if err == nil {
		configPath = filepath.Join(repoRoot, ".ttmp.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			relRoot, err := filepath.Rel(repoRoot, absRoot)
			if err != nil {
				relRoot = settings.Root
			}
			relVocab, err := filepath.Rel(repoRoot, vocabFilePath)
			if err != nil {
				relVocab = filepath.Join(relRoot, "vocabulary.yaml")
			}

			cfg := workspace.WorkspaceConfig{
				Root:       relRoot,
				Vocabulary: relVocab,
			}

			data, err := yaml.Marshal(&cfg)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal config: %w", err)
			}

			if err := os.WriteFile(configPath, data, 0644); err != nil {
				return nil, fmt.Errorf("failed to write .ttmp.yaml: %w", err)
			}
		}
	}

	if configPath == "" {
		status = "initialized"
	}

	return &InitResult{
		Root:         settings.Root,
		Vocabulary:   vocabFilePath,
		DocmgrIgnore: ignorePath,
		ConfigPath:   configPath,
		Status:       status,
	}, nil
}

func (c *InitCommand) Run(
	ctx context.Context,
	parsedValues *values.Values,
) error {
	settings := &InitSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.initializeWorkspace(settings)
	if err != nil {
		return err
	}

	fmt.Printf("Docs root initialized at %s\n", result.Root)
	fmt.Printf("- Vocabulary: %s\n", result.Vocabulary)
	fmt.Printf("- .docmgrignore: %s\n", result.DocmgrIgnore)
	if result.ConfigPath != "" {
		fmt.Printf("- Config: %s\n", result.ConfigPath)
	}
	fmt.Printf("- Status: %s\n", result.Status)

	return nil
}

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
	addItem(&vocab.DocTypes, "skill", "Skill documentation (what it's for and when to use it)")

	// Intent
	addItem(&vocab.Intent, "long-term", "Likely to persist")
	addItem(&vocab.Intent, "short-term", "Short-term documentation for active work")
	addItem(&vocab.Intent, "throwaway", "Temporary/experimental documentation")

	// Status
	addItem(&vocab.Status, "draft", "Initial draft state")
	addItem(&vocab.Status, "active", "Active work in progress")
	addItem(&vocab.Status, "review", "Ready for review")
	addItem(&vocab.Status, "complete", "Work completed")
	addItem(&vocab.Status, "archived", "Archived/completed work")

	// Persist
	repoRoot, err := workspace.FindRepositoryRoot()
	if err != nil {
		return err
	}
	if err := SaveVocabulary(vocab, repoRoot); err != nil {
		return err
	}
	return nil
}
