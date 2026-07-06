package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/go-go-golems/docmgr/internal/tasksmd"
	"github.com/go-go-golems/docmgr/internal/templates"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// Internal task representation
type parsedTask struct {
	TaskIndex int
	LineIndex int
	Checked   bool
	Text      string
	// StableID is the invisible "<!-- t:xxxx -->" marker ID, empty for
	// unmarked (legacy) tasks that are addressed by position only.
	StableID string
}

// DisplayID returns the identifier shown to users: the stable ID when the
// task carries a marker, else the 1-based position.
func (t parsedTask) DisplayID() string {
	if t.StableID != "" {
		return t.StableID
	}
	return strconv.Itoa(t.TaskIndex)
}

// removed unused regex to satisfy linter; parsing handled in code below

func loadTasksFile(ctx context.Context, root string, ticket string, tasksFile string) (string, []string, []parsedTask, error) {
	if ctx == nil {
		return "", nil, nil, fmt.Errorf("nil context")
	}
	var path string
	if tasksFile != "" {
		path = tasksFile
	} else {
		root = workspace.ResolveRoot(root)
		if strings.TrimSpace(ticket) == "" {
			return "", nil, nil, fmt.Errorf("must specify --ticket when --tasks-file is not set")
		}
		p, _, err := findTasksFileViaWorkspace(ctx, root, ticket)
		if err != nil {
			return "", nil, nil, fmt.Errorf("failed to resolve tasks file: %w", err)
		}
		path = p
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to read tasks file: %w", err)
	}
	// Split into lines preserving endings on write via Join("\n")
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	scanner.Split(bufio.ScanLines)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	tasks := parseTasksFromLines(lines)
	return path, lines, tasks, nil
}

func findTasksFileViaWorkspace(ctx context.Context, rootOverride string, ticketID string) (string, string, error) {
	if ctx == nil {
		return "", "", fmt.Errorf("nil context")
	}
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return "", "", fmt.Errorf("empty ticket id")
	}

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: rootOverride})
	if err != nil {
		return "", "", fmt.Errorf("discover workspace: %w", err)
	}
	resolvedRoot := ws.Context().Root

	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return "", resolvedRoot, fmt.Errorf("init workspace index: %w", err)
	}

	// NOTE: tasks.md does not have frontmatter by default, so it may not be indexed by QueryDocs.
	// We therefore derive the ticket directory from the canonical index.md (DocType=index) and
	// then join tasks.md directly.
	ticketDir, err := resolveTicketDirViaWorkspace(ctx, ws, ticketID)
	if err != nil {
		return "", resolvedRoot, err
	}
	return filepath.Join(ticketDir, "tasks.md"), resolvedRoot, nil
}

func parseTasksFromLines(lines []string) []parsedTask {
	tasks := []parsedTask{}
	idx := 0
	for i, l := range lines {
		// match "- [ ] text" or "- [x] text" (case-insensitive)
		if strings.HasPrefix(strings.TrimSpace(l), "- [") || strings.HasPrefix(strings.TrimSpace(l), "* [") {
			// determine checked
			trimmed := strings.TrimSpace(l)
			checked := false
			if strings.HasPrefix(strings.ToLower(trimmed), "- [x]") || strings.HasPrefix(strings.ToLower(trimmed), "* [x]") {
				checked = true
			}
			// extract text after closing bracket
			pos := strings.Index(trimmed, "]")
			text := ""
			if pos >= 0 {
				text = strings.TrimSpace(trimmed[pos+1:])
			}
			stableID, cleanText := tasksmd.ExtractStableID(text)
			idx++
			tasks = append(tasks, parsedTask{TaskIndex: idx, LineIndex: i, Checked: checked, Text: cleanText, StableID: stableID})
		}
	}
	return tasks
}

func formatTaskLine(checked bool, text string, stableID string) string {
	mark := " "
	if checked {
		mark = "x"
	}
	if stableID != "" {
		return fmt.Sprintf("- [%s] %s %s", mark, text, tasksmd.FormatStableIDMarker(stableID))
	}
	return fmt.Sprintf("- [%s] %s", mark, text)
}

// existingStableIDs returns the set of stable IDs already present in tasks.
func existingStableIDs(tasks []parsedTask) map[string]struct{} {
	out := map[string]struct{}{}
	for _, t := range tasks {
		if t.StableID != "" {
			out[t.StableID] = struct{}{}
		}
	}
	return out
}

// renderTaskTable renders the current tasks as a small table so callers that
// passed an unknown --id can see the real identifiers.
func renderTaskTable(tasks []parsedTask) string {
	if len(tasks) == 0 {
		return "  (no tasks)"
	}
	var b strings.Builder
	for _, t := range tasks {
		mark := " "
		if t.Checked {
			mark = "x"
		}
		fmt.Fprintf(&b, "  [%s] [%s] %s\n", t.DisplayID(), mark, t.Text)
	}
	return strings.TrimRight(b.String(), "\n")
}

