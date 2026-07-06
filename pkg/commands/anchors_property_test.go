package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/types"
)

// anchorsFixture is a temp go.work workspace with two repos:
//
//	<ws>/go.work                    use ./repoA ./repoB
//	<ws>/repoA                      the "current" repo (.git, go.mod, ttmp/)
//	<ws>/repoB/pkg/lib.go           sibling go.work member file
//	<outside>/ext.go                absolute path outside the workspace
type anchorsFixture struct {
	wsRoot   string
	repoA    string
	repoB    string
	outside  string
	docsRoot string
	ticket1  string // index.md path (rel to repoA)
	ticket2  string
}

func newAnchorsFixture(t *testing.T) anchorsFixture {
	t.Helper()
	tmp := t.TempDir()
	outside := t.TempDir()
	repoA := filepath.Join(tmp, "repoA")
	repoB := filepath.Join(tmp, "repoB")
	docsRoot := filepath.Join(repoA, "ttmp")

	writeAnchorsFile(t, filepath.Join(tmp, "go.work"), "go 1.23\n\nuse (\n\t./repoA\n\t./repoB\n)\n")
	if err := os.MkdirAll(filepath.Join(repoA, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	writeAnchorsFile(t, filepath.Join(repoA, "go.mod"), "module example.com/repoa\n")
	writeAnchorsFile(t, filepath.Join(repoB, "go.mod"), "module example.com/repob\n")
	writeAnchorsFile(t, filepath.Join(repoA, "backend", "api.go"), "package backend\n")
	writeAnchorsFile(t, filepath.Join(repoA, "backend", "chatapi.go"), "package backend\n")
	writeAnchorsFile(t, filepath.Join(repoA, "backend", "legacy.go"), "package backend\n")
	writeAnchorsFile(t, filepath.Join(repoB, "pkg", "lib.go"), "package pkg\n")
	writeAnchorsFile(t, filepath.Join(outside, "ext.go"), "package ext\n")

	ticket1 := filepath.Join("ttmp", "2026", "07", "05", "TICK-1--anchors", "index.md")
	ticket2 := filepath.Join("ttmp", "2026", "07", "05", "TICK-2--other", "index.md")
	// Pre-seed TICK-1 with a legacy entry (existing file) and a legacy missing
	// entry: relate must preserve both as-is; --fix-anchors migrates only the
	// first.
	writeAnchorsFile(t, filepath.Join(repoA, ticket1), `---
Title: Anchor Property Test
Ticket: TICK-1
DocType: index
Status: active
Intent: long-term
Topics: [testing]
RelatedFiles:
  - Path: backend/legacy.go
    Note: legacy entry that resolves
  - Path: backend/legacy-gone.go
    Note: legacy entry that does not resolve
LastUpdated: 2026-07-05T00:00:00Z
---

# TICK-1
`)
	writeAnchorsFile(t, filepath.Join(repoA, ticket2), `---
Title: Second Ticket
Ticket: TICK-2
DocType: index
Status: active
Intent: long-term
Topics: [testing]
RelatedFiles:
  - Path: backend/chatapi.go
    Note: only chatapi here
LastUpdated: 2026-07-05T00:00:00Z
---

# TICK-2
`)

	oldCwd, _ := os.Getwd()
	if err := os.Chdir(repoA); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldCwd) })

	return anchorsFixture{
		wsRoot:   tmp,
		repoA:    repoA,
		repoB:    repoB,
		outside:  outside,
		docsRoot: docsRoot,
		ticket1:  ticket1,
		ticket2:  ticket2,
	}
}

func writeAnchorsFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

type rowCollector struct{ rows []types.Row }

func (c *rowCollector) AddRow(_ context.Context, row types.Row) error {
	c.rows = append(c.rows, row)
	return nil
}
func (c *rowCollector) Close(context.Context) error { return nil }

