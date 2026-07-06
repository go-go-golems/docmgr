package paths

import (
	"path"
	"path/filepath"
	"strings"
)

// Scheme identifies an explicit path anchor persisted in frontmatter.
//
// Anchored paths make the meaning of a RelatedFiles entry explicit instead of
// leaving readers to guess which base directory a bare string was relative to
// (design doc DOCMGR-200 §8.1, decision D1):
//
//	repo://pkg/foo.go                 relative to the repository root
//	ws://glazed/pkg/fields.go         relative to a go.work workspace member
//	docs://2026/07/05/T/design/01.md  relative to the docs root (ttmp)
//	doc://../reference/01-diary.md    relative to the referencing doc's dir
//	abs:///home/user/x.go             absolute path (escape hatch)
//
// Bare strings (no scheme) are "legacy" and keep resolving through the
// historical multi-anchor guessing logic in Resolver.Normalize.
type Scheme string

const (
	SchemeRepo Scheme = "repo"
	SchemeWs   Scheme = "ws"
	SchemeDocs Scheme = "docs"
	SchemeDoc  Scheme = "doc"
	SchemeAbs  Scheme = "abs"
	// SchemeLegacy marks a bare string without an explicit anchor.
	SchemeLegacy Scheme = ""
)

// AnchoredPath is a parsed anchored path string.
type AnchoredPath struct {
	Scheme Scheme
	// Member is the go.work workspace member directory name (ws:// only).
	Member string
	// Rel is the slash-separated path relative to the anchor base.
	// For SchemeAbs it is the absolute path itself.
	Rel string
}

// IsAnchored reports whether raw carries one of the known anchor schemes.
func IsAnchored(raw string) bool {
	_, ok := ParseAnchored(raw)
	return ok
}

// ParseAnchored parses an anchored path string. ok=false means the string is
// a legacy bare path (or uses an unknown scheme) and must be resolved through
// the legacy normalization logic.
func ParseAnchored(raw string) (AnchoredPath, bool) {
	s := strings.TrimSpace(raw)
	i := strings.Index(s, "://")
	if i <= 0 {
		return AnchoredPath{}, false
	}
	scheme := Scheme(strings.ToLower(s[:i]))
	rest := s[i+len("://"):]

	switch scheme {
	case SchemeLegacy:
		return AnchoredPath{}, false
	case SchemeRepo, SchemeDocs, SchemeDoc:
		return AnchoredPath{Scheme: scheme, Rel: cleanAnchoredRel(rest)}, true
	case SchemeWs:
		rest = strings.TrimLeft(rest, "/")
		member, rel, found := strings.Cut(rest, "/")
		member = strings.TrimSpace(member)
		if member == "" {
			return AnchoredPath{}, false
		}
		if !found {
			rel = ""
		}
		return AnchoredPath{Scheme: SchemeWs, Member: member, Rel: cleanAnchoredRel(rel)}, true
	case SchemeAbs:
		p := strings.TrimSpace(rest)
		if p == "" {
			return AnchoredPath{}, false
		}
		if !strings.HasPrefix(p, "/") && !filepath.IsAbs(filepath.FromSlash(p)) {
			// abs:// requires an absolute payload; tolerate a missing leading
			// slash on POSIX by re-adding it.
			p = "/" + p
		}
		return AnchoredPath{Scheme: SchemeAbs, Rel: path.Clean(filepath.ToSlash(p))}, true
	default:
		// Unknown scheme (http://, https://, ...): treat as legacy string.
		return AnchoredPath{}, false
	}
}

// String renders the anchored path back into its canonical string form.
func (a AnchoredPath) String() string {
	switch a.Scheme {
	case SchemeRepo, SchemeDocs, SchemeDoc:
		return string(a.Scheme) + "://" + a.Rel
	case SchemeWs:
		if a.Rel == "" {
			return "ws://" + a.Member
		}
		return "ws://" + a.Member + "/" + a.Rel
	case SchemeAbs:
		return "abs://" + a.Rel
	case SchemeLegacy:
		return a.Rel
	default:
		return a.Rel
	}
}

func cleanAnchoredRel(rel string) string {
	rel = strings.TrimSpace(rel)
	rel = filepath.ToSlash(rel)
	rel = strings.TrimLeft(rel, "/")
	if rel == "" {
		return ""
	}
	cleaned := path.Clean(rel)
	if cleaned == "." {
		return ""
	}
	return cleaned
}

// AnchoredFor chooses the tightest containing anchor for an absolute path
// (write-side rule, design doc §8.1):
//
//  1. inside the repository        -> repo://<rel>
//  2. inside the go.work workspace -> ws://<member>/<rel>
//  3. inside the docs root         -> docs://<rel>
//  4. anywhere else                -> abs://<abs>
//
// It never emits doc:// or repo-escaping ../ chains.
func (r *Resolver) AnchoredFor(absPath string) AnchoredPath {
	absPath = filepath.Clean(strings.TrimSpace(absPath))
	if absPath == "" || !filepath.IsAbs(absPath) {
		return AnchoredPath{Scheme: SchemeLegacy, Rel: filepath.ToSlash(absPath)}
	}
	if rel := relativeWithin(absPath, r.repoRoot); rel != "" && rel != "." {
		return AnchoredPath{Scheme: SchemeRepo, Rel: rel}
	}
	if r.wsRoot != "" {
		if rel := relativeWithin(absPath, r.wsRoot); rel != "" && rel != "." {
			member, rest, _ := strings.Cut(rel, "/")
			if member != "" {
				return AnchoredPath{Scheme: SchemeWs, Member: member, Rel: rest}
			}
		}
	}
	if rel := relativeWithin(absPath, r.docsRoot); rel != "" && rel != "." {
		return AnchoredPath{Scheme: SchemeDocs, Rel: rel}
	}
	return AnchoredPath{Scheme: SchemeAbs, Rel: filepath.ToSlash(absPath)}
}
