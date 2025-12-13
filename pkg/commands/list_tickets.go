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
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/mattn/go-isatty"
)

// countTasksInTicket loads tasks.md under a ticket and returns (open, done)
func countTasksInTicket(ticketDir string) (int, int) {
	path := filepath.Join(ticketDir, "tasks.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, 0
	}
	lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	tasks := parseTasksFromLines(lines)
	done := 0
	for _, t := range tasks {
		if t.Checked {
			done++
		}
	}
	open := len(tasks) - done
	if open < 0 {
		open = 0
	}
	return open, done
}

// ListTicketsCommand lists ticket workspaces
type ListTicketsCommand struct {
	*cmds.CommandDescription
}

// ListTicketsSettings holds the parameters for the list tickets command
type ListTicketsSettings struct {
	Root   string `glazed.parameter:"root"`
	Ticket string `glazed.parameter:"ticket"`
	Status string `glazed.parameter:"status"`
	// Schema printing flags (human mode only)
	PrintTemplateSchema bool   `glazed.parameter:"print-template-schema"`
	SchemaFormat        string `glazed.parameter:"schema-format"`
}

func NewListTicketsCommand() (*ListTicketsCommand, error) {
	return &ListTicketsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"tickets",
			cmds.WithShort("List ticket workspaces"),
			cmds.WithLong(`Lists all ticket workspaces in the root directory.

Columns:
  ticket,title,status,topics,path,last_updated

Examples:
  # Human output
  docmgr list tickets
  docmgr list tickets --ticket MEN-3475
  docmgr list tickets --status active

  # Scriptable (paths only)
  docmgr list tickets --with-glaze-output --select path

  # CSV of selected fields without headers
  docmgr list tickets --with-glaze-output --output csv --with-headers=false --fields ticket,path
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
			),
		),
	}, nil
}

func (c *ListTicketsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ListTicketsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// If only printing template schema, skip all other processing and output
	if settings.PrintTemplateSchema {
		type TicketInfo struct {
			Ticket      string
			Title       string
			Status      string
			Topics      []string
			Path        string
			LastUpdated string
		}
		templateData := map[string]interface{}{
			"TotalTickets": 0,
			"Tickets": []TicketInfo{
				{
					Ticket:      "",
					Title:       "",
					Status:      "",
					Topics:      []string{},
					Path:        "",
					LastUpdated: "",
				},
			},
			"Rows": []map[string]interface{}{
				{
					"ticket":       "",
					"title":        "",
					"status":       "",
					"topics":       "",
					"tasks_open":   0,
					"tasks_done":   0,
					"path":         "",
					"last_updated": "",
				},
			},
			"Fields": []string{"ticket", "title", "status", "topics", "path", "last_updated"},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	root, tickets, err := queryTicketIndexDocs(ctx, settings.Root, settings.Ticket, settings.Status)
	if err != nil {
		return err
	}
	settings.Root = root

	for _, t := range tickets {
		row := types.NewRow(
			types.MRP(ColTicket, t.Ticket),
			types.MRP(ColTitle, t.Title),
			types.MRP(ColStatus, t.Status),
			types.MRP(ColTopics, strings.Join(t.Topics, ", ")),
			types.MRP(ColTasksOpen, t.TasksOpen),
			types.MRP(ColTasksDone, t.TasksDone),
			types.MRP(ColPath, t.Path),
			types.MRP(ColLastUpdated, t.LastUpdated.Format("2006-01-02 15:04")),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to add ticket row for %s: %w", t.Ticket, err)
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &ListTicketsCommand{}

// Implement BareCommand for human-friendly output
func (c *ListTicketsCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &ListTicketsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// If only printing template schema, skip all other processing and output
	if settings.PrintTemplateSchema {
		type TicketInfo struct {
			Ticket      string
			Title       string
			Status      string
			Topics      []string
			Path        string
			LastUpdated string
		}
		templateData := map[string]interface{}{
			"TotalTickets": 0,
			"Tickets": []TicketInfo{
				{
					Ticket:      "",
					Title:       "",
					Status:      "",
					Topics:      []string{},
					Path:        "",
					LastUpdated: "",
				},
			},
			"Rows": []map[string]interface{}{
				{
					"ticket":       "",
					"title":        "",
					"status":       "",
					"topics":       "",
					"tasks_open":   0,
					"tasks_done":   0,
					"path":         "",
					"last_updated": "",
				},
			},
			"Fields": []string{"ticket", "title", "status", "topics", "path", "last_updated"},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	root, tickets, err := queryTicketIndexDocs(ctx, settings.Root, settings.Ticket, settings.Status)
	if err != nil {
		return err
	}
	settings.Root = root

	// Markdown-formatted sections for human-friendly output
	if len(tickets) == 0 {
		fmt.Println("No tickets found.")
		return nil
	}
	var b strings.Builder
	rootDisplay := settings.Root
	if abs, err := filepath.Abs(settings.Root); err == nil {
		rootDisplay = abs
	}
	if rootDisplay != "" {
		fmt.Fprintf(&b, "Docs root: `%s`\nPaths are relative to this root.\n\n", rootDisplay)
	}

	b.WriteString(fmt.Sprintf("## Tickets (%d)\n\n", len(tickets)))
	for _, t := range tickets {
		topics := "—"
		if len(t.Topics) > 0 {
			topics = strings.Join(t.Topics, ", ")
		}
		fmt.Fprintf(&b, "### %s — %s\n", t.Ticket, t.Title)
		fmt.Fprintf(&b, "- Status: **%s**\n", t.Status)
		fmt.Fprintf(&b, "- Topics: %s\n", topics)
		fmt.Fprintf(&b, "- Tasks: %d open / %d done\n", t.TasksOpen, t.TasksDone)
		fmt.Fprintf(&b, "- Updated: %s\n", t.LastUpdated.Format("2006-01-02 15:04"))
		fmt.Fprintf(&b, "- Path: `%s`\n\n", t.Path)
	}
	content := b.String()

	// If stdout is a TTY, render with glamour for nicer presentation
	fd := os.Stdout.Fd()
	if isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd) {
		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(0),
		)
		if err == nil {
			if rendered, err2 := r.Render(content); err2 == nil {
				fmt.Print(rendered)
			} else {
				fmt.Print(content)
			}
		} else {
			fmt.Print(content)
		}
	} else {
		// Fallback: print raw markdown
		fmt.Print(content)
	}

	// Render postfix template if it exists
	// Build template data struct
	type TicketInfo struct {
		Ticket      string
		Title       string
		Status      string
		Topics      []string
		Path        string
		LastUpdated string
	}

	ticketInfos := make([]TicketInfo, 0, len(tickets))
	for _, t := range tickets {
		topics := t.Topics
		if topics == nil {
			topics = []string{}
		}
		ticketInfos = append(ticketInfos, TicketInfo{
			Ticket:      t.Ticket,
			Title:       t.Title,
			Status:      t.Status,
			Topics:      topics,
			Path:        t.Path,
			LastUpdated: t.LastUpdated.Format("2006-01-02 15:04"),
		})
	}

	// Build rows for template (same as Glaze rows)
	rows := make([]map[string]interface{}, 0, len(tickets))
	fields := []string{"ticket", "title", "status", "topics", "path", "last_updated"}
	for _, t := range tickets {
		topicsStr := strings.Join(t.Topics, ", ")
		rows = append(rows, map[string]interface{}{
			"ticket":       t.Ticket,
			"title":        t.Title,
			"status":       t.Status,
			"topics":       topicsStr,
			"tasks_open":   t.TasksOpen,
			"tasks_done":   t.TasksDone,
			"path":         t.Path,
			"last_updated": t.LastUpdated.Format("2006-01-02 15:04"),
		})
	}

	templateData := map[string]interface{}{
		"TotalTickets": len(tickets),
		"Tickets":      ticketInfos,
		"Rows":         rows,
		"Fields":       fields,
	}

	// Try verb path: ["list", "tickets"]
	verbCandidates := [][]string{
		{"list", "tickets"},
	}
	settingsMap := map[string]interface{}{
		"root":   settings.Root,
		"ticket": settings.Ticket,
		"status": settings.Status,
	}
	_ = templates.RenderVerbTemplate(verbCandidates, rootDisplay, settingsMap, templateData)

	return nil
}

var _ cmds.BareCommand = &ListTicketsCommand{}

type ticketIndexDoc struct {
	Ticket      string
	Title       string
	Status      string
	Topics      []string
	Path        string // relative to docs root (slash-separated) when possible; fallback abs
	TicketDir   string // absolute OS path to the ticket directory
	LastUpdated time.Time
	TasksOpen   int
	TasksDone   int
}

func queryTicketIndexDocs(ctx context.Context, rootOverride string, ticketFilter string, statusFilter string) (string, []ticketIndexDoc, error) {
	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: rootOverride})
	if err != nil {
		return "", nil, fmt.Errorf("failed to discover workspace: %w", err)
	}
	root := ws.Context().Root
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return "", nil, fmt.Errorf("root directory does not exist: %s", root)
	}
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return "", nil, fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeRepo},
		Filters: workspace.DocFilters{
			Ticket:  strings.TrimSpace(ticketFilter),
			Status:  strings.TrimSpace(statusFilter),
			DocType: "index",
		},
		Options: workspace.DocQueryOptions{
			IncludeErrors:       false,
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByLastUpdated,
			Reverse:             true,
		},
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to query docs: %w", err)
	}

	out := make([]ticketIndexDoc, 0, len(res.Docs))
	for _, h := range res.Docs {
		if h.Doc == nil {
			continue
		}
		ticketDirAbs := filepath.Clean(filepath.Dir(filepath.FromSlash(h.Path)))
		open, done := countTasksInTicket(ticketDirAbs)

		relPath := ticketDirAbs
		if rel, err := filepath.Rel(root, ticketDirAbs); err == nil {
			relPath = rel
		}
		relPath = filepath.ToSlash(relPath)

		out = append(out, ticketIndexDoc{
			Ticket:      h.Doc.Ticket,
			Title:       h.Doc.Title,
			Status:      h.Doc.Status,
			Topics:      h.Doc.Topics,
			Path:        relPath,
			TicketDir:   ticketDirAbs,
			LastUpdated: h.Doc.LastUpdated,
			TasksOpen:   open,
			TasksDone:   done,
		})
	}

	// Order by LastUpdated (newest first).
	sort.Slice(out, func(i, j int) bool {
		return out[i].LastUpdated.After(out[j].LastUpdated)
	})

	return root, out, nil
}