// resolveTaskRefs maps user-supplied task references (stable IDs or 1-based
// positions) to tasks. Unknown references produce an error that includes the
// current task table, so callers can see the real IDs.
func resolveTaskRefs(tasks []parsedTask, refs []string) ([]parsedTask, error) {
	byStableID := map[string]parsedTask{}
	byIndex := map[int]parsedTask{}
	for _, t := range tasks {
		if t.StableID != "" {
			byStableID[t.StableID] = t
		}
		byIndex[t.TaskIndex] = t
	}

	var out []parsedTask
	seen := map[int]struct{}{} // dedupe by line index
	var unknown []string
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		var t parsedTask
		var ok bool
		if t, ok = byStableID[ref]; !ok {
			if n, err := strconv.Atoi(ref); err == nil {
				t, ok = byIndex[n]
			}
		}
		if !ok {
			unknown = append(unknown, ref)
			continue
		}
		if _, dup := seen[t.LineIndex]; dup {
			continue
		}
		seen[t.LineIndex] = struct{}{}
		out = append(out, t)
	}
	if len(unknown) > 0 {
		return nil, fmt.Errorf("task id(s) not found: %s\nuse a stable id (e.g. ab12) or the 1-based position from 'docmgr task list'. Current tasks:\n%s",
			strings.Join(unknown, ", "), renderTaskTable(tasks))
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no target task specified")
	}
	return out, nil
}

func displayIDList(tasks []parsedTask) string {
	ids := make([]string, 0, len(tasks))
	for _, t := range tasks {
		ids = append(ids, t.DisplayID())
	}
	return strings.Join(ids, ",")
}

// tasks list
type TasksListCommand struct{ *cmds.CommandDescription }

type TasksListSettings struct {
	Ticket              string `glazed:"ticket"`
	Root                string `glazed:"root"`
	TasksFile           string `glazed:"tasks-file"`
	PrintTemplateSchema bool   `glazed:"print-template-schema"`
	SchemaFormat        string `glazed:"schema-format"`
}

func NewTasksListCommand() (*TasksListCommand, error) {
	cmd := cmds.NewCommandDescription(
		"list",
		cmds.WithShort("List tasks from tasks.md"),
		cmds.WithLong(`List checkbox tasks found in the ticket's tasks.md.

Columns:
  id,index,checked,text

'id' is the stable task ID (from the invisible '<!-- t:xxxx -->' marker) when
present, else the 1-based position. Stamp markers onto old task files with
'docmgr task migrate --ticket <ID>'.

Examples:
  # Human output
  docmgr task list --ticket MEN-4242

  # Scriptable (CSV without headers)
  docmgr task list --ticket MEN-4242 --with-glaze-output --output csv --with-headers=false --fields id,text
`),
		cmds.WithFlags(
			fields.New("ticket", fields.TypeString, fields.WithHelp("Ticket identifier (if --tasks-file not set)"), fields.WithDefault("")),
			fields.New("root", fields.TypeString, fields.WithHelp("Root directory for docs"), fields.WithDefault("ttmp")),
			fields.New("tasks-file", fields.TypeString, fields.WithHelp("Path to tasks.md (overrides --ticket)"), fields.WithDefault("")),
			fields.New("print-template-schema", fields.TypeBool, fields.WithHelp("Print template schema after output (human mode only)"), fields.WithDefault(false)),
			fields.New("schema-format", fields.TypeString, fields.WithHelp("Template schema output format: json|yaml"), fields.WithDefault("json")),
		),
	)
	return &TasksListCommand{CommandDescription: cmd}, nil
}

func (c *TasksListCommand) RunIntoGlazeProcessor(ctx context.Context, pl *values.Values, gp middlewares.Processor) error {
	s := &TasksListSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse tasks list settings: %w", err)
	}

	// Apply config root if present
	s.Root = workspace.ResolveRoot(s.Root)

	// If only printing template schema, skip all other processing and output
	if s.PrintTemplateSchema {
		type TaskInfo struct {
			ID      string
			Index   int
			Checked bool
			Text    string
		}
		templateData := map[string]interface{}{
			"TotalTasks": 0,
			"OpenTasks":  0,
			"DoneTasks":  0,
			"TasksFile":  "",
			"Tasks": []TaskInfo{
				{
					ID:      "",
					Index:   0,
					Checked: false,
					Text:    "",
				},
			},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, s.SchemaFormat)
		return nil
	}

	_, _, tasks, err := loadTasksFile(ctx, s.Root, s.Ticket, s.TasksFile)
	if err != nil {
		return fmt.Errorf("failed to load tasks from file: %w", err)
	}
	for _, t := range tasks {
		row := types.NewRow(
			types.MRP("id", t.DisplayID()),
			types.MRP(ColIndex, t.TaskIndex),
			types.MRP(ColChecked, t.Checked),
			types.MRP(ColText, t.Text),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to emit tasks list row %d: %w", t.TaskIndex, err)
		}
	}
	return nil
}

