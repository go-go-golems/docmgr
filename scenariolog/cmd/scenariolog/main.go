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

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "scenariolog",
		Short: "Scenario logging flight recorder (sqlite + artifacts + FTS)",
	}

	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newRunCmd())
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

func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}
}


