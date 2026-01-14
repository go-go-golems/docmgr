package commands

import (
	"fmt"
	"path/filepath"
	"strings"
)

type skillLoadCommandContext struct {
	// EffectiveRoot is the docs root that the current command is using (absolute).
	EffectiveRoot string
	// DefaultRoot is what `docmgr` would resolve by default for --root=ttmp in this environment.
	// If EffectiveRoot == DefaultRoot, we omit --root in generated commands.
	DefaultRoot string
	// TicketFilter is the current --ticket filter (if any). We include --ticket only
	// when it materially helps disambiguation (and we aren't already using a full path).
	TicketFilter string

	// TitleCounts counts normalized titles (without "Skill:" prefix) across the listing.
	TitleCounts map[string]int
	// NameCounts counts normalized skill names across the listing.
	NameCounts map[string]int
	// SlugCounts counts filename slugs across the listing.
	SlugCounts map[string]int
}

// buildSkillLoadCommand returns a short, resilient, copy/pasteable command that loads a specific skill.
//
// Identifier order:
//
//	skill name -> filename slug -> title (without "Skill:") -> full path
//
// Extra rule: if multiple skills share the same title in this listing, we use full path.
func buildSkillLoadCommand(ctx skillLoadCommandContext, docTitle string, skillName string, docPath string) string {
	titleNoPrefix := strings.TrimSpace(stripSkillPrefix(docTitle))
	if titleNoPrefix == "" {
		titleNoPrefix = strings.TrimSpace(docTitle)
	}
	titleKey := strings.ToLower(titleNoPrefix)

	slug := skillSlugFromPath(docPath)
	slugKey := strings.ToLower(slug)

	nameKey := strings.ToLower(strings.TrimSpace(skillName))

	titleDup := ctx.TitleCounts != nil && ctx.TitleCounts[titleKey] > 1
	nameDup := ctx.NameCounts != nil && nameKey != "" && ctx.NameCounts[nameKey] > 1

	query := ""
	queryIsPath := false

	// If names clash in the listing, go straight to full path.
	if titleDup {
		query = docPath
		queryIsPath = true
	} else if nameKey != "" && ctx.NameCounts != nil && !nameDup {
		query = strings.TrimSpace(skillName)
	} else if slug != "" && ctx.SlugCounts != nil && ctx.SlugCounts[slugKey] == 1 {
		query = slug
	} else {
		// Title fallback (without "Skill:" prefix).
		if titleNoPrefix != "" && ctx.TitleCounts != nil && ctx.TitleCounts[titleKey] == 1 {
			query = titleNoPrefix
		} else {
			query = docPath
			queryIsPath = true
		}
	}

	parts := []string{"docmgr skill show"}

	// Only include --root if it differs from the default root.
	if strings.TrimSpace(ctx.EffectiveRoot) != "" && strings.TrimSpace(ctx.DefaultRoot) != "" {
		if filepath.Clean(ctx.EffectiveRoot) != filepath.Clean(ctx.DefaultRoot) {
			parts = append(parts, fmt.Sprintf("--root %q", ctx.EffectiveRoot))
		}
	}

	// Only include --ticket when it helps disambiguation and we aren't already using a path.
	if !queryIsPath && strings.TrimSpace(ctx.TicketFilter) != "" {
		parts = append(parts, fmt.Sprintf("--ticket %q", strings.TrimSpace(ctx.TicketFilter)))
	}

	// Prefer unquoted slugs; quote titles/paths.
	if query == slug && query != "" && !strings.ContainsAny(query, " \t\"'") {
		parts = append(parts, query)
	} else {
		parts = append(parts, fmt.Sprintf("%q", query))
	}

	return strings.Join(parts, " ")
}

func skillSlugFromPath(docPath string) string {
	base := filepath.Base(docPath)
	if strings.EqualFold(base, "skill.yaml") || strings.EqualFold(base, "skill.yml") {
		base = filepath.Base(filepath.Dir(docPath))
	}
	base = stripMDExt(base)
	// Prefer canonical NN-/NNN- stripping used by docmgr.
	if stripped, _, _ := stripNumericPrefix(base); stripped != "" {
		base = stripped
	}
	return strings.TrimSpace(base)
}

func stripMDExt(s string) string {
	trimmed := strings.TrimSpace(s)
	lower := strings.ToLower(trimmed)
	for _, ext := range []string{".md", ".yaml", ".yml"} {
		if strings.HasSuffix(lower, ext) {
			return strings.TrimSpace(trimmed[:len(trimmed)-len(ext)])
		}
	}
	return trimmed
}
