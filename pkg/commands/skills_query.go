package commands

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/internal/skills"
	"github.com/go-go-golems/docmgr/internal/workspace"
)

var (
	skillPrefixRe = regexp.MustCompile(`(?i)^\s*skill\s*:\s*`)
	multiSpaceRe  = regexp.MustCompile(`\s+`)
	nonSlugCharRe = regexp.MustCompile(`[^a-z0-9-]+`)
)

func stripSkillPrefix(s string) string {
	return strings.TrimSpace(skillPrefixRe.ReplaceAllString(strings.TrimSpace(s), ""))
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
	Handle skills.PlanHandle

	TitleLower         string
	TitleNoPrefixLower string
	TitleSlugLower     string
	TitleNoPrefixSlug  string

	NameLower     string
	NameSlugLower string
	SlugLower     string

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
	return filepath.IsAbs(q) ||
		strings.Contains(q, "/") ||
		strings.Contains(q, "\\") ||
		strings.HasSuffix(lower, ".md") ||
		strings.HasSuffix(lower, ".yaml") ||
		strings.HasSuffix(lower, ".yml")
}

func buildSkillCandidates(ws *workspace.Workspace, handles []skills.PlanHandle, queryRaw string) []skillCandidate {
	queryRaw = strings.TrimSpace(queryRaw)
	queryLower := strings.ToLower(queryRaw)
	queryNoPrefixLower := strings.ToLower(stripSkillPrefix(queryRaw))
	querySlugLower := slugifyLower(queryRaw)
	querySlugNoPrefixLower := slugifyLower(stripSkillPrefix(queryRaw))

	var queryPathNorm paths.NormalizedPath
	if looksLikePathQuery(queryRaw) {
		queryPathNorm = ws.Resolver().Normalize(queryRaw)
	}

	var candidates []skillCandidate
	for _, handle := range handles {
		if handle.Plan == nil {
			continue
		}
		title := strings.TrimSpace(handle.Plan.DisplayTitle())
		titleLower := strings.ToLower(title)
		titleNoPrefix := stripSkillPrefix(title)
		titleNoPrefixLower := strings.ToLower(titleNoPrefix)
		name := strings.TrimSpace(handle.Plan.Skill.Name)
		nameLower := strings.ToLower(name)
		nameSlugLower := slugifyLower(name)
		slug := strings.ToLower(skillSlugFromPath(handle.Path))

		pathNorm := ws.Resolver().Normalize(handle.Path)

		cand := skillCandidate{
			Handle:             handle,
			TitleLower:         titleLower,
			TitleNoPrefixLower: titleNoPrefixLower,
			TitleSlugLower:     slugifyLower(title),
			TitleNoPrefixSlug:  slugifyLower(titleNoPrefix),
			NameLower:          nameLower,
			NameSlugLower:      nameSlugLower,
			SlugLower:          slug,
			PathNorm:           pathNorm,
		}

		if looksLikePathQuery(queryRaw) && !queryPathNorm.Empty() {
			if queryPathNorm.Exists {
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
				if paths.MatchPaths(queryPathNorm, cand.PathNorm) {
					cand.Score = 900
					cand.Why = "path-fuzzy"
				}
			}
		}

		if cand.Score == 0 && nameLower != "" && nameLower == queryLower {
			cand.Score = 880
			cand.Why = "name-exact"
		}
		if cand.Score == 0 && titleLower == queryLower {
			cand.Score = 860
			cand.Why = "title-exact"
		}
		if cand.Score == 0 && titleNoPrefixLower != "" && titleNoPrefixLower == queryNoPrefixLower {
			cand.Score = 840
			cand.Why = "title-no-prefix-exact"
		}
		if cand.Score == 0 && (slug == queryLower || slug == querySlugLower || slug == querySlugNoPrefixLower) {
			cand.Score = 830
			cand.Why = "slug-exact"
		}
		if cand.Score == 0 && (nameSlugLower == querySlugLower || nameSlugLower == querySlugNoPrefixLower) {
			cand.Score = 820
			cand.Why = "name-slug-exact"
		}
		if cand.Score == 0 && (cand.TitleSlugLower == querySlugLower || cand.TitleNoPrefixSlug == querySlugLower || cand.TitleNoPrefixSlug == querySlugNoPrefixLower) {
			cand.Score = 810
			cand.Why = "title-slug-exact"
		}

		if cand.Score == 0 && queryLower != "" && strings.Contains(titleLower, queryLower) {
			cand.Score = 600
			cand.Why = "title-contains"
		}
		if cand.Score == 0 && queryNoPrefixLower != "" && titleNoPrefixLower != "" && strings.Contains(titleNoPrefixLower, queryNoPrefixLower) {
			cand.Score = 590
			cand.Why = "title-no-prefix-contains"
		}
		if cand.Score == 0 && queryLower != "" && strings.Contains(nameLower, queryLower) {
			cand.Score = 580
			cand.Why = "name-contains"
		}
		if cand.Score == 0 && querySlugLower != "" && strings.Contains(cand.TitleSlugLower, querySlugLower) {
			cand.Score = 570
			cand.Why = "slug-contains"
		}

		if cand.Score > 0 {
			candidates = append(candidates, cand)
		}
	}

	return candidates
}

func sortSkillCandidates(candidates []skillCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		ti := candidates[i].Handle.Plan.DisplayTitle()
		tj := candidates[j].Handle.Plan.DisplayTitle()
		if ti != tj {
			return ti < tj
		}
		return candidates[i].Handle.DisplayPath < candidates[j].Handle.DisplayPath
	})
}
