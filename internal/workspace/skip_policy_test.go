package workspace

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

type fakeDirEntry struct {
	name  string
	isDir bool
}

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return f.isDir }
func (f fakeDirEntry) Type() fs.FileMode          { return 0 }
func (f fakeDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

func TestDefaultIngestSkipDir(t *testing.T) {
	cases := []struct {
		name string
		skip bool
	}{
		{name: ".meta", skip: true},
		{name: "_templates", skip: true},
		{name: "_guidelines", skip: true},
		{name: "archive", skip: false},
		{name: "scripts", skip: false},
		{name: "sources", skip: false},
		{name: "design-doc", skip: false},
	}
	for _, c := range cases {
		got := DefaultIngestSkipDir("", fakeDirEntry{name: c.name, isDir: true})
		if got != c.skip {
			t.Fatalf("name=%q: expected skip=%v, got %v", c.name, c.skip, got)
		}
	}
}

func TestComputePathTags_ControlDocsRequireSiblingIndex(t *testing.T) {
	tmp := t.TempDir()

	// Ticket root: has index.md + tasks.md
	ticketDir := filepath.Join(tmp, "TICKET-1--x")
	if err := os.MkdirAll(ticketDir, 0o755); err != nil {
		t.Fatalf("mkdir ticketDir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ticketDir, "index.md"), []byte("---\n---\n"), 0o644); err != nil {
		t.Fatalf("write index.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ticketDir, "tasks.md"), []byte("---\n---\n"), 0o644); err != nil {
		t.Fatalf("write tasks.md: %v", err)
	}

	tags := ComputePathTags(filepath.Join(ticketDir, "tasks.md"))
	if !tags.IsControlDoc {
		t.Fatalf("expected tasks.md at ticket root to be IsControlDoc=true")
	}

	// sources/README.md: should NOT be tagged as control doc unless sources has its own index.md.
	sourcesDir := filepath.Join(ticketDir, "sources")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatalf("mkdir sourcesDir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcesDir, "README.md"), []byte("---\n---\n"), 0o644); err != nil {
		t.Fatalf("write sources README.md: %v", err)
	}
	tags = ComputePathTags(filepath.Join(sourcesDir, "README.md"))
	if tags.IsControlDoc {
		t.Fatalf("expected sources/README.md to be IsControlDoc=false (no sibling index.md)")
	}
}

func TestComputePathTags_PathSegments(t *testing.T) {
	tmp := t.TempDir()
	ticketDir := filepath.Join(tmp, "TICKET-1--x")
	if err := os.MkdirAll(ticketDir, 0o755); err != nil {
		t.Fatalf("mkdir ticketDir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ticketDir, "index.md"), []byte("---\n---\n"), 0o644); err != nil {
		t.Fatalf("write index.md: %v", err)
	}

	cases := []struct {
		rel  string
		want PathTags
	}{
		{rel: "index.md", want: PathTags{IsIndex: true}},
		{rel: "archive/a.md", want: PathTags{IsArchivedPath: true}},
		{rel: "scripts/a.md", want: PathTags{IsScriptsPath: true}},
		{rel: "sources/a.md", want: PathTags{IsSourcesPath: true}},
	}
	for _, c := range cases {
		p := filepath.Join(ticketDir, filepath.FromSlash(c.rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(p), err)
		}
		if err := os.WriteFile(p, []byte("---\n---\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
		got := ComputePathTags(p)
		if c.want.IsIndex && !got.IsIndex {
			t.Fatalf("rel=%q: expected IsIndex=true", c.rel)
		}
		if c.want.IsArchivedPath && !got.IsArchivedPath {
			t.Fatalf("rel=%q: expected IsArchivedPath=true", c.rel)
		}
		if c.want.IsScriptsPath && !got.IsScriptsPath {
			t.Fatalf("rel=%q: expected IsScriptsPath=true", c.rel)
		}
		if c.want.IsSourcesPath && !got.IsSourcesPath {
			t.Fatalf("rel=%q: expected IsSourcesPath=true", c.rel)
		}
	}
}


