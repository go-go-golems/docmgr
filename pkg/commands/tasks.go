package commands

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
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

var taskLineRegex = regexp.MustCompile(`^\s*[-*]\s*\[(?i:[x ])\]\s*(.*)$`)

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
type TasksListCommand struct { *cmds.CommandDescription }

type TasksListSettings struct {
    Ticket    string `glazed.parameter:"ticket"`
    Root      string `glazed.parameter:"root"`
    TasksFile string `glazed.parameter:"tasks-file"`
}

func NewTasksListCommand() (*TasksListCommand, error) {
    cmd := cmds.NewCommandDescription(
        "list",
        cmds.WithShort("List tasks from tasks.md"),
        cmds.WithLong(`List checkbox tasks found in the ticket's tasks.md.`),
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
    if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil { return err }
    path, _, tasks, err := loadTasksFile(s.Root, s.Ticket, s.TasksFile)
    if err != nil { return err }
    for _, t := range tasks {
        row := types.NewRow(
            types.MRP("index", t.TaskIndex),
            types.MRP("checked", t.Checked),
            types.MRP("text", t.Text),
            types.MRP("file", path),
        )
        if err := gp.AddRow(ctx, row); err != nil { return err }
    }
    return nil
}

var _ cmds.GlazeCommand = &TasksListCommand{}

// tasks add
type TasksAddCommand struct { *cmds.CommandDescription }

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

func (c *TasksAddCommand) RunIntoGlazeProcessor(ctx context.Context, pl *layers.ParsedLayers, gp middlewares.Processor) error {
    s := &TasksAddSettings{}
    if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil { return err }
    path, lines, tasks, err := loadTasksFile(s.Root, s.Ticket, s.TasksFile)
    if err != nil { return err }
    newLine := formatTaskLine(false, s.Text)
    if s.After <= 0 || len(tasks) == 0 {
        lines = append(lines, newLine)
    } else {
        // insert after task with TaskIndex == s.After
        insertAt := len(lines)
        for _, t := range tasks {
            if t.TaskIndex == s.After {
                insertAt = t.LineIndex + 1
            }
        }
        if insertAt >= len(lines) {
            lines = append(lines, newLine)
        } else {
            // insert in place
            lines = append(lines[:insertAt], append([]string{newLine}, lines[insertAt:]...)...)
        }
    }
    if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil { return err }
    row := types.NewRow(types.MRP("file", path), types.MRP("status", "task added"))
    return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &TasksAddCommand{}

// tasks check
type TasksCheckCommand struct { *cmds.CommandDescription }

type TasksCheckSettings struct {
    Ticket    string `glazed.parameter:"ticket"`
    Root      string `glazed.parameter:"root"`
    TasksFile string `glazed.parameter:"tasks-file"`
    ID        int    `glazed.parameter:"id"`
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
            parameters.NewParameterDefinition("id", parameters.ParameterTypeInteger, parameters.WithHelp("Task index (from 'tasks list')"), parameters.WithDefault(0)),
            parameters.NewParameterDefinition("match", parameters.ParameterTypeString, parameters.WithHelp("Substring to match a task if --id not set"), parameters.WithDefault("")),
        ),
    )
    return &TasksCheckCommand{CommandDescription: cmd}, nil
}

func (c *TasksCheckCommand) RunIntoGlazeProcessor(ctx context.Context, pl *layers.ParsedLayers, gp middlewares.Processor) error {
    s := &TasksCheckSettings{}
    if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil { return err }
    path, lines, tasks, err := loadTasksFile(s.Root, s.Ticket, s.TasksFile)
    if err != nil { return err }
    target := -1
    if s.ID > 0 {
        target = s.ID
    } else if s.Match != "" {
        // pick first containing
        for _, t := range tasks {
            if strings.Contains(strings.ToLower(t.Text), strings.ToLower(s.Match)) { target = t.TaskIndex; break }
        }
    }
    if target <= 0 { return fmt.Errorf("no target task specified") }
    for _, t := range tasks {
        if t.TaskIndex == target {
            lines[t.LineIndex] = formatTaskLine(true, t.Text)
            break
        }
    }
    if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil { return err }
    row := types.NewRow(types.MRP("file", path), types.MRP("status", "task checked"), types.MRP("id", target))
    return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &TasksCheckCommand{}

// tasks uncheck
type TasksUncheckCommand struct { *cmds.CommandDescription }

type TasksUncheckSettings struct {
    Ticket    string `glazed.parameter:"ticket"`
    Root      string `glazed.parameter:"root"`
    TasksFile string `glazed.parameter:"tasks-file"`
    ID        int    `glazed.parameter:"id"`
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
            parameters.NewParameterDefinition("id", parameters.ParameterTypeInteger, parameters.WithHelp("Task index (from 'tasks list')"), parameters.WithDefault(0)),
            parameters.NewParameterDefinition("match", parameters.ParameterTypeString, parameters.WithHelp("Substring to match a task if --id not set"), parameters.WithDefault("")),
        ),
    )
    return &TasksUncheckCommand{CommandDescription: cmd}, nil
}

