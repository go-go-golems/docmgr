package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/searchsvc"
	"github.com/go-go-golems/docmgr/internal/templates"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgr"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// SearchCommand searches documents by content and metadata
type SearchCommand struct {
	*cmds.CommandDescription
}

// SearchSettings holds the parameters for the search command
type SearchSettings struct {
	Query               string   `glazed:"query"`
	Ticket              string   `glazed:"ticket"`
	Topics              []string `glazed:"topics"`
	DocType             string   `glazed:"doc-type"`
	Status              string   `glazed:"status"`
	OrderBy             string   `glazed:"order-by"`
	Files               bool     `glazed:"files"`
	File                string   `glazed:"file"`
	Dir                 string   `glazed:"dir"`
	ExternalSource      string   `glazed:"external-source"`
	Since               string   `glazed:"since"`
	Until               string   `glazed:"until"`
	CreatedSince        string   `glazed:"created-since"`
	UpdatedSince        string   `glazed:"updated-since"`
	Root                string   `glazed:"root"`
	PrintTemplateSchema bool     `glazed:"print-template-schema"`
	SchemaFormat        string   `glazed:"schema-format"`
}

func NewSearchCommand() (*SearchCommand, error) {
	return &SearchCommand{
		CommandDescription: cmds.NewCommandDescription(
			"search",
			cmds.WithShort("Search documents by content and metadata"),
			cmds.WithLong(`Search documents by full-text content and metadata filters.

The search command supports:
- Full-text search across document content
- Metadata filtering (ticket, topics, doc-type, status)
- File suggestions using heuristics (--files flag)
- Reverse lookup: find docs for a file/directory (--file, --dir)
- External source search (--external-source)
- Date range filtering (--since, --until, --created-since, --updated-since)

Examples:
  # Full-text search
  docmgr search --query "authentication"

  # Filter by metadata
  docmgr search --query "database" --topics backend --doc-type design-doc
  docmgr search --query "database" --topics backend,storage --doc-type design-doc --status review

  # Reverse lookup: find docs that reference a file or directory
  docmgr search --file pkg/commands/add.go
  docmgr search --dir pkg/commands/

  # Time-based filters (relative or absolute)
  docmgr search --updated-since "2 weeks ago"
  docmgr search --created-since "2025-01-01" --until "2025-01-31"
`),
			cmds.WithFlags(
				fields.New(
					"query",
					fields.TypeString,
					fields.WithHelp("Search query text (searches document content)"),
					fields.WithDefault(""),
				),
				fields.New(
					"ticket",
					fields.TypeString,
					fields.WithHelp("Filter by ticket identifier"),
					fields.WithDefault(""),
				),
				fields.New(
					"topics",
					fields.TypeStringList,
					fields.WithHelp("Filter by topics (comma-separated, matches any)"),
					fields.WithDefault([]string{}),
				),
				fields.New(
					"doc-type",
					fields.TypeString,
					fields.WithHelp("Filter by document type"),
					fields.WithDefault(""),
				),
				fields.New(
					"status",
					fields.TypeString,
					fields.WithHelp("Filter by status"),
					fields.WithDefault(""),
				),
				fields.New(
					"order-by",
					fields.TypeString,
					fields.WithHelp("Order results by: path|last_updated|rank"),
					fields.WithDefault("path"),
				),
				fields.New(
					"files",
					fields.TypeBool,
					fields.WithHelp("Suggest related files using heuristics (git + ripgrep)"),
					fields.WithDefault(false),
				),
				fields.New(
					"file",
					fields.TypeString,
					fields.WithHelp("Find documents that reference this file path"),
					fields.WithDefault(""),
				),
				fields.New(
					"dir",
					fields.TypeString,
					fields.WithHelp("Find documents in this directory or referencing files in it"),
					fields.WithDefault(""),
				),
				fields.New(
					"external-source",
					fields.TypeString,
					fields.WithHelp("Find documents that reference this external source URL"),
					fields.WithDefault(""),
				),
				fields.New(
					"since",
					fields.TypeString,
					fields.WithHelp("Find documents updated since this date (relative: '2 weeks ago', 'last month', or absolute: '2025-01-01')"),
					fields.WithDefault(""),
				),
				fields.New(
					"until",
					fields.TypeString,
					fields.WithHelp("Find documents updated until this date (relative: '2 weeks ago', 'last month', or absolute: '2025-01-01')"),
					fields.WithDefault(""),
				),
				fields.New(
					"created-since",
					fields.TypeString,
					fields.WithHelp("Find documents created since this date (relative: '2 weeks ago', 'last month', or absolute: '2025-01-01')"),
					fields.WithDefault(""),
				),
				fields.New(
					"updated-since",
					fields.TypeString,
					fields.WithHelp("Find documents updated since this date (relative: '2 weeks ago', 'last month', or absolute: '2025-01-01')"),
					fields.WithDefault(""),
				),
				fields.New(
					"root",
					fields.TypeString,
					fields.WithHelp("Root directory for docs"),
					fields.WithDefault("ttmp"),
				),
				fields.New(
					"print-template-schema",
					fields.TypeBool,
					fields.WithHelp("Print template schema after output (human mode only)"),
					fields.WithDefault(false),
				),
				fields.New(
					"schema-format",
					fields.TypeString,
					fields.WithHelp("Template schema output format: json|yaml"),
					fields.WithDefault("json"),
				),
			),
		),
	}, nil
}

