package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-go-golems/docmgr/internal/skills"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
)

// SkillExportCommand exports a skill plan into an Agent Skills package.
type SkillExportCommand struct {
	*cmds.CommandDescription
}

// SkillExportSettings holds the parameters for skill export.
type SkillExportSettings struct {
	Root        string `glazed:"root"`
	Ticket      string `glazed:"ticket"`
	Skill       string `glazed:"skill"`
	Query       string `glazed:"query"`
	OutDir      string `glazed:"out-dir"`
	OutputSkill string `glazed:"output-skill"`
	Force       bool   `glazed:"force"`
}

func NewSkillExportCommand() (*SkillExportCommand, error) {
	return &SkillExportCommand{
		CommandDescription: cmds.NewCommandDescription(
			"export",
			cmds.WithShort("Export a skill plan as an Agent Skills package"),
			cmds.WithLong(`Exports a skill.yaml plan as a standard Agent Skills package (.skill).

This resolves file sources, captures binary help output, writes SKILL.md + references,
and optionally packages the result as a .skill archive when --output-skill is provided.

Examples:
  docmgr skill export glaze-help --output-skill dist/glaze-help.skill
  docmgr skill export ttmp/skills/glaze-help/skill.yaml --out-dir dist
  docmgr skill export api-design --ticket MEN-4242 --out-dir dist --output-skill dist/api-design.skill
`),
			cmds.WithArguments(
				fields.New(
					"query",
					fields.TypeString,
					fields.WithHelp("Skill query (name/title/slug/path)"),
					fields.WithRequired(false),
				),
			),
			cmds.WithFlags(
				fields.New(
					"root",
					fields.TypeString,
					fields.WithHelp("Root directory for docs"),
					fields.WithDefault("ttmp"),
				),
				fields.New(
					"ticket",
					fields.TypeString,
					fields.WithHelp("Include ticket-scoped skills for this ticket (workspace skills still included)"),
					fields.WithDefault(""),
				),
				fields.New(
					"skill",
					fields.TypeString,
					fields.WithHelp("Skill query (title/slug/path). Deprecated in favor of positional argument, but still supported."),
					fields.WithRequired(false),
				),
				fields.New(
					"out-dir",
					fields.TypeString,
					fields.WithHelp("Optional output directory for the expanded skill (defaults to temp dir)"),
					fields.WithDefault(""),
				),
				fields.New(
					"output-skill",
					fields.TypeString,
					fields.WithHelp("Optional output path for the .skill package (only created when set)"),
					fields.WithDefault(""),
				),
				fields.New(
					"force",
					fields.TypeBool,
					fields.WithHelp("Overwrite output files when they already exist"),
					fields.WithDefault(false),
				),
			),
		),
	}, nil
}

// Run implements BareCommand.
func (c *SkillExportCommand) Run(ctx context.Context, parsedValues *values.Values) error {
	settings := &SkillExportSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	query := strings.TrimSpace(settings.Skill)
	if query == "" {
		query = strings.TrimSpace(settings.Query)
	}
	if query == "" {
		return errors.New("skill query is required")
	}

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

	activeTicketOnly := strings.TrimSpace(settings.Ticket) == ""
	ticketStatusByID := map[string]string{}
	if _, ticketIndexDocs, err := queryTicketIndexDocs(ctx, settings.Root, "", ""); err == nil {
		for _, t := range ticketIndexDocs {
			id := strings.TrimSpace(t.Ticket)
			if id == "" {
				continue
			}
			ticketStatusByID[id] = strings.TrimSpace(t.Status)
		}
	}

	filtered := make([]skills.PlanHandle, 0, len(handles))
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
		filtered = append(filtered, handle)
	}

	candidates := buildSkillCandidates(ws, filtered, query)
	if len(candidates) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no skills found matching %q\n\n", query)
		fmt.Fprintf(os.Stderr, "Tip: Try matching by name, title, or path. Examples:\n")
		fmt.Fprintf(os.Stderr, "  docmgr skill export %q\n", query)
		fmt.Fprintf(os.Stderr, "  docmgr skill export ttmp/skills/<skill>/skill.yaml\n\n")
		return fmt.Errorf("no skills found matching %q", query)
	}

	sortSkillCandidates(candidates)

	if len(candidates) > 1 && candidates[0].Score == candidates[1].Score {
		fmt.Fprintf(os.Stdout, "Multiple skills match %q. Load one of these:\n\n", query)
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
			fmt.Printf("  Load: %s\n", buildSkillLoadCommand(loadCtx, plan.DisplayTitle(), plan.Skill.Name, cand.Handle.DisplayPath))
			fmt.Println()
		}
		return fmt.Errorf("multiple skills match %q", query)
	}

	h := candidates[0].Handle

	result, err := skills.ExportPlan(ctx, ws, h, skills.ExportOptions{
		OutDir:          settings.OutDir,
		OutputSkillPath: settings.OutputSkill,
		Force:           settings.Force,
	})
	if err != nil {
		return err
	}

	if result.PackagePath != "" {
		fmt.Fprintf(os.Stdout, "Exported skill to %s\n", result.PackagePath)
	} else {
		fmt.Fprintln(os.Stdout, "No .skill output requested (use --output-skill to create one)")
	}
	fmt.Fprintf(os.Stdout, "Skill directory: %s\n", result.SkillDir)
	return nil
}

var _ cmds.BareCommand = &SkillExportCommand{}
