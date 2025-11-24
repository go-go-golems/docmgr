package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

	// Apply config root if present
	settings.Root = workspace.ResolveRoot(settings.Root)

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

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	workspaces, err := workspace.CollectTicketWorkspaces(settings.Root, nil)
	if err != nil {
		return fmt.Errorf("failed to discover ticket workspaces: %w", err)
	}

	// Filter and sort by last updated (newest first)
	filtered := make([]workspace.TicketWorkspace, 0, len(workspaces))
	for _, ws := range workspaces {
		doc := ws.Doc
		if doc == nil {
			continue
		}

		// Apply filters
		if settings.Ticket != "" && !strings.Contains(doc.Ticket, settings.Ticket) {
			continue
		}
		if settings.Status != "" && doc.Status != settings.Status {
			continue
		}

		filtered = append(filtered, ws)
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Doc.LastUpdated.After(filtered[j].Doc.LastUpdated)
	})

	for _, ws := range filtered {
		doc := ws.Doc
		open, done := countTasksInTicket(ws.Path)
		row := types.NewRow(
			types.MRP(ColTicket, doc.Ticket),
			types.MRP(ColTitle, doc.Title),
			types.MRP(ColStatus, doc.Status),
			types.MRP(ColTopics, strings.Join(doc.Topics, ", ")),
			types.MRP(ColTasksOpen, open),
			types.MRP(ColTasksDone, done),
			types.MRP(ColPath, ws.Path),
			types.MRP(ColLastUpdated, doc.LastUpdated.Format("2006-01-02 15:04")),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to add ticket row for %s: %w", doc.Ticket, err)
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

	// Apply config root if present
	settings.Root = workspace.ResolveRoot(settings.Root)

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

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	workspaces, err := workspace.CollectTicketWorkspaces(settings.Root, nil)
	if err != nil {
		return fmt.Errorf("failed to discover ticket workspaces: %w", err)
	}

	// Filter and sort by last updated (newest first)
	filtered := make([]workspace.TicketWorkspace, 0, len(workspaces))
	for _, ws := range workspaces {
		doc := ws.Doc
		if doc == nil {
			continue
		}
		if settings.Ticket != "" && !strings.Contains(doc.Ticket, settings.Ticket) {
			continue
		}
		if settings.Status != "" && doc.Status != settings.Status {
			continue
		}
		filtered = append(filtered, ws)
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Doc.LastUpdated.After(filtered[j].Doc.LastUpdated)
	})

	// Markdown-formatted sections for human-friendly output
	if len(filtered) == 0 {
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

	b.WriteString(fmt.Sprintf("## Tickets (%d)\n\n", len(filtered)))
	for _, ws := range filtered {
		doc := ws.Doc
		open, done := countTasksInTicket(ws.Path)
		topics := "—"
		if len(doc.Topics) > 0 {
			topics = strings.Join(doc.Topics, ", ")
		}
		relPath := ws.Path
		absRoot := rootDisplay
		if abs, err := filepath.Abs(ws.Path); err == nil {
			if absRoot != "" {
				if rel, err2 := filepath.Rel(absRoot, abs); err2 == nil && rel != "" && rel != "." {
					relPath = rel
				} else if err2 == nil && rel == "." {
					relPath = "."
				} else {
					relPath = abs
				}
			} else {
				relPath = abs
			}
		}
		fmt.Fprintf(&b, "### %s — %s\n", doc.Ticket, doc.Title)
		fmt.Fprintf(&b, "- Status: **%s**\n", doc.Status)
		fmt.Fprintf(&b, "- Topics: %s\n", topics)
		fmt.Fprintf(&b, "- Tasks: %d open / %d done\n", open, done)
		fmt.Fprintf(&b, "- Updated: %s\n", doc.LastUpdated.Format("2006-01-02 15:04"))
		fmt.Fprintf(&b, "- Path: `%s`\n\n", relPath)
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

	ticketInfos := make([]TicketInfo, 0, len(filtered))
	for _, ws := range filtered {
		doc := ws.Doc
		topics := doc.Topics
		if topics == nil {
			topics = []string{}
		}
		relPath := ws.Path
		if abs, err := filepath.Abs(ws.Path); err == nil {
			if rootDisplay != "" {
				if rel, err2 := filepath.Rel(rootDisplay, abs); err2 == nil && rel != "" && rel != "." {
					relPath = rel
				} else if err2 == nil && rel == "." {
					relPath = "."
				} else {
					relPath = abs
				}
			} else {
				relPath = abs
			}
		}
		ticketInfos = append(ticketInfos, TicketInfo{
			Ticket:      doc.Ticket,
			Title:       doc.Title,
			Status:      doc.Status,
			Topics:      topics,
			Path:        relPath,
			LastUpdated: doc.LastUpdated.Format("2006-01-02 15:04"),
		})
	}

	// Build rows for template (same as Glaze rows)
	rows := make([]map[string]interface{}, 0, len(filtered))
	fields := []string{"ticket", "title", "status", "topics", "path", "last_updated"}
	for _, ws := range filtered {
		doc := ws.Doc
		open, done := countTasksInTicket(ws.Path)
		topicsStr := strings.Join(doc.Topics, ", ")
		relPath := ws.Path
		if abs, err := filepath.Abs(ws.Path); err == nil {
			if rootDisplay != "" {
				if rel, err2 := filepath.Rel(rootDisplay, abs); err2 == nil && rel != "" && rel != "." {
					relPath = rel
				} else if err2 == nil && rel == "." {
					relPath = "."
				} else {
					relPath = abs
				}
			} else {
				relPath = abs
			}
		}
		rows = append(rows, map[string]interface{}{
			"ticket":       doc.Ticket,
			"title":        doc.Title,
			"status":       doc.Status,
			"topics":       topicsStr,
			"tasks_open":   open,
			"tasks_done":   done,
			"path":         relPath,
			"last_updated": doc.LastUpdated.Format("2006-01-02 15:04"),
		})
	}

	templateData := map[string]interface{}{
		"TotalTickets": len(filtered),
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
