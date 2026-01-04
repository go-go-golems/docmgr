package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/documents"
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

// CreateTicketCommand creates a new ticket workspace under the docs root
type CreateTicketCommand struct {
	*cmds.CommandDescription
}

const DefaultTicketPathTemplate = "{{YYYY}}/{{MM}}/{{DD}}/{{TICKET}}--{{SLUG}}"

// CreateTicketSettings holds the parameters for the create-ticket command
type CreateTicketSettings struct {
	Ticket       string   `glazed.parameter:"ticket"`
	Title        string   `glazed.parameter:"title"`
	Topics       []string `glazed.parameter:"topics"`
	Root         string   `glazed.parameter:"root"`
	Force        bool     `glazed.parameter:"force"`
	PathTemplate string   `glazed.parameter:"path-template"`
}

type CreateTicketResult struct {
	Ticket       string
	Title        string
	Path         string
	Root         string
	Directories  []string
	FilesCreated []string
}

func NewCreateTicketCommand() (*CreateTicketCommand, error) {
	return &CreateTicketCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create-ticket",
			cmds.WithShort("Create a new ticket workspace under the docs root"),
			cmds.WithLong(`Creates a new ticket workspace directory with the standard structure.

Examples:
  # Create a ticket (default root + default date-based path template)
  docmgr ticket create-ticket --ticket MEN-3475 --title "Chat API cleanup" --topics chat,llm-workflow

  # Create with multiple topics
  docmgr ticket create-ticket --ticket MEN-8888 --title "WebSocket reconnection plan" --topics chat,backend,websocket

  # Create under a custom path template (relative to --root)
  docmgr ticket create-ticket --ticket MEN-9999 --title "Scratch ticket for experiments" \
    --root ttmp --path-template "examples/{{TICKET}}--{{SLUG}}"
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

	result, err := c.createTicketWorkspace(settings)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("ticket", result.Ticket),
		types.MRP("path", result.Path),
		types.MRP("title", result.Title),
		types.MRP("status", "created"),
	)

	return gp.AddRow(ctx, row)
}

func (c *CreateTicketCommand) createTicketWorkspace(settings *CreateTicketSettings) (*CreateTicketResult, error) {
	settings.Root = workspace.ResolveRoot(settings.Root)

	slug := utils.SlugifyTitleForTicket(settings.Ticket, settings.Title)
	now := time.Now()
	ticketPath, err := renderTicketPath(settings.Root, settings.PathTemplate, settings.Ticket, slug, settings.Title, now)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve ticket directory: %w", err)
	}

	dirList := []string{
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

	for _, dir := range dirList {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

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
	indexBody := fmt.Sprintf("# %s\n\nDocument workspace for %s.\n", settings.Title, settings.Ticket)
	if tpl, ok := templates.LoadTemplate(settings.Root, "index"); ok {
		_, body := templates.ExtractFrontmatterAndBody(tpl)
		doc.Title = settings.Title
		indexBody = templates.RenderTemplateBody(body, &doc)
	}
	if err := documents.WriteDocumentWithFrontmatter(indexPath, &doc, indexBody, settings.Force); err != nil {
		return nil, fmt.Errorf("failed to write index.md: %w", err)
	}

	files := []string{indexPath}

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

- Add documents: `+"`docmgr doc add --ticket %s --doc-type design-doc --title \"My Design\"`"+`
- Import sources: `+"`docmgr import file --ticket %s --file /path/to/doc.md`"+`
- Update metadata: `+"`docmgr meta update --ticket %s --field Status --value review`"+`
`, settings.Title, settings.Ticket, settings.Ticket, settings.Ticket, settings.Ticket)

	if err := writeFileIfNotExists(readmePath, []byte(readmeContent), settings.Force); err != nil {
		return nil, fmt.Errorf("failed to write README.md: %w", err)
	}
	files = append(files, readmePath)

	tasksPath := filepath.Join(ticketPath, "tasks.md")
	tasksContent := `# Tasks

## TODO

- [ ] Add tasks here

`
	if err := writeFileIfNotExists(tasksPath, []byte(tasksContent), settings.Force); err != nil {
		return nil, fmt.Errorf("failed to write tasks.md: %w", err)
	}
	files = append(files, tasksPath)

	changelogPath := filepath.Join(ticketPath, "changelog.md")
	changelogContent := fmt.Sprintf(`# Changelog

## %s

- Initial workspace created

`, now.Format("2006-01-02"))
	if err := writeFileIfNotExists(changelogPath, []byte(changelogContent), settings.Force); err != nil {
		return nil, fmt.Errorf("failed to write changelog.md: %w", err)
	}
	files = append(files, changelogPath)

	return &CreateTicketResult{
		Ticket:       settings.Ticket,
		Title:        settings.Title,
		Path:         ticketPath,
		Root:         settings.Root,
		Directories:  dirList,
		FilesCreated: files,
	}, nil
}

func (c *CreateTicketCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &CreateTicketSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.createTicketWorkspace(settings)
	if err != nil {
		return err
	}

	absRoot := workspace.ResolveRoot(settings.Root)
	if !filepath.IsAbs(absRoot) {
		if cwd, err := os.Getwd(); err == nil {
			absRoot = filepath.Join(cwd, absRoot)
		}
	}
	relPath := result.Path
	if rel, err := filepath.Rel(absRoot, result.Path); err == nil {
		relPath = filepath.ToSlash(rel)
	}

	fmt.Printf("Docs root: `%s`\n\n", absRoot)
	fmt.Printf("## Ticket Workspace Created\n\n")
	fmt.Printf("- Ticket: %s\n", result.Ticket)
	fmt.Printf("- Title: %s\n", result.Title)
	fmt.Printf("- Path: `%s`\n", relPath)
	fmt.Printf("- Directories: %d\n", len(result.Directories))
	fmt.Printf("- Files: %d\n", len(result.FilesCreated))

	fmt.Printf("\n### Created Directories\n")
	for _, d := range result.Directories {
		if rel, err := filepath.Rel(absRoot, d); err == nil {
			fmt.Printf("- `%s`\n", filepath.ToSlash(rel))
		} else {
			fmt.Printf("- `%s`\n", d)
		}
	}

	fmt.Printf("\n### Created Files\n")
	for _, f := range result.FilesCreated {
		if rel, err := filepath.Rel(absRoot, f); err == nil {
			fmt.Printf("- `%s`\n", filepath.ToSlash(rel))
		} else {
			fmt.Printf("- `%s`\n", f)
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &CreateTicketCommand{}
var _ cmds.BareCommand = &CreateTicketCommand{}
var _ cmds.GlazeCommand = &CreateTicketCommand{}
var _ cmds.BareCommand = &CreateTicketCommand{}

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

var _ cmds.GlazeCommand = &CreateTicketCommand{}
