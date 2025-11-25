package completion

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/commands"
)

// ActionTickets completes known ticket IDs discovered under the resolved root.
func ActionTickets() carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		root := workspace.ResolveRoot("ttmp")
		workspaces, err := workspace.CollectTicketWorkspaces(root, nil)
		if err != nil {
			return carapace.ActionMessage(fmt.Sprintf("failed to collect tickets: %v", err))
		}
		vals := []string{}
		for _, tw := range workspaces {
			ticket := ""
			title := ""
			if tw.Doc != nil {
				ticket = strings.TrimSpace(tw.Doc.Ticket)
				title = strings.TrimSpace(tw.Doc.Title)
			}
			if ticket == "" {
				// Fall back to directory name if frontmatter missing
				base := filepath.Base(tw.Path)
				// Try to extract ticket slug from path segment
				parts := strings.SplitN(base, "-", 2)
				if len(parts) > 0 {
					ticket = parts[0]
				} else {
					ticket = base
				}
			}
			if title == "" && tw.Doc != nil {
				title = tw.Doc.Summary
			}
			if title == "" {
				title = tw.Path
			}
			vals = append(vals, ticket, title)
		}
		return carapace.ActionValuesDescribed(vals...)
	})
}

// ActionVocab returns an action for a vocabulary category (topics, docTypes, status, intent).
func ActionVocab(category string) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		v, err := commands.LoadVocabulary()
		if err != nil {
			return carapace.ActionMessage(fmt.Sprintf("failed to load vocabulary: %v", err))
		}
		vals := []string{}
		switch category {
		case "topics":
			for _, it := range v.Topics {
				vals = append(vals, it.Slug, it.Description)
			}
		case "docTypes":
			for _, it := range v.DocTypes {
				vals = append(vals, it.Slug, it.Description)
			}
		case "status":
			for _, it := range v.Status {
				vals = append(vals, it.Slug, it.Description)
			}
		case "intent":
			for _, it := range v.Intent {
				vals = append(vals, it.Slug, it.Description)
			}
		default:
			return carapace.ActionMessage("unknown vocabulary category")
		}
		return carapace.ActionValuesDescribed(vals...)
	})
}

// ActionTopics completes topics, supporting comma-separated multi values.
func ActionTopics() carapace.Action {
	return carapace.ActionMultiParts(",", func(c carapace.Context) carapace.Action {
		return ActionVocab("topics")
	})
}

// ActionDocTypes completes document types.
func ActionDocTypes() carapace.Action {
	return ActionVocab("docTypes")
}

// ActionStatus completes status values.
func ActionStatus() carapace.Action {
	return ActionVocab("status")
}

// ActionIntent completes intent values.
func ActionIntent() carapace.Action {
	return ActionVocab("intent")
}

// ActionFiles completes file paths.
func ActionFiles() carapace.Action {
	return carapace.ActionFiles()
}

// ActionDirectories completes directory paths.
func ActionDirectories() carapace.Action {
	return carapace.ActionDirectories()
}

// parseFlags extracts simple --flag value pairs from args (no equals handling).
func parseFlags(args []string) map[string]string {
	result := map[string]string{}
	for i := 0; i < len(args); i++ {
		token := args[i]
		if strings.HasPrefix(token, "--") && len(token) > 2 && token != "--" {
			name := strings.TrimPrefix(token, "--")
			// next token if present and not another flag
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				result[name] = args[i+1]
				i++
			} else {
				result[name] = ""
			}
		}
	}
	return result
}

// ActionTaskIDs completes task indices from a tasks.md resolved by --tasks-file or --ticket/--root.
func ActionTaskIDs() carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		flags := parseFlags(c.Args)
		root := flags["root"]
		if root == "" {
			root = "ttmp"
		}
		root = workspace.ResolveRoot(root)

		tasksFile := flags["tasks-file"]
		if tasksFile == "" {
			// Try to locate from ticket
			ticket := flags["ticket"]
			if ticket == "" {
				return carapace.ActionMessage("specify --ticket or --tasks-file for id completion")
			}
			// First pass: scan root entries for directory containing ticket substring
			if entries, err := os.ReadDir(root); err == nil {
				lower := strings.ToLower(ticket)
				for _, e := range entries {
					if e.IsDir() && strings.Contains(strings.ToLower(e.Name()), lower) {
						tasksFile = filepath.Join(root, e.Name(), "tasks.md")
						break
					}
				}
			}
			if tasksFile == "" {
				return carapace.ActionMessage("could not resolve tasks file for ticket")
			}
		}
		// Read tasks.md and extract indices
		f, err := os.Open(tasksFile)
		if err != nil {
			return carapace.ActionMessage(fmt.Sprintf("failed to read tasks file: %v", err))
		}
		defer func() { _ = f.Close() }()
		sc := bufio.NewScanner(f)
		idx := 0
		vals := []string{}
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if strings.HasPrefix(line, "- [") || strings.HasPrefix(line, "* [") {
				idx++
				vals = append(vals, fmt.Sprintf("%d", idx))
			}
		}
		if err := sc.Err(); err != nil {
			return carapace.ActionMessage(fmt.Sprintf("failed to scan tasks file: %v", err))
		}
		// Ensure sorted and unique (should be)
		sort.Strings(vals)
		return carapace.ActionValues(vals...)
	})
}

// ActionMetaFields completes meta update field names.
func ActionMetaFields() carapace.Action {
	return carapace.ActionValues(
		"Title",
		"Ticket",
		"Status",
		"Topics",
		"DocType",
		"Intent",
		"Owners",
		"RelatedFiles",
		"ExternalSources",
		"Summary",
	)
}

// ActionMetaValue completes meta update value based on --field flag on command line.
func ActionMetaValue() carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		flags := parseFlags(c.Args)
		switch strings.ToLower(flags["field"]) {
		case "status":
			return ActionStatus()
		case "intent":
			return ActionIntent()
		case "topics":
			return ActionTopics()
		case "doctype":
			return ActionDocTypes()
		default:
			return carapace.ActionValues() // freeform
		}
	})
}
