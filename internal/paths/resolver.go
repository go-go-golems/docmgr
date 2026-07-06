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
	// AnchorWs marks a path resolved against a go.work workspace member (ws://).
	AnchorWs Anchor = "ws"
	// AnchorAbs marks an explicitly absolute path (abs://).
	AnchorAbs Anchor = "abs"
)

// ResolverOptions configures a Resolver instance.
type ResolverOptions struct {
	DocsRoot  string
	DocPath   string
	ConfigDir string
	RepoRoot  string
	// WorkspaceRoot is the directory containing go.work (for ws:// anchors).
	// When empty it is auto-detected by walking up from the repo root.
	WorkspaceRoot string
}

// Resolver normalizes user-provided paths into canonical representations.
type Resolver struct {
	docDir     string
	docsRoot   string
	configDir  string
	docsParent string
	repoRoot   string
	wsRoot     string
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

	wsRoot := absDir(opts.WorkspaceRoot)
	if wsRoot == "" {
		wsRoot = findWorkspaceRoot(repoRoot, docDir, docsRoot, configDir)
	}

	return &Resolver{
		docDir:     docDir,
		docsRoot:   docsRoot,
		configDir:  configDir,
		docsParent: docsParent,
		repoRoot:   repoRoot,
		wsRoot:     wsRoot,
	}
}

// Resolve is the single entry point for turning any persisted path string —
// anchored (repo://, ws://, docs://, doc://, abs://) or legacy bare string —
// into all known representations, with an honest os.Stat-based Exists for
// every anchor (design doc DOCMGR-200 §8.1).
func (r *Resolver) Resolve(raw string) NormalizedPath {
	return r.Normalize(raw)
}

// ResolveNoFS is Resolve without filesystem access (Exists is always false).
// For anchored inputs the anchor choice is identical to Resolve because the
// anchor is explicit; only legacy bare strings can differ (first-valid-base
// instead of first-existing-base).
func (r *Resolver) ResolveNoFS(raw string) NormalizedPath {
	return r.NormalizeNoFS(raw)
}

// resolveAnchored resolves an explicitly anchored path. Anchors are explicit,
// so no containment checks apply: doc:// MAY escape the repository.
func (r *Resolver) resolveAnchored(a AnchoredPath) NormalizedPath {
	canonical, anchor, absPath := r.anchoredTarget(a)
	if absPath == "" {
		return unresolvedAnchoredPath(canonical, anchor)
	}

	n := r.buildResult(absPath, anchor)
	// Anchored strings are canonical by construction and must round-trip.
	n.Original = canonical
	n.OriginalClean = canonical
	n.Canonical = canonical
	return n
}

// resolveAnchoredNoFS resolves an explicitly anchored path without filesystem
// access. Keep this separate from resolveAnchored instead of using a boolean
// flag so tainted search-query inputs cannot appear to flow into os.Stat.
func (r *Resolver) resolveAnchoredNoFS(a AnchoredPath) NormalizedPath {
	canonical, anchor, absPath := r.anchoredTarget(a)
	if absPath == "" {
		return unresolvedAnchoredPath(canonical, anchor)
	}

	n := r.buildResultWithExists(absPath, anchor, false)
	// Anchored strings are canonical by construction and must round-trip.
	n.Original = canonical
	n.OriginalClean = canonical
	n.Canonical = canonical
	return n
}

func (r *Resolver) anchoredTarget(a AnchoredPath) (string, Anchor, string) {
	var base string
	var anchor Anchor
	var absPath string

	switch a.Scheme {
	case SchemeLegacy:
		// Legacy strings never reach resolveAnchored; treated as unknown.
	case SchemeRepo:
		base, anchor = r.repoRoot, AnchorRepo
	case SchemeWs:
		anchor = AnchorWs
		wsBase := r.wsRoot
		if wsBase == "" {
			// Design fallback: without a go.work the workspace root is the repo root.
			wsBase = r.repoRoot
		}
		if wsBase != "" {
			base = filepath.Join(wsBase, a.Member)
		}
	case SchemeDocs:
		base, anchor = r.docsRoot, AnchorDocsRoot
	case SchemeDoc:
		base, anchor = r.docDir, AnchorDoc
	case SchemeAbs:
		anchor = AnchorAbs
		absPath = filepath.Clean(filepath.FromSlash(a.Rel))
	}

	if absPath == "" && base != "" {
		absPath = filepath.Clean(filepath.Join(base, filepath.FromSlash(a.Rel)))
	}

	return a.String(), anchor, absPath
}

