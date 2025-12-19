package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/templates"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// SkillListCommand lists skills
type SkillListCommand struct {
	*cmds.CommandDescription
}

// SkillListSettings holds the parameters for the skill list command
type SkillListSettings struct {
	Root                string   `glazed.parameter:"root"`
	Ticket              string   `glazed.parameter:"ticket"`
	Topics              []string `glazed.parameter:"topics"`
	File                string   `glazed.parameter:"file"`
	Dir                 string   `glazed.parameter:"dir"`
	PrintTemplateSchema bool     `glazed.parameter:"print-template-schema"`
	SchemaFormat        string   `glazed.parameter:"schema-format"`
}

func NewSkillListCommand() (*SkillListCommand, error) {
	return &SkillListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List skills"),
			cmds.WithLong(`Lists skills (documents with DocType=skill) with their WhatFor, WhenToUse, topics, and related files.

Skills are structured documentation artifacts that provide information about what a skill is for and when to use it.

Columns:
  skill,what_for,when_to_use,topics,related_paths,path

Examples:
  # Human output
  docmgr skill list
  docmgr skill list --ticket 001-ADD-CLAUDE-SKILLS
  docmgr skill list --topics backend,tooling
  docmgr skill list --file pkg/commands/add.go
  docmgr skill list --dir pkg/commands/

  # Scriptable (JSON)
  docmgr skill list --with-glaze-output --output json
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by ticket identifier"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"topics",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Filter by topics (comma-separated, matches any)"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithHelp("Find skills that reference this file path"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"dir",
					parameters.ParameterTypeString,
					parameters.WithHelp("Find skills that reference files in this directory"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"print-template-schema",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Print template schema after output (human mode only)"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"schema-format",
					parameters.ParameterTypeString,
					parameters.WithHelp("Template schema output format: json|yaml"),
					parameters.WithDefault("json"),
				),
			),
		),
	}, nil
}

func (c *SkillListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &SkillListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	fileQueryRaw := strings.TrimSpace(settings.File)
	dirQueryRaw := strings.TrimSpace(settings.Dir)

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return fmt.Errorf("failed to discover workspace: %w", err)
	}
	settings.Root = ws.Context().Root
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	scope := workspace.Scope{Kind: workspace.ScopeRepo}
	if strings.TrimSpace(settings.Ticket) != "" {
		scope = workspace.Scope{Kind: workspace.ScopeTicket, TicketID: strings.TrimSpace(settings.Ticket)}
	}

	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: scope,
		Filters: workspace.DocFilters{
			DocType:   "skill",
			Ticket:    strings.TrimSpace(settings.Ticket),
			TopicsAny: settings.Topics,
			RelatedFile: func() []string {
				if fileQueryRaw != "" {
					return []string{fileQueryRaw}
				}
				return nil
			}(),
			RelatedDir: func() []string {
				if dirQueryRaw != "" {
					return []string{dirQueryRaw}
				}
				return nil
			}(),
		},
		Options: workspace.DocQueryOptions{
			IncludeBody:        false,
			IncludeErrors:      false,
			IncludeDiagnostics: false,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to query skills: %w", err)
	}

	for _, handle := range res.Docs {
		if handle.Doc == nil {
			continue
		}
		doc := handle.Doc

		// Extract related file paths
		relatedPaths := make([]string, 0, len(doc.RelatedFiles))
		for _, rf := range doc.RelatedFiles {
			relatedPaths = append(relatedPaths, rf.Path)
		}

		row := types.NewRow(
			types.MRP("skill", doc.Title),
			types.MRP("what_for", doc.WhatFor),
			types.MRP("when_to_use", doc.WhenToUse),
			types.MRP("topics", strings.Join(doc.Topics, ",")),
			types.MRP("related_paths", strings.Join(relatedPaths, ",")),
			types.MRP("path", handle.Path),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &SkillListCommand{}

// Implement BareCommand for human-friendly output
func (c *SkillListCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &SkillListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = workspace.ResolveRoot(settings.Root)

	// If only printing template schema, skip all other processing and output
	if settings.PrintTemplateSchema {
		type SkillResult struct {
			Skill        string
			WhatFor      string
			WhenToUse    string
			Topics       []string
			RelatedPaths []string
			Path         string
		}
		templateData := map[string]interface{}{
			"TotalResults": 0,
			"Results": []SkillResult{
				{
					Skill:        "",
					WhatFor:      "",
					WhenToUse:    "",
					Topics:       []string{},
					RelatedPaths: []string{},
					Path:         "",
				},
			},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	fileQueryRaw := strings.TrimSpace(settings.File)
	dirQueryRaw := strings.TrimSpace(settings.Dir)

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return fmt.Errorf("failed to discover workspace: %w", err)
	}
	settings.Root = ws.Context().Root
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	scope := workspace.Scope{Kind: workspace.ScopeRepo}
	if strings.TrimSpace(settings.Ticket) != "" {
		scope = workspace.Scope{Kind: workspace.ScopeTicket, TicketID: strings.TrimSpace(settings.Ticket)}
	}

	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: scope,
		Filters: workspace.DocFilters{
			DocType:   "skill",
			Ticket:    strings.TrimSpace(settings.Ticket),
			TopicsAny: settings.Topics,
			RelatedFile: func() []string {
				if fileQueryRaw != "" {
					return []string{fileQueryRaw}
				}
				return nil
			}(),
			RelatedDir: func() []string {
				if dirQueryRaw != "" {
					return []string{dirQueryRaw}
				}
				return nil
			}(),
		},
		Options: workspace.DocQueryOptions{
			IncludeBody:        false,
			IncludeErrors:      false,
			IncludeDiagnostics: false,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to query skills: %w", err)
	}

	// Build template data
	type SkillResult struct {
		Skill        string
		WhatFor      string
		WhenToUse    string
		Topics       []string
		RelatedPaths []string
		Path         string
	}

	results := make([]SkillResult, 0, len(res.Docs))
	for _, handle := range res.Docs {
		if handle.Doc == nil {
			continue
		}
		doc := handle.Doc

		// Extract related file paths
		relatedPaths := make([]string, 0, len(doc.RelatedFiles))
		for _, rf := range doc.RelatedFiles {
			relatedPaths = append(relatedPaths, rf.Path)
		}

		results = append(results, SkillResult{
			Skill:        doc.Title,
			WhatFor:      doc.WhatFor,
			WhenToUse:    doc.WhenToUse,
			Topics:       doc.Topics,
			RelatedPaths: relatedPaths,
			Path:         handle.Path,
		})
	}

	// Human-friendly output
	for _, result := range results {
		fmt.Printf("Skill: %s\n", result.Skill)
		if result.WhatFor != "" {
			fmt.Printf("  What for: %s\n", result.WhatFor)
		}
		if result.WhenToUse != "" {
			fmt.Printf("  When to use: %s\n", result.WhenToUse)
		}
		if len(result.Topics) > 0 {
			fmt.Printf("  Topics: %s\n", strings.Join(result.Topics, ", "))
		}
		if len(result.RelatedPaths) > 0 {
			fmt.Printf("  Related files: %s\n", strings.Join(result.RelatedPaths, ", "))
		}
		fmt.Printf("  Path: %s\n", result.Path)
		fmt.Println()
	}

	// Render verb template if it exists
	templateData := map[string]interface{}{
		"TotalResults": len(results),
		"Results":      results,
	}
	absRoot := settings.Root
	if !filepath.IsAbs(absRoot) {
		if cwd, err := os.Getwd(); err == nil {
			absRoot = filepath.Join(cwd, absRoot)
		}
	}
	verbCandidates := [][]string{
		{"skill", "list"},
	}
	settingsMap := map[string]interface{}{
		"root":   settings.Root,
		"ticket": settings.Ticket,
		"topics": settings.Topics,
		"file":   settings.File,
		"dir":    settings.Dir,
	}
	_ = templates.RenderVerbTemplate(verbCandidates, absRoot, settingsMap, templateData)

	return nil
}

var _ cmds.BareCommand = &SkillListCommand{}

