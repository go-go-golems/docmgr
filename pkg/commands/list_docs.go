package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/go-go-golems/docmgr/internal/templates"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgr"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/mattn/go-isatty"
)

// ListDocsCommand lists individual documents
type ListDocsCommand struct {
	*cmds.CommandDescription
}

// ListDocsSettings holds the parameters for the list docs command
type ListDocsSettings struct {
	Root    string   `glazed.parameter:"root"`
	Ticket  string   `glazed.parameter:"ticket"`
	Status  string   `glazed.parameter:"status"`
	DocType string   `glazed.parameter:"doc-type"`
	Topics  []string `glazed.parameter:"topics"`
	// Schema printing flags (human mode only)
	PrintTemplateSchema bool   `glazed.parameter:"print-template-schema"`
	SchemaFormat        string `glazed.parameter:"schema-format"`
}

func NewListDocsCommand() (*ListDocsCommand, error) {
	return &ListDocsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List individual documents"),
			cmds.WithLong(`Lists all individual documents across all workspaces.

Columns:
  ticket,doc_type,title,status,topics,path,last_updated

Examples:
  # Human output
  docmgr doc list
  docmgr list docs
  docmgr list docs --ticket MEN-3475
  docmgr list docs --doc-type design-doc
  docmgr list docs --topics chat,backend

  # Scriptable (paths only)
  docmgr list docs --ticket MEN-3475 --with-glaze-output --select path

  # TSV subset
  docmgr list docs --ticket MEN-3475 --with-glaze-output --output tsv --fields doc_type,title,path
`),
			cmds.WithFlags(
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
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by ticket identifier"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"status",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by status"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"doc-type",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by document type"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"topics",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Filter by topics (comma-separated, matches any)"),
					parameters.WithDefault([]string{}),
				),
			),
		),
	}, nil
}

func (c *ListDocsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ListDocsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// If only printing template schema, skip all other processing and output
	if settings.PrintTemplateSchema {
		type DocInfo struct {
			DocType string
			Title   string
			Status  string
			Topics  []string
			Updated string
			Path    string
		}
		type TicketInfo struct {
			Ticket string
			Docs   []DocInfo
		}
		templateData := map[string]interface{}{
			"TotalDocs":    0,
			"TotalTickets": 0,
			"Tickets": []TicketInfo{
				{
					Ticket: "",
					Docs:   []DocInfo{{}},
				},
			},
			"Rows": []map[string]interface{}{
				{
					"ticket":       "",
					"doc_type":     "",
					"title":        "",
					"status":       "",
					"topics":       "",
					"path":         "",
					"last_updated": "",
				},
			},
			"Fields": []string{"ticket", "doc_type", "title", "status", "topics", "path", "last_updated"},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return fmt.Errorf("failed to discover workspace: %w", err)
	}
	settings.Root = ws.Context().Root
	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeRepo},
		Filters: workspace.DocFilters{
			Ticket:    settings.Ticket,
			Status:    settings.Status,
			DocType:   settings.DocType,
			TopicsAny: settings.Topics,
		},
		Options: workspace.DocQueryOptions{
			IncludeErrors:       false,
			IncludeDiagnostics:  true,
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByPath,
			Reverse:             false,
			IncludeBody:         false,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to query docs: %w", err)
	}

	for _, h := range res.Docs {
		// Skip index.md files (those are tickets, use list tickets for those)
		if filepath.Base(h.Path) == "index.md" {
			continue
		}
		if h.Doc == nil {
			continue
		}

		relPath, err := filepath.Rel(settings.Root, h.Path)
		if err != nil {
			relPath = h.Path
		}
		relPath = filepath.ToSlash(relPath)

		row := types.NewRow(
			types.MRP(ColTicket, h.Doc.Ticket),
			types.MRP(ColDocType, h.Doc.DocType),
			types.MRP(ColTitle, h.Doc.Title),
			types.MRP(ColStatus, h.Doc.Status),
			types.MRP(ColTopics, strings.Join(h.Doc.Topics, ", ")),
			types.MRP(ColPath, relPath),
			types.MRP(ColLastUpdated, h.Doc.LastUpdated.Format("2006-01-02 15:04")),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to add document row for %s: %w", relPath, err)
		}
	}

	for i := range res.Diagnostics {
		docmgr.RenderTaxonomy(ctx, &res.Diagnostics[i])
	}

	return nil
}

