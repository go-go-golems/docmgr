package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/docmgr/internal/workspace"
)

// resolveDocRef resolves a user-supplied --doc reference to an absolute file path.
//
// Matching precedence (deterministic; first hit wins):
//  1. absolute path
//  2. path relative to the current working directory
//  3. path relative to the repository root
//  4. path relative to the docs root
//  5. same as (4) with a duplicated docs-root prefix stripped (kills the
//     observed "ttmp/ttmp/..." double-join)
//  6. unique suffix match against the indexed doc paths (last resort)
//
// ws may be nil; in that case a workspace is discovered lazily only if the
// suffix-match fallback is needed. On ambiguity the error lists candidates.
func resolveDocRef(ctx context.Context, ws *workspace.Workspace, rootOverride string, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("doc path is required")
	}

	if filepath.IsAbs(raw) {
		p := filepath.Clean(raw)
		if fileExists(p) {
			return p, nil
		}
		return "", fmt.Errorf("doc not found: %s", p)
	}

	rel := filepath.FromSlash(raw)

	docsRoot := ""
	repoRoot := ""
	if ws != nil {
		docsRoot = ws.Context().Root
		repoRoot = ws.Context().RepoRoot
	} else {
		docsRoot = workspace.ResolveRoot(rootOverride)
		if rr, err := workspace.FindRepositoryRoot(); err == nil {
			repoRoot = rr
		}
	}

	var candidates []string
	// (2) cwd-relative
	if abs, err := filepath.Abs(rel); err == nil {
		candidates = append(candidates, abs)
	}
	// (3) repo-relative
	if repoRoot != "" {
		candidates = append(candidates, filepath.Join(repoRoot, rel))
	}
	// (4) docs-root-relative
	if docsRoot != "" {
		candidates = append(candidates, filepath.Join(docsRoot, rel))
		// (5) docs-root-relative with duplicated docs-root basename prefixes
		// stripped (e.g. --doc ttmp/2026/... or the observed ttmp/ttmp/...
		// double-join, resolved against a docs root already named ttmp).
		rootBase := filepath.Base(filepath.Clean(docsRoot))
		slashRel := filepath.ToSlash(rel)
		for rootBase != "" && strings.HasPrefix(slashRel, rootBase+"/") {
			slashRel = strings.TrimPrefix(slashRel, rootBase+"/")
			candidates = append(candidates, filepath.Join(docsRoot, filepath.FromSlash(slashRel)))
		}
	}

	seen := map[string]struct{}{}
	for _, cand := range candidates {
		cand = filepath.Clean(cand)
		if _, ok := seen[cand]; ok {
			continue
		}
		seen[cand] = struct{}{}
		if fileExists(cand) {
			return cand, nil
		}
	}

	// (6) unique suffix match against the docs index.
	matches, err := suffixMatchIndexedDocs(ctx, ws, rootOverride, raw)
	if err == nil {
		switch len(matches) {
		case 1:
			return filepath.FromSlash(matches[0]), nil
		case 0:
			// fall through to the not-found error
		default:
			sort.Strings(matches)
			return "", fmt.Errorf("doc reference %q is ambiguous; candidates: %s", raw, strings.Join(matches, ", "))
		}
	}

	return "", fmt.Errorf("doc not found: %q (tried absolute, cwd-relative, repo-relative, and docs-root-relative paths; run 'docmgr doc list --ticket <ticket>' to see doc paths)", raw)
}

// suffixMatchIndexedDocs returns indexed doc paths ending in the given
// (slash-normalized) relative path.
func suffixMatchIndexedDocs(ctx context.Context, ws *workspace.Workspace, rootOverride string, raw string) ([]string, error) {
	suffix := strings.Trim(filepath.ToSlash(filepath.Clean(strings.TrimSpace(raw))), "/")
	if suffix == "" || suffix == "." {
		return nil, fmt.Errorf("empty doc reference")
	}

	if ws == nil {
		discovered, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: rootOverride})
		if err != nil {
			return nil, err
		}
		if err := discovered.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
			return nil, err
		}
		ws = discovered
	}

	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeRepo},
		Options: workspace.DocQueryOptions{
			IncludeErrors:       true,
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeSourcesPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByPath,
		},
	})
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, h := range res.Docs {
		p := filepath.ToSlash(strings.TrimSpace(h.Path))
		if p == "" {
			continue
		}
		if p == suffix || strings.HasSuffix(p, "/"+suffix) {
			matches = append(matches, p)
		}
	}
	return matches, nil
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
