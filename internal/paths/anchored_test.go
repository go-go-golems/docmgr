package paths

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFileP writes a file, creating parent directories.
func writeFileP(t *testing.T, path, content string) {
	t.Helper()
	mustMkdir(t, filepath.Dir(path))
	writeFile(t, path, content)
}

// wsFixture builds a temp go.work workspace with two repos:
//
//	<tmp>/go.work                 (use ./repoA ./repoB)
//	<tmp>/repoA/                  (.git, go.mod, ttmp docs root, a doc, pkg/main.go)
//	<tmp>/repoB/pkg/lib.go        (sibling member)
//	<tmp>/outside.go              (inside wsRoot but not a member path segment issue: still ws://)
type wsFixture struct {
	wsRoot   string
	repoA    string
	repoB    string
	docsRoot string
	docPath  string
	resolver *Resolver
}

func newWsFixture(t *testing.T) wsFixture {
	t.Helper()
	tmp := t.TempDir()
	repoA := filepath.Join(tmp, "repoA")
	repoB := filepath.Join(tmp, "repoB")
	docsRoot := filepath.Join(repoA, "ttmp")
	docPath := filepath.Join(docsRoot, "2026/07/05/TICK-1--x/index.md")

	writeFileP(t, filepath.Join(tmp, "go.work"), "go 1.23\n\nuse (\n\t./repoA\n\t./repoB\n)\n")
	mustMkdir(t, filepath.Join(repoA, ".git"))
	writeFileP(t, filepath.Join(repoA, "go.mod"), "module example.com/repoa\n")
	writeFileP(t, filepath.Join(repoB, "go.mod"), "module example.com/repob\n")
	writeFileP(t, docPath, "# index\n")
	writeFileP(t, filepath.Join(repoA, "pkg", "main.go"), "package pkg\n")
	writeFileP(t, filepath.Join(repoB, "pkg", "lib.go"), "package pkg\n")

	resolver := NewResolver(ResolverOptions{
		DocsRoot:  docsRoot,
		DocPath:   docPath,
		ConfigDir: repoA,
		RepoRoot:  repoA,
		// WorkspaceRoot intentionally omitted: exercise go.work auto-detection.
	})

	return wsFixture{
		wsRoot:   tmp,
		repoA:    repoA,
		repoB:    repoB,
		docsRoot: docsRoot,
		docPath:  docPath,
		resolver: resolver,
	}
}

func TestParseAnchoredRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in     string
		want   AnchoredPath
		wantOK bool
	}{
		{"repo://pkg/foo.go", AnchoredPath{Scheme: SchemeRepo, Rel: "pkg/foo.go"}, true},
		{"ws://glazed/pkg/fields.go", AnchoredPath{Scheme: SchemeWs, Member: "glazed", Rel: "pkg/fields.go"}, true},
		{"docs://2026/07/05/T/design/01.md", AnchoredPath{Scheme: SchemeDocs, Rel: "2026/07/05/T/design/01.md"}, true},
		{"doc://../reference/01.md", AnchoredPath{Scheme: SchemeDoc, Rel: "../reference/01.md"}, true},
		{"abs:///home/user/x.go", AnchoredPath{Scheme: SchemeAbs, Rel: "/home/user/x.go"}, true},
		{"  repo://pkg/foo.go  ", AnchoredPath{Scheme: SchemeRepo, Rel: "pkg/foo.go"}, true},
		{"http://example.com/x", AnchoredPath{}, false},
		{"pkg/foo.go", AnchoredPath{}, false},
		{"/abs/path.go", AnchoredPath{}, false},
		{"://x", AnchoredPath{}, false},
	}
	for _, tc := range cases {
		got, ok := ParseAnchored(tc.in)
		if ok != tc.wantOK {
			t.Fatalf("ParseAnchored(%q) ok=%v, want %v", tc.in, ok, tc.wantOK)
		}
		if !ok {
			continue
		}
		if got != tc.want {
			t.Fatalf("ParseAnchored(%q) = %+v, want %+v", tc.in, got, tc.want)
		}
		// Round-trip through String and re-parse.
		again, ok2 := ParseAnchored(got.String())
		if !ok2 || again != got {
			t.Fatalf("round-trip failed for %q: %+v -> %q -> %+v", tc.in, got, got.String(), again)
		}
	}
}

