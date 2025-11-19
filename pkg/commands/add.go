package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/templates"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/models"
	"github.com/go-go-golems/docmgr/pkg/utils"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// AddCommand adds a new document to a workspace
type AddCommand struct {
	*cmds.CommandDescription
}

// AddSettings holds the parameters for the add command
type AddSettings struct {
	Ticket          string   `glazed.parameter:"ticket"`
	DocType         string   `glazed.parameter:"doc-type"`
	Title           string   `glazed.parameter:"title"`
	Root            string   `glazed.parameter:"root"`
	Topics          []string `glazed.parameter:"topics"`
	Owners          []string `glazed.parameter:"owners"`
	Status          string   `glazed.parameter:"status"`
	Intent          string   `glazed.parameter:"intent"`
	ExternalSources []string `glazed.parameter:"external-sources"`
	Summary         string   `glazed.parameter:"summary"`
	RelatedFiles    []string `glazed.parameter:"related-files"`
}

func NewAddCommand() (*AddCommand, error) {
	return &AddCommand{
		CommandDescription: cmds.NewCommandDescription(
			"add",
			cmds.WithShort("Add a new document to a workspace"),
			cmds.WithLong(`Creates a new document in the subdirectory named after its doc-type.

Example:
  docmgr add --ticket MEN-3475 --doc-type design-doc --title "Draft Architecture"
  docmgr add --ticket MEN-3475 --doc-type reference --title "API Contracts"
  docmgr add --ticket MEN-3475 --doc-type til       --title "Today I Learned — Hydration"
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Ticket identifier"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"doc-type",
					parameters.ParameterTypeString,
					parameters.WithHelp("Document type (per vocabulary; stored under <doc-type>/ subdir"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"title",
					parameters.ParameterTypeString,
					parameters.WithHelp("Document title"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"topics",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Comma-separated list of topics (overrides ticket defaults)"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"owners",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Comma-separated list of owners (overrides ticket defaults)"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"status",
					parameters.ParameterTypeString,
					parameters.WithHelp("Status (overrides ticket default)"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"intent",
					parameters.ParameterTypeString,
					parameters.WithHelp("Intent (overrides ticket default)"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"external-sources",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Comma-separated list of external sources (URLs)"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"summary",
					parameters.ParameterTypeString,
					parameters.WithHelp("Short summary for the document"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"related-files",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Comma-separated list of related files to seed frontmatter"),
					parameters.WithDefault([]string{}),
				),
			),
		),
	}, nil
}

func (c *AddCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &AddSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = workspace.ResolveRoot(settings.Root)
	// Echo resolved context prior to write
	cfgPath, _ := workspace.FindTTMPConfigPath()
	vocabPath, _ := workspace.ResolveVocabularyPath()
	absRoot := settings.Root
	if !filepath.IsAbs(absRoot) {
		if cwd, err := os.Getwd(); err == nil {
			absRoot = filepath.Join(cwd, absRoot)
		}
	}
	fmt.Printf("root=%s config=%s vocabulary=%s\n", absRoot, cfgPath, vocabPath)

	// Find the ticket directory
	ticketDir, err := findTicketDirectory(settings.Root, settings.Ticket)
	if err != nil {
		return fmt.Errorf("failed to find ticket directory: %w", err)
	}

	// Use doc-type slug directly as subdirectory name
	subdir := settings.DocType

	// Ensure target subdirectory exists
	targetDir := filepath.Join(ticketDir, subdir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	// Create filename from title with slugification and numeric prefix
	slug := utils.Slugify(settings.Title)
	docPath, err := buildPrefixedDocPath(targetDir, slug)
	if err != nil {
		return fmt.Errorf("failed to allocate prefixed filename: %w", err)
	}

	// Final guard: ensure file does not already exist (buildPrefixedDocPath should avoid collisions)
	if _, err := os.Stat(docPath); err == nil {
		return fmt.Errorf("document already exists: %s", docPath)
	}

	// Read ticket metadata
	indexPath := filepath.Join(ticketDir, "index.md")
	ticketDoc, err := readDocumentFrontmatter(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read ticket metadata: %w", err)
	}

	// Create document with frontmatter
	// Defaults from ticket, then override with flags if provided
	topics := ticketDoc.Topics
	if len(settings.Topics) > 0 {
		var ts []string
		for _, t := range settings.Topics {
			t = strings.TrimSpace(t)
			if t != "" {
				ts = append(ts, t)
			}
		}
		topics = ts
	}

	owners := ticketDoc.Owners
	if len(settings.Owners) > 0 {
		var os_ []string
		for _, o := range settings.Owners {
			o = strings.TrimSpace(o)
			if o != "" {
				os_ = append(os_, o)
			}
		}
		owners = os_
	}

	status := ticketDoc.Status
	if settings.Status != "" {
		status = settings.Status
	}

	intent := ticketDoc.Intent
	if intent == "" {
		intent = "long-term"
	}
	if settings.Intent != "" {
		intent = settings.Intent
	}

	external := []string{}
	if len(settings.ExternalSources) > 0 {
		for _, s := range settings.ExternalSources {
			s = strings.TrimSpace(s)
			if s != "" {
				external = append(external, s)
			}
		}
	}

	var rfs models.RelatedFiles
	if len(settings.RelatedFiles) > 0 {
		for _, f := range settings.RelatedFiles {
			f = strings.TrimSpace(f)
			if f != "" {
				rfs = append(rfs, models.RelatedFile{Path: f})
			}
		}
	}

	doc := models.Document{
		Title:           settings.Title,
		Ticket:          settings.Ticket,
		Status:          status,
		Topics:          topics,
		DocType:         settings.DocType,
		Intent:          intent,
		Owners:          owners,
		RelatedFiles:    rfs,
		ExternalSources: external,
		Summary:         settings.Summary,
		LastUpdated:     time.Now(),
	}

	// Try to load and render a template body
	content := ""
	if tpl, ok := templates.LoadTemplate(settings.Root, settings.DocType); ok {
		_, body := templates.ExtractFrontmatterAndBody(tpl)
		// Ensure title is set on doc for placeholders
		doc.Title = settings.Title
		content = templates.RenderTemplateBody(body, &doc)
	} else {
		content = fmt.Sprintf("# %s\n\n<!-- Add your content here -->\n", settings.Title)
	}

	if err := writeDocumentWithFrontmatter(docPath, &doc, content, false); err != nil {
		return fmt.Errorf("failed to write document: %w", err)
	}

	row := types.NewRow(
		types.MRP("ticket", settings.Ticket),
		types.MRP("doc_type", settings.DocType),
		types.MRP("title", settings.Title),
		types.MRP("path", docPath),
		types.MRP("status", "created"),
	)

	return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &AddCommand{}