var _ cmds.GlazeCommand = &TasksListCommand{}

// Implement BareCommand for human-friendly output
func (c *TasksListCommand) Run(ctx context.Context, pl *values.Values) error {
	s := &TasksListSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse tasks list settings: %w", err)
	}

	// Apply config root if present
	s.Root = workspace.ResolveRoot(s.Root)

	// If only printing template schema, skip all other processing and output
	if s.PrintTemplateSchema {
		type TaskInfo struct {
			ID      string
			Index   int
			Checked bool
			Text    string
		}
		templateData := map[string]interface{}{
			"TotalTasks": 0,
			"OpenTasks":  0,
			"DoneTasks":  0,
			"TasksFile":  "",
			"Tasks": []TaskInfo{
				{
					ID:      "",
					Index:   0,
					Checked: false,
					Text:    "",
				},
			},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, s.SchemaFormat)
		return nil
	}

	path, _, tasks, err := loadTasksFile(ctx, s.Root, s.Ticket, s.TasksFile)
	if err != nil {
		return fmt.Errorf("failed to load tasks from file: %w", err)
	}
	if len(tasks) == 0 {
		target := "--ticket " + s.Ticket
		if strings.TrimSpace(s.Ticket) == "" {
			target = "--tasks-file " + path
		}
		fmt.Printf("No tasks yet (%s). Add one with: docmgr task add %s --text \"...\"\n", path, target)
	}
	for _, t := range tasks {
		mark := " "
		if t.Checked {
			mark = "x"
		}
		fmt.Printf("[%s] [%s] %s\n", t.DisplayID(), mark, t.Text)
	}

	// Render postfix template if it exists
	// Build template data struct
	type TaskInfo struct {
		ID      string
		Index   int
		Checked bool
		Text    string
	}

	taskInfos := make([]TaskInfo, 0, len(tasks))
	openTasks := 0
	doneTasks := 0
	for _, t := range tasks {
		taskInfos = append(taskInfos, TaskInfo{
			ID:      t.DisplayID(),
			Index:   t.TaskIndex,
			Checked: t.Checked,
			Text:    t.Text,
		})
		if t.Checked {
			doneTasks++
		} else {
			openTasks++
		}
	}

	templateData := map[string]interface{}{
		"TotalTasks": len(tasks),
		"OpenTasks":  openTasks,
		"DoneTasks":  doneTasks,
		"TasksFile":  path,
		"Tasks":      taskInfos,
	}

	// Try verb path: ["tasks", "list"]
	verbCandidates := [][]string{
		{"tasks", "list"},
	}
	settingsMap := map[string]interface{}{
		"root":      s.Root,
		"ticket":    s.Ticket,
		"tasksFile": s.TasksFile,
	}
	absRoot := s.Root
	if abs, err := filepath.Abs(s.Root); err == nil {
		absRoot = abs
	}
	_ = templates.RenderVerbTemplate(verbCandidates, absRoot, settingsMap, templateData)

	return nil
}

var _ cmds.BareCommand = &TasksListCommand{}

// tasks add
type TasksAddCommand struct{ *cmds.CommandDescription }

type TasksAddSettings struct {
	Ticket    string `glazed:"ticket"`
	Root      string `glazed:"root"`
	TasksFile string `glazed:"tasks-file"`
	Text      string `glazed:"text"`
	After     int    `glazed:"after"`
}

func NewTasksAddCommand() (*TasksAddCommand, error) {
	cmd := cmds.NewCommandDescription(
		"add",
		cmds.WithShort("Add a task to tasks.md"),
		cmds.WithLong(`Add a new checkbox task to the ticket's tasks.md.

Examples:
  # Append a new task
  docmgr task add --ticket MEN-4242 --text "Write design doc"

  # Insert after an existing task index
  docmgr task add --ticket MEN-4242 --text "Add tests" --after 1
`),
		cmds.WithFlags(
			fields.New("ticket", fields.TypeString, fields.WithHelp("Ticket identifier (if --tasks-file not set)"), fields.WithDefault("")),
			fields.New("root", fields.TypeString, fields.WithHelp("Root directory for docs"), fields.WithDefault("ttmp")),
			fields.New("tasks-file", fields.TypeString, fields.WithHelp("Path to tasks.md (overrides --ticket)"), fields.WithDefault("")),
			fields.New("text", fields.TypeString, fields.WithHelp("Task text to add"), fields.WithRequired(true)),
			fields.New("after", fields.TypeInteger, fields.WithHelp("Insert after given task index (0=append)"), fields.WithDefault(0)),
		),
	)
	return &TasksAddCommand{CommandDescription: cmd}, nil
}