func runRelateForTest(t *testing.T, doc string, fileNotes []string) {
	t.Helper()
	cmd, err := NewRelateCommand()
	if err != nil {
		t.Fatalf("NewRelateCommand: %v", err)
	}
	section, ok := cmd.GetDefaultSection()
	if !ok {
		t.Fatal("relate command missing default section")
	}
	parsed := values.New()
	sectionValues, err := values.NewSectionValues(
		section,
		values.WithFieldValue("ticket", ""),
		values.WithFieldValue("doc", doc),
		values.WithFieldValue("remove-files", []string{}),
		values.WithFieldValue("file-note", fileNotes),
		values.WithFieldValue("suggest", false),
		values.WithFieldValue("apply-suggestions", false),
		values.WithFieldValue("from-git", false),
		values.WithFieldValue("query", ""),
		values.WithFieldValue("topics", []string{}),
		values.WithFieldValue("root", "ttmp"),
	)
	if err != nil {
		t.Fatalf("NewSectionValues: %v", err)
	}
	parsed.Set(schema.DefaultSlug, sectionValues)

	collector := &rowCollector{}
	if err := cmd.RunIntoGlazeProcessor(context.Background(), parsed, collector); err != nil {
		t.Fatalf("relate failed: %v", err)
	}
}

func runDoctorForTest(t *testing.T, all bool, ticket string, fixAnchors bool) []types.Row {
	t.Helper()
	cmd, err := NewDoctorCommand()
	if err != nil {
		t.Fatalf("NewDoctorCommand: %v", err)
	}
	section, ok := cmd.GetDefaultSection()
	if !ok {
		t.Fatal("doctor command missing default section")
	}
	parsed := values.New()
	sectionValues, err := values.NewSectionValues(
		section,
		values.WithFieldValue("ticket", ticket),
		values.WithFieldValue("root", "ttmp"),
		values.WithFieldValue("all", all),
		values.WithFieldValue("doc", ""),
		values.WithFieldValue("ignore-dir", []string{}),
		values.WithFieldValue("ignore-glob", []string{}),
		values.WithFieldValue("stale-after", 100000),
		values.WithFieldValue("fail-on", "none"),
		values.WithFieldValue("diagnostics-json", ""),
		values.WithFieldValue("fix-anchors", fixAnchors),
		values.WithFieldValue("print-template-schema", false),
		values.WithFieldValue("schema-format", "json"),
	)
	if err != nil {
		t.Fatalf("NewSectionValues: %v", err)
	}
	parsed.Set(schema.DefaultSlug, sectionValues)

	collector := &rowCollector{}
	if err := cmd.RunIntoGlazeProcessor(context.Background(), parsed, collector); err != nil {
		t.Fatalf("doctor failed: %v", err)
	}
	return collector.rows
}

func doctorIssuesForPath(rows []types.Row, issue string) []string {
	var out []string
	for _, row := range rows {
		iv, _ := row.Get("issue")
		if fmt.Sprint(iv) != issue {
			continue
		}
		mv, _ := row.Get("message")
		out = append(out, fmt.Sprint(mv))
	}
	return out
}

