package main

import (
	"context"
	"fmt"
	"os"

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

func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}
}