func TestResolveAnchoredMatrix(t *testing.T) {
	f := newWsFixture(t)

	cases := []struct {
		name       string
		in         string
		wantAbs    string
		wantExists bool
		wantAnchor Anchor
	}{
		{"repo existing", "repo://pkg/main.go", filepath.Join(f.repoA, "pkg", "main.go"), true, AnchorRepo},
		{"repo missing", "repo://pkg/missing.go", filepath.Join(f.repoA, "pkg", "missing.go"), false, AnchorRepo},
		{"ws sibling existing", "ws://repoB/pkg/lib.go", filepath.Join(f.repoB, "pkg", "lib.go"), true, AnchorWs},
		{"ws sibling missing", "ws://repoB/pkg/nope.go", filepath.Join(f.repoB, "pkg", "nope.go"), false, AnchorWs},
		{"docs existing", "docs://2026/07/05/TICK-1--x/index.md", f.docPath, true, AnchorDocsRoot},
		{"docs missing", "docs://2026/07/05/TICK-1--x/none.md", filepath.Join(f.docsRoot, "2026/07/05/TICK-1--x/none.md"), false, AnchorDocsRoot},
		{"doc escaping repo", "doc://../../../../../../repoB/pkg/lib.go", filepath.Join(f.repoB, "pkg", "lib.go"), true, AnchorDoc},
		{"abs existing", "abs://" + filepath.ToSlash(filepath.Join(f.repoB, "pkg", "lib.go")), filepath.Join(f.repoB, "pkg", "lib.go"), true, AnchorAbs},
		{"abs missing", "abs:///nonexistent/definitely/missing.go", "/nonexistent/definitely/missing.go", false, AnchorAbs},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n := f.resolver.Resolve(tc.in)
			if n.Abs != filepath.ToSlash(tc.wantAbs) {
				t.Fatalf("Resolve(%q).Abs = %q, want %q", tc.in, n.Abs, filepath.ToSlash(tc.wantAbs))
			}
			if n.Exists != tc.wantExists {
				t.Fatalf("Resolve(%q).Exists = %v, want %v", tc.in, n.Exists, tc.wantExists)
			}
			if n.Anchor != tc.wantAnchor {
				t.Fatalf("Resolve(%q).Anchor = %q, want %q", tc.in, n.Anchor, tc.wantAnchor)
			}
			// Anchored inputs must agree between Resolve and ResolveNoFS on the
			// anchor and absolute path (the anchor is explicit).
			nfs := f.resolver.ResolveNoFS(tc.in)
			if nfs.Abs != n.Abs || nfs.Anchor != n.Anchor {
				t.Fatalf("ResolveNoFS(%q) = {Abs:%q Anchor:%q}, want {Abs:%q Anchor:%q}", tc.in, nfs.Abs, nfs.Anchor, n.Abs, n.Anchor)
			}
			// Canonical must round-trip to the same absolute path.
			n2 := f.resolver.Resolve(n.Canonical)
			if n2.Abs != n.Abs {
				t.Fatalf("Canonical %q re-resolves to %q, want %q", n.Canonical, n2.Abs, n.Abs)
			}
		})
	}
}

func TestAnchoredForTightestAnchor(t *testing.T) {
	f := newWsFixture(t)

	outside := t.TempDir()
	writeFileP(t, filepath.Join(outside, "ext.go"), "package ext\n")

	cases := []struct {
		name string
		abs  string
		want string
	}{
		{"inside repo", filepath.Join(f.repoA, "pkg", "main.go"), "repo://pkg/main.go"},
		{"inside docs root (repo wins)", f.docPath, "repo://ttmp/2026/07/05/TICK-1--x/index.md"},
		{"sibling go.work member", filepath.Join(f.repoB, "pkg", "lib.go"), "ws://repoB/pkg/lib.go"},
		{"outside everything", filepath.Join(outside, "ext.go"), "abs://" + filepath.ToSlash(filepath.Join(outside, "ext.go"))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := f.resolver.AnchoredFor(tc.abs).String()
			if got != tc.want {
				t.Fatalf("AnchoredFor(%q) = %q, want %q", tc.abs, got, tc.want)
			}
			// Write→resolve agreement: the anchored form resolves back to abs.
			n := f.resolver.Resolve(got)
			if n.Abs != filepath.ToSlash(tc.abs) {
				t.Fatalf("Resolve(AnchoredFor(%q)) = %q, want the original path", tc.abs, n.Abs)
			}
		})
	}
}

func TestAnchoredForDocsRootOutsideRepo(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	docsRoot := filepath.Join(tmp, "docs")
	docFile := filepath.Join(docsRoot, "2026/07/05/T--x/index.md")
	mustMkdir(t, filepath.Join(repo, ".git"))
	writeFileP(t, docFile, "# x\n")

	resolver := NewResolver(ResolverOptions{
		DocsRoot: docsRoot,
		RepoRoot: repo,
	})
	got := resolver.AnchoredFor(docFile).String()
	// tmp has no go.work, so docs:// is the tightest anchor.
	want := "docs://2026/07/05/T--x/index.md"
	if got != want {
		t.Fatalf("AnchoredFor(%q) = %q, want %q", docFile, got, want)
	}
	n := resolver.Resolve(got)
	if !n.Exists || n.Abs != filepath.ToSlash(docFile) {
		t.Fatalf("Resolve(%q) = {Abs:%q Exists:%v}, want existing %q", got, n.Abs, n.Exists, filepath.ToSlash(docFile))
	}
}

func TestFindWorkspaceRootFrom(t *testing.T) {
	f := newWsFixture(t)
	if got := FindWorkspaceRootFrom(f.repoA); got != f.wsRoot {
		t.Fatalf("FindWorkspaceRootFrom(%q) = %q, want %q", f.repoA, got, f.wsRoot)
	}
	if got := FindWorkspaceRootFrom(filepath.Join(f.repoB, "pkg")); got != f.wsRoot {
		t.Fatalf("FindWorkspaceRootFrom(repoB/pkg) = %q, want %q", got, f.wsRoot)
	}
	empty := t.TempDir()
	if got := FindWorkspaceRootFrom(empty); got != "" {
		t.Fatalf("FindWorkspaceRootFrom(%q) = %q, want empty", empty, got)
	}
}