// applyTaskAdd inserts the new task line and returns the tasks file path, the
// 1-based index of the added task, and its stable ID.
func (c *TasksAddCommand) applyTaskAdd(ctx context.Context, s *TasksAddSettings) (string, int, string, error) {
	path, lines, tasks, err := loadTasksFile(ctx, s.Root, s.Ticket, s.TasksFile)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to load tasks file: %w", err)
	}
	stableID := tasksmd.NewStableID(existingStableIDs(tasks))
	newLine := formatTaskLine(false, strings.TrimSpace(s.Text), stableID)
	if s.After <= 0 || len(tasks) == 0 {
		lines = append(lines, newLine)
	} else {
		insertAt := len(lines)
		for _, t := range tasks {
			if t.TaskIndex == s.After {
				insertAt = t.LineIndex + 1
			}
		}
		if insertAt >= len(lines) {
			lines = append(lines, newLine)
		} else {
			lines = append(lines[:insertAt], append([]string{newLine}, lines[insertAt:]...)...)
		}
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		return "", 0, "", fmt.Errorf("failed to write tasks file %s: %w", path, err)
	}
	newIndex := 0
	for _, t := range parseTasksFromLines(lines) {
		if t.StableID == stableID {
			newIndex = t.TaskIndex
		}
	}
	return path, newIndex, stableID, nil
}

func (c *TasksAddCommand) Run(ctx context.Context, pl *values.Values) error {
	s := &TasksAddSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse tasks add settings: %w", err)
	}
	path, _, stableID, err := c.applyTaskAdd(ctx, s)
	if err != nil {
		return err
	}
	fmt.Printf("Task %s added to %s\n", stableID, path)
	printReminder("Reminder: update the changelog and relate changed files with notes if needed.")
	return nil
}

func (c *TasksAddCommand) RunIntoGlazeProcessor(ctx context.Context, pl *values.Values, gp middlewares.Processor) error {
	s := &TasksAddSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse tasks add settings: %w", err)
	}
	path, newIndex, stableID, err := c.applyTaskAdd(ctx, s)
	if err != nil {
		return err
	}
	row := types.NewRow(
		types.MRP("ticket", s.Ticket),
		types.MRP("tasks_file", path),
		types.MRP("id", stableID),
		types.MRP("index", newIndex),
		types.MRP("text", s.Text),
		types.MRP("status", "added"),
	)
	return gp.AddRow(ctx, row)
}

var _ cmds.BareCommand = &TasksAddCommand{}
var _ cmds.GlazeCommand = &TasksAddCommand{}

// tasks check
type TasksCheckCommand struct{ *cmds.CommandDescription }

type TasksCheckSettings struct {
	Ticket    string   `glazed:"ticket"`
	Root      string   `glazed:"root"`
	TasksFile string   `glazed:"tasks-file"`
	IDs       []string `glazed:"id"`
	Match     string   `glazed:"match"`
}

func NewTasksCheckCommand() (*TasksCheckCommand, error) {
	cmd := cmds.NewCommandDescription(
		"check",
		cmds.WithShort("Mark a task as done"),
		cmds.WithLong(`Mark a checkbox task as completed in tasks.md.

--id accepts stable task IDs (e.g. ab12, shown by 'task list') or 1-based
positions.

Examples:
  docmgr task check --ticket MEN-4242 --id ab12
  docmgr task check --ticket MEN-4242 --id 1,2
`),
		cmds.WithFlags(
			fields.New("ticket", fields.TypeString, fields.WithHelp("Ticket identifier (if --tasks-file not set)"), fields.WithDefault("")),
			fields.New("root", fields.TypeString, fields.WithHelp("Root directory for docs"), fields.WithDefault("ttmp")),
			fields.New("tasks-file", fields.TypeString, fields.WithHelp("Path to tasks.md (overrides --ticket)"), fields.WithDefault("")),
			fields.New("id", fields.TypeStringList, fields.WithHelp("Task id(s): stable ID or 1-based position, comma-separated (from 'tasks list')")),
			fields.New("match", fields.TypeString, fields.WithHelp("Substring to match a task if --id not set"), fields.WithDefault("")),
		),
	)
	return &TasksCheckCommand{CommandDescription: cmd}, nil
}

