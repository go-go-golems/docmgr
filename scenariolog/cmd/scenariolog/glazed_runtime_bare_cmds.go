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
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
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
	DBPath string `glazed.parameter:"db"`
}

func NewInitBareCommand() *InitBareCommand {
	return &InitBareCommand{
		CommandDescription: cmds.NewCommandDescription(
			"init",
			cmds.WithShort("Initialize or migrate the scenario sqlite database"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"db",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to sqlite database file"),
					parameters.WithRequired(true),
				),
			),
		),
	}
}

var _ cmds.BareCommand = &InitBareCommand{}

func (c *InitBareCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := &InitSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
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
	DBPath   string            `glazed.parameter:"db"`
	RootDir  string            `glazed.parameter:"root-dir"`
	Suite    string            `glazed.parameter:"suite"`
	RunID    string            `glazed.parameter:"run-id"`
	KV       map[string]string `glazed.parameter:"kv"`
}

func NewRunStartBareCommand() *RunStartBareCommand {
	return &RunStartBareCommand{
		CommandDescription: cmds.NewCommandDescription(
			"start",
			cmds.WithShort("Start a new run and print the run_id"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"db",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to sqlite database file"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"root-dir",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for this scenario run (where artifacts live)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"suite",
					parameters.ParameterTypeString,
					parameters.WithHelp("Suite name (optional)"),
				),
				parameters.NewParameterDefinition(
					"run-id",
					parameters.ParameterTypeString,
					parameters.WithHelp("Explicit run id (optional; otherwise generated)"),
				),
				parameters.NewParameterDefinition(
					"kv",
					parameters.ParameterTypeKeyValue,
					parameters.WithHelp("KV tags to attach to the run (repeatable). Format: key:value, or --kv @file.json/@file.yaml for a map"),
				),
			),
		),
	}
}

var _ cmds.BareCommand = &RunStartBareCommand{}

func (c *RunStartBareCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := &RunStartSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
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
	DBPath   string `glazed.parameter:"db"`
	RunID    string `glazed.parameter:"run-id"`
	ExitCode int    `glazed.parameter:"exit-code"`
}

func NewRunEndBareCommand() *RunEndBareCommand {
	return &RunEndBareCommand{
		CommandDescription: cmds.NewCommandDescription(
			"end",
			cmds.WithShort("Finalize a run (exit code + duration)"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"db",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to sqlite database file"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"run-id",
					parameters.ParameterTypeString,
					parameters.WithHelp("Run id to finalize"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"exit-code",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Exit code for the run (default 0)"),
					parameters.WithDefault(0),
				),
			),
		),
	}
}

var _ cmds.BareCommand = &RunEndBareCommand{}

func (c *RunEndBareCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := &RunEndSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
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
	DBPath     string            `glazed.parameter:"db"`
	RunID      string            `glazed.parameter:"run-id"`
	RootDir    string            `glazed.parameter:"root-dir"`
	WorkDir    string            `glazed.parameter:"work-dir"`
	LogDir     string            `glazed.parameter:"log-dir"`
	StepNum    int               `glazed.parameter:"step-num"`
	StepName   string            `glazed.parameter:"name"`
	ScriptPath string            `glazed.parameter:"script-path"`
	KV         map[string]string `glazed.parameter:"kv"`
	Command    []string          `glazed.parameter:"cmd"`
}

func NewExecBareCommand() *ExecBareCommand {
	return &ExecBareCommand{
		CommandDescription: cmds.NewCommandDescription(
			"exec",
			cmds.WithShort("Execute a step command, capture stdout/stderr, and log results to sqlite"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"db",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to sqlite database file"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"run-id",
					parameters.ParameterTypeString,
					parameters.WithHelp("Run id to attach this step to"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"root-dir",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for this scenario run (used for path normalization)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"work-dir",
					parameters.ParameterTypeString,
					parameters.WithHelp("Working directory for the executed command (defaults to current directory)"),
				),
				parameters.NewParameterDefinition(
					"log-dir",
					parameters.ParameterTypeString,
					parameters.WithHelp("Log directory (relative to root-dir unless absolute; must already exist)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"step-num",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Step number (used for ordering + filenames)"),
					parameters.WithDefault(0),
				),
				parameters.NewParameterDefinition(
					"name",
					parameters.ParameterTypeString,
					parameters.WithHelp("Step name"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"script-path",
					parameters.ParameterTypeString,
					parameters.WithHelp("Script path (optional)"),
				),
				parameters.NewParameterDefinition(
					"kv",
					parameters.ParameterTypeKeyValue,
					parameters.WithHelp("KV tags to attach to this step (repeatable). Format: key:value, or --kv @file.json/@file.yaml for a map"),
				),
			),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"cmd",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Command to execute (argv)"),
					parameters.WithRequired(true),
				),
			),
		),
	}
}

var _ cmds.BareCommand = &ExecBareCommand{}

func (c *ExecBareCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := &ExecSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
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


