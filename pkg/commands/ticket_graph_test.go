package commands

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

func buildTicketGraphCobra(t *testing.T) *cobra.Command {
	t.Helper()

	cmdImpl, err := NewTicketGraphCommand()
	if err != nil {
		t.Fatalf("NewTicketGraphCommand: %v", err)
	}
	cobraCmd, err := common.BuildCommand(
		cmdImpl,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
	)
	if err != nil {
		t.Fatalf("BuildCommand: %v", err)
	}
	cobraCmd.SetErr(io.Discard)
	cobraCmd.SilenceErrors = true
	cobraCmd.SilenceUsage = true
	return cobraCmd
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = buf.ReadFrom(r)
		close(done)
	}()

	runErr := fn()

	_ = w.Close()
	<-done
	os.Stdout = origStdout
	_ = r.Close()

	return buf.String(), runErr
}

func TestSanitizeMermaidLabel(t *testing.T) {
	in := "Hello \"world\"\nA|B [C]"
	got := sanitizeMermaidLabel(in, 1000)
	if strings.Contains(got, "\n") {
		t.Fatalf("expected newlines to be escaped, got %q", got)
	}
	if strings.Contains(got, "\"") {
		t.Fatalf("expected quotes to be sanitized, got %q", got)
	}
	if strings.Contains(got, "|") {
		t.Fatalf("expected pipes to be sanitized, got %q", got)
	}
	if strings.Contains(got, "[") || strings.Contains(got, "]") {
		t.Fatalf("expected brackets to be sanitized, got %q", got)
	}
	if !strings.Contains(got, "\\n") {
		t.Fatalf("expected newline escape sequence, got %q", got)
	}
}

func TestShortHashIsStable(t *testing.T) {
	a := shortHash("abc")
	b := shortHash("abc")
	c := shortHash("abcd")
	if a != b {
		t.Fatalf("expected stable hash, got %q vs %q", a, b)
	}
	if a == c {
		t.Fatalf("expected different hash for different input, got %q", a)
	}
	if len(a) != 10 {
		t.Fatalf("expected short hash length 10, got %d (%q)", len(a), a)
	}
}