// applyTaskMark sets the checked state for the targeted tasks. It returns the
// tasks file path, the affected tasks, and the updated task list.
func applyTaskMark(ctx context.Context, root string, ticket string, tasksFile string, ids []string, match string, checked bool) (string, []parsedTask, []parsedTask, error) {
	path, lines, tasks, err := loadTasksFile(ctx, root, ticket, tasksFile)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to load tasks file: %w", err)
	}
	var targets []parsedTask
	if len(ids) > 0 {
		targets, err = resolveTaskRefs(tasks, ids)
		if err != nil {
			return "", nil, nil, err
		}
	} else if match != "" {
		for _, t := range tasks {
			if strings.Contains(strings.ToLower(t.Text), strings.ToLower(match)) {
				targets = []parsedTask{t}
				break
			}
		}
	}
	if len(targets) == 0 {
		return "", nil, nil, fmt.Errorf("no target task specified")
	}
	for _, t := range targets {
		lines[t.LineIndex] = formatTaskLine(checked, t.Text, t.StableID)
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		return "", nil, nil, fmt.Errorf("failed to write tasks file %s: %w", path, err)
	}
	return path, targets, parseTasksFromLines(lines), nil
}

func (c *TasksCheckCommand) Run(ctx context.Context, pl *values.Values) error {
	s := &TasksCheckSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse tasks check settings: %w", err)
	}
	path, targets, updatedTasks, err := applyTaskMark(ctx, s.Root, s.Ticket, s.TasksFile, s.IDs, s.Match, true)
	if err != nil {
		return err
	}
	if len(targets) > 1 {
		fmt.Printf("Tasks checked: %s (file=%s)\n", displayIDList(targets), path)
	} else {
		fmt.Printf("Task checked: %s (file=%s)\n", displayIDList(targets), path)
	}

	// Check if all tasks are now done
	allDone := true
	hasTasks := len(updatedTasks) > 0
	for _, t := range updatedTasks {
		if !t.Checked {
			allDone = false
			break
		}
	}

	// Suggest ticket close if all tasks are done
	if allDone && hasTasks && s.Ticket != "" {
		fmt.Printf("All tasks complete! Consider closing the ticket: docmgr ticket close --ticket %s\n", s.Ticket)
	} else {
		printReminder("Reminder: update the changelog and relate changed files with notes if needed.")
	}
	return nil
}

func (c *TasksCheckCommand) RunIntoGlazeProcessor(ctx context.Context, pl *values.Values, gp middlewares.Processor) error {
	s := &TasksCheckSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse tasks check settings: %w", err)
	}
	path, targets, updatedTasks, err := applyTaskMark(ctx, s.Root, s.Ticket, s.TasksFile, s.IDs, s.Match, true)
	if err != nil {
		return err
	}

	// Check if all tasks are now done
	allDone := true
	hasTasks := len(updatedTasks) > 0
	openTasks := 0
	doneTasks := 0
	for _, t := range updatedTasks {
		if t.Checked {
			doneTasks++
		} else {
			openTasks++
			allDone = false
		}
	}

	// Emit structured output
	row := types.NewRow(
		types.MRP("ticket", s.Ticket),
		types.MRP("tasks_file", path),
		types.MRP("all_tasks_done", allDone && hasTasks),
		types.MRP("open_tasks", openTasks),
		types.MRP("done_tasks", doneTasks),
		types.MRP("total_tasks", len(updatedTasks)),
		types.MRP("checked_ids", displayIDList(targets)),
	)
	return gp.AddRow(ctx, row)
}

var _ cmds.BareCommand = &TasksCheckCommand{}
var _ cmds.GlazeCommand = &TasksCheckCommand{}

// tasks uncheck
type TasksUncheckCommand struct{ *cmds.CommandDescription }

type TasksUncheckSettings struct {
	Ticket    string   `glazed:"ticket"`
	Root      string   `glazed:"root"`
	TasksFile string   `glazed:"tasks-file"`
	IDs       []string `glazed:"id"`
	Match     string   `glazed:"match"`
}

func NewTasksUncheckCommand() (*TasksUncheckCommand, error) {
	cmd := cmds.NewCommandDescription(
		"uncheck",
		cmds.WithShort("Mark a task as not done"),
		cmds.WithLong(`Mark a checkbox task as incomplete in tasks.md.

--id accepts stable task IDs (e.g. ab12, shown by 'task list') or 1-based
positions.

Examples:
  docmgr task uncheck --ticket MEN-4242 --id ab12
  docmgr task uncheck --ticket MEN-4242 --id 1,2
`),
		cmds.WithFlags(
			fields.New("ticket", fields.TypeString, fields.WithHelp("Ticket identifier (if --tasks-file not set)"), fields.WithDefault("")),
			fields.New("root", fields.TypeString, fields.WithHelp("Root directory for docs"), fields.WithDefault("ttmp")),
			fields.New("tasks-file", fields.TypeString, fields.WithHelp("Path to tasks.md (overrides --ticket)"), fields.WithDefault("")),
			fields.New("id", fields.TypeStringList, fields.WithHelp("Task id(s): stable ID or 1-based position, comma-separated (from 'tasks list')")),
			fields.New("match", fields.TypeString, fields.WithHelp("Substring to match a task if --id not set"), fields.WithDefault("")),
		),
	)
	return &TasksUncheckCommand{CommandDescription: cmd}, nil
}

