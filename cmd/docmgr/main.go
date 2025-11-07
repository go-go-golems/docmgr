package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/docmgr/pkg/doc"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "docmgr",
		Short: "Document Manager for LLM Workflows",
		Long: `docmgr is a structured document manager for managing documentation
in LLM-assisted workflows. It provides commands to create, organize,
and validate documentation workspaces with metadata and external source support.

Helpful docs (built-in):

  - Quick usage:          docmgr help how-to-use
  - Initial setup guide:  docmgr help how-to-setup
  - List all embedded docs: docmgr help --all`,
	}

	// Setup Glazed help system and load embedded docs
	helpSystem := help.NewHelpSystem()
	_ = doc.AddDocToHelpSystem(helpSystem)
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	// Create init command (initialize docs root)
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

	// Create configure command (write .ttmp.yaml)
	configureCmd, err := commands.NewConfigureCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating configure command: %v\n", err)
		os.Exit(1)
	}
	cobraConfigureCmd, err := cli.BuildCobraCommand(configureCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building configure command: %v\n", err)
		os.Exit(1)
	}

	// Create create-ticket command
	createTicketCmd, err := commands.NewCreateTicketCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating create-ticket command: %v\n", err)
		os.Exit(1)
	}
	cobraCreateTicketCmd, err := cli.BuildCobraCommand(createTicketCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building create-ticket command: %v\n", err)
		os.Exit(1)
	}

	// Create list parent command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List tickets and documents",
		Long:  "List ticket workspaces or individual documents",
	}

	// Create list tickets command
	listTicketsCmd, err := commands.NewListTicketsCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating list tickets command: %v\n", err)
		os.Exit(1)
	}

	cobraListTicketsCmd, err := cli.BuildCobraCommand(listTicketsCmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building list tickets command: %v\n", err)
		os.Exit(1)
	}

	// Create list docs command
	listDocsCmd, err := commands.NewListDocsCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating list docs command: %v\n", err)
		os.Exit(1)
	}

	cobraListDocsCmd, err := cli.BuildCobraCommand(listDocsCmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building list docs command: %v\n", err)
		os.Exit(1)
	}

	listCmd.AddCommand(cobraListTicketsCmd)
	listCmd.AddCommand(cobraListDocsCmd)

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

	// Create meta parent command
	metaCmd := &cobra.Command{
		Use:   "meta",
		Short: "Manage document metadata",
		Long:  "Update and manage document frontmatter metadata",
	}

	// Create meta update command
	metaUpdateCmd, err := commands.NewMetaUpdateCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating meta update command: %v\n", err)
		os.Exit(1)
	}

	cobraMetaUpdateCmd, err := cli.BuildCobraCommand(metaUpdateCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building meta update command: %v\n", err)
		os.Exit(1)
	}

	metaCmd.AddCommand(cobraMetaUpdateCmd)

	// Create vocab parent command
	vocabCmd := &cobra.Command{
		Use:   "vocab",
		Short: "Manage vocabulary",
		Long:  "Manage vocabulary entries in doc/vocabulary.yaml",
	}

	// Create vocab list command
	vocabListCmd, err := commands.NewVocabListCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating vocab list command: %v\n", err)
		os.Exit(1)
	}

	cobraVocabListCmd, err := cli.BuildCobraCommand(vocabListCmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building vocab list command: %v\n", err)
		os.Exit(1)
	}

	// Create vocab add command
	vocabAddCmd, err := commands.NewVocabAddCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating vocab add command: %v\n", err)
		os.Exit(1)
	}

	cobraVocabAddCmd, err := cli.BuildCobraCommand(vocabAddCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building vocab add command: %v\n", err)
		os.Exit(1)
	}

	vocabCmd.AddCommand(cobraVocabListCmd)
	vocabCmd.AddCommand(cobraVocabAddCmd)

	// Create search command
	searchCmd, err := commands.NewSearchCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating search command: %v\n", err)
		os.Exit(1)
	}

	cobraSearchCmd, err := cli.BuildCobraCommand(searchCmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building search command: %v\n", err)
		os.Exit(1)
	}

	// Create guidelines command
	guidelinesCmd, err := commands.NewGuidelinesCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating guidelines command: %v\n", err)
		os.Exit(1)
	}

	cobraGuidelinesCmd, err := cli.BuildCobraCommand(guidelinesCmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building guidelines command: %v\n", err)
		os.Exit(1)
	}

	// Create relate command
	relateCmd, err := commands.NewRelateCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating relate command: %v\n", err)
		os.Exit(1)
	}

	cobraRelateCmd, err := cli.BuildCobraCommand(relateCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building relate command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraInitCmd)
	rootCmd.AddCommand(cobraConfigureCmd)
	rootCmd.AddCommand(cobraCreateTicketCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(cobraAddCmd)
	rootCmd.AddCommand(cobraDoctorCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(metaCmd)
	rootCmd.AddCommand(vocabCmd)
	rootCmd.AddCommand(cobraSearchCmd)
	rootCmd.AddCommand(cobraGuidelinesCmd)
	rootCmd.AddCommand(cobraRelateCmd)

	// Create renumber command
	renumberCmd, err := commands.NewRenumberCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating renumber command: %v\n", err)
		os.Exit(1)
	}
	cobraRenumberCmd, err := cli.BuildCobraCommand(renumberCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building renumber command: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(cobraRenumberCmd)

	// Create layout-fix command
	layoutFixCmd, err := commands.NewLayoutFixCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating layout-fix command: %v\n", err)
		os.Exit(1)
	}
	cobraLayoutFixCmd, err := cli.BuildCobraCommand(layoutFixCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building layout-fix command: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(cobraLayoutFixCmd)

	// Create status command
	statusCmd, err := commands.NewStatusCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating status command: %v\n", err)
		os.Exit(1)
	}

	cobraStatusCmd, err := cli.BuildCobraCommand(statusCmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building status command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraStatusCmd)

	// Create changelog parent command
	changelogCmd := &cobra.Command{
		Use:   "changelog",
		Short: "Manage ticket changelogs",
		Long:  "Append entries and manage changelog.md within a ticket workspace",
	}

	chgUpdate, err := commands.NewChangelogUpdateCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating changelog update command: %v\n", err)
		os.Exit(1)
	}
	cobraChgUpdate, err := cli.BuildCobraCommand(chgUpdate,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building changelog update command: %v\n", err)
		os.Exit(1)
	}
	changelogCmd.AddCommand(cobraChgUpdate)
	rootCmd.AddCommand(changelogCmd)

	// Create tasks parent command
	tasksCmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage tasks in tasks.md",
		Long:  "Manage checkbox tasks in a ticket's tasks.md file",
	}

	// tasks list
	_tasksList, err := commands.NewTasksListCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tasks list command: %v\n", err)
		os.Exit(1)
	}
	cobraTasksList, err := cli.BuildCobraCommand(_tasksList,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building tasks list command: %v\n", err)
		os.Exit(1)
	}

	// tasks add
	_tasksAdd, err := commands.NewTasksAddCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tasks add command: %v\n", err)
		os.Exit(1)
	}
	cobraTasksAdd, err := cli.BuildCobraCommand(_tasksAdd,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building tasks add command: %v\n", err)
		os.Exit(1)
	}

	// tasks check
	_tasksCheck, err := commands.NewTasksCheckCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tasks check command: %v\n", err)
		os.Exit(1)
	}
	cobraTasksCheck, err := cli.BuildCobraCommand(_tasksCheck,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building tasks check command: %v\n", err)
		os.Exit(1)
	}

	// tasks uncheck
	_tasksUncheck, err := commands.NewTasksUncheckCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tasks uncheck command: %v\n", err)
		os.Exit(1)
	}
	cobraTasksUncheck, err := cli.BuildCobraCommand(_tasksUncheck,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building tasks uncheck command: %v\n", err)
		os.Exit(1)
	}

	// tasks edit
	_tasksEdit, err := commands.NewTasksEditCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tasks edit command: %v\n", err)
		os.Exit(1)
	}
	cobraTasksEdit, err := cli.BuildCobraCommand(_tasksEdit,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building tasks edit command: %v\n", err)
		os.Exit(1)
	}

	// tasks remove
	_tasksRemove, err := commands.NewTasksRemoveCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tasks remove command: %v\n", err)
		os.Exit(1)
	}
	cobraTasksRemove, err := cli.BuildCobraCommand(_tasksRemove,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building tasks remove command: %v\n", err)
		os.Exit(1)
	}

	tasksCmd.AddCommand(cobraTasksList)
	tasksCmd.AddCommand(cobraTasksAdd)
	tasksCmd.AddCommand(cobraTasksCheck)
	tasksCmd.AddCommand(cobraTasksUncheck)
	tasksCmd.AddCommand(cobraTasksEdit)
	tasksCmd.AddCommand(cobraTasksRemove)

	rootCmd.AddCommand(tasksCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
