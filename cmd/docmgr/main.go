package main

import (
	"fmt"
	"os"

	appcmds "github.com/go-go-golems/docmgr/cmd/docmgr/cmds"
	"github.com/go-go-golems/docmgr/pkg/doc"
	"github.com/go-go-golems/glazed/pkg/help"
)

func main() {
	helpSystem := help.NewHelpSystem()
	_ = doc.AddDocToHelpSystem(helpSystem)

	rootCmd, err := appcmds.NewRootCommand(helpSystem)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error constructing root command: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