func (c *TasksUncheckCommand) Run(ctx context.Context, pl *values.Values) error {
	s := &TasksUncheckSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse tasks uncheck settings: %w", err)
	}
	path, targets, _, err := applyTaskMark(ctx, s.Root, s.Ticket, s.TasksFile, s.IDs, s.Match, false)
	if err != nil {
		return err
	}
	if len(targets) > 1 {
		fmt.Printf("Tasks unchecked: %s (file=%s)\n", displayIDList(targets), path)
	} else {
		fmt.Printf("Task unchecked: %s (file=%s)\n", displayIDList(targets), path)
	}
	printReminder("Reminder: update the changelog and relate changed files with notes if needed.")
	return nil
}

func (c *TasksUncheckCommand) RunIntoGlazeProcessor(ctx context.Context, pl *values.Values, gp middlewares.Processor) error {
	s := &TasksUncheckSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse tasks uncheck settings: %w", err)
	}
	path, targets, updatedTasks, err := applyTaskMark(ctx, s.Root, s.Ticket, s.TasksFile, s.IDs, s.Match, false)
	if err != nil {
		return err
	}
	openTasks := 0
	doneTasks := 0
	for _, t := range updatedTasks {
		if t.Checked {
			doneTasks++
		} else {
			openTasks++
		}
	}
	row := types.NewRow(
		types.MRP("ticket", s.Ticket),
		types.MRP("tasks_file", path),
		types.MRP("open_tasks", openTasks),
		types.MRP("done_tasks", doneTasks),
		types.MRP("total_tasks", len(updatedTasks)),
		types.MRP("unchecked_ids", displayIDList(targets)),
	)
	return gp.AddRow(ctx, row)
}

var _ cmds.BareCommand = &TasksUncheckCommand{}
var _ cmds.GlazeCommand = &TasksUncheckCommand{}

// tasks edit
type TasksEditCommand struct{ *cmds.CommandDescription }

type TasksEditSettings struct {
	Ticket    string `glazed:"ticket"`
	Root      string `glazed:"root"`
	TasksFile string `glazed:"tasks-file"`
	ID        string `glazed:"id"`
	Text      string `glazed:"text"`
}

func NewTasksEditCommand() (*TasksEditCommand, error) {
	cmd := cmds.NewCommandDescription(
		"edit",
		cmds.WithShort("Edit a task's text"),
		cmds.WithLong(`Edit the text of a checkbox task in tasks.md.

--id accepts a stable task ID (e.g. ab12, shown by 'task list') or a 1-based
position.

Examples:
  docmgr task edit --ticket MEN-4242 --id ab12 --text "Updated task text"
`),
		cmds.WithFlags(
			fields.New("ticket", fields.TypeString, fields.WithHelp("Ticket identifier (if --tasks-file not set)"), fields.WithDefault("")),
			fields.New("root", fields.TypeString, fields.WithHelp("Root directory for docs"), fields.WithDefault("ttmp")),
			fields.New("tasks-file", fields.TypeString, fields.WithHelp("Path to tasks.md (overrides --ticket)"), fields.WithDefault("")),
			fields.New("id", fields.TypeString, fields.WithHelp("Task id: stable ID or 1-based position (from 'tasks list')"), fields.WithRequired(true)),
			fields.New("text", fields.TypeString, fields.WithHelp("New task text"), fields.WithRequired(true)),
		),
	)
	return &TasksEditCommand{CommandDescription: cmd}, nil
}

func (c *TasksEditCommand) RunIntoGlazeProcessor(ctx context.Context, pl *values.Values, gp middlewares.Processor) error {
	s := &TasksEditSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse tasks edit settings: %w", err)
	}
	path, err := editTaskLine(ctx, s)
	if err != nil {
		return err
	}
	row := types.NewRow(types.MRP("file", path), types.MRP("status", "task edited"), types.MRP("id", s.ID))
	if err := gp.AddRow(ctx, row); err != nil {
		return fmt.Errorf("failed to emit tasks edit row for %s id %s: %w", path, s.ID, err)
	}
	return nil
}

