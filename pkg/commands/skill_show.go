package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/go-go-golems/docmgr/internal/paths"
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
	Root   string `glazed.parameter:"root"`
	Ticket string `glazed.parameter:"ticket"`
	// Skill is the legacy/explicit flag-based query (docmgr skill show --skill <query>)
	Skill string `glazed.parameter:"skill"`
	// Query is the positional argument (docmgr skill show <query>)
	Query string `glazed.parameter:"query"`
}

func NewSkillShowCommand() (*SkillShowCommand, error) {
	return &SkillShowCommand{
		CommandDescription: cmds.NewCommandDescription(
			"show",
			cmds.WithShort("Show detailed information about a skill"),
			cmds.WithLong(`Shows detailed information about a specific skill.

The skill can be located by a resilient multi-strategy matcher:
  - Title match (with or without the "Skill:" prefix)
  - Filename/slug match (including numeric prefixes like 01-foo.md)
  - Path match (absolute, repo-relative, docs-root-relative; file or directory)

If multiple skills match, candidates are listed with exact commands to load each one.

Examples:
  docmgr skill show "API Design"
  docmgr skill show api-design
  docmgr skill show ttmp/skills/test-driven-development.md
  docmgr skill show --skill "Skill: Test-Driven Development"
  docmgr skill show --ticket MEN-4242 api-design
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
					parameters.WithHelp("Limit search to a ticket workspace (useful when skills clash)"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"skill",
					parameters.ParameterTypeString,
					parameters.WithHelp("Skill query (title/slug/path). Deprecated in favor of positional argument, but still supported."),
					parameters.WithRequired(false),
				),
			),
		),
	}, nil
}

var (
	skillPrefixRe    = regexp.MustCompile(`(?i)^\s*skill\s*:\s*`)
	looseNumPrefixRe = regexp.MustCompile(`(?i)^\s*\d+[-_ ]+`)
	extMDRe          = regexp.MustCompile(`(?i)\.md$`)
	multiSpaceRe     = regexp.MustCompile(`\s+`)
	nonSlugCharRe    = regexp.MustCompile(`[^a-z0-9-]+`)
)

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

func stripSkillPrefix(s string) string {
	return strings.TrimSpace(skillPrefixRe.ReplaceAllString(strings.TrimSpace(s), ""))
}

func stripLeadingNumericPrefixLoose(s string) string {
	s = strings.TrimSpace(s)
	// Prefer the canonical NN-/NNN- stripping used elsewhere in docmgr.
	if stripped, _, _ := stripNumericPrefix(s); stripped != s {
		return strings.TrimSpace(stripped)
	}
	// Fallback: accept looser formats (e.g. "1-foo", "01_foo", "01 foo").
	return strings.TrimSpace(looseNumPrefixRe.ReplaceAllString(s, ""))
}

func stripMDExt(s string) string {
	return strings.TrimSpace(extMDRe.ReplaceAllString(strings.TrimSpace(s), ""))
}

func normalizeSpacesLower(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	s = multiSpaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func slugifyLower(s string) string {
	s = normalizeSpacesLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = nonSlugCharRe.ReplaceAllString(s, "")
	s = strings.Trim(s, "-")
	s = multiSpaceRe.ReplaceAllString(s, "-")
	return s
}

type skillCandidate struct {
	Handle workspace.DocHandle

	TitleLower         string
	TitleNoPrefixLower string
	TitleSlugLower     string
	TitleNoPrefixSlug  string

	FileStemLower      string
	FileStemNoNumLower string

	PathNorm paths.NormalizedPath

	Score int
	Why   string
}

func looksLikePathQuery(q string) bool {
	q = strings.TrimSpace(q)
	if q == "" {
		return false
	}
	lower := strings.ToLower(q)
	// We only want path scoring when the user likely provided a path, because
	// otherwise generic words (e.g. "websocket") can accidentally match directory names.
	return filepath.IsAbs(q) ||
		strings.Contains(q, "/") ||
		strings.Contains(q, "\\") ||
		strings.HasSuffix(lower, ".md")
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
		return fmt.Errorf("failed to discover workspace: %w", err)
	}
	settings.Root = ws.Context().Root
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: true}); err != nil {
		return fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	scope := workspace.Scope{Kind: workspace.ScopeRepo}
	if strings.TrimSpace(settings.Ticket) != "" {
		scope = workspace.Scope{Kind: workspace.ScopeTicket, TicketID: strings.TrimSpace(settings.Ticket)}
	}

	// Query skills
	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: scope,
		Filters: workspace.DocFilters{
			DocType: "skill",
			Ticket:  strings.TrimSpace(settings.Ticket),
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

	// When no --ticket is provided, only consider skills belonging to active-ish tickets.
	// Ticket-specific skills from completed/archived tickets are excluded by default.
	// Workspace-level skills (no Ticket) are always included.
	activeTicketOnly := strings.TrimSpace(settings.Ticket) == ""
	ticketStatusByID := map[string]string{}
	if activeTicketOnly {
		if _, ticketIndexDocs, err := queryTicketIndexDocs(ctx, settings.Root, "", ""); err == nil {
			for _, t := range ticketIndexDocs {
				if strings.TrimSpace(t.Ticket) == "" {
					continue
				}
				ticketStatusByID[strings.TrimSpace(t.Ticket)] = strings.TrimSpace(t.Status)
			}
		}
	}

	queryRaw := strings.TrimSpace(query)
	queryLower := strings.ToLower(queryRaw)
	queryNoPrefixLower := strings.ToLower(stripSkillPrefix(queryRaw))
	queryStemLower := strings.ToLower(stripMDExt(filepath.Base(queryRaw)))
	queryStemNoNumLower := strings.ToLower(stripLeadingNumericPrefixLoose(queryStemLower))
	querySlugLower := slugifyLower(queryRaw)
	querySlugNoPrefixLower := slugifyLower(stripSkillPrefix(queryRaw))

	var queryPathNorm paths.NormalizedPath
	if looksLikePathQuery(queryRaw) {
		queryPathNorm = ws.Resolver().Normalize(queryRaw)
	}

	var candidates []skillCandidate
	for _, handle := range res.Docs {
		if handle.Doc == nil {
			continue
		}

		if activeTicketOnly && strings.TrimSpace(handle.Doc.Ticket) != "" {
			if st, ok := ticketStatusByID[strings.TrimSpace(handle.Doc.Ticket)]; ok {
				if !isTicketActiveForSkillDefaultFilter(st) {
					continue
				}
			}
		}

		title := strings.TrimSpace(handle.Doc.Title)
		titleLower := strings.ToLower(title)
		titleNoPrefix := stripSkillPrefix(title)
		titleNoPrefixLower := strings.ToLower(titleNoPrefix)

		fileStem := stripMDExt(filepath.Base(handle.Path))
		fileStemNoNum := stripLeadingNumericPrefixLoose(fileStem)

		pathNorm := ws.Resolver().Normalize(handle.Path)

		cand := skillCandidate{
			Handle:             handle,
			TitleLower:         titleLower,
			TitleNoPrefixLower: titleNoPrefixLower,
			TitleSlugLower:     slugifyLower(title),
			TitleNoPrefixSlug:  slugifyLower(titleNoPrefix),
			FileStemLower:      strings.ToLower(fileStem),
			FileStemNoNumLower: strings.ToLower(fileStemNoNum),
			PathNorm:           pathNorm,
		}

		// Path matching: prefer existing, then representation overlap / suffix / substring.
		if looksLikePathQuery(queryRaw) && !queryPathNorm.Empty() {
			if queryPathNorm.Exists {
				// If query resolves to a directory, match skills under it.
				if queryPathNorm.Abs != "" {
					if fi, err := os.Stat(filepath.FromSlash(queryPathNorm.Abs)); err == nil && fi.IsDir() {
						if paths.DirectoryMatch(queryPathNorm, cand.PathNorm) {
							cand.Score = 1000
							cand.Why = "path-dir"
						}
					}
				}
				if cand.Score == 0 && paths.MatchPaths(queryPathNorm, cand.PathNorm) {
					cand.Score = 1100
					cand.Why = "path-file"
				}
			} else {
				// Non-existent path-like input: still allow fuzzy match against indexed paths.
				if paths.MatchPaths(queryPathNorm, cand.PathNorm) {
					cand.Score = 900
					cand.Why = "path-fuzzy"
				}
			}
		}

		// Exact title match (case-insensitive).
		if cand.Score == 0 && titleLower == queryLower {
			cand.Score = 800
			cand.Why = "title-exact"
		}
		// Exact title match without "Skill:" prefix.
		if cand.Score == 0 && titleNoPrefixLower != "" && titleNoPrefixLower == queryNoPrefixLower {
			cand.Score = 780
			cand.Why = "title-no-prefix-exact"
		}

		// Exact filename/slug match.
		if cand.Score == 0 && (cand.FileStemLower == queryLower || cand.FileStemLower == queryStemLower || cand.FileStemNoNumLower == queryStemNoNumLower) {
			cand.Score = 760
			cand.Why = "filename-exact"
		}
		if cand.Score == 0 && (cand.FileStemLower == querySlugLower || cand.FileStemNoNumLower == querySlugLower || cand.FileStemNoNumLower == querySlugNoPrefixLower) {
			cand.Score = 750
			cand.Why = "filename-slug-exact"
		}

		// Slugified title match.
		if cand.Score == 0 && (cand.TitleSlugLower == querySlugLower || cand.TitleNoPrefixSlug == querySlugLower || cand.TitleNoPrefixSlug == querySlugNoPrefixLower) {
			cand.Score = 740
			cand.Why = "title-slug-exact"
		}

		// Contains matches (lower priority).
		if cand.Score == 0 && queryLower != "" && strings.Contains(titleLower, queryLower) {
			cand.Score = 600
			cand.Why = "title-contains"
		}
		if cand.Score == 0 && queryNoPrefixLower != "" && titleNoPrefixLower != "" && strings.Contains(titleNoPrefixLower, queryNoPrefixLower) {
			cand.Score = 590
			cand.Why = "title-no-prefix-contains"
		}
		if cand.Score == 0 && queryLower != "" && strings.Contains(cand.FileStemLower, queryLower) {
			cand.Score = 580
			cand.Why = "filename-contains"
		}
		if cand.Score == 0 && querySlugLower != "" && (strings.Contains(cand.TitleSlugLower, querySlugLower) || strings.Contains(cand.FileStemNoNumLower, querySlugLower)) {
			cand.Score = 570
			cand.Why = "slug-contains"
		}

		if cand.Score > 0 {
			candidates = append(candidates, cand)
		}
	}

	if len(candidates) == 0 {
		// Provide a helpful, actionable list (same UX principle as skill list).
		fmt.Fprintf(os.Stderr, "Error: no skills found matching %q\n\n", queryRaw)
		fmt.Fprintf(os.Stderr, "Tip: Try matching by title, filename, or path. Examples:\n")
		fmt.Fprintf(os.Stderr, "  docmgr skill show %q\n", queryRaw)
		fmt.Fprintf(os.Stderr, "  docmgr skill show --skill %q\n", queryRaw)
		fmt.Fprintf(os.Stderr, "  docmgr skill show ttmp/skills/<skill>.md\n\n")
		return fmt.Errorf("no skills found matching %q", queryRaw)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		// Stabilize ordering for deterministic output.
		ti := ""
		tj := ""
		if candidates[i].Handle.Doc != nil {
			ti = candidates[i].Handle.Doc.Title
		}
		if candidates[j].Handle.Doc != nil {
			tj = candidates[j].Handle.Doc.Title
		}
		if ti != tj {
			return ti < tj
		}
		return candidates[i].Handle.Path < candidates[j].Handle.Path
	})

	// If top score ties, treat as ambiguity and print candidate load commands.
	if len(candidates) > 1 && candidates[0].Score == candidates[1].Score {
		fmt.Fprintf(os.Stdout, "Multiple skills match %q. Load one of these:\n\n", queryRaw)
		defaultRoot := workspace.ResolveRoot("ttmp")
		_, ticketIndexDocs, _ := queryTicketIndexDocs(ctx, settings.Root, "", "")
		ticketTitleByID := map[string]string{}
		for _, t := range ticketIndexDocs {
			if strings.TrimSpace(t.Ticket) == "" {
				continue
			}
			ticketTitleByID[strings.TrimSpace(t.Ticket)] = strings.TrimSpace(t.Title)
		}

		// Build a uniqueness index for load command generation.
		titleCounts := map[string]int{}
		slugCounts := map[string]int{}
		for _, cand := range candidates {
			if cand.Handle.Doc == nil {
				continue
			}
			titleNoPrefix := strings.TrimSpace(stripSkillPrefix(cand.Handle.Doc.Title))
			if titleNoPrefix == "" {
				titleNoPrefix = strings.TrimSpace(cand.Handle.Doc.Title)
			}
			titleCounts[strings.ToLower(titleNoPrefix)]++
			slugCounts[strings.ToLower(skillSlugFromPath(cand.Handle.Path))]++
		}

		loadCtx := skillLoadCommandContext{
			EffectiveRoot: settings.Root,
			DefaultRoot:   defaultRoot,
			TicketFilter:  strings.TrimSpace(settings.Ticket),
			TitleCounts:   titleCounts,
			SlugCounts:    slugCounts,
		}

		for _, cand := range candidates {
			doc := cand.Handle.Doc
			if doc == nil {
				continue
			}
			fmt.Printf("Skill: %s\n", doc.Title)
			if strings.TrimSpace(doc.Ticket) != "" {
				tt := ticketTitleByID[strings.TrimSpace(doc.Ticket)]
				if strings.TrimSpace(tt) != "" {
					fmt.Printf("  Ticket: %s — %s\n", strings.TrimSpace(doc.Ticket), strings.TrimSpace(tt))
				} else {
					fmt.Printf("  Ticket: %s\n", strings.TrimSpace(doc.Ticket))
				}
			}
			fmt.Printf("  Load: %s\n", buildSkillLoadCommand(loadCtx, doc.Title, cand.Handle.Path))
			fmt.Println()
		}
		return fmt.Errorf("multiple skills match %q", queryRaw)
	}

	// Display skill details (best match).
	h := candidates[0].Handle
	doc := h.Doc

	fmt.Printf("Title: %s\n", doc.Title)
	if doc.Ticket != "" {
		ticketTitle := ""
		if _, tickets, err := queryTicketIndexDocs(ctx, settings.Root, doc.Ticket, ""); err == nil && len(tickets) > 0 {
			ticketTitle = strings.TrimSpace(tickets[0].Title)
		}
		if ticketTitle != "" {
			fmt.Printf("Ticket: %s — %s\n", doc.Ticket, ticketTitle)
		} else {
			fmt.Printf("Ticket: %s\n", doc.Ticket)
		}
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
