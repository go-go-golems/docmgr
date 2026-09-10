package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type TicketCloseCommand struct{ *cmds.CommandDescription }
type TicketCloseSettings struct {
	Ticket         string `glazed:"ticket"`
	Root           string `glazed:"root"`
	Status         string `glazed:"status"`
	Intent         string `glazed:"intent"`
	ChangelogEntry string `glazed:"changelog-entry"`
	OperationID    string `glazed:"operation-id"`
}

func NewTicketCloseCommand() (*TicketCloseCommand, error) {
	return &TicketCloseCommand{cmds.NewCommandDescription("close",
		cmds.WithShort("Close a ticket with recoverable status and changelog projections"),
		cmds.WithLong(`Preflights both files and journals the operation before writing history, then status.
This is recoverable, not atomic multi-file visibility. Cooperating close/milestone
operations share a ticket lock. Other editors do not participate in that lock.
An already matching status/intent is a no-op unless an explicit operation ID is supplied.
Use --operation-id for retry identity. After an interrupted default close, retry
with the reported ID and original options. Conflicting human edits stop recovery.
Open tasks produce a warning, not a failure. See: docmgr help milestone-workflows.`),
		cmds.WithFlags(fields.New("ticket", fields.TypeString, fields.WithRequired(true)), fields.New("root", fields.TypeString, fields.WithDefault("ttmp")), fields.New("status", fields.TypeString, fields.WithDefault("complete")), fields.New("intent", fields.TypeString), fields.New("changelog-entry", fields.TypeString, fields.WithDefault("Ticket closed")), fields.New("operation-id", fields.TypeString, fields.WithHelp("Stable retry ID; otherwise generated and reported"))))}, nil
}
func operationTicketDir(ctx context.Context, root, ticket string) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("nil context")
	}
	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: workspace.ResolveRoot(root)})
	if err != nil {
		return "", err
	}
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return "", err
	}
	return resolveTicketDirViaWorkspace(ctx, ws, ticket)
}
func (c *TicketCloseCommand) execute(ctx context.Context, parsed *values.Values) (CloseResult, error) {
	s := TicketCloseSettings{}
	if err := parsed.DecodeSectionInto(schema.DefaultSlug, &s); err != nil {
		return CloseResult{}, err
	}
	dir, err := operationTicketDir(ctx, s.Root, s.Ticket)
	if err != nil {
		return CloseResult{}, err
	}
	return CloseTicket(ctx, dir, s.OperationID, CloseRequest{Ticket: s.Ticket, Status: s.Status, Intent: s.Intent, Entry: s.ChangelogEntry})
}
func (c *TicketCloseCommand) RunIntoGlazeProcessor(ctx context.Context, parsed *values.Values, gp middlewares.Processor) error {
	r, err := c.execute(ctx, parsed)
	if err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("status", r.Status), types.MRP("intent", r.Intent), types.MRP("open_tasks", r.OpenTasks), types.MRP("done_tasks", r.DoneTasks), types.MRP("all_tasks_done", r.OpenTasks == 0 && r.DoneTasks > 0), types.MRP("receipt", r.Receipt)))
}
func (c *TicketCloseCommand) Run(ctx context.Context, parsed *values.Values) error {
	r, err := c.execute(ctx, parsed)
	if err != nil {
		return err
	}
	if r.OpenTasks > 0 {
		fmt.Fprintf(os.Stderr, "Warning: Not all tasks are done (%d open, %d done). Closing anyway.\n", r.OpenTasks, r.DoneTasks)
	}
	fmt.Printf("Ticket status: %s; operation: %s (%s)\n", r.Status, r.Receipt.OperationID, r.Receipt.State)
	return nil
}

var _ cmds.GlazeCommand = &TicketCloseCommand{}
var _ cmds.BareCommand = &TicketCloseCommand{}
