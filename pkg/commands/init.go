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
	"gopkg.in/yaml.v3"
)

// InitCommand creates a new document workspace
type InitCommand struct {
	*cmds.CommandDescription
}

// InitSettings holds the parameters for the init command
type InitSettings struct {
	Ticket string   `glazed.parameter:"ticket"`
	Title  string   `glazed.parameter:"title"`
	Topics []string `glazed.parameter:"topics"`
	Root   string   `glazed.parameter:"root"`
}

func NewInitCommand() (*InitCommand, error) {
	return &InitCommand{
		CommandDescription: cmds.NewCommandDescription(
			"init",
			cmds.WithShort("Initialize a new document workspace"),
			cmds.WithLong(`Creates a new document workspace with the standard directory structure.

Example:
  docmgr init MEN-3475 --title "Chat API cleanup" --topics chat,llm-workflow
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Ticket identifier (e.g., MEN-3475)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"title",
					parameters.ParameterTypeString,
					parameters.WithHelp("Document title"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"topics",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Comma-separated list of topics"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs"),
					parameters.WithDefault("docs"),
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

	// Create slug from title
	slug := strings.ToLower(strings.ReplaceAll(settings.Title, " ", "-"))
	dirName := fmt.Sprintf("%s-%s", settings.Ticket, slug)
	ticketPath := filepath.Join(settings.Root, "active", dirName)

	// Create directory structure
	dirs := []string{
		ticketPath,
		filepath.Join(ticketPath, "design"),
		filepath.Join(ticketPath, "reference"),
		filepath.Join(ticketPath, "playbooks"),
		filepath.Join(ticketPath, "scripts"),
		filepath.Join(ticketPath, "sources"),
		filepath.Join(ticketPath, ".meta"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create index.md with frontmatter
	doc := models.Document{
		Title:           settings.Title,
		Ticket:          settings.Ticket,
		Status:          "active",
		Topics:          settings.Topics,
		DocType:         "index",
		Intent:          "long-term",
		Owners:          []string{},
		RelatedFiles:    []string{},
		ExternalSources: []string{},
		Summary:         "",
		LastUpdated:     time.Now(),
	}

	indexPath := filepath.Join(ticketPath, "index.md")
	if err := writeDocumentWithFrontmatter(indexPath, &doc, fmt.Sprintf("# %s\n\nDocument workspace for %s.\n", settings.Title, settings.Ticket)); err != nil {
		return fmt.Errorf("failed to write index.md: %w", err)
	}

	// Create README.md
	readmePath := filepath.Join(ticketPath, "README.md")
	readmeContent := fmt.Sprintf(`# %s

This is the document workspace for ticket %s.

## Structure

- **design/**: Design documents and architecture notes
- **reference/**: Reference documentation and API contracts
- **playbooks/**: Operational playbooks and procedures
- **scripts/**: Utility scripts and automation
- **sources/**: External sources and imported documents

## Getting Started

Use docmgr commands to manage this workspace:

- Add documents: ` + "`docmgr add design-doc \"My Design\"`" + `
- Import sources: ` + "`docmgr import file path/to/doc.md`" + `
- Update metadata: ` + "`docmgr meta update --field Status --value review`" + `
`, settings.Title, settings.Ticket)

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	// Output result
	row := types.NewRow(
		types.MRP("ticket", settings.Ticket),
		types.MRP("path", ticketPath),
		types.MRP("title", settings.Title),
		types.MRP("status", "created"),
	)

	return gp.AddRow(ctx, row)
}

func writeDocumentWithFrontmatter(path string, doc *models.Document, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write frontmatter
	if _, err := f.WriteString("---\n"); err != nil {
		return err
	}

	encoder := yaml.NewEncoder(f)
	if err := encoder.Encode(doc); err != nil {
		return err
	}
	encoder.Close()

	if _, err := f.WriteString("---\n\n"); err != nil {
		return err
	}

	// Write content
	if _, err := f.WriteString(content); err != nil {
		return err
	}

	return nil
}

var _ cmds.GlazeCommand = &InitCommand{}
