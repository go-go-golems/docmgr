package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolverNormalizeRelativeToRepo(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	docRoot := filepath.Join(repo, "docmgr", "ttmp")
	docPath := filepath.Join(docRoot, "2025/11/26/TICKET/index.md")

	mustMkdir(t, filepath.Dir(docPath))
	writeFile(t, docPath, "# test\n")

	targetFile := filepath.Join(repo, "pkg", "commands", "relate.go")
	mustMkdir(t, filepath.Dir(targetFile))
	writeFile(t, targetFile, "// relate\n")

	cfgDir := repo

	resolver := NewResolver(ResolverOptions{
		DocsRoot:  docRoot,
		DocPath:   docPath,
		ConfigDir: cfgDir,
		RepoRoot:  repo,
	})

	docDir := filepath.Dir(docPath)
	targetRel, err := filepath.Rel(docDir, targetFile)
	if err != nil {
		t.Fatalf("failed to compute doc-relative path: %v", err)
	}

	normalized := resolver.Normalize(targetRel)

	if normalized.Canonical != "pkg/commands/relate.go" {
		t.Fatalf("expected canonical repo-relative path, got %q", normalized.Canonical)
	}
	if normalized.Best() != "pkg/commands/relate.go" {
		t.Fatalf("expected best representation to be repo-relative, got %q", normalized.Best())
	}
	if normalized.Anchor != AnchorDoc {
		t.Fatalf("expected anchor %s, got %s", AnchorDoc, normalized.Anchor)
	}
	if !normalized.Exists {
		t.Fatalf("expected file to exist")
	}
}

func TestMatchPathsWithSuffix(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	docRoot := filepath.Join(repo, "ttmp")
	docPath := filepath.Join(docRoot, "2025/11/26/TICKET/index.md")

	mustMkdir(t, filepath.Dir(docPath))
	writeFile(t, docPath, "# test\n")

	targetFile := filepath.Join(repo, "foobar", "foobar.md")
	mustMkdir(t, filepath.Dir(targetFile))
	writeFile(t, targetFile, "content")

	resolver := NewResolver(ResolverOptions{
		DocsRoot: docRoot,
		DocPath:  docPath,
		RepoRoot: repo,
	})

	docDir := filepath.Dir(docPath)
	targetRel, err := filepath.Rel(docDir, targetFile)
	if err != nil {
		t.Fatalf("failed to compute relative path: %v", err)
	}

	target := resolver.Normalize(targetRel)
	query := resolver.Normalize("foobar.md")

	if !MatchPaths(query, target) {
		t.Fatalf("expected suffix-based match")
	}
}

func TestDirectoryMatch(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	docRoot := filepath.Join(repo, "ttmp")
	docPath := filepath.Join(docRoot, "2025/11/26/TICKET/index.md")

	mustMkdir(t, filepath.Dir(docPath))
	writeFile(t, docPath, "# test\n")

	targetFile := filepath.Join(repo, "backend", "api", "service.go")
	mustMkdir(t, filepath.Dir(targetFile))
	writeFile(t, targetFile, "package api")

	resolver := NewResolver(ResolverOptions{
		DocsRoot: docRoot,
		DocPath:  docPath,
		RepoRoot: repo,
	})

	dir := resolver.Normalize("backend/api")

	docDir := filepath.Dir(docPath)
	targetRel, err := filepath.Rel(docDir, targetFile)
	if err != nil {
		t.Fatalf("failed to compute relative path: %v", err)
	}
	target := resolver.Normalize(targetRel)

	if !DirectoryMatch(dir, target) {
		t.Fatalf("expected directory match for backend/api")
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("failed to create dir %s: %v", path, err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}
