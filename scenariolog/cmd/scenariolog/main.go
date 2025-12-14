package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/docmgr/scenariolog/pkg/doc"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
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

func newRootCmd() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "scenariolog",
		Short: "Scenario logging flight recorder (sqlite + artifacts + FTS)",
	}

	helpSystem := help.NewHelpSystem()
	if err := doc.AddDocToHelpSystem(helpSystem); err != nil {
		return nil, err
	}
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	if err := addRuntimeBareCommands(rootCmd); err != nil {
		return nil, err
	}
	if err := addGlazedCommands(rootCmd); err != nil {
		return nil, err
	}
	return rootCmd, nil
}

func main() {
	cmd, err := newRootCmd()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}
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


