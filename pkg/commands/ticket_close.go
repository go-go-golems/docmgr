package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// TicketCloseCommand closes a ticket by updating status, optional intent, and changelog
type TicketCloseCommand struct {
	*cmds.CommandDescription
}

// TicketCloseSettings holds the parameters for the ticket close command
type TicketCloseSettings struct {
	Ticket         string `glazed.parameter:"ticket"`
	Root           string `glazed.parameter:"root"`
	Status         string `glazed.parameter:"status"`
	Intent         string `glazed.parameter:"intent"`
	ChangelogEntry string `glazed.parameter:"changelog-entry"`
}

func NewTicketCloseCommand() (*TicketCloseCommand, error) {
	return &TicketCloseCommand{
		CommandDescription: cmds.NewCommandDescription(
			"close",
			cmds.WithShort("Close a ticket by updating status, optional intent, and changelog"),
			cmds.WithLong(`Atomically closes a ticket by:
  • Updating Status (default: "complete", override with --status)
  • Optionally updating Intent (via --intent)
  • Appending a changelog entry (default: "Ticket closed")
  • Updating LastUpdated timestamp

The command checks if all tasks are done and warns if not, but does not fail.

Examples:
  # Close with defaults
  docmgr ticket close --ticket DOCMGR-CLOSE

  # Close with custom status
  docmgr ticket close --ticket DOCMGR-CLOSE --status archived

  # Close with intent and custom changelog message
  docmgr ticket close --ticket DOCMGR-CLOSE --intent long-term --changelog-entry "All tasks completed, ready for review"

  # Structured output for automation
  docmgr ticket close --ticket DOCMGR-CLOSE --with-glaze-output --output json
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Ticket identifier"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"status",
					parameters.ParameterTypeString,
					parameters.WithHelp("Status value (default: 'complete')"),
					parameters.WithDefault("complete"),
				),
				parameters.NewParameterDefinition(
					"intent",
					parameters.ParameterTypeString,
					parameters.WithHelp("Intent value (optional, defaults from config or omitted)"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"changelog-entry",
					parameters.ParameterTypeString,
					parameters.WithHelp("Changelog entry message (default: 'Ticket closed')"),
					parameters.WithDefault("Ticket closed"),
				),
			),
		),
	}, nil
}

func (c *TicketCloseCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	if ctx == nil {
		return fmt.Errorf("nil context")
	}
	settings := &TicketCloseSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Resolve root
	settings.Root = workspace.ResolveRoot(settings.Root)

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return fmt.Errorf("failed to discover workspace: %w", err)
	}
	settings.Root = ws.Context().Root
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	// Find ticket directory (Workspace+QueryDocs-backed)
	ticketDir, err := resolveTicketDirViaWorkspace(ctx, ws, settings.Ticket)
	if err != nil {
		return fmt.Errorf("failed to find ticket directory: %w", err)
	}

	// Check if all tasks are done
	openTasks, doneTasks := countTasksInTicket(ticketDir)
	allTasksDone := openTasks == 0 && (openTasks+doneTasks) > 0

	// Read current index.md
	indexPath := filepath.Join(ticketDir, "index.md")
	doc, content, err := documents.ReadDocumentWithFrontmatter(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read ticket index: %w", err)
	}

	// Track what was updated
	operations := map[string]bool{
		"status_updated":    false,
		"intent_updated":    false,
		"changelog_updated": false,
	}

	// Update status
	if settings.Status != "" {
		doc.Status = settings.Status
		operations["status_updated"] = true
	}

	// Update intent if provided
	if settings.Intent != "" {
		doc.Intent = settings.Intent
		operations["intent_updated"] = true
	}

	// Update LastUpdated
	doc.LastUpdated = time.Now()

	// Write updated index.md
	if err := documents.WriteDocumentWithFrontmatter(indexPath, doc, content, true); err != nil {
		return fmt.Errorf("failed to write ticket index: %w", err)
	}

	// Update changelog
	changelogPath := filepath.Join(ticketDir, "changelog.md")
	changelogEntry := settings.ChangelogEntry
	if changelogEntry == "" {
		changelogEntry = "Ticket closed"
	}

	// Ensure changelog exists
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		_ = os.MkdirAll(filepath.Dir(changelogPath), 0755)
		_ = os.WriteFile(changelogPath, []byte("# Changelog\n\n"), 0644)
	}

	// Append changelog entry
	today := time.Now().Format("2006-01-02")
	entryText := fmt.Sprintf("\n## %s\n\n%s\n\n", today, changelogEntry)
	fp, err := os.OpenFile(changelogPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open changelog: %w", err)
	}
	defer func() { _ = fp.Close() }()
	if _, err := fp.WriteString(entryText); err != nil {
		return fmt.Errorf("failed to write changelog entry: %w", err)
	}
	operations["changelog_updated"] = true

	// Emit structured output
	row := types.NewRow(
		types.MRP("ticket", settings.Ticket),
		types.MRP("all_tasks_done", allTasksDone),
		types.MRP("open_tasks", openTasks),
		types.MRP("done_tasks", doneTasks),
		types.MRP("status", doc.Status),
		types.MRP("intent", doc.Intent),
		types.MRP("operations", operations),
		types.MRP("status_updated", operations["status_updated"]),
		types.MRP("intent_updated", operations["intent_updated"]),
		types.MRP("changelog_updated", operations["changelog_updated"]),
	)
	return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &TicketCloseCommand{}

// Run implements BareCommand for human-friendly output
func (c *TicketCloseCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	if ctx == nil {
		return fmt.Errorf("nil context")
	}
	settings := &TicketCloseSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Resolve root
	settings.Root = workspace.ResolveRoot(settings.Root)

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return fmt.Errorf("failed to discover workspace: %w", err)
	}
	settings.Root = ws.Context().Root
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	// Find ticket directory (Workspace+QueryDocs-backed)
	ticketDir, err := resolveTicketDirViaWorkspace(ctx, ws, settings.Ticket)
	if err != nil {
		return fmt.Errorf("failed to find ticket directory: %w", err)
	}

	// Check if all tasks are done
	openTasks, doneTasks := countTasksInTicket(ticketDir)
	allTasksDone := openTasks == 0 && (openTasks+doneTasks) > 0

	// Warn if not all tasks are done
	if !allTasksDone && (openTasks+doneTasks) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: Not all tasks are done (%d open, %d done). Closing anyway.\n", openTasks, doneTasks)
	}

	// Read current index.md
	indexPath := filepath.Join(ticketDir, "index.md")
	doc, content, err := documents.ReadDocumentWithFrontmatter(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read ticket index: %w", err)
	}

	oldStatus := doc.Status
	oldIntent := doc.Intent

	// Update status
	if settings.Status != "" {
		doc.Status = settings.Status
	}

	// Update intent if provided
	if settings.Intent != "" {
		doc.Intent = settings.Intent
	}

	// Update LastUpdated
	doc.LastUpdated = time.Now()

	// Write updated index.md
	if err := documents.WriteDocumentWithFrontmatter(indexPath, doc, content, true); err != nil {
		return fmt.Errorf("failed to write ticket index: %w", err)
	}

	// Update changelog
	changelogPath := filepath.Join(ticketDir, "changelog.md")
	changelogEntry := settings.ChangelogEntry
	if changelogEntry == "" {
		changelogEntry = "Ticket closed"
	}

	// Ensure changelog exists
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		_ = os.MkdirAll(filepath.Dir(changelogPath), 0755)
		_ = os.WriteFile(changelogPath, []byte("# Changelog\n\n"), 0644)
	}

	// Append changelog entry
	today := time.Now().Format("2006-01-02")
	entryText := fmt.Sprintf("\n## %s\n\n%s\n\n", today, changelogEntry)
	fp, err := os.OpenFile(changelogPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open changelog: %w", err)
	}
	defer func() { _ = fp.Close() }()
	if _, err := fp.WriteString(entryText); err != nil {
		return fmt.Errorf("failed to write changelog entry: %w", err)
	}

	// Print human-friendly output
	var changes []string
	if oldStatus != doc.Status {
		changes = append(changes, fmt.Sprintf("Status: %s → %s", oldStatus, doc.Status))
	}
	if settings.Intent != "" && oldIntent != doc.Intent {
		changes = append(changes, fmt.Sprintf("Intent: %s → %s", oldIntent, doc.Intent))
	}
	changes = append(changes, "Changelog updated", "LastUpdated refreshed")

	fmt.Printf("Ticket %s closed successfully.\n", settings.Ticket)
	if len(changes) > 0 {
		fmt.Printf("Changes: %s\n", strings.Join(changes, ", "))
	}
	fmt.Printf("Changelog: %s\n", changelogPath)

	return nil
}

var _ cmds.BareCommand = &TicketCloseCommand{}
