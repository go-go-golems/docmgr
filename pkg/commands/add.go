package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docmgr/docmgr/pkg/models"
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
	Ticket  string `glazed.parameter:"ticket"`
	DocType string `glazed.parameter:"doc-type"`
	Title   string `glazed.parameter:"title"`
	Root    string `glazed.parameter:"root"`
}

func NewAddCommand() (*AddCommand, error) {
	return &AddCommand{
		CommandDescription: cmds.NewCommandDescription(
			"add",
			cmds.WithShort("Add a new document to a workspace"),
			cmds.WithLong(`Creates a new document in the appropriate subdirectory of a workspace.

Example:
  docmgr add --ticket MEN-3475 --doc-type design-doc --title "Draft Architecture"
  docmgr add --ticket MEN-3475 --doc-type reference --title "API Contracts"
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
					parameters.WithHelp("Document type (design-doc, reference, playbook)"),
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
    settings.Root = ResolveRoot(settings.Root)

    // Find the ticket directory
	ticketDir, err := findTicketDirectory(settings.Root, settings.Ticket)
	if err != nil {
		return fmt.Errorf("failed to find ticket directory: %w", err)
	}

	// Determine subdirectory based on doc type
	var subdir string
	switch settings.DocType {
	case "design-doc":
		subdir = "design"
	case "reference":
		subdir = "reference"
	case "playbook":
		subdir = "playbooks"
	default:
		return fmt.Errorf("unknown document type: %s (use: design-doc, reference, playbook)", settings.DocType)
	}

	// Create filename from title
	slug := strings.ToLower(strings.ReplaceAll(settings.Title, " ", "-"))
	filename := fmt.Sprintf("%s.md", slug)
	docPath := filepath.Join(ticketDir, subdir, filename)

	// Check if file already exists
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
	doc := models.Document{
		Title:           settings.Title,
		Ticket:          settings.Ticket,
		Status:          ticketDoc.Status,
		Topics:          ticketDoc.Topics,
		DocType:         settings.DocType,
		Intent:          "long-term",
		Owners:          ticketDoc.Owners,
		RelatedFiles:    []string{},
		ExternalSources: []string{},
		Summary:         "",
		LastUpdated:     time.Now(),
	}

	content := fmt.Sprintf("# %s\n\n<!-- Add your content here -->\n", settings.Title)
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
