package workspace

import (
	"strings"

	"github.com/go-go-golems/docmgr/internal/paths"
)

// RelatedFileNormalized contains the normalized keys we persist for a single RelatedFiles entry.
//
// Paths v2 (design doc DOCMGR-200 §8.1 / Phase 2): the index stores the output of the ONE
// resolver — the resolved absolute path plus the anchor that produced it — instead of six
// divergent representations. Reverse lookups match on the absolute path (exact) or on
// case-sensitive whole-segment suffixes of it.
type RelatedFileNormalized struct {
	// Abs is the resolved absolute path (slash-separated). Empty when the entry
	// could not be resolved against any anchor.
	Abs string
	// RepoRelative is the repo-relative form when the resolved path is inside the
	// repository (kept for display).
	RepoRelative string
	// Anchor is the anchor that produced Abs (repo/ws/docs/doc/abs/config/...).
	// Empty for absolute legacy inputs.
	Anchor string
}

func normalizeRelatedFile(resolver *paths.Resolver, raw string) RelatedFileNormalized {
	raw = strings.TrimSpace(raw)
	if raw == "" || resolver == nil {
		return RelatedFileNormalized{}
	}

	// Resolve (not ResolveNoFS): legacy bare strings pick their anchor by existence,
	// which is exactly what doctor and relate do — write, index, doctor and resolve
	// must tell one consistent story.
	n := resolver.Resolve(raw)
	return RelatedFileNormalized{
		Abs:          strings.TrimSpace(n.Abs),
		RepoRelative: strings.TrimSpace(n.RepoRelative),
		Anchor:       strings.TrimSpace(string(n.Anchor)),
	}
}