func (c *TasksUncheckCommand) RunIntoGlazeProcessor(ctx context.Context, pl *layers.ParsedLayers, gp middlewares.Processor) error {
    s := &TasksUncheckSettings{}
    if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil { return err }
    path, lines, tasks, err := loadTasksFile(s.Root, s.Ticket, s.TasksFile)
    if err != nil { return err }
    target := -1
    if s.ID > 0 { target = s.ID } else if s.Match != "" {
        for _, t := range tasks { if strings.Contains(strings.ToLower(t.Text), strings.ToLower(s.Match)) { target = t.TaskIndex; break } }
    }
    if target <= 0 { return fmt.Errorf("no target task specified") }
    for _, t := range tasks {
        if t.TaskIndex == target {
            lines[t.LineIndex] = formatTaskLine(false, t.Text)
            break
        }
    }
    if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil { return err }
    row := types.NewRow(types.MRP("file", path), types.MRP("status", "task unchecked"), types.MRP("id", target))
    return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &TasksUncheckCommand{}

// tasks edit
type TasksEditCommand struct { *cmds.CommandDescription }

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
    if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil { return err }
    path, lines, tasks, err := loadTasksFile(s.Root, s.Ticket, s.TasksFile)
    if err != nil { return err }
    var target *parsedTask
    for i := range tasks {
        if tasks[i].TaskIndex == s.ID { target = &tasks[i]; break }
    }
    if target == nil { return fmt.Errorf("task id not found: %d", s.ID) }
    lines[target.LineIndex] = formatTaskLine(target.Checked, s.Text)
    if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil { return err }
    row := types.NewRow(types.MRP("file", path), types.MRP("status", "task edited"), types.MRP("id", s.ID))
    return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &TasksEditCommand{}

// tasks remove
type TasksRemoveCommand struct { *cmds.CommandDescription }

type TasksRemoveSettings struct {
    Ticket    string `glazed.parameter:"ticket"`
    Root      string `glazed.parameter:"root"`
    TasksFile string `glazed.parameter:"tasks-file"`
    ID        int    `glazed.parameter:"id"`
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
            parameters.NewParameterDefinition("id", parameters.ParameterTypeInteger, parameters.WithHelp("Task index (from 'tasks list')"), parameters.WithRequired(true)),
        ),
    )
    return &TasksRemoveCommand{CommandDescription: cmd}, nil
}

func (c *TasksRemoveCommand) RunIntoGlazeProcessor(ctx context.Context, pl *layers.ParsedLayers, gp middlewares.Processor) error {
    s := &TasksRemoveSettings{}
    if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil { return err }
    path, lines, tasks, err := loadTasksFile(s.Root, s.Ticket, s.TasksFile)
    if err != nil { return err }
    // find line index to remove
    lineIdx := -1
    for _, t := range tasks {
        if t.TaskIndex == s.ID { lineIdx = t.LineIndex; break }
    }
    if lineIdx < 0 { return fmt.Errorf("task id not found: %d", s.ID) }
    // remove line
    newLines := append([]string{}, lines[:lineIdx]...)
    newLines = append(newLines, lines[lineIdx+1:]...)
    if err := os.WriteFile(path, []byte(strings.Join(newLines, "\n")+"\n"), 0644); err != nil { return err }
    row := types.NewRow(types.MRP("file", path), types.MRP("status", "task removed"), types.MRP("id", s.ID))
    return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &TasksRemoveCommand{}


