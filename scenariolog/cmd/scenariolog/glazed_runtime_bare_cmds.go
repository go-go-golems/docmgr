package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/go-go-golems/docmgr/scenariolog/internal/scenariolog"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func addRuntimeBareCommands(rootCmd *cobra.Command) error {
	initCmd, err := cli.BuildCobraCommand(NewInitBareCommand())
	if err != nil {
		return err
	}
	execCmd, err := cli.BuildCobraCommand(NewExecBareCommand())
	if err != nil {
		return err
	}

	runGroup := &cobra.Command{
		Use:   "run",
		Short: "Manage scenario runs (start/end)",
	}
	runStartCmd, err := cli.BuildCobraCommand(NewRunStartBareCommand())
	if err != nil {
		return err
	}
	runEndCmd, err := cli.BuildCobraCommand(NewRunEndBareCommand())
	if err != nil {
		return err
	}
	runGroup.AddCommand(runStartCmd, runEndCmd)

	rootCmd.AddCommand(initCmd, execCmd, runGroup)
	return nil
}

type InitBareCommand struct {
	*cmds.CommandDescription
}

type InitSettings struct {
	DBPath string `glazed:"db"`
}

func NewInitBareCommand() *InitBareCommand {
	return &InitBareCommand{
		CommandDescription: cmds.NewCommandDescription(
			"init",
			cmds.WithShort("Initialize or migrate the scenario sqlite database"),
			cmds.WithFlags(
				fields.New(
					"db",
					fields.TypeString,
					fields.WithHelp("Path to sqlite database file"),
					fields.WithRequired(true),
				),
			),
		),
	}
}

var _ cmds.BareCommand = &InitBareCommand{}

func (c *InitBareCommand) Run(ctx context.Context, parsedValues *values.Values) error {
	s := &InitSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// On CTRL-C, cancel the context so sqlite operations stop promptly.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	db, err := scenariolog.Open(ctx, s.DBPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	return scenariolog.Migrate(ctx, db)
}

type RunStartBareCommand struct {
	*cmds.CommandDescription
}

type RunStartSettings struct {
	DBPath  string            `glazed:"db"`
	RootDir string            `glazed:"root-dir"`
	Suite   string            `glazed:"suite"`
	RunID   string            `glazed:"run-id"`
	KV      map[string]string `glazed:"kv"`
}

func NewRunStartBareCommand() *RunStartBareCommand {
	return &RunStartBareCommand{
		CommandDescription: cmds.NewCommandDescription(
			"start",
			cmds.WithShort("Start a new run and print the run_id"),
			cmds.WithFlags(
				fields.New(
					"db",
					fields.TypeString,
					fields.WithHelp("Path to sqlite database file"),
					fields.WithRequired(true),
				),
				fields.New(
					"root-dir",
					fields.TypeString,
					fields.WithHelp("Root directory for this scenario run (where artifacts live)"),
					fields.WithRequired(true),
				),
				fields.New(
					"suite",
					fields.TypeString,
					fields.WithHelp("Suite name (optional)"),
				),
				fields.New(
					"run-id",
					fields.TypeString,
					fields.WithHelp("Explicit run id (optional; otherwise generated)"),
				),
				fields.New(
					"kv",
					fields.TypeKeyValue,
					fields.WithHelp("KV tags to attach to the run (repeatable). Format: key:value, or --kv @file.json/@file.yaml for a map"),
				),
			),
		),
	}
}

var _ cmds.BareCommand = &RunStartBareCommand{}