func (c *SearchCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp middlewares.Processor,
) error {
	settings := &SearchSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = workspace.ResolveRoot(settings.Root)

	// If only printing template schema, skip all other processing and output
	if settings.PrintTemplateSchema {
		type SearchResult struct {
			Ticket  string
			Title   string
			DocType string
			Status  string
			Topics  []string
			Path    string
			Snippet string
		}
		templateData := map[string]interface{}{
			"Query":        "",
			"TotalResults": 0,
			"Results": []SearchResult{
				{
					Ticket:  "",
					Title:   "",
					DocType: "",
					Status:  "",
					Topics:  []string{},
					Path:    "",
					Snippet: "",
				},
			},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	// If --files flag is set, suggest files instead of searching documents
	if settings.Files {
		return c.suggestFiles(ctx, settings, gp)
	}

	// Validate that we have at least a query or some filters
	if settings.Query == "" && settings.Ticket == "" && len(settings.Topics) == 0 && settings.DocType == "" && settings.Status == "" &&
		settings.File == "" && settings.Dir == "" && settings.ExternalSource == "" &&
		settings.Since == "" && settings.Until == "" && settings.CreatedSince == "" && settings.UpdatedSince == "" {
		return fmt.Errorf("must provide at least a query or filter")
	}

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	fileQueryRaw := strings.TrimSpace(settings.File)

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return fmt.Errorf("failed to discover workspace: %w", err)
	}
	settings.Root = ws.Context().Root
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: true}); err != nil {
		return fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	orderBy := workspace.OrderBy(strings.TrimSpace(settings.OrderBy))
	if orderBy == "" {
		orderBy = workspace.OrderByPath
	}

	resp, err := searchsvc.SearchDocs(ctx, ws, searchsvc.SearchQuery{
		TextQuery:           strings.TrimSpace(settings.Query),
		Ticket:              strings.TrimSpace(settings.Ticket),
		Topics:              settings.Topics,
		DocType:             strings.TrimSpace(settings.DocType),
		Status:              strings.TrimSpace(settings.Status),
		File:                strings.TrimSpace(settings.File),
		Dir:                 strings.TrimSpace(settings.Dir),
		ExternalSource:      strings.TrimSpace(settings.ExternalSource),
		Since:               strings.TrimSpace(settings.Since),
		Until:               strings.TrimSpace(settings.Until),
		CreatedSince:        strings.TrimSpace(settings.CreatedSince),
		UpdatedSince:        strings.TrimSpace(settings.UpdatedSince),
		OrderBy:             orderBy,
		Reverse:             false,
		IncludeArchivedPath: true,
		IncludeScriptsPath:  true,
		IncludeControlDocs:  true,
		IncludeDiagnostics:  true,
		IncludeErrors:       false,
	})
	if err != nil {
		return err
	}

	for _, r := range resp.Results {
		row := types.NewRow(
			types.MRP("ticket", r.Ticket),
			types.MRP("title", r.Title),
			types.MRP("doc_type", r.DocType),
			types.MRP("status", r.Status),
			types.MRP("topics", strings.Join(r.Topics, ", ")),
			types.MRP("path", r.Path),
			types.MRP("snippet", r.Snippet),
		)
		if fileQueryRaw != "" {
			if len(r.MatchedFiles) > 0 {
				row.Set("file", strings.Join(r.MatchedFiles, ", "))
			}
			if len(r.MatchedNotes) > 0 {
				row.Set("file_note", strings.Join(r.MatchedNotes, " | "))
			}
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to emit search result for %s: %w", r.Path, err)
		}
	}

	for i := range resp.Diagnostics {
		docmgr.RenderTaxonomy(ctx, &resp.Diagnostics[i])
	}

	return nil
}