func TestFindGitRootFrom(t *testing.T) {
	f := newWsFixture(t)
	if got := FindGitRootFrom(filepath.Join(f.repoA, "pkg")); got != f.repoA {
		t.Fatalf("FindGitRootFrom = %q, want %q", got, f.repoA)
	}
	// .git file with valid gitdir pointer (worktree layout).
	tmp := t.TempDir()
	wt := filepath.Join(tmp, "wt")
	gd := filepath.Join(tmp, "gitdir")
	mustMkdir(t, wt)
	mustMkdir(t, gd)
	writeFileP(t, filepath.Join(wt, ".git"), "gitdir: "+gd+"\n")
	if got := FindGitRootFrom(wt); got != wt {
		t.Fatalf("FindGitRootFrom(worktree) = %q, want %q", got, wt)
	}
}

func TestMatchPathsStrictness(t *testing.T) {
	f := newWsFixture(t)

	writeFileP(t, filepath.Join(f.repoA, "backend", "api.go"), "package backend\n")
	writeFileP(t, filepath.Join(f.repoA, "backend", "chatapi.go"), "package backend\n")

	apiQuery := f.resolver.Resolve("api.go")
	target := f.resolver.Resolve("backend/api.go")
	chatTarget := f.resolver.Resolve("backend/chatapi.go")

	if !MatchPaths(apiQuery, target) {
		t.Fatalf("expected basename query api.go to match backend/api.go")
	}
	if MatchPaths(apiQuery, chatTarget) {
		t.Fatalf("substring containment must be gone: api.go must NOT match chatapi.go")
	}

	// Case-sensitive: API.go must not match api.go.
	upperQuery := f.resolver.Resolve("API.go")
	if MatchPaths(upperQuery, target) {
		t.Fatalf("suffix matching must be case-sensitive: API.go must NOT match api.go")
	}

	// Whole-segment suffix with multiple segments.
	multi := f.resolver.Resolve("backend/api.go")
	if !MatchPaths(multi, target) {
		t.Fatalf("expected backend/api.go to match itself via abs equality")
	}

	// Different parents sharing a basename must not match on full paths.
	writeFileP(t, filepath.Join(f.repoA, "other", "api.go"), "package other\n")
	otherTarget := f.resolver.Resolve("other/api.go")
	if MatchPaths(target, otherTarget) {
		t.Fatalf("backend/api.go must NOT match other/api.go")
	}

	// Anchored query matches legacy entry for the same file.
	anchoredQuery := f.resolver.Resolve("repo://backend/api.go")
	if !MatchPaths(anchoredQuery, target) {
		t.Fatalf("expected repo://backend/api.go to match legacy backend/api.go")
	}

	// Cross-repo: ws:// entry matches the sibling file's absolute path.
	wsEntry := f.resolver.Resolve("ws://repoB/pkg/lib.go")
	absQuery := f.resolver.Resolve(filepath.Join(f.repoB, "pkg", "lib.go"))
	if !MatchPaths(wsEntry, absQuery) {
		t.Fatalf("expected ws://repoB/pkg/lib.go to match its absolute path")
	}
}

func TestLegacyDocRelativeStillEscapesViaDocAnchorOnlyWhenAnchored(t *testing.T) {
	f := newWsFixture(t)

	// Legacy ../ chain to a sibling repo: the doc anchor must stay inside the
	// repo for legacy strings (historical behavior), so this does NOT resolve
	// to the sibling file...
	rel, err := filepath.Rel(filepath.Dir(f.docPath), filepath.Join(f.repoB, "pkg", "lib.go"))
	if err != nil {
		t.Fatalf("rel: %v", err)
	}
	legacy := f.resolver.Resolve(filepath.ToSlash(rel))
	if legacy.Exists {
		t.Fatalf("legacy ../ chain across repos should not resolve to an existing file, got %+v", legacy)
	}
	// ...but the explicit doc:// anchor may escape.
	anchored := f.resolver.Resolve("doc://" + filepath.ToSlash(rel))
	if !anchored.Exists {
		t.Fatalf("doc:// anchor must be allowed to escape the repo, got %+v", anchored)
	}
	if anchored.Abs != filepath.ToSlash(filepath.Join(f.repoB, "pkg", "lib.go")) {
		t.Fatalf("doc:// resolved to %q", anchored.Abs)
	}
}

func TestNormalizeLegacyStillWorks(t *testing.T) {
	f := newWsFixture(t)
	n := f.resolver.Resolve("pkg/main.go")
	if !n.Exists || n.Anchor != AnchorRepo || n.Canonical != "pkg/main.go" {
		t.Fatalf("legacy repo-relative resolution regressed: %+v", n)
	}
	if _, err := os.Stat(filepath.FromSlash(n.Abs)); err != nil {
		t.Fatalf("stat %q: %v", n.Abs, err)
	}
}