var _ cmds.GlazeCommand = &TasksEditCommand{}

func (c *TasksEditCommand) Run(ctx context.Context, pl *values.Values) error {
	s := &TasksEditSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse tasks edit settings: %w", err)
	}
	path, err := editTaskLine(ctx, s)
	if err != nil {
		return err
	}
	fmt.Printf("Task %s updated in %s\n", s.ID, path)
	printReminder("Reminder: update the changelog and relate changed files with notes if needed.")
	return nil
}

var _ cmds.BareCommand = &TasksEditCommand{}

func editTaskLine(ctx context.Context, s *TasksEditSettings) (string, error) {
	path, lines, tasks, err := loadTasksFile(ctx, s.Root, s.Ticket, s.TasksFile)
	if err != nil {
		return "", fmt.Errorf("failed to load tasks file: %w", err)
	}
	targets, err := resolveTaskRefs(tasks, []string{s.ID})
	if err != nil {
		return "", err
	}
	target := targets[0]
	lines[target.LineIndex] = formatTaskLine(target.Checked, strings.TrimSpace(s.Text), target.StableID)
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		return "", fmt.Errorf("failed to write tasks file %s: %w", path, err)
	}
	return path, nil
}

// tasks remove
type TasksRemoveCommand struct{ *cmds.CommandDescription }

type TasksRemoveSettings struct {
	Ticket    string   `glazed:"ticket"`
	Root      string   `glazed:"root"`
	TasksFile string   `glazed:"tasks-file"`
	IDs       []string `glazed:"id"`
}

func NewTasksRemoveCommand() (*TasksRemoveCommand, error) {
	cmd := cmds.NewCommandDescription(
		"remove",
		cmds.WithShort("Remove a task"),
		cmds.WithLong(`Remove a checkbox task from tasks.md.

--id accepts stable task IDs (e.g. ab12, shown by 'task list') or 1-based
positions.

Examples:
  docmgr task remove --ticket MEN-4242 --id ab12
  docmgr task remove --ticket MEN-4242 --id 3,4
`),
		cmds.WithFlags(
			fields.New("ticket", fields.TypeString, fields.WithHelp("Ticket identifier (if --tasks-file not set)"), fields.WithDefault("")),
			fields.New("root", fields.TypeString, fields.WithHelp("Root directory for docs"), fields.WithDefault("ttmp")),
			fields.New("tasks-file", fields.TypeString, fields.WithHelp("Path to tasks.md (overrides --ticket)"), fields.WithDefault("")),
			fields.New("id", fields.TypeStringList, fields.WithHelp("Task id(s): stable ID or 1-based position, comma-separated (from 'tasks list')"), fields.WithRequired(true)),
		),
	)
	return &TasksRemoveCommand{CommandDescription: cmd}, nil
}

// applyTaskRemove deletes the targeted task lines and returns the tasks file
// path, the removed tasks, and remaining tasks.
func (c *TasksRemoveCommand) applyTaskRemove(ctx context.Context, s *TasksRemoveSettings) (string, []parsedTask, []parsedTask, error) {
	path, lines, tasks, err := loadTasksFile(ctx, s.Root, s.Ticket, s.TasksFile)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to load tasks file: %w", err)
	}
	if len(s.IDs) == 0 {
		return "", nil, nil, fmt.Errorf("no target task specified")
	}
	targets, err := resolveTaskRefs(tasks, s.IDs)
	if err != nil {
		return "", nil, nil, err
	}
	lineIdxs := make([]int, 0, len(targets))
	for _, t := range targets {
		lineIdxs = append(lineIdxs, t.LineIndex)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(lineIdxs)))
	newLines := append([]string{}, lines...)
	for _, idx := range lineIdxs {
		newLines = append(newLines[:idx], newLines[idx+1:]...)
	}
	if err := os.WriteFile(path, []byte(strings.Join(newLines, "\n")+"\n"), 0644); err != nil {
		return "", nil, nil, fmt.Errorf("failed to write tasks file %s: %w", path, err)
	}
	return path, targets, parseTasksFromLines(newLines), nil
}

func (c *TasksRemoveCommand) Run(ctx context.Context, pl *values.Values) error {
	s := &TasksRemoveSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse tasks remove settings: %w", err)
	}
	path, removed, _, err := c.applyTaskRemove(ctx, s)
	if err != nil {
		return err
	}
	if len(removed) > 1 {
		fmt.Printf("Tasks removed: %s (file=%s)\n", displayIDList(removed), path)
	} else {
		fmt.Printf("Task removed: %s (file=%s)\n", displayIDList(removed), path)
	}
	return nil
}

