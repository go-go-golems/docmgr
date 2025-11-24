package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/templates"
	"github.com/go-go-golems/docmgr/internal/workspace"
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
	Category            string `glazed.parameter:"category"`
	Root                string `glazed.parameter:"root"`
	PrintTemplateSchema bool   `glazed.parameter:"print-template-schema"`
	SchemaFormat        string `glazed.parameter:"schema-format"`
}

func NewVocabListCommand() (*VocabListCommand, error) {
	return &VocabListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List vocabulary entries"),
			cmds.WithLong(`Lists vocabulary entries from the workspace vocabulary file.

The vocabulary path is resolved from .ttmp.yaml if configured via 'vocabulary'.
By default, it is '<root>/vocabulary.yaml' (root defaults to 'ttmp').

Columns:
  category,slug,description

Examples:
  # Human output
  docmgr vocab list
  docmgr vocab list --category topics
  docmgr vocab list --category docTypes

  # Scriptable (JSON)
  docmgr vocab list --with-glaze-output --output json
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"category",
					parameters.ParameterTypeString,
					parameters.WithHelp("Category to list (topics, docTypes, intent). Leave empty to list all."),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"print-template-schema",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Print template schema after output (human mode only)"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"schema-format",
					parameters.ParameterTypeString,
					parameters.WithHelp("Template schema output format: json|yaml"),
					parameters.WithDefault("json"),
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
	// Echo resolved context
	root := workspace.ResolveRoot(settings.Root)
	cfgPath, _ := workspace.FindTTMPConfigPath()
	vocabPath, _ := workspace.ResolveVocabularyPath()
	absRoot := root
	if !filepath.IsAbs(absRoot) {
		if cwd, err := os.Getwd(); err == nil {
			absRoot = filepath.Join(cwd, absRoot)
		}
	}
	fmt.Printf("root=%s config=%s vocabulary=%s\n", absRoot, cfgPath, vocabPath)

	vocab, err := LoadVocabulary()
	if err != nil {
		return fmt.Errorf("failed to load vocabulary: %w", err)
	}

	category := strings.ToLower(settings.Category)

	if category == "" || category == "topics" {
		for _, item := range vocab.Topics {
			row := types.NewRow(
				types.MRP(ColCategory, "topics"),
				types.MRP(ColSlug, item.Slug),
				types.MRP(ColDescription, item.Description),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	if category == "" || category == "doctypes" || category == "doc-types" {
		for _, item := range vocab.DocTypes {
			row := types.NewRow(
				types.MRP(ColCategory, "docTypes"),
				types.MRP(ColSlug, item.Slug),
				types.MRP(ColDescription, item.Description),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	if category == "" || category == "intent" {
		for _, item := range vocab.Intent {
			row := types.NewRow(
				types.MRP(ColCategory, "intent"),
				types.MRP(ColSlug, item.Slug),
				types.MRP(ColDescription, item.Description),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	if category == "" || category == "status" {
		for _, item := range vocab.Status {
			row := types.NewRow(
				types.MRP(ColCategory, "status"),
				types.MRP(ColSlug, item.Slug),
				types.MRP(ColDescription, item.Description),
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

	// Apply config root if present
	settings.Root = workspace.ResolveRoot(settings.Root)

	// If only printing template schema, skip all other processing and output
	if settings.PrintTemplateSchema {
		type VocabItem struct {
			Slug        string
			Description string
		}
		templateData := map[string]interface{}{
			"Category": "",
			"Topics": []VocabItem{
				{Slug: "", Description: ""},
			},
			"DocTypes": []VocabItem{
				{Slug: "", Description: ""},
			},
			"Intent": []VocabItem{
				{Slug: "", Description: ""},
			},
			"Status": []VocabItem{
				{Slug: "", Description: ""},
			},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	// Echo resolved context
	root := settings.Root
	cfgPath, _ := workspace.FindTTMPConfigPath()
	vocabPath, _ := workspace.ResolveVocabularyPath()
	absRoot := root
	if !filepath.IsAbs(absRoot) {
		if cwd, err := os.Getwd(); err == nil {
			absRoot = filepath.Join(cwd, absRoot)
		}
	}
	fmt.Printf("root=%s config=%s vocabulary=%s\n", absRoot, cfgPath, vocabPath)

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
	if category == "" || category == "status" {
		for _, item := range vocab.Status {
			fmt.Printf("status: %s — %s\n", item.Slug, item.Description)
		}
	}

	// Render postfix template if it exists
	// Build template data struct
	type VocabItem struct {
		Slug        string
		Description string
	}

	templateData := map[string]interface{}{
		"Category": settings.Category,
		"Topics":   make([]VocabItem, 0),
		"DocTypes": make([]VocabItem, 0),
		"Intent":   make([]VocabItem, 0),
		"Status":   make([]VocabItem, 0),
	}

	if category == "" || category == "topics" {
		topics := make([]VocabItem, 0, len(vocab.Topics))
		for _, item := range vocab.Topics {
			topics = append(topics, VocabItem{Slug: item.Slug, Description: item.Description})
		}
		templateData["Topics"] = topics
	}
	if category == "" || category == "doctypes" || category == "doc-types" {
		docTypes := make([]VocabItem, 0, len(vocab.DocTypes))
		for _, item := range vocab.DocTypes {
			docTypes = append(docTypes, VocabItem{Slug: item.Slug, Description: item.Description})
		}
		templateData["DocTypes"] = docTypes
	}
	if category == "" || category == "intent" {
		intent := make([]VocabItem, 0, len(vocab.Intent))
		for _, item := range vocab.Intent {
			intent = append(intent, VocabItem{Slug: item.Slug, Description: item.Description})
		}
		templateData["Intent"] = intent
	}
	if category == "" || category == "status" {
		status := make([]VocabItem, 0, len(vocab.Status))
		for _, item := range vocab.Status {
			status = append(status, VocabItem{Slug: item.Slug, Description: item.Description})
		}
		templateData["Status"] = status
	}

	// Try verb path: ["vocab", "list"]
	verbCandidates := [][]string{
		{"vocab", "list"},
	}
	settingsMap := map[string]interface{}{
		"root":     settings.Root,
		"category": settings.Category,
	}
	_ = templates.RenderVerbTemplate(verbCandidates, absRoot, settingsMap, templateData)

	return nil
}

var _ cmds.BareCommand = &VocabListCommand{}
