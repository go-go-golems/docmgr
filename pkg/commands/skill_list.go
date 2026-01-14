package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/internal/skills"
	"github.com/go-go-golems/docmgr/internal/templates"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
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
			cmds.WithLong(`Lists skill.yaml plans with their WhatFor, WhenToUse, topics, and related files.

Skills are plan-based documentation artifacts that package references and help output
into Agent Skills format.

Columns:
  skill,what_for,when_to_use,topics,related_paths,path,load_command

Examples:
  # Human output
  docmgr skill list
  docmgr skill list --ticket 001-ADD-CLAUDE-SKILLS
  docmgr skill list --topics backend,tooling
  docmgr skill list --file glazed/pkg/doc/topics/01-help-system.md
  docmgr skill list --dir glazed/pkg/doc/topics/

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
					parameters.WithHelp("Include ticket-scoped skills for this ticket (workspace skills still included)"),
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

	results, err := collectSkillListResults(ctx, settings)
	if err != nil {
		return err
	}

	for _, result := range results {
		row := types.NewRow(
			types.MRP("skill", result.Skill),
			types.MRP("what_for", result.WhatFor),
			types.MRP("when_to_use", result.WhenToUse),
			types.MRP("topics", strings.Join(result.Topics, ",")),
			types.MRP("related_paths", strings.Join(result.RelatedPaths, ",")),
			types.MRP("path", result.Path),
			types.MRP("load_command", result.LoadCommand),
			types.MRP("ticket", result.Ticket),
			types.MRP("ticket_title", result.TicketTitle),
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
			LoadCommand  string
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
					LoadCommand:  "",
				},
			},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	results, err := collectSkillListResults(ctx, settings)
	if err != nil {
		return err
	}

	for _, result := range results {
		fmt.Printf("Skill: %s\n", result.Skill)
		if strings.TrimSpace(result.Ticket) != "" {
			line := strings.TrimSpace(result.Ticket)
			if strings.TrimSpace(result.TicketTitle) != "" {
				line = fmt.Sprintf("%s — %s", line, strings.TrimSpace(result.TicketTitle))
			}
			fmt.Printf("  Ticket: %s\n", line)
		}
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
		fmt.Printf("  Load: %s\n", result.LoadCommand)
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

type skillListResult struct {
	Skill        string
	WhatFor      string
	WhenToUse    string
	Topics       []string
	RelatedPaths []string
	Path         string
	LoadCommand  string
	Ticket       string
	TicketTitle  string
}

func collectSkillListResults(ctx context.Context, settings *SkillListSettings) ([]skillListResult, error) {
	if settings == nil {
		return nil, errors.New("settings required")
	}
	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return nil, fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return nil, errors.Wrap(err, "failed to discover workspace")
	}
	settings.Root = ws.Context().Root
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return nil, errors.Wrap(err, "failed to initialize workspace index")
	}

	_, ticketIndexDocs, _ := queryTicketIndexDocs(ctx, settings.Root, "", "")
	ticketTitleByID := map[string]string{}
	ticketStatusByID := map[string]string{}
	for _, t := range ticketIndexDocs {
		ticketID := strings.TrimSpace(t.Ticket)
		if ticketID == "" {
			continue
		}
		ticketTitleByID[ticketID] = strings.TrimSpace(t.Title)
		ticketStatusByID[ticketID] = strings.TrimSpace(t.Status)
	}

	handles, err := skills.DiscoverPlans(ctx, ws, skills.DiscoverOptions{
		TicketID:          strings.TrimSpace(settings.Ticket),
		IncludeWorkspace:  true,
		IncludeAllTickets: strings.TrimSpace(settings.Ticket) == "",
	})
	if err != nil {
		return nil, err
	}

	fileQueryRaw := strings.TrimSpace(settings.File)
	dirQueryRaw := strings.TrimSpace(settings.Dir)

	var fileQuery paths.NormalizedPath
	var dirQuery paths.NormalizedPath
	if fileQueryRaw != "" {
		fileQuery = ws.Resolver().Normalize(fileQueryRaw)
	}
	if dirQueryRaw != "" {
		dirQuery = ws.Resolver().Normalize(dirQueryRaw)
	}

	activeTicketOnly := strings.TrimSpace(settings.Ticket) == ""
	var filtered []skills.PlanHandle
	for _, handle := range handles {
		if handle.Plan == nil {
			continue
		}
		if activeTicketOnly && strings.TrimSpace(handle.TicketID) != "" {
			st := strings.ToLower(strings.TrimSpace(ticketStatusByID[strings.TrimSpace(handle.TicketID)]))
			if !isTicketActiveForSkillDefaultFilter(st) {
				continue
			}
		}
		if !matchesAnyTopic(handle.Plan.Skill.Topics, settings.Topics) {
			continue
		}
		if fileQueryRaw != "" && !matchesFileSources(fileQuery, handle.SourceFiles) {
			continue
		}
		if dirQueryRaw != "" && !matchesDirSources(dirQuery, handle.SourceFiles) {
			continue
		}
		filtered = append(filtered, handle)
	}

	// Build a uniqueness index for load command generation.
	titleCounts := map[string]int{}
	nameCounts := map[string]int{}
	slugCounts := map[string]int{}
	for _, handle := range filtered {
		if handle.Plan == nil {
			continue
		}
		titleNoPrefix := strings.TrimSpace(stripSkillPrefix(handle.Plan.DisplayTitle()))
		if titleNoPrefix == "" {
			titleNoPrefix = strings.TrimSpace(handle.Plan.DisplayTitle())
		}
		titleCounts[strings.ToLower(titleNoPrefix)]++
		nameCounts[strings.ToLower(strings.TrimSpace(handle.Plan.Skill.Name))]++
		slugCounts[strings.ToLower(skillSlugFromPath(handle.Path))]++
	}

	loadCtx := skillLoadCommandContext{
		EffectiveRoot: settings.Root,
		DefaultRoot:   workspace.ResolveRoot("ttmp"),
		TicketFilter:  strings.TrimSpace(settings.Ticket),
		TitleCounts:   titleCounts,
		NameCounts:    nameCounts,
		SlugCounts:    slugCounts,
	}

	results := make([]skillListResult, 0, len(filtered))
	for _, handle := range filtered {
		if handle.Plan == nil {
			continue
		}
		plan := handle.Plan

		relatedPaths := make([]string, 0, len(handle.SourceFiles))
		for _, rf := range handle.SourceFiles {
			relatedPaths = append(relatedPaths, rf.Path)
		}

		results = append(results, skillListResult{
			Skill:        plan.DisplayTitle(),
			WhatFor:      plan.Skill.WhatFor,
			WhenToUse:    plan.Skill.WhenToUse,
			Topics:       plan.Skill.Topics,
			RelatedPaths: relatedPaths,
			Path:         handle.DisplayPath,
			LoadCommand:  buildSkillLoadCommand(loadCtx, plan.DisplayTitle(), plan.Skill.Name, handle.DisplayPath),
			Ticket:       handle.TicketID,
			TicketTitle:  ticketTitleByID[strings.TrimSpace(handle.TicketID)],
		})
	}

	return results, nil
}

func matchesAnyTopic(planTopics []string, filter []string) bool {
	if len(filter) == 0 {
		return true
	}
	if len(planTopics) == 0 {
		return false
	}
	wanted := make([]string, 0, len(filter))
	for _, topic := range filter {
		t := strings.ToLower(strings.TrimSpace(topic))
		if t != "" {
			wanted = append(wanted, t)
		}
	}
	if len(wanted) == 0 {
		return true
	}
	for _, topic := range planTopics {
		topicLower := strings.ToLower(strings.TrimSpace(topic))
		for _, wantedTopic := range wanted {
			if topicLower == wantedTopic {
				return true
			}
		}
	}
	return false
}

func matchesFileSources(query paths.NormalizedPath, sources []skills.SourceFile) bool {
	if query.Empty() {
		return false
	}
	for _, source := range sources {
		if paths.MatchPaths(query, source.Normalized) {
			return true
		}
	}
	return false
}

func matchesDirSources(query paths.NormalizedPath, sources []skills.SourceFile) bool {
	if query.Empty() {
		return false
	}
	for _, source := range sources {
		if paths.DirectoryMatch(query, source.Normalized) {
			return true
		}
	}
	return false
}
