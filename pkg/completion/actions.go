package completion

import (
	"fmt"
	"path/filepath"
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


