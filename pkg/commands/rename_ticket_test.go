package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/glazed/pkg/cli"
)

func writeMarkdown(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdirs failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
}

func TestRenameTicketCommand_EndToEnd(t *testing.T) {
	// Setup temp workspace
	tmp := t.TempDir()
	root := filepath.Join(tmp, "ttmp")
	oldTicket := "MEN-1234"
	newTicket := "MEN-5678"
	ticketDir := filepath.Join(root, "2025", "11", "18", oldTicket+"-my-ticket")

	// Minimal index with frontmatter including Ticket and DocType
	index := strings.TrimSpace(`---
Title: Test Ticket
Ticket: MEN-1234
DocType: index
Status: active
Topics: []
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-18
---

# Index
`)
	writeMarkdown(t, filepath.Join(ticketDir, "index.md"), index)

	// A secondary document under a subdir
	doc := strings.TrimSpace(`---
Title: Design One
Ticket: MEN-1234
DocType: design-doc
Status: draft
Topics: []
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-18
---

# Design One
`)
	writeMarkdown(t, filepath.Join(ticketDir, "design-doc", "01-intro.md"), doc)

	// Build cobra command
	cmdImpl, err := NewRenameTicketCommand()
	if err != nil {
		t.Fatalf("create command failed: %v", err)
	}
	cobraCmd, err := cli.BuildCobraCommand(cmdImpl)
	if err != nil {
		t.Fatalf("build cobra command failed: %v", err)
	}

	// Dry run first: nothing should change
	cobraCmd.SetArgs([]string{
		"--ticket", oldTicket,
		"--new-ticket", newTicket,
		"--root", root,
		"--dry-run",
	})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("dry-run execute failed: %v", err)
	}
	if _, err := os.Stat(ticketDir); err != nil {
		t.Fatalf("dry-run should not move directory, but got error: %v", err)
	}

	// Real rename
	// Rebuild cobra command to avoid persisted flag values from previous run
	cmdImpl2, err := NewRenameTicketCommand()
	if err != nil {
		t.Fatalf("create command (second) failed: %v", err)
	}
	cobraCmd2, err := cli.BuildCobraCommand(cmdImpl2)
	if err != nil {
		t.Fatalf("build cobra command (second) failed: %v", err)
	}
	cobraCmd2.SetArgs([]string{
		"--ticket", oldTicket,
		"--new-ticket", newTicket,
		"--root", root,
	})
	if err := cobraCmd2.Execute(); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	newDir := filepath.Join(root, "2025", "11", "18", newTicket+"-my-ticket")
	if _, err := os.Stat(newDir); err != nil {
		t.Fatalf("expected new directory to exist: %v", err)
	}
	if _, err := os.Stat(ticketDir); !os.IsNotExist(err) {
		t.Fatalf("expected old directory to be gone, got err=%v", err)
	}

	// Validate frontmatter updated in index.md
	idxPath := filepath.Join(newDir, "index.md")
	idxDoc, _, err := documents.ReadDocumentWithFrontmatter(idxPath)
	if err != nil {
		t.Fatalf("parse index frontmatter failed: %v", err)
	}
	if idxDoc.Ticket != newTicket {
		t.Fatalf("index Ticket not updated, got %s", idxDoc.Ticket)
	}

	// Validate frontmatter updated in sub doc
	docPath := filepath.Join(newDir, "design-doc", "01-intro.md")
	subDoc, _, err := documents.ReadDocumentWithFrontmatter(docPath)
	if err != nil {
		t.Fatalf("parse sub doc failed: %v", err)
	}
	if subDoc.Ticket != newTicket {
		t.Fatalf("sub doc Ticket not updated, got %s", subDoc.Ticket)
	}
}