// TestAnchorsWriteIndexDoctorResolveAgreement is the paths-v2 property test
// (design doc DOCMGR-200 §9 Phase 2): what relate writes, what the index
// stores, what doctor validates and what Resolve returns must be one
// consistent story across anchor × {existing,missing} × location.
func TestAnchorsWriteIndexDoctorResolveAgreement(t *testing.T) {
	f := newAnchorsFixture(t)

	siblingExisting := filepath.Join(f.repoB, "pkg", "lib.go")
	siblingMissing := filepath.Join(f.repoB, "pkg", "nope.go")
	outsideExisting := filepath.Join(f.outside, "ext.go")
	outsideMissing := filepath.Join(f.outside, "gone.go")

	commaNote := "covers sections 4.4, 8.1 (comma preserved)"

	runRelateForTest(t, f.ticket1, []string{
		"backend/api.go:" + commaNote,
		"backend/missing.go:in-repo missing file",
		siblingExisting + ":sibling repo file",
		siblingMissing + ":sibling repo missing file",
		outsideExisting + ":outside workspace file",
		outsideMissing + ":outside workspace missing file",
	})

	// --- 1. Write side: frontmatter carries anchored paths (tightest anchor),
	// legacy entries preserved verbatim.
	doc, _, err := documents.ReadDocumentWithFrontmatter(f.ticket1)
	if err != nil {
		t.Fatalf("read frontmatter: %v", err)
	}
	got := map[string]string{}
	for _, rf := range doc.RelatedFiles {
		got[rf.Path] = rf.Note
	}
	expectedWrites := map[string]struct {
		abs    string
		exists bool
	}{
		"repo://backend/api.go":                      {filepath.Join(f.repoA, "backend", "api.go"), true},
		"repo://backend/missing.go":                  {filepath.Join(f.repoA, "backend", "missing.go"), false},
		"ws://repoB/pkg/lib.go":                      {siblingExisting, true},
		"ws://repoB/pkg/nope.go":                     {siblingMissing, false},
		"abs://" + filepath.ToSlash(outsideExisting): {outsideExisting, true},
		"abs://" + filepath.ToSlash(outsideMissing):  {outsideMissing, false},
		// preserved legacy entries
		"backend/legacy.go":      {filepath.Join(f.repoA, "backend", "legacy.go"), true},
		"backend/legacy-gone.go": {filepath.Join(f.repoA, "backend", "legacy-gone.go"), false},
	}
	if len(got) != len(expectedWrites) {
		t.Fatalf("expected %d RelatedFiles entries, got %d: %v", len(expectedWrites), len(got), got)
	}
	for path := range expectedWrites {
		if _, ok := got[path]; !ok {
			t.Fatalf("expected frontmatter to contain %q, got %v", path, got)
		}
	}
	if got["repo://backend/api.go"] != commaNote {
		t.Fatalf("comma note mangled: %q", got["repo://backend/api.go"])
	}
	for p := range got {
		if strings.Contains(p, "../") {
			t.Fatalf("repo-escaping ../ chain written: %q", p)
		}
	}

	// --- 2 + 4. Index + Resolve agreement: for each entry the index's norm_abs
	// equals what the doc-anchored resolver resolves, and Exists is honest.
	ws, err := workspace.DiscoverWorkspace(context.Background(), workspace.DiscoverOptions{RootOverride: "ttmp"})
	if err != nil {
		t.Fatalf("DiscoverWorkspace: %v", err)
	}
	if ws.Context().WorkspaceRoot != f.wsRoot {
		t.Fatalf("workspace root = %q, want %q", ws.Context().WorkspaceRoot, f.wsRoot)
	}
	if err := ws.InitIndex(context.Background(), workspace.BuildIndexOptions{}); err != nil {
		t.Fatalf("InitIndex: %v", err)
	}

	docAbs := filepath.Join(f.repoA, f.ticket1)
	resolver := paths.NewResolver(paths.ResolverOptions{
		DocsRoot:      ws.Context().Root,
		DocPath:       docAbs,
		ConfigDir:     ws.Context().ConfigDir,
		RepoRoot:      ws.Context().RepoRoot,
		WorkspaceRoot: ws.Context().WorkspaceRoot,
	})

	for path, want := range expectedWrites {
		n := resolver.Resolve(path)
		if n.Abs != filepath.ToSlash(want.abs) {
			t.Fatalf("Resolve(%q).Abs = %q, want %q", path, n.Abs, filepath.ToSlash(want.abs))
		}
		if n.Exists != want.exists {
			t.Fatalf("Resolve(%q).Exists = %v, want %v", path, n.Exists, want.exists)
		}

		var normAbs string
		err := ws.DB().QueryRowContext(context.Background(),
			`SELECT COALESCE(norm_abs,'') FROM related_files WHERE raw_path = ?`, path,
		).Scan(&normAbs)
		if err != nil {
			t.Fatalf("index row for %q: %v", path, err)
		}
		if normAbs != filepath.ToSlash(want.abs) {
			t.Fatalf("index norm_abs for %q = %q, want %q", path, normAbs, filepath.ToSlash(want.abs))
		}
	}

	// --- 3. Reverse lookup through the index: absolute sibling path, anchored
	// form, repo-relative form and bare basename all find TICK-1's index.
	queries := []string{
		siblingExisting,
		"ws://repoB/pkg/lib.go",
		"backend/api.go",
		"api.go",
		outsideExisting,
	}
	for _, q := range queries {
		res, err := ws.QueryDocs(context.Background(), workspace.DocQuery{
			Scope:   workspace.Scope{Kind: workspace.ScopeRepo},
			Filters: workspace.DocFilters{RelatedFile: []string{q}},
		})
		if err != nil {
			t.Fatalf("QueryDocs(%q): %v", q, err)
		}
		found := false
		for _, h := range res.Docs {
			if strings.HasSuffix(filepath.ToSlash(h.Path), "TICK-1--anchors/index.md") {
				found = true
			}
			if strings.HasSuffix(filepath.ToSlash(h.Path), "TICK-2--other/index.md") && q == "api.go" {
				t.Fatalf("reverse lookup %q matched TICK-2 (chatapi.go) — substring matching is back", q)
			}
		}
		if !found {
			t.Fatalf("reverse lookup %q did not find TICK-1 index", q)
		}
	}
	// Negative: chatapi.go must find only TICK-2.
	res, err := ws.QueryDocs(context.Background(), workspace.DocQuery{
		Scope:   workspace.Scope{Kind: workspace.ScopeRepo},
		Filters: workspace.DocFilters{RelatedFile: []string{"chatapi.go"}},
	})
	if err != nil {
		t.Fatalf("QueryDocs(chatapi.go): %v", err)
	}
	for _, h := range res.Docs {
		if strings.HasSuffix(filepath.ToSlash(h.Path), "TICK-1--anchors/index.md") {
			t.Fatalf("chatapi.go query matched TICK-1, which only has api.go")
		}
	}

	// --- 4. Doctor: existing entries (including cross-repo ws:// and abs://)
	// validate cleanly; only genuinely-missing files warn.
	rows := runDoctorForTest(t, false, "TICK-1", false)
	missing := doctorIssuesForPath(rows, "missing_related_file")
	wantMissing := []string{
		"repo://backend/missing.go",
		"ws://repoB/pkg/nope.go",
		"abs://" + filepath.ToSlash(outsideMissing),
		"backend/legacy-gone.go",
	}
	if len(missing) != len(wantMissing) {
		t.Fatalf("expected %d missing_related_file warnings, got %d: %v", len(wantMissing), len(missing), missing)
	}
	for _, w := range wantMissing {
		found := false
		for _, m := range missing {
			if strings.Contains(m, w) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected missing_related_file for %q, got %v", w, missing)
		}
	}
	for _, present := range []string{"repo://backend/api.go", "ws://repoB/pkg/lib.go", "abs://" + filepath.ToSlash(outsideExisting), "backend/legacy.go"} {
		for _, m := range missing {
			if strings.Contains(m, present) {
				t.Fatalf("doctor flagged existing file %q as missing (self-contradiction is back): %v", present, m)
			}
		}
	}

	// --- 5. doctor --fix-anchors migrates the resolvable legacy entry and
	// leaves the unresolvable one as legacy with a warning.
	rows = runDoctorForTest(t, false, "TICK-1", true)
	if migrated := doctorIssuesForPath(rows, "anchors_migrated"); len(migrated) != 1 {
		t.Fatalf("expected 1 anchors_migrated row, got %v", migrated)
	}
	skipped := doctorIssuesForPath(rows, "anchor_migration_skipped")
	if len(skipped) != 1 || !strings.Contains(skipped[0], "backend/legacy-gone.go") {
		t.Fatalf("expected anchor_migration_skipped for backend/legacy-gone.go, got %v", skipped)
	}

	doc, _, err = documents.ReadDocumentWithFrontmatter(f.ticket1)
	if err != nil {
		t.Fatalf("re-read frontmatter: %v", err)
	}
	pathsAfter := map[string]bool{}
	for _, rf := range doc.RelatedFiles {
		pathsAfter[rf.Path] = true
	}
	if !pathsAfter["repo://backend/legacy.go"] {
		t.Fatalf("expected legacy.go migrated to repo://backend/legacy.go, got %v", pathsAfter)
	}
	if pathsAfter["backend/legacy.go"] {
		t.Fatalf("legacy.go should have been rewritten, got %v", pathsAfter)
	}
	if !pathsAfter["backend/legacy-gone.go"] {
		t.Fatalf("unresolvable legacy entry must stay legacy, got %v", pathsAfter)
	}
}
