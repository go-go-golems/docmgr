package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/docmgr/scenariolog/internal/scenariolog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exited with code %d", e.Code)
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "scenariolog",
		Short: "Scenario logging flight recorder (sqlite + artifacts + FTS)",
	}

	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newExecCmd())
	rootCmd.AddCommand(newSearchCmd())
	return rootCmd
}

func newInitCmd() *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize or migrate the scenario sqlite database",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				return errors.New("--db is required")
			}

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			db, err := scenariolog.Open(ctx, dbPath)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			if err := scenariolog.Migrate(ctx, db); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Path to sqlite database file")
	return cmd
}

func newRunCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Manage scenario runs (start/end)",
	}
	runCmd.AddCommand(newRunStartCmd())
	runCmd.AddCommand(newRunEndCmd())
	return runCmd
}

func newRunStartCmd() *cobra.Command {
	var dbPath string
	var rootDir string
	var suite string
	var runID string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a new run and print the run_id",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				return errors.New("--db is required")
			}
			if rootDir == "" {
				return errors.New("--root-dir is required")
			}

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			db, err := scenariolog.Open(ctx, dbPath)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			if err := scenariolog.Migrate(ctx, db); err != nil {
				return err
			}

			now := time.Now()
			if runID == "" {
				runID = scenariolog.NewRunID(now)
			}

			if err := scenariolog.StartRun(ctx, db, runID, rootDir, suite, now); err != nil {
				return err
			}

			fmt.Println(runID)
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Path to sqlite database file")
	cmd.Flags().StringVar(&rootDir, "root-dir", "", "Root directory for this scenario run (where artifacts live)")
	cmd.Flags().StringVar(&suite, "suite", "", "Suite name (optional)")
	cmd.Flags().StringVar(&runID, "run-id", "", "Explicit run id (optional; otherwise generated)")
	return cmd
}

func newRunEndCmd() *cobra.Command {
	var dbPath string
	var runID string
	var exitCode int

	cmd := &cobra.Command{
		Use:   "end",
		Short: "Finalize a run (exit code + duration)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				return errors.New("--db is required")
			}
			if runID == "" {
				return errors.New("--run-id is required")
			}

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			db, err := scenariolog.Open(ctx, dbPath)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			if err := scenariolog.Migrate(ctx, db); err != nil {
				return err
			}

			if err := scenariolog.EndRun(ctx, db, runID, exitCode, time.Now()); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Path to sqlite database file")
	cmd.Flags().StringVar(&runID, "run-id", "", "Run id to finalize")
	cmd.Flags().IntVar(&exitCode, "exit-code", 0, "Exit code for the run (default 0)")
	return cmd
}

func newExecCmd() *cobra.Command {
	var dbPath string
	var runID string
	var rootDir string
	var logDir string
	var stepNum int
	var stepName string
	var scriptPath string

	cmd := &cobra.Command{
		Use:   "exec -- <command> [args...]",
		Short: "Execute a step command, capture stdout/stderr, and log results to sqlite",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				return errors.New("--db is required")
			}
			if runID == "" {
				return errors.New("--run-id is required")
			}
			if rootDir == "" {
				return errors.New("--root-dir is required")
			}
			if logDir == "" {
				return errors.New("--log-dir is required")
			}
			if stepName == "" {
				return errors.New("--name is required")
			}

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			db, err := scenariolog.Open(ctx, dbPath)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			if err := scenariolog.Migrate(ctx, db); err != nil {
				return err
			}

			res, err := scenariolog.ExecStep(ctx, db, scenariolog.ExecStepSpec{
				RunID:      runID,
				RootDir:    rootDir,
				LogDir:     logDir,
				StepNum:    stepNum,
				StepName:   stepName,
				ScriptPath: scriptPath,
				Command:    args,
			})
			if err != nil {
				return err
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
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Path to sqlite database file")
	cmd.Flags().StringVar(&runID, "run-id", "", "Run id to attach this step to")
	cmd.Flags().StringVar(&rootDir, "root-dir", "", "Root directory for this scenario run (used for cwd + path normalization)")
	cmd.Flags().StringVar(&logDir, "log-dir", "", "Log directory (relative to root-dir unless absolute; must already exist)")
	cmd.Flags().IntVar(&stepNum, "step-num", 0, "Step number (used for ordering + filenames)")
	cmd.Flags().StringVar(&stepName, "name", "", "Step name")
	cmd.Flags().StringVar(&scriptPath, "script-path", "", "Script path (optional)")

	return cmd
}

func newSearchCmd() *cobra.Command {
	var dbPath string
	var runID string
	var query string
	var limit int

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search indexed log lines (FTS5)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				return errors.New("--db is required")
			}
			if runID == "" {
				return errors.New("--run-id is required")
			}
			if query == "" {
				return errors.New("--query is required")
			}

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			db, err := scenariolog.Open(ctx, dbPath)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			if err := scenariolog.Migrate(ctx, db); err != nil {
				return err
			}

			hits, err := scenariolog.SearchFTS(ctx, db, runID, query, limit)
			if err != nil {
				return err
			}

			// Simple tabular output for now; we can switch to Glazed output later.
			for _, h := range hits {
				fmt.Printf("%d\t%d\t%s\n", h.ArtifactID, h.LineNum, h.Text)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Path to sqlite database file")
	cmd.Flags().StringVar(&runID, "run-id", "", "Run id to search within")
	cmd.Flags().StringVar(&query, "query", "", "FTS query string")
	cmd.Flags().IntVar(&limit, "limit", 100, "Max number of hits to return")
	return cmd
}

func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		var ee *ExitError
		if errors.As(err, &ee) {
			_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", ee.Error())
			os.Exit(ee.Code)
		}
		_, _ = fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}
}