func (c *RunStartBareCommand) Run(ctx context.Context, parsedValues *values.Values) error {
	s := &RunStartSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	db, err := scenariolog.Open(ctx, s.DBPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := scenariolog.Migrate(ctx, db); err != nil {
		return err
	}

	now := time.Now()
	runID := s.RunID
	if runID == "" {
		runID = scenariolog.NewRunID(now)
	}

	if err := scenariolog.StartRun(ctx, db, runID, s.RootDir, s.Suite, now); err != nil {
		return err
	}

	// Allow user-provided kv to override auto-emitted provenance tags.
	for k, v := range s.KV {
		if err := scenariolog.SetKV(ctx, db, runID, "", "", k, v); err != nil {
			return err
		}
	}

	fmt.Fprintln(os.Stdout, runID)
	return nil
}

type RunEndBareCommand struct {
	*cmds.CommandDescription
}

type RunEndSettings struct {
	DBPath   string `glazed:"db"`
	RunID    string `glazed:"run-id"`
	ExitCode int    `glazed:"exit-code"`
}

func NewRunEndBareCommand() *RunEndBareCommand {
	return &RunEndBareCommand{
		CommandDescription: cmds.NewCommandDescription(
			"end",
			cmds.WithShort("Finalize a run (exit code + duration)"),
			cmds.WithFlags(
				fields.New(
					"db",
					fields.TypeString,
					fields.WithHelp("Path to sqlite database file"),
					fields.WithRequired(true),
				),
				fields.New(
					"run-id",
					fields.TypeString,
					fields.WithHelp("Run id to finalize"),
					fields.WithRequired(true),
				),
				fields.New(
					"exit-code",
					fields.TypeInteger,
					fields.WithHelp("Exit code for the run (default 0)"),
					fields.WithDefault(0),
				),
			),
		),
	}
}

var _ cmds.BareCommand = &RunEndBareCommand{}

func (c *RunEndBareCommand) Run(ctx context.Context, parsedValues *values.Values) error {
	s := &RunEndSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	db, err := scenariolog.Open(ctx, s.DBPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := scenariolog.Migrate(ctx, db); err != nil {
		return err
	}

	return scenariolog.EndRun(ctx, db, s.RunID, s.ExitCode, time.Now())
}

type ExecBareCommand struct {
	*cmds.CommandDescription
}

type ExecSettings struct {
	DBPath     string            `glazed:"db"`
	RunID      string            `glazed:"run-id"`
	RootDir    string            `glazed:"root-dir"`
	WorkDir    string            `glazed:"work-dir"`
	LogDir     string            `glazed:"log-dir"`
	StepNum    int               `glazed:"step-num"`
	StepName   string            `glazed:"name"`
	ScriptPath string            `glazed:"script-path"`
	KV         map[string]string `glazed:"kv"`
	Command    []string          `glazed:"cmd"`
}

func NewExecBareCommand() *ExecBareCommand {
	return &ExecBareCommand{
		CommandDescription: cmds.NewCommandDescription(
			"exec",
			cmds.WithShort("Execute a step command, capture stdout/stderr, and log results to sqlite"),
			cmds.WithFlags(
				fields.New(
					"db",
					fields.TypeString,
					fields.WithHelp("Path to sqlite database file"),
					fields.WithRequired(true),
				),
				fields.New(
					"run-id",
					fields.TypeString,
					fields.WithHelp("Run id to attach this step to"),
					fields.WithRequired(true),
				),
				fields.New(
					"root-dir",
					fields.TypeString,
					fields.WithHelp("Root directory for this scenario run (used for path normalization)"),
					fields.WithRequired(true),
				),
				fields.New(
					"work-dir",
					fields.TypeString,
					fields.WithHelp("Working directory for the executed command (defaults to current directory)"),
				),
				fields.New(
					"log-dir",
					fields.TypeString,
					fields.WithHelp("Log directory (relative to root-dir unless absolute; must already exist)"),
					fields.WithRequired(true),
				),
				fields.New(
					"step-num",
					fields.TypeInteger,
					fields.WithHelp("Step number (used for ordering + filenames)"),
					fields.WithDefault(0),
				),
				fields.New(
					"name",
					fields.TypeString,
					fields.WithHelp("Step name"),
					fields.WithRequired(true),
				),
				fields.New(
					"script-path",
					fields.TypeString,
					fields.WithHelp("Script path (optional)"),
				),
				fields.New(
					"kv",
					fields.TypeKeyValue,
					fields.WithHelp("KV tags to attach to this step (repeatable). Format: key:value, or --kv @file.json/@file.yaml for a map"),
				),
			),
			cmds.WithArguments(
				fields.New(
					"cmd",
					fields.TypeStringList,
					fields.WithHelp("Command to execute (argv)"),
					fields.WithRequired(true),
				),
			),
		),
	}
}

var _ cmds.BareCommand = &ExecBareCommand{}

func (c *ExecBareCommand) Run(ctx context.Context, parsedValues *values.Values) error {
	s := &ExecSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}
	if len(s.Command) == 0 {
		return errors.New("cmd is required")
	}

	// On CTRL-C, cancel the context so ExecStep can terminate the full process group
	// and still finalize sqlite rows.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	db, err := scenariolog.Open(ctx, s.DBPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := scenariolog.Migrate(ctx, db); err != nil {
		return err
	}

	res, err := scenariolog.ExecStep(ctx, db, scenariolog.ExecStepSpec{
		RunID:      s.RunID,
		RootDir:    s.RootDir,
		WorkDir:    s.WorkDir,
		LogDir:     s.LogDir,
		StepNum:    s.StepNum,
		StepName:   s.StepName,
		ScriptPath: s.ScriptPath,
		Command:    s.Command,
	})
	if err != nil {
		return err
	}

	for k, v := range s.KV {
		if err := scenariolog.SetKV(ctx, db, s.RunID, res.StepID, "", k, v); err != nil {
			return err
		}
	}

	// Human-friendly summary.
	fmt.Fprintf(os.Stderr, "[scenariolog] step=%s exit=%d duration_ms=%d stdout=%s stderr=%s\n",
		res.StepID, res.ExitCode, res.DurationMs, res.StdoutPath, res.StderrPath)

	if res.ExitCode != 0 {
		return &ExitError{
			Code: res.ExitCode,
			Err:  fmt.Errorf("step %s exited with code %d", res.StepID, res.ExitCode),
		}
	}

	return nil
}
