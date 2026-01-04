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

type AddResult struct {
	Ticket         string
	DocType        string
	Title          string
	DocPath        string
	DocStatus      string
	Topics         []string
	Owners         []string
	Intent         string
	GuidelineText  string
	GuidelineTitle string
	Root           string
	ConfigPath     string
	VocabularyPath string
}

func NewAddCommand() (*AddCommand, error) {
	return &AddCommand{
		CommandDescription: cmds.NewCommandDescription(
			"add",
			cmds.WithShort("Add a new document to a workspace"),
			cmds.WithLong(`Creates a new document in the subdirectory named after its doc-type.

Examples:
  # Create a design doc in a ticket workspace
  docmgr doc add --ticket MEN-3475 --doc-type design-doc --title "Draft Architecture"

  # Override ticket defaults (topics/owners) for this one doc
  docmgr doc add --ticket MEN-3475 --doc-type reference --title "API Contracts" \
    --topics api,backend --owners manuel,alice

  # Seed multiple external sources + related files in frontmatter
  docmgr doc add --ticket MEN-3475 --doc-type reference --title "Trace Links" \
    --external-sources "https://example.com/spec,https://github.com/org/repo/issues/123" \
    --related-files "pkg/commands/add.go,pkg/commands/relate.go"
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

	result, err := c.createDocument(ctx, settings)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("ticket", result.Ticket),
		types.MRP("doc_type", result.DocType),
		types.MRP("title", result.Title),
		types.MRP("path", result.DocPath),
		types.MRP("status", "created"),
	)

	return gp.AddRow(ctx, row)
}

func (c *AddCommand) createDocument(ctx context.Context, settings *AddSettings) (*AddResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil context")
	}
	settings.Root = workspace.ResolveRoot(settings.Root)
	cfgPath, _ := workspace.FindTTMPConfigPath()
	vocabPath, _ := workspace.ResolveVocabularyPath()
	absRoot := settings.Root
	if !filepath.IsAbs(absRoot) {
		if cwd, err := os.Getwd(); err == nil {
			absRoot = filepath.Join(cwd, absRoot)
		}
	}

	// Ticket discovery is now Workspace+QueryDocs-backed (no legacy walkers).
	ticketDir, resolvedRoot, err := findTicketDirectoryViaWorkspace(ctx, settings.Root, settings.Ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to find ticket directory: %w", err)
	}
	// Keep Root consistent with Workspace resolution (affects templates/guidelines lookup + output paths).
	if strings.TrimSpace(resolvedRoot) != "" {
		settings.Root = resolvedRoot
		absRoot = resolvedRoot
	}

	subdir := settings.DocType
	targetDir := filepath.Join(ticketDir, subdir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	slug := utils.Slugify(settings.Title)
	docPath, err := buildPrefixedDocPath(targetDir, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate prefixed filename: %w", err)
	}
	if _, err := os.Stat(docPath); err == nil {
		return nil, fmt.Errorf("document already exists: %s", docPath)
	}

	indexPath := filepath.Join(ticketDir, "index.md")
	ticketDoc, err := readDocumentFrontmatter(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ticket metadata: %w", err)
	}

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

	content := ""
	if tpl, ok := templates.LoadTemplate(settings.Root, settings.DocType); ok {
		_, body := templates.ExtractFrontmatterAndBody(tpl)
		doc.Title = settings.Title
		content = templates.RenderTemplateBody(body, &doc)
	}
	// If no template found, content remains empty - document will have only frontmatter

	if err := documents.WriteDocumentWithFrontmatter(docPath, &doc, content, false); err != nil {
		return nil, fmt.Errorf("failed to write document: %w", err)
	}

	guidelineText := ""
	if guideline, ok := templates.LoadGuideline(settings.Root, settings.DocType); ok {
		guidelineText = guideline
	}
	// No guideline found - that's fine, just don't show any

	return &AddResult{
		Ticket:         settings.Ticket,
		DocType:        settings.DocType,
		Title:          settings.Title,
		DocPath:        docPath,
		DocStatus:      doc.Status,
		Topics:         doc.Topics,
		Owners:         doc.Owners,
		Intent:         doc.Intent,
		GuidelineText:  guidelineText,
		GuidelineTitle: settings.DocType,
		Root:           absRoot,
		ConfigPath:     cfgPath,
		VocabularyPath: vocabPath,
	}, nil
}

func (c *AddCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &AddSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.createDocument(ctx, settings)
	if err != nil {
		return err
	}

	relPath := result.DocPath
	if rel, err := filepath.Rel(result.Root, result.DocPath); err == nil {
		relPath = filepath.ToSlash(rel)
	}

	topics := "—"
	if len(result.Topics) > 0 {
		topics = strings.Join(result.Topics, ", ")
	}
	owners := "—"
	if len(result.Owners) > 0 {
		owners = strings.Join(result.Owners, ", ")
	}

	fmt.Printf("Docs root: `%s`\n", result.Root)
	if result.ConfigPath != "" {
		fmt.Printf("Config: `%s`\n", result.ConfigPath)
	}
	if result.VocabularyPath != "" {
		fmt.Printf("Vocabulary: `%s`\n", result.VocabularyPath)
	}
	fmt.Printf("\n## Document Created\n\n")
	fmt.Printf("- Ticket: %s\n", result.Ticket)
	fmt.Printf("- Doc type: %s\n", result.DocType)
	fmt.Printf("- Title: %s\n", result.Title)
	fmt.Printf("- Status: %s\n", result.DocStatus)
	fmt.Printf("- Intent: %s\n", result.Intent)
	fmt.Printf("- Topics: %s\n", topics)
	fmt.Printf("- Owners: %s\n", owners)
	fmt.Printf("- Path: `%s`\n", relPath)

	if result.GuidelineText != "" {
		fmt.Printf("\n### Guidelines for %s\n\n%s\n", result.GuidelineTitle, result.GuidelineText)
	} else {
		fmt.Printf("\n(No guidelines found for doc-type %s. Use `docmgr doc guidelines --list` to view available types.)\n", result.DocType)
	}

	return nil
}

var _ cmds.GlazeCommand = &AddCommand{}
var _ cmds.BareCommand = &AddCommand{}

func findTicketDirectoryViaWorkspace(ctx context.Context, rootOverride string, ticketID string) (string, string, error) {
	if ctx == nil {
		return "", "", fmt.Errorf("nil context")
	}
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return "", "", fmt.Errorf("empty ticket id")
	}

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: rootOverride})
	if err != nil {
		return "", "", fmt.Errorf("discover workspace: %w", err)
	}
	resolvedRoot := ws.Context().Root

	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return "", resolvedRoot, fmt.Errorf("init workspace index: %w", err)
	}

	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope:   workspace.Scope{Kind: workspace.ScopeTicket, TicketID: ticketID},
		Filters: workspace.DocFilters{DocType: "index"},
		Options: workspace.DocQueryOptions{
			IncludeErrors:       false,
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByPath,
		},
	})
	if err != nil {
		return "", resolvedRoot, fmt.Errorf("query ticket index doc: %w", err)
	}
	if len(res.Docs) == 0 {
		return "", resolvedRoot, fmt.Errorf("ticket not found: %s", ticketID)
	}
	if len(res.Docs) > 1 {
		return "", resolvedRoot, fmt.Errorf("ambiguous ticket index doc for %s (got %d)", ticketID, len(res.Docs))
	}

	p := strings.TrimSpace(res.Docs[0].Path)
	if p == "" {
		return "", resolvedRoot, fmt.Errorf("ticket index doc has empty path for %s", ticketID)
	}
	ticketDir := filepath.Clean(filepath.Dir(filepath.FromSlash(p)))
	if ticketDir == "" || ticketDir == "." {
		return "", resolvedRoot, fmt.Errorf("failed to derive ticket dir from %q", p)
	}

	return ticketDir, resolvedRoot, nil
}