func (c *TasksRemoveCommand) RunIntoGlazeProcessor(ctx context.Context, pl *values.Values, gp middlewares.Processor) error {
	s := &TasksRemoveSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse tasks remove settings: %w", err)
	}
	path, removed, remaining, err := c.applyTaskRemove(ctx, s)
	if err != nil {
		return err
	}
	row := types.NewRow(
		types.MRP("ticket", s.Ticket),
		types.MRP("tasks_file", path),
		types.MRP("removed_ids", displayIDList(removed)),
		types.MRP("total_tasks", len(remaining)),
		types.MRP("status", "removed"),
	)
	return gp.AddRow(ctx, row)
}

var _ cmds.BareCommand = &TasksRemoveCommand{}
var _ cmds.GlazeCommand = &TasksRemoveCommand{}

// tasks migrate: stamp stable ID markers onto unmarked tasks.
type TasksMigrateCommand struct{ *cmds.CommandDescription }

type TasksMigrateSettings struct {
	Ticket    string `glazed:"ticket"`
	Root      string `glazed:"root"`
	TasksFile string `glazed:"tasks-file"`
}

func NewTasksMigrateCommand() (*TasksMigrateCommand, error) {
	cmd := cmds.NewCommandDescription(
		"migrate",
		cmds.WithShort("Stamp stable IDs onto tasks that lack them"),
		cmds.WithLong(`Add invisible stable-ID markers ('<!-- t:xxxx -->') to every checkbox task
in tasks.md that does not have one yet. Tasks created by 'docmgr task add'
already carry markers; this migrates older, hand-written task lists so that
'task check/uncheck/edit/remove --id' can address tasks by stable ID instead
of position.

Examples:
  docmgr task migrate --ticket MEN-4242
`),
		cmds.WithFlags(
			fields.New("ticket", fields.TypeString, fields.WithHelp("Ticket identifier (if --tasks-file not set)"), fields.WithDefault("")),
			fields.New("root", fields.TypeString, fields.WithHelp("Root directory for docs"), fields.WithDefault("ttmp")),
			fields.New("tasks-file", fields.TypeString, fields.WithHelp("Path to tasks.md (overrides --ticket)"), fields.WithDefault("")),
		),
	)
	return &TasksMigrateCommand{CommandDescription: cmd}, nil
}

// applyTaskMigrate stamps markers and returns the tasks file path, the number
// of stamped tasks, and the updated task list.
func applyTaskMigrate(ctx context.Context, root string, ticket string, tasksFile string) (string, int, []parsedTask, error) {
	path, lines, tasks, err := loadTasksFile(ctx, root, ticket, tasksFile)
	if err != nil {
		return "", 0, nil, fmt.Errorf("failed to load tasks file: %w", err)
	}
	existing := existingStableIDs(tasks)
	stamped := 0
	for _, t := range tasks {
		if t.StableID != "" {
			continue
		}
		id := tasksmd.NewStableID(existing)
		existing[id] = struct{}{}
		lines[t.LineIndex] = formatTaskLine(t.Checked, t.Text, id)
		stamped++
	}
	if stamped > 0 {
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
			return "", 0, nil, fmt.Errorf("failed to write tasks file %s: %w", path, err)
		}
	}
	return path, stamped, parseTasksFromLines(lines), nil
}

func (c *TasksMigrateCommand) Run(ctx context.Context, pl *values.Values) error {
	s := &TasksMigrateSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse tasks migrate settings: %w", err)
	}
	path, stamped, tasks, err := applyTaskMigrate(ctx, s.Root, s.Ticket, s.TasksFile)
	if err != nil {
		return err
	}
	if stamped == 0 {
		fmt.Printf("All %d task(s) already have stable IDs (file=%s)\n", len(tasks), path)
		return nil
	}
	fmt.Printf("Stamped stable IDs onto %d task(s) (file=%s)\n", stamped, path)
	return nil
}

func (c *TasksMigrateCommand) RunIntoGlazeProcessor(ctx context.Context, pl *values.Values, gp middlewares.Processor) error {
	s := &TasksMigrateSettings{}
	if err := pl.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse tasks migrate settings: %w", err)
	}
	path, stamped, tasks, err := applyTaskMigrate(ctx, s.Root, s.Ticket, s.TasksFile)
	if err != nil {
		return err
	}
	row := types.NewRow(
		types.MRP("ticket", s.Ticket),
		types.MRP("tasks_file", path),
		types.MRP("stamped", stamped),
		types.MRP("total_tasks", len(tasks)),
		types.MRP("status", "migrated"),
	)
	return gp.AddRow(ctx, row)
}

var _ cmds.BareCommand = &TasksMigrateCommand{}
var _ cmds.GlazeCommand = &TasksMigrateCommand{}
