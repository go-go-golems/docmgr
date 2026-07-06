package commands

import "fmt"

// verboseEnabled gates the workspace banner and coaching/reminder output.
// It is set from the root command's persistent --verbose flag.
var verboseEnabled bool

// SetVerbose toggles verbose (banner + reminder) output for bare-mode commands.
func SetVerbose(v bool) { verboseEnabled = v }

// VerboseEnabled reports whether verbose output was requested.
func VerboseEnabled() bool { return verboseEnabled }

// printWorkspaceBanner prints the legacy "Docs root / Config / Vocabulary" banner.
// It is silent unless --verbose is set (design D4: output diet).
func printWorkspaceBanner(root string, configPath string, vocabularyPath string) {
	if !verboseEnabled {
		return
	}
	if root != "" {
		fmt.Printf("Docs root: `%s`\n", root)
	}
	if configPath != "" {
		fmt.Printf("Config: `%s`\n", configPath)
	}
	if vocabularyPath != "" {
		fmt.Printf("Vocabulary: `%s`\n", vocabularyPath)
	}
}

// printReminder prints coaching reminders only in verbose mode.
func printReminder(msg string) {
	if !verboseEnabled {
		return
	}
	fmt.Println(msg)
}
