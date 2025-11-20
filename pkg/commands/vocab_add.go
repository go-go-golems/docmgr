package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/models"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// VocabAddCommand adds a vocabulary entry
type VocabAddCommand struct {
	*cmds.CommandDescription
}

// VocabAddSettings holds the parameters for the vocab add command
type VocabAddSettings struct {
	Category    string `glazed.parameter:"category"`
	Slug        string `glazed.parameter:"slug"`
	Description string `glazed.parameter:"description"`
	Root        string `glazed.parameter:"root"`
}

type VocabAddResult struct {
	Category       string
	Slug           string
	Description    string
	VocabularyPath string
	Root           string
	ConfigPath     string
}

func NewVocabAddCommand() (*VocabAddCommand, error) {
	return &VocabAddCommand{
		CommandDescription: cmds.NewCommandDescription(
			"add",
			cmds.WithShort("Add a vocabulary entry"),
			cmds.WithLong(`Adds a new entry to the workspace vocabulary file.

The vocabulary path is resolved from .ttmp.yaml if configured via 'vocabulary'.
By default, it is '<root>/vocabulary.yaml' (root defaults to 'ttmp').

Example:
  docmgr vocab add --category topics --slug observability --description "Logging and metrics"
  docmgr vocab add --category docTypes --slug working-note --description "Free-form notes"
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"category",
					parameters.ParameterTypeString,
					parameters.WithHelp("Category (topics, docTypes, intent)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"slug",
					parameters.ParameterTypeString,
					parameters.WithHelp("Vocabulary slug (lowercase, no spaces)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"description",
					parameters.ParameterTypeString,
					parameters.WithHelp("Description of the vocabulary entry"),
					parameters.WithRequired(true),
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

func (c *VocabAddCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &VocabAddSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.addVocabularyEntry(settings)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("category", result.Category),
		types.MRP("slug", result.Slug),
		types.MRP("description", result.Description),
		types.MRP("vocabulary_path", result.VocabularyPath),
		types.MRP("status", "added"),
	)

	return gp.AddRow(ctx, row)
}

// findRepoRoot finds the repository root by walking up from current directory
// unified repo root detection moved to workspace.FindRepositoryRoot() in config.go

var _ cmds.GlazeCommand = &VocabAddCommand{}

func (c *VocabAddCommand) addVocabularyEntry(settings *VocabAddSettings) (*VocabAddResult, error) {
	vocab, err := LoadVocabulary()
	if err != nil {
		return nil, fmt.Errorf("failed to load vocabulary: %w", err)
	}

	repoRoot, err := workspace.FindRepositoryRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	cfgPath, _ := workspace.FindTTMPConfigPath()
	vocabPath, _ := workspace.ResolveVocabularyPath()
	root := workspace.ResolveRoot(settings.Root)
	absRoot := root
	if !filepath.IsAbs(absRoot) {
		if cwd, err := os.Getwd(); err == nil {
			absRoot = filepath.Join(cwd, absRoot)
		}
	}

	newItem := models.VocabItem{
		Slug:        strings.ToLower(settings.Slug),
		Description: settings.Description,
	}

	category := strings.ToLower(settings.Category)
	var categoryItems *[]models.VocabItem

	switch category {
	case "topics":
		categoryItems = &vocab.Topics
	case "doctypes", "doc-types":
		categoryItems = &vocab.DocTypes
	case "intent":
		categoryItems = &vocab.Intent
	case "status":
		categoryItems = &vocab.Status
	default:
		return nil, fmt.Errorf("invalid category: %s (must be topics, docTypes, intent, or status)", category)
	}

	for _, item := range *categoryItems {
		if item.Slug == newItem.Slug {
			return nil, fmt.Errorf("slug '%s' already exists in category '%s'", newItem.Slug, category)
		}
	}

	*categoryItems = append(*categoryItems, newItem)

	if err := SaveVocabulary(vocab, repoRoot); err != nil {
		return nil, fmt.Errorf("failed to save vocabulary: %w", err)
	}

	return &VocabAddResult{
		Category:       category,
		Slug:           newItem.Slug,
		Description:    newItem.Description,
		VocabularyPath: vocabPath,
		Root:           absRoot,
		ConfigPath:     cfgPath,
	}, nil
}

func (c *VocabAddCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &VocabAddSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.addVocabularyEntry(settings)
	if err != nil {
		return err
	}

	fmt.Printf("Docs root: `%s`\n", result.Root)
	if result.ConfigPath != "" {
		fmt.Printf("Config: `%s`\n", result.ConfigPath)
	}
	if result.VocabularyPath != "" {
		fmt.Printf("Vocabulary: `%s`\n", result.VocabularyPath)
	}
	fmt.Printf("\nVocabulary entry added:\n")
	fmt.Printf("- Category: %s\n", result.Category)
	fmt.Printf("- Slug: %s\n", result.Slug)
	fmt.Printf("- Description: %s\n", result.Description)

	return nil
}

var _ cmds.BareCommand = &VocabAddCommand{}