// suggestFiles suggests related files using heuristics (git + ripgrep)
func (c *SearchCommand) suggestFiles(
	ctx context.Context,
	settings *SearchSettings,
	gp middlewares.Processor,
) error {
	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return fmt.Errorf("failed to discover workspace: %w", err)
	}
	settings.Root = ws.Context().Root
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	suggestions, err := searchsvc.SuggestFiles(ctx, ws, searchsvc.SuggestFilesQuery{
		Ticket: strings.TrimSpace(settings.Ticket),
		Topics: settings.Topics,
		Query:  strings.TrimSpace(settings.Query),
	})
	if err != nil {
		return err
	}

	for _, s := range suggestions {
		row := types.NewRow(
			types.MRP("file", s.File),
			types.MRP("source", s.Source),
			types.MRP("reason", s.Reason),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to emit suggestion for %s: %w", s.File, err)
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &SearchCommand{}

// Implement BareCommand for human-friendly output
func (c *SearchCommand) Run(
	ctx context.Context,
	parsedValues *values.Values,
) error {
	settings := &SearchSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}
	settings.Root = workspace.ResolveRoot(settings.Root)

	// If only printing template schema, skip all other processing and output
	if settings.PrintTemplateSchema {
		type SearchResult struct {
			Ticket  string
			Title   string
			DocType string
			Status  string
			Topics  []string
			Path    string
			Snippet string
		}
		templateData := map[string]interface{}{
			"Query":        "",
			"TotalResults": 0,
			"Results": []SearchResult{
				{
					Ticket:  "",
					Title:   "",
					DocType: "",
					Status:  "",
					Topics:  []string{},
					Path:    "",
					Snippet: "",
				},
			},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	// Suggest files mode
	if settings.Files {
		ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
		if err != nil {
			return fmt.Errorf("failed to discover workspace: %w", err)
		}
		settings.Root = ws.Context().Root
		if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
			return fmt.Errorf("failed to initialize workspace index: %w", err)
		}

		suggestions, err := searchsvc.SuggestFiles(ctx, ws, searchsvc.SuggestFilesQuery{
			Ticket: strings.TrimSpace(settings.Ticket),
			Topics: settings.Topics,
			Query:  strings.TrimSpace(settings.Query),
		})
		if err != nil {
			return err
		}
		for _, s := range suggestions {
			fmt.Printf("%s — %s (source=%s)\n", s.File, s.Reason, s.Source)
		}
		return nil
	}

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return fmt.Errorf("failed to discover workspace: %w", err)
	}
	settings.Root = ws.Context().Root
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: true}); err != nil {
		return fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	orderBy := workspace.OrderBy(strings.TrimSpace(settings.OrderBy))
	if orderBy == "" {
		orderBy = workspace.OrderByPath
	}

	resp, err := searchsvc.SearchDocs(ctx, ws, searchsvc.SearchQuery{
		TextQuery:           strings.TrimSpace(settings.Query),
		Ticket:              strings.TrimSpace(settings.Ticket),
		Topics:              settings.Topics,
		DocType:             strings.TrimSpace(settings.DocType),
		Status:              strings.TrimSpace(settings.Status),
		File:                strings.TrimSpace(settings.File),
		Dir:                 strings.TrimSpace(settings.Dir),
		ExternalSource:      strings.TrimSpace(settings.ExternalSource),
		Since:               strings.TrimSpace(settings.Since),
		Until:               strings.TrimSpace(settings.Until),
		CreatedSince:        strings.TrimSpace(settings.CreatedSince),
		UpdatedSince:        strings.TrimSpace(settings.UpdatedSince),
		OrderBy:             orderBy,
		Reverse:             false,
		IncludeArchivedPath: true,
		IncludeScriptsPath:  true,
		IncludeControlDocs:  true,
		IncludeDiagnostics:  false,
		IncludeErrors:       false,
	})
	if err != nil {
		return err
	}

	// Print human output
	for _, result := range resp.Results {
		if strings.TrimSpace(settings.File) != "" {
			extra := ""
			if len(result.MatchedFiles) > 0 {
				extra += " file=" + strings.Join(result.MatchedFiles, ", ")
			}
			if len(result.MatchedNotes) > 0 {
				extra += " note=" + strings.Join(result.MatchedNotes, " | ")
			}
			fmt.Printf("%s — %s [%s] :: %s%s\n", result.Path, result.Title, result.Ticket, result.Snippet, extra)
		} else {
			fmt.Printf("%s — %s [%s] :: %s\n", result.Path, result.Title, result.Ticket, result.Snippet)
		}
	}

	// Render postfix template if it exists
	// Build template data struct
	type SearchResult struct {
		Ticket  string
		Title   string
		DocType string
		Status  string
		Topics  []string
		Path    string
		Snippet string
	}

	searchResults := make([]SearchResult, 0, len(resp.Results))
	for _, result := range resp.Results {
		topics := result.Topics
		if topics == nil {
			topics = []string{}
		}
		searchResults = append(searchResults, SearchResult{
			Ticket:  result.Ticket,
			Title:   result.Title,
			DocType: result.DocType,
			Status:  result.Status,
			Topics:  topics,
			Path:    result.Path,
			Snippet: result.Snippet,
		})
	}

	templateData := map[string]interface{}{
		"Query":        settings.Query,
		"TotalResults": resp.Total,
		"Results":      searchResults,
	}

	// Try verb path: ["doc", "search"] or ["search"]
	verbCandidates := [][]string{
		{"doc", "search"},
		{"search"},
	}
	settingsMap := map[string]interface{}{
		"root":           settings.Root,
		"query":          settings.Query,
		"ticket":         settings.Ticket,
		"topics":         settings.Topics,
		"docType":        settings.DocType,
		"status":         settings.Status,
		"file":           settings.File,
		"dir":            settings.Dir,
		"externalSource": settings.ExternalSource,
	}
	absRoot := settings.Root
	if abs, err := filepath.Abs(settings.Root); err == nil {
		absRoot = abs
	}
	_ = templates.RenderVerbTemplate(verbCandidates, absRoot, settingsMap, templateData)

	return nil
}

var _ cmds.BareCommand = &SearchCommand{}
