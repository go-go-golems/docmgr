package paths

import (
	"os"
	"path/filepath"
	"strings"
)

// Anchor represents the base a path was resolved against.
type Anchor string

const (
	AnchorUnknown    Anchor = ""
	AnchorRepo       Anchor = "repo"
	AnchorDoc        Anchor = "doc"
	AnchorConfig     Anchor = "config"
	AnchorDocsRoot   Anchor = "docs-root"
	AnchorDocsParent Anchor = "docs-parent"
)

// ResolverOptions configures a Resolver instance.
type ResolverOptions struct {
	DocsRoot  string
	DocPath   string
	ConfigDir string
	RepoRoot  string
}

// Resolver normalizes user-provided paths into canonical representations.
type Resolver struct {
	docDir     string
	docsRoot   string
	configDir  string
	docsParent string
	repoRoot   string
}

// NormalizedPath contains all known representations of a path.
type NormalizedPath struct {
	Original      string
	OriginalClean string
	Canonical     string
	Abs           string
	RepoRelative  string
	DocsRelative  string
	DocRelative   string
	Anchor        Anchor
	Exists        bool
}

// NewResolver builds a Resolver with best-effort absolute anchors.
func NewResolver(opts ResolverOptions) *Resolver {
	docDir := absDir(filepath.Dir(opts.DocPath))
	docsRoot := absDir(opts.DocsRoot)
	configDir := absDir(opts.ConfigDir)
	var docsParent string
	if docsRoot != "" {
		docsParent = absDir(filepath.Dir(docsRoot))
	}

	repoRoot := absDir(opts.RepoRoot)
	if repoRoot == "" {
		repoRoot = findRepositoryRoot(docDir, docsRoot, configDir)
	}

	return &Resolver{
		docDir:     docDir,
		docsRoot:   docsRoot,
		configDir:  configDir,
		docsParent: docsParent,
		repoRoot:   repoRoot,
	}
}

// Normalize resolves a raw path string into a NormalizedPath.
func (r *Resolver) Normalize(raw string) NormalizedPath {
	result := NormalizedPath{
		Original:      raw,
		OriginalClean: toSlash(strings.TrimSpace(raw)),
	}
	cleaned := cleanInput(raw)
	if cleaned == "" {
		return result
	}

	if filepath.IsAbs(cleaned) {
		return r.buildResult(cleaned, AnchorUnknown, true)
	}

	bases := []struct {
		path   string
		anchor Anchor
	}{
		{path: r.repoRoot, anchor: AnchorRepo},
		{path: r.docDir, anchor: AnchorDoc},
		{path: r.configDir, anchor: AnchorConfig},
		{path: r.docsRoot, anchor: AnchorDocsRoot},
		{path: r.docsParent, anchor: AnchorDocsParent},
	}

	var fallback NormalizedPath
	for _, base := range bases {
		if base.path == "" {
			continue
		}
		absPath := filepath.Clean(filepath.Join(base.path, cleaned))
		if base.anchor == AnchorRepo && r.repoRoot != "" {
			if rel, err := filepath.Rel(r.repoRoot, absPath); err != nil || strings.HasPrefix(rel, "..") {
				continue
			}
		}
		normalized := r.buildResult(absPath, base.anchor, false)
		if normalized.Abs == "" {
			continue
		}
		if fallback.Canonical == "" {
			fallback = normalized
		}
		if normalized.Exists {
			return normalized
		}
	}

	if fallback.Canonical != "" {
		return fallback
	}

	// Could not resolve against any base; fall back to cleaned relative path.
	result.Canonical = toSlash(cleaned)
	return result
}

// Representations returns the set of comparable strings for the path.
func (n NormalizedPath) Representations() []string {
	values := []string{
		n.Canonical,
		n.RepoRelative,
		n.DocsRelative,
		n.DocRelative,
		n.Abs,
		n.OriginalClean,
	}
	return uniqueStrings(values...)
}

// Suffixes returns the trailing path segments (up to maxSegments) for fuzzy matches.
func (n NormalizedPath) Suffixes(maxSegments int) []string {
	source := firstNonEmpty(n.Canonical, n.RepoRelative, n.Abs, n.OriginalClean)
	if source == "" {
		return nil
	}
	segments := strings.Split(source, "/")
	if len(segments) == 0 {
		return nil
	}
	if maxSegments > len(segments) {
		maxSegments = len(segments)
	}
	var out []string
	for i := 1; i <= maxSegments; i++ {
		start := len(segments) - i
		suffix := strings.Join(segments[start:], "/")
		out = append(out, suffix)
	}
	return uniqueStrings(out...)
}

// Empty reports whether the normalized path has no usable representations.
func (n NormalizedPath) Empty() bool {
	return strings.TrimSpace(n.Canonical) == "" &&
		strings.TrimSpace(n.Abs) == "" &&
		strings.TrimSpace(n.OriginalClean) == ""
}

