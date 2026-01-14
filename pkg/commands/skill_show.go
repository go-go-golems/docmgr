package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-go-golems/docmgr/internal/skills"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
)

// SkillShowCommand shows detailed information about a skill
type SkillShowCommand struct {
	*cmds.CommandDescription
}

// SkillShowSettings holds command parameters
type SkillShowSettings struct {
	Root    string `glazed.parameter:"root"`
	Ticket  string `glazed.parameter:"ticket"`
	Skill   string `glazed.parameter:"skill"`
	Query   string `glazed.parameter:"query"`
	Resolve bool   `glazed.parameter:"resolve"`
}

func NewSkillShowCommand() (*SkillShowCommand, error) {
	return &SkillShowCommand{
		CommandDescription: cmds.NewCommandDescription(
			"show",
			cmds.WithShort("Show detailed information about a skill"),
			cmds.WithLong(`Shows detailed information about a specific skill plan.

The skill can be located by a resilient multi-strategy matcher:
  - Name or title match (with or without the "Skill:" prefix)
  - Slug match (skill name or plan directory name)
  - Path match (absolute, repo-relative, docs-root-relative; file or directory)

If multiple skills match, candidates are listed with exact commands to load each one.

Examples:
  docmgr skill show "API Design"
  docmgr skill show api-design
  docmgr skill show ttmp/skills/glaze-help/skill.yaml
  docmgr skill show --skill "Skill: Test-Driven Development"
  docmgr skill show --ticket MEN-4242 api-design
  docmgr skill show api-design --resolve
`),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"query",
					parameters.ParameterTypeString,
					parameters.WithHelp("Skill query (title/slug/path)"),
					parameters.WithRequired(false),
				),
			),
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
					"skill",
					parameters.ParameterTypeString,
					parameters.WithHelp("Skill query (title/slug/path). Deprecated in favor of positional argument, but still supported."),
					parameters.WithRequired(false),
				),
				parameters.NewParameterDefinition(
					"resolve",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Resolve sources (reads files and runs binary help commands)"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

// isTicketActiveForSkillDefaultFilter decides whether a ticket status should be treated as
// "active enough" to keep ticket-scoped skills visible when the user did NOT specify --ticket.
//
// Rationale: "review" and "draft" are still in-progress states; only completed/archived tickets
// should be hidden by default to reduce noise.
func isTicketActiveForSkillDefaultFilter(st string) bool {
	switch strings.ToLower(strings.TrimSpace(st)) {
	case "active", "review", "draft":
		return true
	default:
		return false
	}
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

	query := strings.TrimSpace(settings.Skill)
	if query == "" {
		query = strings.TrimSpace(settings.Query)
	}
	if query == "" {
		return fmt.Errorf("skill query is required (provide a positional <query> or --skill)")
	}

	// Apply config root if present (consistent with other verbs like skill list).
	settings.Root = workspace.ResolveRoot(settings.Root)
	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return errors.Wrap(err, "failed to discover workspace")
	}
	settings.Root = ws.Context().Root
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return errors.Wrap(err, "failed to initialize workspace index")
	}

	handles, err := skills.DiscoverPlans(ctx, ws, skills.DiscoverOptions{
		TicketID:          strings.TrimSpace(settings.Ticket),
		IncludeWorkspace:  true,
		IncludeAllTickets: strings.TrimSpace(settings.Ticket) == "",
	})
	if err != nil {
		return err
	}

	// When no --ticket is provided, only consider skills belonging to active-ish tickets.
	activeTicketOnly := strings.TrimSpace(settings.Ticket) == ""

	// Preload ticket index docs once per invocation (used for default filtering + ticket title display).
	ticketStatusByID := map[string]string{}
	ticketTitleByID := map[string]string{}
	{
		ticketFilter := strings.TrimSpace(settings.Ticket)
		if activeTicketOnly {
			ticketFilter = ""
		}
		if _, ticketIndexDocs, err := queryTicketIndexDocs(ctx, settings.Root, ticketFilter, ""); err == nil {
			for _, t := range ticketIndexDocs {
				id := strings.TrimSpace(t.Ticket)
				if id == "" {
					continue
				}
				ticketStatusByID[id] = strings.TrimSpace(t.Status)
				ticketTitleByID[id] = strings.TrimSpace(t.Title)
			}
		}
	}

	filtered := make([]skills.PlanHandle, 0, len(handles))
	for _, handle := range handles {
		if handle.Plan == nil {
			continue
		}
		if activeTicketOnly && strings.TrimSpace(handle.TicketID) != "" {
			if st, ok := ticketStatusByID[strings.TrimSpace(handle.TicketID)]; ok {
				if !isTicketActiveForSkillDefaultFilter(st) {
					continue
				}
			}
		}
		filtered = append(filtered, handle)
	}

	queryRaw := strings.TrimSpace(query)
	candidates := buildSkillCandidates(ws, filtered, queryRaw)

	if len(candidates) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no skills found matching %q\n\n", queryRaw)
		fmt.Fprintf(os.Stderr, "Tip: Try matching by name, title, or path. Examples:\n")
		fmt.Fprintf(os.Stderr, "  docmgr skill show %q\n", queryRaw)
		fmt.Fprintf(os.Stderr, "  docmgr skill show --skill %q\n", queryRaw)
		fmt.Fprintf(os.Stderr, "  docmgr skill show ttmp/skills/<skill>/skill.yaml\n\n")
		return fmt.Errorf("no skills found matching %q", queryRaw)
	}

	sortSkillCandidates(candidates)

	if len(candidates) > 1 && candidates[0].Score == candidates[1].Score {
		fmt.Fprintf(os.Stdout, "Multiple skills match %q. Load one of these:\n\n", queryRaw)
		defaultRoot := workspace.ResolveRoot("ttmp")

		titleCounts := map[string]int{}
		nameCounts := map[string]int{}
		slugCounts := map[string]int{}
		for _, cand := range candidates {
			if cand.Handle.Plan == nil {
				continue
			}
			titleNoPrefix := strings.TrimSpace(stripSkillPrefix(cand.Handle.Plan.DisplayTitle()))
			if titleNoPrefix == "" {
				titleNoPrefix = strings.TrimSpace(cand.Handle.Plan.DisplayTitle())
			}
			titleCounts[strings.ToLower(titleNoPrefix)]++
			nameCounts[strings.ToLower(strings.TrimSpace(cand.Handle.Plan.Skill.Name))]++
			slugCounts[strings.ToLower(skillSlugFromPath(cand.Handle.Path))]++
		}

		loadCtx := skillLoadCommandContext{
			EffectiveRoot: settings.Root,
			DefaultRoot:   defaultRoot,
			TicketFilter:  strings.TrimSpace(settings.Ticket),
			TitleCounts:   titleCounts,
			NameCounts:    nameCounts,
			SlugCounts:    slugCounts,
		}

		for _, cand := range candidates {
			plan := cand.Handle.Plan
			if plan == nil {
				continue
			}
			fmt.Printf("Skill: %s\n", plan.DisplayTitle())
			fmt.Printf("  Name: %s\n", plan.Skill.Name)
			if strings.TrimSpace(cand.Handle.TicketID) != "" {
				tt := ticketTitleByID[strings.TrimSpace(cand.Handle.TicketID)]
				if strings.TrimSpace(tt) != "" {
					fmt.Printf("  Ticket: %s — %s\n", strings.TrimSpace(cand.Handle.TicketID), strings.TrimSpace(tt))
				} else {
					fmt.Printf("  Ticket: %s\n", strings.TrimSpace(cand.Handle.TicketID))
				}
			}
			fmt.Printf("  Load: %s\n", buildSkillLoadCommand(loadCtx, plan.DisplayTitle(), plan.Skill.Name, cand.Handle.DisplayPath))
			fmt.Println()
		}
		return fmt.Errorf("multiple skills match %q", queryRaw)
	}

	h := candidates[0].Handle
	plan := h.Plan

	fmt.Printf("Title: %s\n", plan.DisplayTitle())
	fmt.Printf("Name: %s\n", plan.Skill.Name)
	if strings.TrimSpace(h.TicketID) != "" {
		ticketTitle := strings.TrimSpace(ticketTitleByID[strings.TrimSpace(h.TicketID)])
		if ticketTitle != "" {
			fmt.Printf("Ticket: %s — %s\n", h.TicketID, ticketTitle)
		} else {
			fmt.Printf("Ticket: %s\n", h.TicketID)
		}
	}
	if strings.TrimSpace(plan.Skill.Description) != "" {
		fmt.Printf("Description: %s\n", strings.TrimSpace(plan.Skill.Description))
	}
	if strings.TrimSpace(plan.Skill.WhatFor) != "" {
		fmt.Printf("\nWhat this skill is for:\n%s\n", strings.TrimSpace(plan.Skill.WhatFor))
	}
	if strings.TrimSpace(plan.Skill.WhenToUse) != "" {
		fmt.Printf("\nWhen to use this skill:\n%s\n", strings.TrimSpace(plan.Skill.WhenToUse))
	}
	if len(plan.Skill.Topics) > 0 {
		fmt.Printf("\nTopics: %s\n", strings.Join(plan.Skill.Topics, ", "))
	}
	if strings.TrimSpace(plan.Skill.Compatibility) != "" {
		fmt.Printf("\nCompatibility: %s\n", strings.TrimSpace(plan.Skill.Compatibility))
	}
	if strings.TrimSpace(plan.Skill.License) != "" {
		fmt.Printf("\nLicense: %s\n", strings.TrimSpace(plan.Skill.License))
	}

	if len(plan.Sources) > 0 {
		fmt.Printf("\nSources:\n")
		for _, source := range plan.Sources {
			markers := []string{}
			switch strings.ToLower(strings.TrimSpace(source.Type)) {
			case "file":
				line := fmt.Sprintf("  - file: %s", source.Path)
				if strings.TrimSpace(source.Output) != "" {
					line += fmt.Sprintf(" -> %s", source.Output)
				}
				if source.StripFrontmatter {
					markers = append(markers, "strip-frontmatter")
				}
				if source.AppendToBody {
					markers = append(markers, "append-to-body")
				}
				if len(markers) > 0 {
					line += fmt.Sprintf(" (%s)", strings.Join(markers, ", "))
				}
				fmt.Println(line)
			case "binary-help":
				line := fmt.Sprintf("  - binary-help: %s help %s", source.Binary, source.Topic)
				if strings.TrimSpace(source.Output) != "" {
					line += fmt.Sprintf(" -> %s", source.Output)
				}
				if strings.TrimSpace(source.Wrap) != "" {
					markers = append(markers, fmt.Sprintf("wrap: %s", strings.TrimSpace(source.Wrap)))
				}
				if source.AppendToBody {
					markers = append(markers, "append-to-body")
				}
				if len(markers) > 0 {
					line += fmt.Sprintf(" (%s)", strings.Join(markers, ", "))
				}
				fmt.Println(line)
			default:
				fmt.Printf("  - %s\n", source.Type)
			}
		}
	}

	fmt.Printf("\nPlan Path: %s\n", h.DisplayPath)

	if settings.Resolve {
		resolved, err := skills.ResolvePlan(ctx, ws, h, skills.ResolveOptions{AllowBinary: true})
		if err != nil {
			return err
		}
		fmt.Printf("\nResolved Sources:\n")
		for _, res := range resolved {
			fmt.Printf("\n--- %s (%s) ---\n", res.OutputPath, res.Source.Type)
			_, _ = os.Stdout.Write(res.Content)
			if len(res.Content) > 0 && res.Content[len(res.Content)-1] != '\n' {
				fmt.Println()
			}
		}
	}

	return nil
}

var _ cmds.BareCommand = &SkillShowCommand{}
