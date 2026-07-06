package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/tickets"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// ShowTicketCommand prints one ticket's detail (metadata, docs, tasks, changelog head).
type ShowTicketCommand struct {
	*cmds.CommandDescription
}

type ShowTicketSettings struct {
	Ticket    string `glazed:"ticket"`
	TicketRef string `glazed:"ticket-ref"`
	Root      string `glazed:"root"`
}

type showTicketDocEntry struct {
	Path    string // relative to the ticket directory
	DocType string
	Title   string
}

type showTicketResult struct {
	Ticket        string
	Title         string
	Status        string
	Topics        []string
	Path          string // ticket dir relative to docs root
	LastUpdated   string
	TasksOpen     int
	TasksDone     int
	Docs          []showTicketDocEntry
	ChangelogHead string
}

func NewShowTicketCommand() (*ShowTicketCommand, error) {
	return &ShowTicketCommand{
		CommandDescription: cmds.NewCommandDescription(
			"show",
			cmds.WithShort("Show one ticket's detail (metadata, docs, tasks, changelog)"),
			cmds.WithLong(`Shows a single ticket workspace: metadata, document list, task summary,
and the most recent changelog heading.

The ticket reference is forgiving: exact ID, unique ID prefix, or the ticket
directory name (e.g. "MEN-4242--normalize-chat-api-paths") all work.

Examples:
  docmgr ticket show MEN-4242
  docmgr ticket show --ticket MEN-4242
`),
			cmds.WithFlags(
				fields.New(
					"ticket",
					fields.TypeString,
					fields.WithHelp("Ticket identifier (ID, unique prefix, or directory name)"),
					fields.WithDefault(""),
				),
				fields.New(
					"root",
					fields.TypeString,
					fields.WithHelp("Root directory for docs"),
					fields.WithDefault("ttmp"),
				),
			),
			cmds.WithArguments(
				fields.New(
					"ticket-ref",
					fields.TypeString,
					fields.WithHelp("Ticket reference (alternative to --ticket)"),
					fields.WithDefault(""),
				),
			),
		),
	}, nil
}

func (c *ShowTicketCommand) gather(ctx context.Context, settings *ShowTicketSettings) (*showTicketResult, error) {
	ref := strings.TrimSpace(settings.Ticket)
	if ref == "" {
		ref = strings.TrimSpace(settings.TicketRef)
	}
	if ref == "" {
		return nil, fmt.Errorf("specify a ticket: docmgr ticket show --ticket <ID> (or a positional reference)")
	}

	settings.Root = workspace.ResolveRoot(settings.Root)
	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return nil, fmt.Errorf("failed to discover workspace: %w", err)
	}
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return nil, fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	res, err := tickets.Resolve(ctx, ws, ref)
	if err != nil {
		return nil, err
	}

	open, done := countTasksInTicket(res.TicketDirAbs)

	result := &showTicketResult{
		Ticket:    res.TicketID,
		Path:      res.TicketDirRel,
		TasksOpen: open,
		TasksDone: done,
	}
	if res.IndexDoc != nil {
		result.Title = res.IndexDoc.Title
		result.Status = res.IndexDoc.Status
		result.Topics = res.IndexDoc.Topics
		if !res.IndexDoc.LastUpdated.IsZero() {
			result.LastUpdated = res.IndexDoc.LastUpdated.Format("2006-01-02 15:04")
		}
	}

	docsRes, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeTicket, TicketID: res.TicketID},
		Options: workspace.DocQueryOptions{
			IncludeErrors:       false,
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeSourcesPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByPath,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query ticket docs: %w", err)
	}
	for _, h := range docsRes.Docs {
		if h.Doc == nil {
			continue
		}
		rel := h.Path
		if r, err := filepath.Rel(res.TicketDirAbs, filepath.FromSlash(h.Path)); err == nil && !strings.HasPrefix(r, "..") {
			rel = filepath.ToSlash(r)
		}
		result.Docs = append(result.Docs, showTicketDocEntry{
			Path:    rel,
			DocType: h.Doc.DocType,
			Title:   h.Doc.Title,
		})
	}

	result.ChangelogHead = latestChangelogHeading(filepath.Join(res.TicketDirAbs, "changelog.md"))

	return result, nil
}

// latestChangelogHeading returns the last "## " heading in changelog.md
// (entries are appended, so the last heading is the most recent).
func latestChangelogHeading(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	head := ""
	for _, line := range strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			head = strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
		}
	}
	return head
}

// Run implements cmds.BareCommand.
func (c *ShowTicketCommand) Run(ctx context.Context, pl *values.Values) error {
	settings := &ShowTicketSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.gather(ctx, settings)
	if err != nil {
		return err
	}

	topics := "—"
	if len(result.Topics) > 0 {
		topics = strings.Join(result.Topics, ", ")
	}
	fmt.Printf("%s — %s\n", result.Ticket, result.Title)
	fmt.Printf("status: %s  topics: %s  updated: %s\n", result.Status, topics, result.LastUpdated)
	fmt.Printf("path: %s\n", result.Path)
	fmt.Printf("tasks: %d open / %d done\n", result.TasksOpen, result.TasksDone)
	fmt.Printf("docs (%d):\n", len(result.Docs))
	for _, d := range result.Docs {
		fmt.Printf("  - %s (%s)\n", d.Path, d.DocType)
	}
	if result.ChangelogHead != "" {
		fmt.Printf("changelog: %s\n", result.ChangelogHead)
	}
	return nil
}

// RunIntoGlazeProcessor implements cmds.GlazeCommand.
func (c *ShowTicketCommand) RunIntoGlazeProcessor(ctx context.Context, pl *values.Values, gp middlewares.Processor) error {
	settings := &ShowTicketSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.gather(ctx, settings)
	if err != nil {
		return err
	}

	docPaths := make([]string, 0, len(result.Docs))
	for _, d := range result.Docs {
		docPaths = append(docPaths, d.Path)
	}

	row := types.NewRow(
		types.MRP(ColTicket, result.Ticket),
		types.MRP(ColTitle, result.Title),
		types.MRP(ColStatus, result.Status),
		types.MRP(ColTopics, strings.Join(result.Topics, ", ")),
		types.MRP(ColTasksOpen, result.TasksOpen),
		types.MRP(ColTasksDone, result.TasksDone),
		types.MRP(ColPath, result.Path),
		types.MRP(ColLastUpdated, result.LastUpdated),
		types.MRP("docs", docPaths),
		types.MRP("changelog_head", result.ChangelogHead),
	)
	return gp.AddRow(ctx, row)
}

var _ cmds.BareCommand = &ShowTicketCommand{}
var _ cmds.GlazeCommand = &ShowTicketCommand{}