var _ cmds.GlazeCommand = &ListDocsCommand{}

// Implement BareCommand for human-friendly output
func (c *ListDocsCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &ListDocsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// If only printing template schema, skip all other processing and output
	if settings.PrintTemplateSchema {
		type DocInfo struct {
			DocType string
			Title   string
			Status  string
			Topics  []string
			Updated string
			Path    string
		}
		type TicketInfo struct {
			Ticket string
			Docs   []DocInfo
		}
		templateData := map[string]interface{}{
			"TotalDocs":    0,
			"TotalTickets": 0,
			"Tickets": []TicketInfo{
				{
					Ticket: "",
					Docs:   []DocInfo{{}},
				},
			},
			"Rows": []map[string]interface{}{
				{
					"ticket":       "",
					"doc_type":     "",
					"title":        "",
					"status":       "",
					"topics":       "",
					"path":         "",
					"last_updated": "",
				},
			},
			"Fields": []string{"ticket", "doc_type", "title", "status", "topics", "path", "last_updated"},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return fmt.Errorf("failed to discover workspace: %w", err)
	}
	settings.Root = ws.Context().Root
	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	type docEntry struct {
		ticket      string
		docType     string
		title       string
		status      string
		topics      []string
		lastUpdated time.Time
		path        string
	}

	var entries []docEntry
	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeRepo},
		Filters: workspace.DocFilters{
			Ticket:    settings.Ticket,
			Status:    settings.Status,
			DocType:   settings.DocType,
			TopicsAny: settings.Topics,
		},
		Options: workspace.DocQueryOptions{
			IncludeErrors:       false,
			IncludeDiagnostics:  false, // keep human mode quiet (matches previous behavior)
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByPath,
			Reverse:             false,
			IncludeBody:         false,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to query docs: %w", err)
	}

	for _, h := range res.Docs {
		if filepath.Base(h.Path) == "index.md" {
			continue
		}
		if h.Doc == nil {
			continue
		}
		relPath, err := filepath.Rel(settings.Root, h.Path)
		if err != nil {
			relPath = h.Path
		}
		entries = append(entries, docEntry{
			ticket:      h.Doc.Ticket,
			docType:     h.Doc.DocType,
			title:       h.Doc.Title,
			status:      h.Doc.Status,
			topics:      append([]string{}, h.Doc.Topics...),
			lastUpdated: h.Doc.LastUpdated,
			path:        filepath.ToSlash(relPath),
		})
	}

	if len(entries) == 0 {
		fmt.Println("No documents found.")
		return nil
	}

	absRoot := settings.Root
	if !filepath.IsAbs(absRoot) {
		if cwd, err := os.Getwd(); err == nil {
			absRoot = filepath.Join(cwd, absRoot)
		}
	}

	grouped := map[string][]docEntry{}
	latest := map[string]time.Time{}
	order := []string{}
	for _, entry := range entries {
		if _, ok := grouped[entry.ticket]; !ok {
			grouped[entry.ticket] = []docEntry{}
			order = append(order, entry.ticket)
		}
		grouped[entry.ticket] = append(grouped[entry.ticket], entry)
		if entry.lastUpdated.After(latest[entry.ticket]) {
			latest[entry.ticket] = entry.lastUpdated
		}
	}
	sort.SliceStable(order, func(i, j int) bool {
		return latest[order[i]].After(latest[order[j]])
	})

	var b strings.Builder
	fmt.Fprintf(&b, "Docs root: `%s`\nPaths are relative to this root.\n\n", absRoot)
	fmt.Fprintf(&b, "## Documents (%d)\n\n", len(entries))
	for _, ticket := range order {
		docs := grouped[ticket]
		sort.SliceStable(docs, func(i, j int) bool {
			if docs[i].docType == docs[j].docType {
				return docs[i].title < docs[j].title
			}
			return docs[i].docType < docs[j].docType
		})
		fmt.Fprintf(&b, "### %s (%d docs)\n\n", ticket, len(docs))
		for _, entry := range docs {
			topics := "—"
			if len(entry.topics) > 0 {
				topics = strings.Join(entry.topics, ", ")
			}
			updated := "unknown"
			if !entry.lastUpdated.IsZero() {
				updated = entry.lastUpdated.Format("2006-01-02 15:04")
			}
			docType := entry.docType
			if docType == "" {
				docType = "doc"
			}
			fmt.Fprintf(&b, "- **%s** — %s\n", docType, entry.title)
			fmt.Fprintf(&b, "  - Status: **%s**\n", entry.status)
			fmt.Fprintf(&b, "  - Topics: %s\n", topics)
			fmt.Fprintf(&b, "  - Updated: %s\n", updated)
			fmt.Fprintf(&b, "  - Path: `%s`\n\n", entry.path)
		}
	}

	content := b.String()
	fd := os.Stdout.Fd()
	if isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd) {
		if renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(0),
		); err == nil {
			if rendered, err := renderer.Render(content); err == nil {
				fmt.Print(rendered)
			} else {
				fmt.Print(content)
			}
		} else {
			fmt.Print(content)
		}
	} else {
		fmt.Print(content)
	}

	// Render postfix template if it exists
	// Build template data struct
	type DocInfo struct {
		DocType string
		Title   string
		Status  string
		Topics  []string
		Updated string
		Path    string
	}
	type TicketInfo struct {
		Ticket string
		Docs   []DocInfo
	}

	tickets := make([]TicketInfo, 0, len(order))
	for _, ticket := range order {
		docs := grouped[ticket]
		docInfos := make([]DocInfo, 0, len(docs))
		for _, entry := range docs {
			updated := "unknown"
			if !entry.lastUpdated.IsZero() {
				updated = entry.lastUpdated.Format("2006-01-02 15:04")
			}
			docInfos = append(docInfos, DocInfo{
				DocType: entry.docType,
				Title:   entry.title,
				Status:  entry.status,
				Topics:  entry.topics,
				Updated: updated,
				Path:    entry.path,
			})
		}
		tickets = append(tickets, TicketInfo{
			Ticket: ticket,
			Docs:   docInfos,
		})
	}

	// Build rows for template (same as Glaze rows)
	rows := make([]map[string]interface{}, 0, len(entries))
	fields := []string{"ticket", "doc_type", "title", "status", "topics", "path", "last_updated"}
	for _, entry := range entries {
		topicsStr := strings.Join(entry.topics, ", ")
		updated := "unknown"
		if !entry.lastUpdated.IsZero() {
			updated = entry.lastUpdated.Format("2006-01-02 15:04")
		}
		rows = append(rows, map[string]interface{}{
			"ticket":       entry.ticket,
			"doc_type":     entry.docType,
			"title":        entry.title,
			"status":       entry.status,
			"topics":       topicsStr,
			"path":         entry.path,
			"last_updated": updated,
		})
	}

	templateData := map[string]interface{}{
		"TotalDocs":    len(entries),
		"TotalTickets": len(order),
		"Tickets":      tickets,
		"Rows":         rows,
		"Fields":       fields,
	}

	// Try both possible verb paths: ["doc", "list"] and ["list", "docs"]
	verbCandidates := [][]string{
		{"doc", "list"},
		{"list", "docs"},
	}
	settingsMap := map[string]interface{}{
		"root":    settings.Root,
		"ticket":  settings.Ticket,
		"status":  settings.Status,
		"docType": settings.DocType,
		"topics":  settings.Topics,
	}
	_ = templates.RenderVerbTemplate(verbCandidates, absRoot, settingsMap, templateData)

	return nil
}

var _ cmds.BareCommand = &ListDocsCommand{}