func TestTicketGraph_Depth0(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	root := filepath.Join(repo, "ttmp")
	ticket := "MEN-1"
	ticketDir := filepath.Join(root, "2026", "01", "03", ticket+"--demo")

	if err := os.MkdirAll(filepath.Join(repo, "pkg"), 0o755); err != nil {
		t.Fatalf("mkdir repo/pkg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "pkg", "a.go"), []byte("package pkg\n"), 0o644); err != nil {
		t.Fatalf("write a.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example.com/test\n\ngo 1.24.2\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.MkdirAll(ticketDir, 0o755); err != nil {
		t.Fatalf("mkdir ticket dir: %v", err)
	}

	index := `---
Title: Ticket One
Ticket: MEN-1
DocType: index
Status: active
Topics: []
Owners: []
RelatedFiles:
  - Path: pkg/a.go
    Note: A note
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T00:00:00Z
---

# Ticket One
`
	if err := os.WriteFile(filepath.Join(ticketDir, "index.md"), []byte(index), 0o644); err != nil {
		t.Fatalf("write index.md: %v", err)
	}

	oldCwd, _ := os.Getwd()
	_ = os.Chdir(repo)
	t.Cleanup(func() { _ = os.Chdir(oldCwd) })

	cobraCmd := buildTicketGraphCobra(t)
	cobraCmd.SetArgs([]string{
		"--ticket", ticket,
		"--root", root,
		"--format", "mermaid",
	})
	s, err := captureStdout(t, cobraCmd.Execute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(s, "graph TD") {
		t.Fatalf("expected mermaid header, got: %s", s)
	}
	if !strings.Contains(s, "MEN-1") {
		t.Fatalf("expected ticket path/label to include MEN-1, got: %s", s)
	}
	if !strings.Contains(s, "pkg/a.go") {
		t.Fatalf("expected file node, got: %s", s)
	}
}

func TestTicketGraph_TransitiveDepth1_RepoScope(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	root := filepath.Join(repo, "ttmp")

	if err := os.MkdirAll(filepath.Join(repo, "pkg"), 0o755); err != nil {
		t.Fatalf("mkdir repo/pkg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "pkg", "a.go"), []byte("package pkg\n"), 0o644); err != nil {
		t.Fatalf("write a.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example.com/test\n\ngo 1.24.2\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	// Ticket 1 references pkg/a.go.
	t1Dir := filepath.Join(root, "2026", "01", "03", "MEN-1--demo")
	if err := os.MkdirAll(t1Dir, 0o755); err != nil {
		t.Fatalf("mkdir t1: %v", err)
	}
	t1 := `---
Title: Ticket One
Ticket: MEN-1
DocType: index
Status: active
Topics: []
Owners: []
RelatedFiles:
  - Path: pkg/a.go
    Note: shared
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T00:00:00Z
---

# Ticket One
`
	if err := os.WriteFile(filepath.Join(t1Dir, "index.md"), []byte(t1), 0o644); err != nil {
		t.Fatalf("write t1 index: %v", err)
	}

	// Ticket 2 also references pkg/a.go.
	t2Dir := filepath.Join(root, "2026", "01", "03", "MEN-2--demo")
	if err := os.MkdirAll(filepath.Join(t2Dir, "reference"), 0o755); err != nil {
		t.Fatalf("mkdir t2: %v", err)
	}
	t2 := `---
Title: External Doc
Ticket: MEN-2
DocType: reference
Status: active
Topics: []
Owners: []
RelatedFiles:
  - Path: pkg/a.go
    Note: also shared
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T00:00:00Z
---

# External Doc
`
	if err := os.WriteFile(filepath.Join(t2Dir, "reference", "01-external.md"), []byte(t2), 0o644); err != nil {
		t.Fatalf("write t2 doc: %v", err)
	}

	oldCwd, _ := os.Getwd()
	_ = os.Chdir(repo)
	t.Cleanup(func() { _ = os.Chdir(oldCwd) })

	cobraCmd := buildTicketGraphCobra(t)
	cobraCmd.SetArgs([]string{
		"--ticket", "MEN-1",
		"--root", root,
		"--format", "mermaid",
		"--scope", "repo",
		"--depth", "1",
		"--expand-files=false",
		"--max-nodes", "200",
		"--max-edges", "500",
	})
	s, err := captureStdout(t, cobraCmd.Execute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(s, "MEN-2") {
		t.Fatalf("expected transitive expansion to include external ticket doc, got: %s", s)
	}
}

func TestTicketGraph_TransitiveDepth1_RepoScope_BasenameTriggerKeepsEdge(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	root := filepath.Join(repo, "ttmp")

	if err := os.MkdirAll(filepath.Join(repo, "pkg"), 0o755); err != nil {
		t.Fatalf("mkdir repo/pkg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "pkg", "a.go"), []byte("package pkg\n"), 0o644); err != nil {
		t.Fatalf("write a.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example.com/test\n\ngo 1.24.2\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	// Ticket 1 references the shared file by basename only ("a.go"). This does not exist
	// at repo root, so canonicalizeForGraph yields "a.go"; QueryDocs will still match other
	// docs via suffix matching ("%/a.go").
	t1Dir := filepath.Join(root, "2026", "01", "03", "MEN-1--demo")
	if err := os.MkdirAll(t1Dir, 0o755); err != nil {
		t.Fatalf("mkdir t1: %v", err)
	}
	t1 := `---
Title: Ticket One
Ticket: MEN-1
DocType: index
Status: active
Topics: []
Owners: []
RelatedFiles:
  - Path: a.go
    Note: basename-only
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T00:00:00Z
---

# Ticket One
`
	if err := os.WriteFile(filepath.Join(t1Dir, "index.md"), []byte(t1), 0o644); err != nil {
		t.Fatalf("write t1 index: %v", err)
	}

	// Ticket 2 references the same file with a repo-relative path.
	t2Dir := filepath.Join(root, "2026", "01", "03", "MEN-2--demo")
	if err := os.MkdirAll(filepath.Join(t2Dir, "reference"), 0o755); err != nil {
		t.Fatalf("mkdir t2: %v", err)
	}
	t2 := `---
Title: External Doc
Ticket: MEN-2
DocType: reference
Status: active
Topics: []
Owners: []
RelatedFiles:
  - Path: pkg/a.go
    Note: repo-relative
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-03T00:00:00Z
---

# External Doc
`
	if err := os.WriteFile(filepath.Join(t2Dir, "reference", "01-external.md"), []byte(t2), 0o644); err != nil {
		t.Fatalf("write t2 doc: %v", err)
	}

	oldCwd, _ := os.Getwd()
	_ = os.Chdir(repo)
	t.Cleanup(func() { _ = os.Chdir(oldCwd) })

	cobraCmd := buildTicketGraphCobra(t)
	cobraCmd.SetArgs([]string{
		"--ticket", "MEN-1",
		"--root", root,
		"--format", "mermaid",
		"--scope", "repo",
		"--depth", "1",
		"--expand-files=false",
		"--max-nodes", "200",
		"--max-edges", "500",
	})
	s, err := captureStdout(t, cobraCmd.Execute)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(s, "MEN-2") {
		t.Fatalf("expected transitive expansion to include external ticket doc, got: %s", s)
	}
	if !strings.Contains(s, "pkg/a.go") {
		t.Fatalf("expected suffix-matched trigger to keep an edge to the canonical related file (pkg/a.go), got: %s", s)
	}
}

func TestTicketGraph_DepthRequiresRepoScope(t *testing.T) {
	// Don't exercise the cobra wiring for this error case because the glazed CLI wrapper
	// can call os.Exit on errors; validate the underlying contract directly.
	_, err := buildTicketGraph(context.Background(), &TicketGraphSettings{
		Ticket:    "MEN-1",
		Root:      "ttmp",
		Depth:     1,
		Scope:     "ticket",
		MaxNodes:  1,
		MaxEdges:  1,
		BatchSize: 1,
	})
	if err == nil {
		t.Fatalf("expected error when --depth>0 without --scope repo")
	}
}
