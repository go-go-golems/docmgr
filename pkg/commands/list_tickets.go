package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/glamour"
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
			return err
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

	// Markdown-formatted table for human-friendly output
	if len(filtered) == 0 {
		fmt.Println("No tickets found.")
		return nil
	}
	var b strings.Builder
	b.WriteString("| Ticket | Title | Status | Topics | Tasks (open/done) | Updated | Path |\n")
	b.WriteString("|---|---|---|---|---:|---|---|\n")
	for _, ws := range filtered {
		doc := ws.Doc
		open, done := countTasksInTicket(ws.Path)
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %d/%d | %s | %s |\n",
			doc.Ticket,
			doc.Title,
			doc.Status,
			strings.Join(doc.Topics, ", "),
			open, done,
			doc.LastUpdated.Format("2006-01-02 15:04"),
			ws.Path,
		)
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
				return nil
			}
		}
	}
	// Fallback: print raw markdown
	fmt.Print(content)
	return nil
}

var _ cmds.BareCommand = &ListTicketsCommand{}
