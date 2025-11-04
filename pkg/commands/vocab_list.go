package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// VocabListCommand lists vocabulary entries
type VocabListCommand struct {
	*cmds.CommandDescription
}

// VocabListSettings holds the parameters for the vocab list command
type VocabListSettings struct {
	Category string `glazed.parameter:"category"`
}

func NewVocabListCommand() (*VocabListCommand, error) {
	return &VocabListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List vocabulary entries"),
            cmds.WithLong(`Lists vocabulary entries from the workspace vocabulary file.

The vocabulary path is resolved from .ttmp.yaml if configured via 'vocabulary'.
By default, it is '<root>/vocabulary.yaml' (root defaults to 'ttmp').

Example:
  docmgr vocab list
  docmgr vocab list --category topics
  docmgr vocab list --category docTypes
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"category",
					parameters.ParameterTypeString,
					parameters.WithHelp("Category to list (topics, docTypes, intent). Leave empty to list all."),
					parameters.WithDefault(""),
				),
			),
		),
	}, nil
}

func (c *VocabListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &VocabListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	vocab, err := LoadVocabulary()
	if err != nil {
		return fmt.Errorf("failed to load vocabulary: %w", err)
	}

	category := strings.ToLower(settings.Category)

	if category == "" || category == "topics" {
		for _, item := range vocab.Topics {
			row := types.NewRow(
				types.MRP("category", "topics"),
				types.MRP("slug", item.Slug),
				types.MRP("description", item.Description),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	if category == "" || category == "doctypes" || category == "doc-types" {
		for _, item := range vocab.DocTypes {
			row := types.NewRow(
				types.MRP("category", "docTypes"),
				types.MRP("slug", item.Slug),
				types.MRP("description", item.Description),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	if category == "" || category == "intent" {
		for _, item := range vocab.Intent {
			row := types.NewRow(
				types.MRP("category", "intent"),
				types.MRP("slug", item.Slug),
				types.MRP("description", item.Description),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &VocabListCommand{}

// Implement BareCommand for human-friendly output
func (c *VocabListCommand) Run(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
) error {
    settings := &VocabListSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to parse settings: %w", err)
    }

    vocab, err := LoadVocabulary()
    if err != nil {
        return fmt.Errorf("failed to load vocabulary: %w", err)
    }

    category := strings.ToLower(settings.Category)

    if category == "" || category == "topics" {
        for _, item := range vocab.Topics {
            fmt.Printf("topics: %s — %s\n", item.Slug, item.Description)
        }
    }
    if category == "" || category == "doctypes" || category == "doc-types" {
        for _, item := range vocab.DocTypes {
            fmt.Printf("docTypes: %s — %s\n", item.Slug, item.Description)
        }
    }
    if category == "" || category == "intent" {
        for _, item := range vocab.Intent {
            fmt.Printf("intent: %s — %s\n", item.Slug, item.Description)
        }
    }
    return nil
}

var _ cmds.BareCommand = &VocabListCommand{}