// Best returns the highest-priority representation for display purposes.
func (n NormalizedPath) Best() string {
	return firstNonEmpty(n.Canonical, n.RepoRelative, n.DocsRelative, n.DocRelative, n.Abs, n.OriginalClean)
}

// MatchPaths performs fuzzy matching between two normalized paths.
func MatchPaths(a, b NormalizedPath) bool {
	if a.Empty() || b.Empty() {
		return false
	}
	aSet := toSet(a.Representations())
	bSet := toSet(b.Representations())
	if intersects(aSet, bSet) {
		return true
	}
	aSuffix := toSet(a.Suffixes(3))
	bSuffix := toSet(b.Suffixes(3))
	if intersects(aSuffix, bSuffix) {
		return true
	}
	return containsSubstring(aSet, bSet)
}

// DirectoryMatch reports whether target is inside dir.
func DirectoryMatch(dir NormalizedPath, target NormalizedPath) bool {
	if dir.Empty() || target.Empty() {
		return false
	}
	dirCandidates := dir.Representations()
	targetValues := target.Representations()
	for _, d := range dirCandidates {
		dNorm := normalizeForCompare(d)
		if dNorm == "" {
			continue
		}
		for _, t := range targetValues {
			tNorm := normalizeForCompare(t)
			if tNorm == "" {
				continue
			}
			if hasPathPrefix(tNorm, dNorm) {
				return true
			}
		}
	}
	return false
}

func (r *Resolver) buildResult(absPath string, anchor Anchor, forceExists bool) NormalizedPath {
	if absPath == "" {
		return NormalizedPath{}
	}
	exists := forceExists
	if !forceExists {
		exists = pathExists(absPath)
	}

	absSlash := toSlash(absPath)
	repoRel := relativeWithin(absPath, r.repoRoot)
	docRel := relativeWithin(absPath, r.docDir)
	if docRel == "" {
		docRel = relativeAllowParents(absPath, r.docDir)
	}
	docsRel := relativeWithin(absPath, r.docsRoot)

	canonical := firstNonEmpty(repoRel, docsRel, docRel, absSlash)

	return NormalizedPath{
		Original:      absSlash,
		OriginalClean: absSlash,
		Canonical:     canonical,
		Abs:           absSlash,
		RepoRelative:  repoRel,
		DocsRelative:  docsRel,
		DocRelative:   docRel,
		Anchor:        anchor,
		Exists:        exists,
	}
}

func absDir(path string) string {
	if path == "" {
		return ""
	}
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}
	return filepath.Clean(path)
}

func cleanInput(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	sep := string(filepath.Separator)
	s := strings.ReplaceAll(raw, "\\", sep)
	s = strings.ReplaceAll(s, "/", sep)
	s = expandHome(s)
	s = filepath.Clean(s)
	return s
}

func expandHome(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return path
	}
	prefixes := []string{"~/", "~\\"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			if home, err := os.UserHomeDir(); err == nil {
				return filepath.Join(home, path[len(prefix):])
			}
			break
		}
	}
	return path
}

func findRepositoryRoot(docDir, docsRoot, configDir string) string {
	candidates := []string{docDir, docsRoot, configDir}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if root := walkForRepo(candidate); root != "" {
			return root
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		return walkForRepo(cwd)
	}
	return ""
}

func walkForRepo(start string) string {
	dir := absDir(start)
	for dir != "" {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func relativeWithin(target, base string) string {
	if target == "" || base == "" {
		return ""
	}
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return ""
	}
	if strings.HasPrefix(rel, "..") {
		return ""
	}
	return toSlash(rel)
}

func pathExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func relativeAllowParents(target, base string) string {
	if target == "" || base == "" {
		return ""
	}
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return ""
	}
	return toSlash(rel)
}

func toSlash(path string) string {
	return filepath.ToSlash(strings.TrimSpace(path))
}

func uniqueStrings(values ...string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, v := range values {
		if v = strings.TrimSpace(v); v == "" {
			continue
		}
		key := toSlash(v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return toSlash(v)
		}
	}
	return ""
}

func toSet(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(values))
	for _, v := range values {
		n := normalizeForCompare(v)
		if n == "" {
			continue
		}
		set[n] = struct{}{}
	}
	return set
}

func normalizeForCompare(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.ToLower(filepath.ToSlash(value))
	return value
}

func intersects(a, b map[string]struct{}) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	if len(a) > len(b) {
		a, b = b, a
	}
	for k := range a {
		if _, ok := b[k]; ok {
			return true
		}
	}
	return false
}

func containsSubstring(a, b map[string]struct{}) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	for ka := range a {
		for kb := range b {
			if strings.Contains(ka, kb) || strings.Contains(kb, ka) {
				return true
			}
		}
	}
	return false
}

func hasPathPrefix(value, prefix string) bool {
	if value == "" || prefix == "" {
		return false
	}
	value = strings.Trim(value, "/")
	prefix = strings.Trim(prefix, "/")
	return value == prefix || strings.HasPrefix(value, prefix+"/")
}
