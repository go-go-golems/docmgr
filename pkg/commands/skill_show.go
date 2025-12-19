package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// SkillShowCommand shows detailed information about a skill
type SkillShowCommand struct {
	*cmds.CommandDescription
}

// SkillShowSettings holds command parameters
type SkillShowSettings struct {
	Root  string `glazed.parameter:"root"`
	Skill string `glazed.parameter:"skill"`
}

func NewSkillShowCommand() (*SkillShowCommand, error) {
	return &SkillShowCommand{
		CommandDescription: cmds.NewCommandDescription(
			"show",
			cmds.WithShort("Show detailed information about a skill"),
			cmds.WithLong(`Shows detailed information about a specific skill.

The skill name can be matched by:
  - Exact title match (case-insensitive)
  - Case-insensitive contains match

If multiple skills match, the first match is shown and a warning is printed.

Examples:
  docmgr skill show "API Design"
  docmgr skill show api-design
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"skill",
					parameters.ParameterTypeString,
					parameters.WithHelp("Skill name to show (matched against title)"),
					parameters.WithRequired(true),
				),
			),
		),
	}, nil
}

// Run implements BareCommand (show commands typically use human-friendly output)
func (c *SkillShowCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &SkillShowSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	if strings.TrimSpace(settings.Skill) == "" {
		return fmt.Errorf("skill name is required")
	}

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return fmt.Errorf("failed to discover workspace: %w", err)
	}
	settings.Root = ws.Context().Root
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: true}); err != nil {
		return fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	// Query all skills
	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeRepo},
		Filters: workspace.DocFilters{
			DocType: "skill",
		},
		Options: workspace.DocQueryOptions{
			IncludeBody:        true,
			IncludeErrors:      false,
			IncludeDiagnostics: false,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to query skills: %w", err)
	}

	// Match skills by title (case-insensitive exact or contains)
	queryLower := strings.ToLower(strings.TrimSpace(settings.Skill))
	var matches []workspace.DocHandle
	for _, handle := range res.Docs {
		if handle.Doc == nil {
			continue
		}
		titleLower := strings.ToLower(handle.Doc.Title)
		if titleLower == queryLower || strings.Contains(titleLower, queryLower) {
			matches = append(matches, handle)
		}
	}

	if len(matches) == 0 {
		return fmt.Errorf("no skills found matching %q", settings.Skill)
	}

	// Handle ambiguity: show first match + warn if multiple
	if len(matches) > 1 {
		fmt.Fprintf(os.Stderr, "Warning: Multiple skills found matching %q, showing first match:\n", settings.Skill)
		for _, m := range matches {
			fmt.Fprintf(os.Stderr, "  - %s (%s)\n", m.Doc.Title, m.Path)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Display skill details
	h := matches[0]
	doc := h.Doc

	fmt.Printf("Title: %s\n", doc.Title)
	if doc.Ticket != "" {
		fmt.Printf("Ticket: %s\n", doc.Ticket)
	}
	if doc.Status != "" {
		fmt.Printf("Status: %s\n", doc.Status)
	}
	if doc.WhatFor != "" {
		fmt.Printf("\nWhat this skill is for:\n%s\n", doc.WhatFor)
	}
	if doc.WhenToUse != "" {
		fmt.Printf("\nWhen to use this skill:\n%s\n", doc.WhenToUse)
	}
	if len(doc.Topics) > 0 {
		fmt.Printf("\nTopics: %s\n", strings.Join(doc.Topics, ", "))
	}
	if len(doc.RelatedFiles) > 0 {
		fmt.Printf("\nRelated Files:\n")
		for _, rf := range doc.RelatedFiles {
			fmt.Printf("  - %s", rf.Path)
			if rf.Note != "" {
				fmt.Printf(": %s", rf.Note)
			}
			fmt.Printf("\n")
		}
	}
	fmt.Printf("\nPath: %s\n", filepath.ToSlash(h.Path))
	if h.Body != "" {
		fmt.Printf("\n%s\n", h.Body)
	}

	return nil
}

var _ cmds.BareCommand = &SkillShowCommand{}

