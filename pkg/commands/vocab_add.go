package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	vocab, err := LoadVocabulary()
	if err != nil {
		return fmt.Errorf("failed to load vocabulary: %w", err)
	}

	// Find repo root by looking for vocabulary file or creating doc/ directory
	repoRoot, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
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
	default:
		return fmt.Errorf("invalid category: %s (must be topics, docTypes, or intent)", category)
	}

	// Check if slug already exists
	for _, item := range *categoryItems {
		if item.Slug == newItem.Slug {
			return fmt.Errorf("slug '%s' already exists in category '%s'", newItem.Slug, category)
		}
	}

	// Add new item
	*categoryItems = append(*categoryItems, newItem)

	// Save vocabulary (path resolved via config or defaults)
	if err := SaveVocabulary(vocab, repoRoot); err != nil {
		return fmt.Errorf("failed to save vocabulary: %w", err)
	}

	row := types.NewRow(
		types.MRP("category", category),
		types.MRP("slug", newItem.Slug),
		types.MRP("description", newItem.Description),
		types.MRP("status", "added"),
	)

	return gp.AddRow(ctx, row)
}

// findRepoRoot finds the repository root by walking up from current directory
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// Check for common repo root indicators
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, "doc")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root, use current directory
			return dir, nil
		}
		dir = parent
	}
}

var _ cmds.GlazeCommand = &VocabAddCommand{}
