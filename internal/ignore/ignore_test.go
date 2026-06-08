package ignore

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMatcherBuiltinsIgnoreNestedDependencyDirs(t *testing.T) {
	repo, docs := testRepo(t)
	m := loadTestMatcher(t, repo, docs)

	cases := []struct {
		name  string
		path  string
		isDir bool
		want  bool
	}{
		{name: "node_modules dir", path: filepath.Join(docs, "TICKET", "scripts", "node_modules"), isDir: true, want: true},
		{name: "node_modules descendant", path: filepath.Join(docs, "TICKET", "scripts", "node_modules", "pkg", "README.md"), want: true},
		{name: "pnpm descendant", path: filepath.Join(docs, "TICKET", "scripts", ".pnpm", "pkg", "README.md"), want: true},
		{name: "git descendant", path: filepath.Join(docs, "TICKET", ".git", "config"), want: true},
		{name: "dist descendant", path: filepath.Join(docs, "TICKET", "dist", "bundle.md"), want: true},
		{name: "substring does not match", path: filepath.Join(docs, "TICKET", "scripts", "my-node_modules-cache", "README.md"), want: false},
		{name: "normal doc", path: filepath.Join(docs, "TICKET", "design-doc", "01-plan.md"), want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := m.Match(tc.path, tc.isDir)
			if got.Ignored != tc.want {
				t.Fatalf("Match(%q).Ignored = %v, want %v (decision=%+v)", tc.path, got.Ignored, tc.want, got)
			}
		})
	}
}

func TestMatcherLoadsDocsRootDocmgrignore(t *testing.T) {
	repo, docs := testRepo(t)
	writeFile(t, filepath.Join(docs, FileName), "custom-cache/\n**/draft-*.md\n/root-only.md\n")
	m := loadTestMatcher(t, repo, docs)

	ignored := []string{
		filepath.Join(docs, "TICKET", "custom-cache", "notes.md"),
		filepath.Join(docs, "TICKET", "design-doc", "draft-plan.md"),
		filepath.Join(docs, "root-only.md"),
	}
	for _, path := range ignored {
		decision := m.Match(path, false)
		if !decision.Ignored {
			t.Fatalf("expected %s to be ignored, decision=%+v", path, decision)
		}
		if !decision.Matched {
			t.Fatalf("expected %s to record a match", path)
		}
	}

	included := filepath.Join(docs, "TICKET", "design-doc", "01-plan.md")
	if decision := m.Match(included, false); decision.Ignored {
		t.Fatalf("expected %s to be included, decision=%+v", included, decision)
	}
}

func TestMatcherLoadsNestedDocmgrignore(t *testing.T) {
	repo, docs := testRepo(t)
	ticket := filepath.Join(docs, "2026", "06", "08", "TICKET--demo")
	writeFile(t, filepath.Join(ticket, "scripts", FileName), "screenshots/\n")
	m := loadTestMatcher(t, repo, docs)

	ignored := filepath.Join(ticket, "scripts", "screenshots", "shot.md")
	if decision := m.Match(ignored, false); !decision.Ignored {
		t.Fatalf("expected nested .docmgrignore to ignore %s, decision=%+v", ignored, decision)
	}

	notSibling := filepath.Join(ticket, "reference", "screenshots", "shot.md")
	if decision := m.Match(notSibling, false); decision.Ignored {
		t.Fatalf("nested .docmgrignore should not apply to sibling subtree, decision=%+v", decision)
	}
}

func TestMatcherLoadsRepositoryRootDocmgrignore(t *testing.T) {
	repo, docs := testRepo(t)
	writeFile(t, filepath.Join(repo, FileName), "ttmp/**/generated-*.md\n")
	m := loadTestMatcher(t, repo, docs)

	path := filepath.Join(docs, "TICKET", "generated-report.md")
	if decision := m.Match(path, false); !decision.Ignored {
		t.Fatalf("expected repository .docmgrignore to ignore %s, decision=%+v", path, decision)
	}
}

func TestMatcherRelativePathsResolveAgainstDocsRoot(t *testing.T) {
	repo, docs := testRepo(t)
	writeFile(t, filepath.Join(docs, FileName), "scratch/\n")
	m := loadTestMatcher(t, repo, docs)

	decision := m.Match(filepath.Join("TICKET", "scratch", "notes.md"), false)
	if !decision.Ignored {
		t.Fatalf("expected docs-root-relative path to be ignored, decision=%+v", decision)
	}
}

func TestMatcherRepoRelativeDocsRootPathDoesNotDuplicateDocsRoot(t *testing.T) {
	repo, docs := testRepo(t)
	m := loadTestMatcher(t, repo, docs)

	rel := filepath.Join(filepath.Base(docs), "TICKET", "scripts", "node_modules", "pkg", "README.md")
	decision := m.Match(rel, false)
	if !decision.Ignored {
		t.Fatalf("expected repo-relative docs-root path to be ignored, decision=%+v", decision)
	}
	want := filepath.ToSlash(filepath.Join(docs, "TICKET", "scripts", "node_modules", "pkg", "README.md"))
	if decision.Path != want {
		t.Fatalf("resolved path = %q, want %q", decision.Path, want)
	}
}

func testRepo(t *testing.T) (string, string) {
	t.Helper()
	repo := t.TempDir()
	docs := filepath.Join(repo, "ttmp")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	return repo, docs
}

func loadTestMatcher(t *testing.T, repo string, docs string) *Matcher {
	t.Helper()
	m, err := Load(context.Background(), LoadOptions{
		RepoRoot:       repo,
		DocsRoot:       docs,
		IncludeBuiltin: true,
		IncludeNested:  true,
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return m
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
