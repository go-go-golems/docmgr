package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/go-go-golems/docmgr/internal/operations"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type MilestoneCommand struct{ *cmds.CommandDescription }
type milestoneSettings struct {
	Root     string   `glazed:"root"`
	Ticket   string   `glazed:"ticket"`
	ID       string   `glazed:"operation-id"`
	Summary  string   `glazed:"summary"`
	Phase    string   `glazed:"phase"`
	Next     string   `glazed:"next"`
	Tasks    []string `glazed:"task-id"`
	Evidence string   `glazed:"evidence-file"`
	Expected string   `glazed:"expected-file"`
	Dry      bool     `glazed:"dry-run"`
}

func NewMilestoneCommand() *MilestoneCommand {
	return &MilestoneCommand{cmds.NewCommandDescription("record", cmds.WithShort("Record a retry-safe milestone and project tasks/history"), cmds.WithLong("Bounded ticket-local evidence and recoverable projections. See docmgr help milestone-workflows for consistency, revision and recovery contracts."), cmds.WithFlags(
		fields.New("root", fields.TypeString, fields.WithDefault("ttmp")), fields.New("ticket", fields.TypeString, fields.WithRequired(true)), fields.New("operation-id", fields.TypeString, fields.WithRequired(true)), fields.New("summary", fields.TypeString, fields.WithRequired(true)), fields.New("phase", fields.TypeString, fields.WithDefault("checkpoint")), fields.New("next", fields.TypeString), fields.New("task-id", fields.TypeStringList, fields.WithHelp("Stable task IDs to complete")), fields.New("evidence-file", fields.TypeString, fields.WithHelp("JSON array of kind/path/revision(SHA256)/claim")), fields.New("expected-file", fields.TypeString, fields.WithHelp("JSON map of projection names to expected SHA256 (or missing)")), fields.New("dry-run", fields.TypeBool, fields.WithDefault(false))))}
}
func readOperationJSON(path string, out any) error {
	data, _, err := operations.Read(filepath.Dir(path), filepath.Base(path))
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("expected one JSON value")
	}
	return nil
}
func (c *MilestoneCommand) execute(ctx context.Context, parsed *values.Values) (operations.Receipt, error) {
	s := milestoneSettings{}
	if err := parsed.DecodeSectionInto(schema.DefaultSlug, &s); err != nil {
		return operations.Receipt{}, err
	}
	req := MilestoneRequest{Ticket: s.Ticket, OperationID: s.ID, Summary: s.Summary, Phase: s.Phase, Next: s.Next, TaskIDs: s.Tasks}
	if s.Evidence != "" {
		if err := readOperationJSON(s.Evidence, &req.Evidence); err != nil {
			return operations.Receipt{}, err
		}
	}
	if s.Expected != "" {
		if err := readOperationJSON(s.Expected, &req.Expected); err != nil {
			return operations.Receipt{}, err
		}
	}
	dir, err := operationTicketDir(ctx, s.Root, s.Ticket)
	if err != nil {
		return operations.Receipt{}, err
	}
	return RecordMilestone(ctx, dir, req, s.Dry)
}
func (c *MilestoneCommand) Run(ctx context.Context, parsed *values.Values) error {
	r, err := c.execute(ctx, parsed)
	if err != nil {
		return err
	}
	fmt.Printf("Milestone %s: %s; projections: %v\n", r.OperationID, r.State, r.ChangedPaths)
	return nil
}
func (c *MilestoneCommand) RunIntoGlazeProcessor(ctx context.Context, parsed *values.Values, gp middlewares.Processor) error {
	r, err := c.execute(ctx, parsed)
	if err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("operation_id", r.OperationID), types.MRP("state", r.State), types.MRP("receipt", r)))
}

var _ cmds.BareCommand = &MilestoneCommand{}
var _ cmds.GlazeCommand = &MilestoneCommand{}

type TicketResumeCommand struct{ *cmds.CommandDescription }

func NewTicketResumeCommand() *TicketResumeCommand {
	return &TicketResumeCommand{cmds.NewCommandDescription("resume", cmds.WithShort("Derive current tasks, checkpoint and stale evidence without writing"), cmds.WithFlags(fields.New("root", fields.TypeString, fields.WithDefault("ttmp")), fields.New("ticket", fields.TypeString, fields.WithRequired(true))))}
}
func (c *TicketResumeCommand) execute(ctx context.Context, parsed *values.Values) (ResumeView, error) {
	s := struct {
		Root   string `glazed:"root"`
		Ticket string `glazed:"ticket"`
	}{}
	if err := parsed.DecodeSectionInto(schema.DefaultSlug, &s); err != nil {
		return ResumeView{}, err
	}
	dir, err := operationTicketDir(ctx, s.Root, s.Ticket)
	if err != nil {
		return ResumeView{}, err
	}
	return TicketResume(ctx, dir)
}
func (c *TicketResumeCommand) Run(ctx context.Context, parsed *values.Values) error {
	r, err := c.execute(ctx, parsed)
	if err != nil {
		return err
	}
	fmt.Printf("Ticket %s: %s; phase %s\nNext: %s\n", r.Ticket, r.Status, r.Phase, r.Next)
	for _, task := range r.Remaining {
		fmt.Printf("[ ] %s %s\n", task.StableID, task.Text)
	}
	for _, conflict := range r.Conflicts {
		fmt.Printf("WARNING: %s\n", conflict)
	}
	fmt.Printf("Documents: %v\n", r.Documents)
	return nil
}
func (c *TicketResumeCommand) RunIntoGlazeProcessor(ctx context.Context, parsed *values.Values, gp middlewares.Processor) error {
	r, err := c.execute(ctx, parsed)
	if err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("ticket", r.Ticket), types.MRP("resume", r)))
}

var _ cmds.BareCommand = &TicketResumeCommand{}
var _ cmds.GlazeCommand = &TicketResumeCommand{}
