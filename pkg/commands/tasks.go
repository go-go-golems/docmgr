package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// Internal task representation
type parsedTask struct {
	TaskIndex int
	LineIndex int
	Checked   bool
	Text      string
}

// removed unused regex to satisfy linter; parsing handled in code below

func loadTasksFile(root string, ticket string, tasksFile string) (string, []string, []parsedTask, error) {
	var path string
	if tasksFile != "" {
		path = tasksFile
	} else {
		root = ResolveRoot(root)
		// Prefer simple name-based match to avoid false positives with guideline folders
		// and non-ticket directories that may contain an index.md without frontmatter.
		td := ""
		if ticket != "" {
			if entries, err := os.ReadDir(root); err == nil {
				for _, e := range entries {
					if e.IsDir() && strings.Contains(strings.ToLower(e.Name()), strings.ToLower(ticket)) {
						td = filepath.Join(root, e.Name())
						break
					}
				}
			}
		}
		if td == "" {
			// Fallback to metadata-based resolution
			var err error
			td, err = findTicketDirectory(root, ticket)
			if err != nil {
				return "", nil, nil, fmt.Errorf("failed to find ticket directory: %w", err)
			}
			path = filepath.Join(td, "tasks.md")
		} else {
			path = filepath.Join(td, "tasks.md")
		}
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
			idx++
			tasks = append(tasks, parsedTask{TaskIndex: idx, LineIndex: i, Checked: checked, Text: text})
		}
	}
	return tasks
}

func formatTaskLine(checked bool, text string) string {
	mark := " "
	if checked {
		mark = "x"
	}
	return fmt.Sprintf("- [%s] %s", mark, text)
}

// tasks list
type TasksListCommand struct{ *cmds.CommandDescription }

type TasksListSettings struct {
	Ticket    string `glazed.parameter:"ticket"`
	Root      string `glazed.parameter:"root"`
	TasksFile string `glazed.parameter:"tasks-file"`
}

