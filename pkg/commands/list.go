package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// ListCommand lists document workspaces
type ListCommand struct {
	*cmds.CommandDescription
}

// ListSettings holds the parameters for the list command
type ListSettings struct {
	Root   string `glazed:"root"`
	Ticket string `glazed:"ticket"`
	Status string `glazed:"status"`
}

func NewListCommand() (*ListCommand, error) {
	return &ListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List document workspaces"),
			cmds.WithLong(`Lists all document workspaces in the root directory.

Example:
  docmgr list
  docmgr list --ticket MEN-3475
  docmgr list --status active
`),
			cmds.WithFlags(
				fields.New(
					"root",
					fields.TypeString,
					fields.WithHelp("Root directory for docs"),
					fields.WithDefault("ttmp"),
				),
				fields.New(
					"ticket",
					fields.TypeString,
					fields.WithHelp("Filter by ticket identifier"),
					fields.WithDefault(""),
				),
				fields.New(
					"status",
					fields.TypeString,
					fields.WithHelp("Filter by status"),
					fields.WithDefault(""),
				),
			),
		),
	}, nil
}

func (c *ListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp middlewares.Processor,
) error {
	settings := &ListSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
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
			Ticket:  strings.TrimSpace(settings.Ticket),
			Status:  strings.TrimSpace(settings.Status),
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
		return fmt.Errorf("failed to query docs: %w", err)
	}

	for _, h := range res.Docs {
		doc := h.Doc
		if doc == nil {
			continue
		}
		ticketDirAbs := filepath.Clean(filepath.Dir(filepath.FromSlash(h.Path)))
		relPath := ticketDirAbs
		if rel, err := filepath.Rel(settings.Root, ticketDirAbs); err == nil {
			relPath = rel
		}
		relPath = filepath.ToSlash(relPath)

		row := types.NewRow(
			types.MRP("ticket", doc.Ticket),
			types.MRP("title", doc.Title),
			types.MRP("status", doc.Status),
			types.MRP("topics", strings.Join(doc.Topics, ", ")),
			types.MRP("path", relPath),
			types.MRP("last_updated", doc.LastUpdated.Format("2006-01-02 15:04")),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to add workspace row for %s: %w", doc.Ticket, err)
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &ListCommand{}