func unresolvedAnchoredPath(canonical string, anchor Anchor) NormalizedPath {
	// Anchor base unknown (e.g. docs:// without a docs root).
	return NormalizedPath{
		Original:      canonical,
		OriginalClean: canonical,
		Canonical:     canonical,
		Anchor:        anchor,
	}
}

// Normalize resolves a raw path string into a NormalizedPath.
//
// Anchored inputs (repo://, ws://, docs://, doc://, abs://) resolve directly
// against their explicit base; bare strings keep the legacy multi-anchor
// guessing behavior.
func (r *Resolver) Normalize(raw string) NormalizedPath {
	if a, ok := ParseAnchored(raw); ok {
		return r.resolveAnchored(a)
	}
	result := NormalizedPath{
		Original:      raw,
		OriginalClean: toSlash(strings.TrimSpace(raw)),
	}
	cleaned := cleanInput(raw)
	if cleaned == "" {
		return result
	}

	if filepath.IsAbs(cleaned) {
		return r.buildResult(cleaned, AnchorUnknown)
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
		// Ensure the candidate stays within its anchor base (prevents traversal).
		// Doc-relative paths are allowed to traverse parents, but should still not
		// escape the repository if we know the repo root.
		if base.anchor == AnchorDoc {
			if r.repoRoot != "" {
				if rel, err := filepath.Rel(r.repoRoot, absPath); err != nil || strings.HasPrefix(rel, "..") {
					continue
				}
			}
		} else {
			if rel, err := filepath.Rel(base.path, absPath); err != nil || strings.HasPrefix(rel, "..") {
				continue
			}
		}
		normalized := r.buildResult(absPath, base.anchor)
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

// NormalizeNoFS is like Normalize, but does not touch the filesystem (no Stat).
// This is useful for normalizing user-provided strings for matching and diagnostics
// without turning them into filesystem reads.
func (r *Resolver) NormalizeNoFS(raw string) NormalizedPath {
	if a, ok := ParseAnchored(raw); ok {
		return r.resolveAnchoredNoFS(a)
	}
	result := NormalizedPath{
		Original:      raw,
		OriginalClean: toSlash(strings.TrimSpace(raw)),
	}
	cleaned := cleanInput(raw)
	if cleaned == "" {
		return result
	}

	if filepath.IsAbs(cleaned) {
		return r.buildResultWithExists(cleaned, AnchorUnknown, false)
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

	for _, base := range bases {
		if base.path == "" {
			continue
		}
		absPath := filepath.Clean(filepath.Join(base.path, cleaned))
		// Ensure the candidate stays within its anchor base (prevents traversal).
		// Doc-relative paths are allowed to traverse parents, but should still not
		// escape the repository if we know the repo root.
		if base.anchor == AnchorDoc {
			if r.repoRoot != "" {
				if rel, err := filepath.Rel(r.repoRoot, absPath); err != nil || strings.HasPrefix(rel, "..") {
					continue
				}
			}
		} else {
			if rel, err := filepath.Rel(base.path, absPath); err != nil || strings.HasPrefix(rel, "..") {
				continue
			}
		}
		normalized := r.buildResultWithExists(absPath, base.anchor, false)
		if normalized.Abs == "" {
			continue
		}
		return normalized
	}

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

// MatchPaths reports whether two normalized paths refer to the same file.
//
// Matching is intentionally strict (design doc DOCMGR-200 §8.1 "fuzzy layer
// diet"): first an exact match on the resolved absolute path, then
// case-SENSITIVE suffix matching on whole path segments (one side's full
// path must equal a trailing run of the other side's segments). Substring
// containment is gone: "api.go" no longer matches "chatapi.go".
func MatchPaths(a, b NormalizedPath) bool {
	if a.Empty() || b.Empty() {
		return false
	}
	if a.Abs != "" && a.Abs == b.Abs {
		return true
	}
	for _, x := range a.matchKeys() {
		for _, y := range b.matchKeys() {
			if segmentSuffixMatch(x, y) {
				return true
			}
		}
	}
	return false
}

// matchKeys returns the comparable full-path strings for strict matching.
func (n NormalizedPath) matchKeys() []string {
	// These keys are used only for string comparison in MatchPaths. Avoid
	// filepath normalization here so security scanners do not mistake the
	// comparison-only path keys for filesystem access.
	return uniqueCompareStrings(
		n.Abs,
		n.RepoRelative,
		n.DocsRelative,
		n.DocRelative,
		stripAnchorScheme(n.Canonical),
		n.OriginalClean,
	)
}

func stripAnchorScheme(value string) string {
	if a, ok := ParseAnchored(value); ok {
		switch a.Scheme {
		case SchemeWs:
			if a.Rel == "" {
				return a.Member
			}
			return a.Member + "/" + a.Rel
		case SchemeRepo, SchemeDocs, SchemeDoc, SchemeAbs, SchemeLegacy:
			return a.Rel
		default:
			return a.Rel
		}
	}
	return value
}

// segmentSuffixMatch reports whether one path equals a whole-segment suffix
// of the other (case-sensitive).
func segmentSuffixMatch(x, y string) bool {
	xs := splitSegments(x)
	ys := splitSegments(y)
	if len(xs) == 0 || len(ys) == 0 {
		return false
	}
	if len(xs) > len(ys) {
		xs, ys = ys, xs
	}
	off := len(ys) - len(xs)
	for i := range xs {
		if xs[i] != ys[off+i] {
			return false
		}
	}
	return true
}

func splitSegments(value string) []string {
	value = strings.Trim(strings.TrimSpace(filepath.ToSlash(value)), "/")
	if value == "" {
		return nil
	}
	return strings.Split(value, "/")
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

func (r *Resolver) buildResult(absPath string, anchor Anchor) NormalizedPath {
	if absPath == "" {
		return NormalizedPath{}
	}
	exists := pathExists(absPath)
	return r.buildResultWithExists(absPath, anchor, exists)
}

func (r *Resolver) buildResultWithExists(absPath string, anchor Anchor, exists bool) NormalizedPath {
	if absPath == "" {
		return NormalizedPath{}
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

func findWorkspaceRoot(candidates ...string) string {
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if root := FindWorkspaceRootFrom(candidate); root != "" {
			return root
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		return FindWorkspaceRootFrom(cwd)
	}
	return ""
}

// FindWorkspaceRootFrom walks up from start looking for a go.work file and
// returns the directory containing it ("" when none is found). This is the
// base directory for ws://<member>/<rel> anchors.
func FindWorkspaceRootFrom(start string) string {
	dir := absDir(start)
	for dir != "" {
		if fi, err := os.Stat(filepath.Join(dir, "go.work")); err == nil && !fi.IsDir() {
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

// FindRepoRootFrom walks up from start for the nearest directory containing
// .git or go.mod (the shared repo-root heuristic; see also FindGitRootFrom).
func FindRepoRootFrom(start string) string {
	return walkForRepo(start)
}

// FindGitRootFrom walks up from start looking for a .git directory or a valid
// .git gitfile (worktrees/submodules) and returns the containing directory.
func FindGitRootFrom(start string) string {
	dir := absDir(start)
	for dir != "" {
		gitPath := filepath.Join(dir, ".git")
		if fi, err := os.Stat(gitPath); err == nil {
			if fi.IsDir() {
				return dir
			}
			// .git is a file; validate the gitdir: pointer.
			if b, err := os.ReadFile(gitPath); err == nil {
				line := strings.TrimSpace(string(b))
				if strings.HasPrefix(strings.ToLower(line), "gitdir:") {
					gd := strings.TrimSpace(line[len("gitdir:"):])
					if !filepath.IsAbs(gd) {
						gd = filepath.Join(dir, gd)
					}
					if _, err := os.Stat(gd); err == nil {
						return dir
					}
				}
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
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
	// Resolver.Stat is intentionally limited to the filesystem-aware Resolve path.
	// Search and other lookup-only inputs use ResolveNoFS/NormalizeNoFS, which do
	// not call pathExists; filesystem-aware callers build this path from configured
	// workspace anchors after lexical containment checks.
	// codeql[go/path-injection]
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
	return uniqueWithNormalizer(toSlash, values...)
}

func uniqueCompareStrings(values ...string) []string {
	return uniqueWithNormalizer(slashForCompare, values...)
}

func uniqueWithNormalizer(normalize func(string) string, values ...string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, v := range values {
		if v = strings.TrimSpace(v); v == "" {
			continue
		}
		key := normalize(v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

func slashForCompare(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return toSlash(v)
		}
	}
	return ""
}

func normalizeForCompare(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.ToLower(filepath.ToSlash(value))
	return value
}

func hasPathPrefix(value, prefix string) bool {
	if value == "" || prefix == "" {
		return false
	}
	value = strings.Trim(value, "/")
	prefix = strings.Trim(prefix, "/")
	return value == prefix || strings.HasPrefix(value, prefix+"/")
}