func NewTasksListCommand() (*TasksListCommand, error) {
	cmd := cmds.NewCommandDescription(
		"list",
		cmds.WithShort("List tasks from tasks.md"),
		cmds.WithLong(`List checkbox tasks found in the ticket's tasks.md.

Columns:
  index,checked,text,file

Examples:
  # Human output
  docmgr tasks list --ticket MEN-4242

  # Scriptable (CSV without headers)
  docmgr tasks list --ticket MEN-4242 --with-glaze-output --output csv --with-headers=false --fields index,text
`),
		cmds.WithFlags(
			parameters.NewParameterDefinition("ticket", parameters.ParameterTypeString, parameters.WithHelp("Ticket identifier (if --tasks-file not set)"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("root", parameters.ParameterTypeString, parameters.WithHelp("Root directory for docs"), parameters.WithDefault("ttmp")),
			parameters.NewParameterDefinition("tasks-file", parameters.ParameterTypeString, parameters.WithHelp("Path to tasks.md (overrides --ticket)"), parameters.WithDefault("")),
		),
	)
	return &TasksListCommand{CommandDescription: cmd}, nil
}

func (c *TasksListCommand) RunIntoGlazeProcessor(ctx context.Context, pl *layers.ParsedLayers, gp middlewares.Processor) error {
	s := &TasksListSettings{}
	if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}
	path, _, tasks, err := loadTasksFile(s.Root, s.Ticket, s.TasksFile)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		row := types.NewRow(
			types.MRP(ColIndex, t.TaskIndex),
			types.MRP(ColChecked, t.Checked),
			types.MRP(ColText, t.Text),
			types.MRP(ColFile, path),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

var _ cmds.GlazeCommand = &TasksListCommand{}

// Implement BareCommand for human-friendly output
func (c *TasksListCommand) Run(ctx context.Context, pl *layers.ParsedLayers) error {
	s := &TasksListSettings{}
	if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}
	path, _, tasks, err := loadTasksFile(s.Root, s.Ticket, s.TasksFile)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		mark := " "
		if t.Checked {
			mark = "x"
		}
		fmt.Printf("[%d] [%s] %s (file=%s)\n", t.TaskIndex, mark, t.Text, path)
	}
	return nil
}

var _ cmds.BareCommand = &TasksListCommand{}

// tasks add
type TasksAddCommand struct{ *cmds.CommandDescription }

type TasksAddSettings struct {
	Ticket    string `glazed.parameter:"ticket"`
	Root      string `glazed.parameter:"root"`
	TasksFile string `glazed.parameter:"tasks-file"`
	Text      string `glazed.parameter:"text"`
	After     int    `glazed.parameter:"after"`
}

func NewTasksAddCommand() (*TasksAddCommand, error) {
	cmd := cmds.NewCommandDescription(
		"add",
		cmds.WithShort("Add a task to tasks.md"),
		cmds.WithLong(`Add a new checkbox task to the ticket's tasks.md.`),
		cmds.WithFlags(
			parameters.NewParameterDefinition("ticket", parameters.ParameterTypeString, parameters.WithHelp("Ticket identifier (if --tasks-file not set)"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("root", parameters.ParameterTypeString, parameters.WithHelp("Root directory for docs"), parameters.WithDefault("ttmp")),
			parameters.NewParameterDefinition("tasks-file", parameters.ParameterTypeString, parameters.WithHelp("Path to tasks.md (overrides --ticket)"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("text", parameters.ParameterTypeString, parameters.WithHelp("Task text to add"), parameters.WithRequired(true)),
			parameters.NewParameterDefinition("after", parameters.ParameterTypeInteger, parameters.WithHelp("Insert after given task index (0=append)"), parameters.WithDefault(0)),
		),
	)
	return &TasksAddCommand{CommandDescription: cmd}, nil
}

func (c *TasksAddCommand) Run(ctx context.Context, pl *layers.ParsedLayers) error {
    s := &TasksAddSettings{}
    if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    path, lines, tasks, err := loadTasksFile(s.Root, s.Ticket, s.TasksFile)
    if err != nil {
        return err
    }
    newLine := formatTaskLine(false, s.Text)
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
        return err
    }
    fmt.Printf("Task added to %s\n", path)
    fmt.Println("Reminder: update the changelog and relate changed files with notes if needed.")
    return nil
}

var _ cmds.BareCommand = &TasksAddCommand{}

// tasks check
type TasksCheckCommand struct{ *cmds.CommandDescription }

type TasksCheckSettings struct {
	Ticket    string `glazed.parameter:"ticket"`
	Root      string `glazed.parameter:"root"`
	TasksFile string `glazed.parameter:"tasks-file"`
	IDs       []int  `glazed.parameter:"id"`
	Match     string `glazed.parameter:"match"`
}

func NewTasksCheckCommand() (*TasksCheckCommand, error) {
	cmd := cmds.NewCommandDescription(
		"check",
		cmds.WithShort("Mark a task as done"),
		cmds.WithLong(`Mark a checkbox task as completed in tasks.md.`),
		cmds.WithFlags(
			parameters.NewParameterDefinition("ticket", parameters.ParameterTypeString, parameters.WithHelp("Ticket identifier (if --tasks-file not set)"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("root", parameters.ParameterTypeString, parameters.WithHelp("Root directory for docs"), parameters.WithDefault("ttmp")),
			parameters.NewParameterDefinition("tasks-file", parameters.ParameterTypeString, parameters.WithHelp("Path to tasks.md (overrides --ticket)"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("id", parameters.ParameterTypeIntegerList, parameters.WithHelp("Task index(es), comma-separated (from 'tasks list')")),
			parameters.NewParameterDefinition("match", parameters.ParameterTypeString, parameters.WithHelp("Substring to match a task if --id not set"), parameters.WithDefault("")),
		),
	)
	return &TasksCheckCommand{CommandDescription: cmd}, nil
}

func (c *TasksCheckCommand) Run(ctx context.Context, pl *layers.ParsedLayers) error {
	s := &TasksCheckSettings{}
	if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}
	path, lines, tasks, err := loadTasksFile(s.Root, s.Ticket, s.TasksFile)
	if err != nil {
		return err
	}
	var targets []int
	if len(s.IDs) > 0 {
		targets = s.IDs
	} else if s.Match != "" {
		for _, t := range tasks {
			if strings.Contains(strings.ToLower(t.Text), strings.ToLower(s.Match)) {
				targets = []int{t.TaskIndex}
				break
			}
		}
	}
	if len(targets) == 0 {
		return fmt.Errorf("no target task specified")
	}
	found := map[int]bool{}
	for _, t := range tasks {
		for _, id := range targets {
			if t.TaskIndex == id {
				lines[t.LineIndex] = formatTaskLine(true, t.Text)
				found[id] = true
			}
		}
	}
	var missing []int
	for _, id := range targets {
		if !found[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("task id(s) not found: %v", missing)
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		return err
	}
	idsStr := make([]string, 0, len(targets))
	for _, id := range targets {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
    if len(targets) > 1 {
        fmt.Printf("Tasks checked: %s (file=%s)\n", strings.Join(idsStr, ","), path)
    } else {
        fmt.Printf("Task checked: %s (file=%s)\n", strings.Join(idsStr, ","), path)
    }
    fmt.Println("Reminder: update the changelog and relate changed files with notes if needed.")
    return nil
}

var _ cmds.BareCommand = &TasksCheckCommand{}

// tasks uncheck
type TasksUncheckCommand struct{ *cmds.CommandDescription }

type TasksUncheckSettings struct {
	Ticket    string `glazed.parameter:"ticket"`
	Root      string `glazed.parameter:"root"`
	TasksFile string `glazed.parameter:"tasks-file"`
	IDs       []int  `glazed.parameter:"id"`
	Match     string `glazed.parameter:"match"`
}

func NewTasksUncheckCommand() (*TasksUncheckCommand, error) {
	cmd := cmds.NewCommandDescription(
		"uncheck",
		cmds.WithShort("Mark a task as not done"),
		cmds.WithLong(`Mark a checkbox task as incomplete in tasks.md.`),
		cmds.WithFlags(
			parameters.NewParameterDefinition("ticket", parameters.ParameterTypeString, parameters.WithHelp("Ticket identifier (if --tasks-file not set)"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("root", parameters.ParameterTypeString, parameters.WithHelp("Root directory for docs"), parameters.WithDefault("ttmp")),
			parameters.NewParameterDefinition("tasks-file", parameters.ParameterTypeString, parameters.WithHelp("Path to tasks.md (overrides --ticket)"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("id", parameters.ParameterTypeIntegerList, parameters.WithHelp("Task index(es), comma-separated (from 'tasks list')")),
			parameters.NewParameterDefinition("match", parameters.ParameterTypeString, parameters.WithHelp("Substring to match a task if --id not set"), parameters.WithDefault("")),
		),
	)
	return &TasksUncheckCommand{CommandDescription: cmd}, nil
}

func (c *TasksUncheckCommand) Run(ctx context.Context, pl *layers.ParsedLayers) error {
	s := &TasksUncheckSettings{}
	if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}
	path, lines, tasks, err := loadTasksFile(s.Root, s.Ticket, s.TasksFile)
	if err != nil {
		return err
	}
	var targets []int
	if len(s.IDs) > 0 {
		targets = s.IDs
	} else if s.Match != "" {
		for _, t := range tasks {
			if strings.Contains(strings.ToLower(t.Text), strings.ToLower(s.Match)) {
				targets = []int{t.TaskIndex}
				break
			}
		}
	}
	if len(targets) == 0 {
		return fmt.Errorf("no target task specified")
	}
	found := map[int]bool{}
	for _, t := range tasks {
		for _, id := range targets {
			if t.TaskIndex == id {
				lines[t.LineIndex] = formatTaskLine(false, t.Text)
				found[id] = true
			}
		}
	}
	var missing []int
	for _, id := range targets {
		if !found[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("task id(s) not found: %v", missing)
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		return err
	}
	idsStr := make([]string, 0, len(targets))
	for _, id := range targets {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
    if len(targets) > 1 {
        fmt.Printf("Tasks unchecked: %s (file=%s)\n", strings.Join(idsStr, ","), path)
    } else {
        fmt.Printf("Task unchecked: %s (file=%s)\n", strings.Join(idsStr, ","), path)
    }
    fmt.Println("Reminder: update the changelog and relate changed files with notes if needed.")
    return nil
}

var _ cmds.BareCommand = &TasksUncheckCommand{}

// tasks edit
type TasksEditCommand struct{ *cmds.CommandDescription }

type TasksEditSettings struct {
	Ticket    string `glazed.parameter:"ticket"`
	Root      string `glazed.parameter:"root"`
	TasksFile string `glazed.parameter:"tasks-file"`
	ID        int    `glazed.parameter:"id"`
	Text      string `glazed.parameter:"text"`
}

func NewTasksEditCommand() (*TasksEditCommand, error) {
	cmd := cmds.NewCommandDescription(
		"edit",
		cmds.WithShort("Edit a task's text"),
		cmds.WithLong(`Edit the text of a checkbox task in tasks.md.`),
		cmds.WithFlags(
			parameters.NewParameterDefinition("ticket", parameters.ParameterTypeString, parameters.WithHelp("Ticket identifier (if --tasks-file not set)"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("root", parameters.ParameterTypeString, parameters.WithHelp("Root directory for docs"), parameters.WithDefault("ttmp")),
			parameters.NewParameterDefinition("tasks-file", parameters.ParameterTypeString, parameters.WithHelp("Path to tasks.md (overrides --ticket)"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("id", parameters.ParameterTypeInteger, parameters.WithHelp("Task index (from 'tasks list')"), parameters.WithRequired(true)),
			parameters.NewParameterDefinition("text", parameters.ParameterTypeString, parameters.WithHelp("New task text"), parameters.WithRequired(true)),
		),
	)
	return &TasksEditCommand{CommandDescription: cmd}, nil
}

func (c *TasksEditCommand) RunIntoGlazeProcessor(ctx context.Context, pl *layers.ParsedLayers, gp middlewares.Processor) error {
	s := &TasksEditSettings{}
	if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}
	path, lines, tasks, err := loadTasksFile(s.Root, s.Ticket, s.TasksFile)
	if err != nil {
		return err
	}
	var target *parsedTask
	for i := range tasks {
		if tasks[i].TaskIndex == s.ID {
			target = &tasks[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("task id not found: %d", s.ID)
	}
	lines[target.LineIndex] = formatTaskLine(target.Checked, s.Text)
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		return err
	}
	row := types.NewRow(types.MRP("file", path), types.MRP("status", "task edited"), types.MRP("id", s.ID))
	return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &TasksEditCommand{}

// tasks remove
type TasksRemoveCommand struct{ *cmds.CommandDescription }

type TasksRemoveSettings struct {
	Ticket    string `glazed.parameter:"ticket"`
	Root      string `glazed.parameter:"root"`
	TasksFile string `glazed.parameter:"tasks-file"`
	IDs       []int  `glazed.parameter:"id"`
}

func NewTasksRemoveCommand() (*TasksRemoveCommand, error) {
	cmd := cmds.NewCommandDescription(
		"remove",
		cmds.WithShort("Remove a task"),
		cmds.WithLong(`Remove a checkbox task from tasks.md.`),
		cmds.WithFlags(
			parameters.NewParameterDefinition("ticket", parameters.ParameterTypeString, parameters.WithHelp("Ticket identifier (if --tasks-file not set)"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("root", parameters.ParameterTypeString, parameters.WithHelp("Root directory for docs"), parameters.WithDefault("ttmp")),
			parameters.NewParameterDefinition("tasks-file", parameters.ParameterTypeString, parameters.WithHelp("Path to tasks.md (overrides --ticket)"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("id", parameters.ParameterTypeIntegerList, parameters.WithHelp("Task index(es), comma-separated (from 'tasks list')"), parameters.WithRequired(true)),
		),
	)
	return &TasksRemoveCommand{CommandDescription: cmd}, nil
}

func (c *TasksRemoveCommand) Run(ctx context.Context, pl *layers.ParsedLayers) error {
	s := &TasksRemoveSettings{}
	if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}
	path, lines, tasks, err := loadTasksFile(s.Root, s.Ticket, s.TasksFile)
	if err != nil {
		return err
	}
	if len(s.IDs) == 0 {
		return fmt.Errorf("no target task specified")
	}
	lineIdxs := make([]int, 0, len(s.IDs))
	found := map[int]bool{}
	for _, id := range s.IDs {
		for _, t := range tasks {
			if t.TaskIndex == id {
				lineIdxs = append(lineIdxs, t.LineIndex)
				found[id] = true
				break
			}
		}
	}
	var missing []int
	for _, id := range s.IDs {
		if !found[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("task id(s) not found: %v", missing)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(lineIdxs)))
	newLines := append([]string{}, lines...)
	for _, idx := range lineIdxs {
		newLines = append(newLines[:idx], newLines[idx+1:]...)
	}
	if err := os.WriteFile(path, []byte(strings.Join(newLines, "\n")+"\n"), 0644); err != nil {
		return err
	}
	idsStr := make([]string, 0, len(s.IDs))
	for _, id := range s.IDs {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
    if len(s.IDs) > 1 {
        fmt.Printf("Tasks removed: %s (file=%s)\n", strings.Join(idsStr, ","), path)
    } else {
        fmt.Printf("Task removed: %s (file=%s)\n", strings.Join(idsStr, ","), path)
    }
    return nil
}

var _ cmds.BareCommand = &TasksRemoveCommand{}
