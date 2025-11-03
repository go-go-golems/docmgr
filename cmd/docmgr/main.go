package main

import (
	"fmt"
	"os"

	"github.com/docmgr/docmgr/pkg/commands"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "docmgr",
		Short: "Document Manager for LLM Workflows",
		Long: `docmgr is a structured document manager for managing documentation
in LLM-assisted workflows. It provides commands to create, organize,
and validate documentation workspaces with metadata and external source support.`,
	}

	// Create init command
	initCmd, err := commands.NewInitCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating init command: %v\n", err)
		os.Exit(1)
	}

	cobraInitCmd, err := cli.BuildCobraCommand(initCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building init command: %v\n", err)
		os.Exit(1)
	}

	// Create list command
	listCmd, err := commands.NewListCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating list command: %v\n", err)
		os.Exit(1)
	}

	cobraListCmd, err := cli.BuildCobraCommand(listCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building list command: %v\n", err)
		os.Exit(1)
	}

	// Create add command
	addCmd, err := commands.NewAddCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating add command: %v\n", err)
		os.Exit(1)
	}

	cobraAddCmd, err := cli.BuildCobraCommand(addCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building add command: %v\n", err)
		os.Exit(1)
	}

	// Create doctor command
	doctorCmd, err := commands.NewDoctorCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating doctor command: %v\n", err)
		os.Exit(1)
	}

	cobraDoctorCmd, err := cli.BuildCobraCommand(doctorCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building doctor command: %v\n", err)
		os.Exit(1)
	}

	// Create import parent command
	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import external sources",
		Long:  "Import files and external sources into document workspaces",
	}

	// Create import file command
	importFileCmd, err := commands.NewImportFileCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating import file command: %v\n", err)
		os.Exit(1)
	}

	cobraImportFileCmd, err := cli.BuildCobraCommand(importFileCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building import file command: %v\n", err)
		os.Exit(1)
	}

	importCmd.AddCommand(cobraImportFileCmd)

	// Add all commands to root
	rootCmd.AddCommand(cobraInitCmd)
	rootCmd.AddCommand(cobraListCmd)
	rootCmd.AddCommand(cobraAddCmd)
	rootCmd.AddCommand(cobraDoctorCmd)
	rootCmd.AddCommand(importCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
