package workspace

import (
	"strings"

	"github.com/go-go-golems/docmgr/internal/paths"
)

// RelatedFileNormalized contains the normalized keys we persist for a single RelatedFiles entry.
//
// These keys intentionally include multiple representations so later query logic can implement
// best-effort matching in SQL without having to reconstruct all anchors at query-time.
//
// Spec: §7.3 / §12.1 (normalization + fallback matching strategy).
type RelatedFileNormalized struct {
	Canonical    string
	RepoRelative string
	DocsRelative string
	DocRelative  string
	Abs          string
	Clean        string
	Anchor       string
}

func normalizeRelatedFile(resolver *paths.Resolver, raw string) RelatedFileNormalized {
	raw = strings.TrimSpace(raw)
	if raw == "" || resolver == nil {
		return RelatedFileNormalized{}
	}

	n := resolver.Normalize(raw)
	return RelatedFileNormalized{
		Canonical:    strings.TrimSpace(n.Canonical),
		RepoRelative: strings.TrimSpace(n.RepoRelative),
		DocsRelative: strings.TrimSpace(n.DocsRelative),
		DocRelative:  strings.TrimSpace(n.DocRelative),
		Abs:          strings.TrimSpace(n.Abs),
		Clean:        strings.TrimSpace(normalizeCleanPath(raw)),
		Anchor:       strings.TrimSpace(string(n.Anchor)),
	}
}
