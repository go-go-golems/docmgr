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
	"gopkg.in/yaml.v3"
)

// CreateTicketCommand creates a new ticket workspace under the docs root
type CreateTicketCommand struct {
	*cmds.CommandDescription
}

const DefaultTicketPathTemplate = "{{YYYY}}/{{MM}}/{{DD}}/{{TICKET}}-{{SLUG}}"

// CreateTicketSettings holds the parameters for the create-ticket command
type CreateTicketSettings struct {
	Ticket       string   `glazed.parameter:"ticket"`
	Title        string   `glazed.parameter:"title"`
	Topics       []string `glazed.parameter:"topics"`
	Root         string   `glazed.parameter:"root"`
	Force        bool     `glazed.parameter:"force"`
	PathTemplate string   `glazed.parameter:"path-template"`
}

func NewCreateTicketCommand() (*CreateTicketCommand, error) {
	return &CreateTicketCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create-ticket",
			cmds.WithShort("Create a new ticket workspace under the docs root"),
			cmds.WithLong(`Creates a new ticket workspace directory with the standard structure.

Example:
  docmgr create-ticket --ticket MEN-3475 --title "Chat API cleanup" --topics chat,llm-workflow
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
					parameters.WithHelp("Ticket title"),
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
					parameters.WithHelp("Root directory for docs (defaults to 'ttmp' or .ttmp.yaml root)"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"force",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Force overwrite of existing files"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"path-template",
					parameters.ParameterTypeString,
					parameters.WithHelp("Template for ticket directory relative to root (placeholders: {{YYYY}}, {{MM}}, {{DD}}, {{DATE}}, {{TICKET}}, {{SLUG}}, {{TITLE}})"),
					parameters.WithDefault(DefaultTicketPathTemplate),
				),
			),
		),
	}, nil
}

func (c *CreateTicketCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &CreateTicketSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = workspace.ResolveRoot(settings.Root)

	// Create slug from title
	slug := utils.Slugify(settings.Title)
	now := time.Now()
	ticketPath, err := renderTicketPath(settings.Root, settings.PathTemplate, settings.Ticket, slug, settings.Title, now)
	if err != nil {
		return fmt.Errorf("failed to resolve ticket directory: %w", err)
	}

	// Create directory structure
	dirs := []string{
		ticketPath,
		filepath.Join(ticketPath, "design"),
		filepath.Join(ticketPath, "reference"),
		filepath.Join(ticketPath, "playbooks"),
		filepath.Join(ticketPath, "scripts"),
		filepath.Join(ticketPath, "sources"),
		filepath.Join(ticketPath, ".meta"),
		filepath.Join(ticketPath, "various"),
		filepath.Join(ticketPath, "archive"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create index.md with frontmatter
	// Load config defaults
	cfg, _ := workspace.LoadWorkspaceConfig()

	doc := models.Document{
		Title:   settings.Title,
		Ticket:  settings.Ticket,
		Status:  "active",
		Topics:  settings.Topics,
		DocType: "index",
		Intent: func() string {
			if cfg != nil && cfg.Defaults.Intent != "" {
				return cfg.Defaults.Intent
			}
			return "long-term"
		}(),
		Owners: func() []string {
			if cfg != nil && len(cfg.Defaults.Owners) > 0 {
				return cfg.Defaults.Owners
			}
			return []string{}
		}(),
		RelatedFiles:    models.RelatedFiles{},
		ExternalSources: []string{},
		Summary:         "",
		LastUpdated:     now,
	}

	indexPath := filepath.Join(ticketPath, "index.md")
	// Try to load index template body
	indexBody := fmt.Sprintf("# %s\n\nDocument workspace for %s.\n", settings.Title, settings.Ticket)
	if tpl, ok := templates.LoadTemplate(settings.Root, "index"); ok {
		_, body := templates.ExtractFrontmatterAndBody(tpl)
		// Ensure placeholders are populated from doc
		doc.Title = settings.Title
		indexBody = templates.RenderTemplateBody(body, &doc)
	}
	if err := writeDocumentWithFrontmatter(indexPath, &doc, indexBody, settings.Force); err != nil {
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
- **various/**: Scratch or meeting notes, working notes
- **archive/**: Optional space for deprecated or reference-only artifacts

## Getting Started

Use docmgr commands to manage this workspace:

- Add documents: `+"`docmgr add design-doc \"My Design\"`"+`
- Import sources: `+"`docmgr import file path/to/doc.md`"+`
- Update metadata: `+"`docmgr meta update --field Status --value review`"+`
`, settings.Title, settings.Ticket)

	if err := writeFileIfNotExists(readmePath, []byte(readmeContent), settings.Force); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	// Create tasks.md
	tasksPath := filepath.Join(ticketPath, "tasks.md")
	tasksContent := `# Tasks

## TODO

- [ ] Add tasks here

`
	if err := writeFileIfNotExists(tasksPath, []byte(tasksContent), settings.Force); err != nil {
		return fmt.Errorf("failed to write tasks.md: %w", err)
	}

	// Create changelog.md
	changelogPath := filepath.Join(ticketPath, "changelog.md")
	changelogContent := fmt.Sprintf(`# Changelog

## %s

- Initial workspace created

`, now.Format("2006-01-02"))
	if err := writeFileIfNotExists(changelogPath, []byte(changelogContent), settings.Force); err != nil {
		return fmt.Errorf("failed to write changelog.md: %w", err)
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

func renderTicketPath(root, templateStr, ticket, slug, title string, now time.Time) (string, error) {
	if templateStr == "" {
		templateStr = DefaultTicketPathTemplate
	}
	replacements := map[string]string{
		"{{YYYY}}":   now.Format("2006"),
		"{{MM}}":     now.Format("01"),
		"{{DD}}":     now.Format("02"),
		"{{DATE}}":   now.Format("2006-01-02"),
		"{{TICKET}}": ticket,
		"{{SLUG}}":   slug,
		"{{TITLE}}":  title,
	}
	relative := templateStr
	for placeholder, value := range replacements {
		relative = strings.ReplaceAll(relative, placeholder, value)
	}
	relative = filepath.Clean(relative)
	relative = strings.TrimPrefix(relative, string(os.PathSeparator))
	relative = strings.TrimPrefix(relative, "./")
	if relative == "" || relative == "." {
		return "", fmt.Errorf("path template resolved to empty path")
	}
	if strings.HasPrefix(relative, "..") {
		return "", fmt.Errorf("path template resolves outside root: %s", relative)
	}
	return filepath.Join(root, relative), nil
}

// writeDocumentWithFrontmatter writes a document with frontmatter to a file.
// If the file exists and force is false, it preserves existing frontmatter
// and content without overwriting.
func writeDocumentWithFrontmatter(path string, doc *models.Document, content string, force bool) error {
	// Check if file exists and we're not forcing
	if !force {
		if _, err := os.Stat(path); err == nil {
			// File exists, preserve it
			return nil
		}
	}

	// Write the document
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	// Write frontmatter
	if _, err := f.WriteString("---\n"); err != nil {
		return err
	}

	encoder := yaml.NewEncoder(f)
	if err := encoder.Encode(doc); err != nil {
		return err
	}
	if err := encoder.Close(); err != nil {
		return err
	}

	if _, err := f.WriteString("---\n\n"); err != nil {
		return err
	}

	// Write content
	if _, err := f.WriteString(content); err != nil {
		return err
	}

	return nil
}

var _ cmds.GlazeCommand = &CreateTicketCommand{}
