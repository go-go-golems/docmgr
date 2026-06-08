package commands

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

func TestFindIndexFilesSkipsIgnoredIndexFiles(t *testing.T) {
	tmp := t.TempDir()
	ticketDir := filepath.Join(tmp, "DOC-1--demo")
	writeDoctorTestFile(t, filepath.Join(ticketDir, "index.md"), "root")
	writeDoctorTestFile(t, filepath.Join(ticketDir, "design-doc", "index.md"), "ignored duplicate")
	writeDoctorTestFile(t, filepath.Join(ticketDir, "reference", "index.md"), "real duplicate")

	ignoredFile := filepath.Clean(filepath.Join(ticketDir, "design-doc", "index.md"))
	indexFiles := findIndexFiles(ticketDir, func(path, baseName string, isDir bool) bool {
		return !isDir && filepath.Clean(path) == ignoredFile
	})

	if len(indexFiles) != 2 {
		t.Fatalf("expected 2 non-ignored index files, got %d: %v", len(indexFiles), indexFiles)
	}
	for _, path := range indexFiles {
		if filepath.Clean(path) == ignoredFile {
			t.Fatalf("ignored index file was returned: %v", indexFiles)
		}
	}
}

func writeDoctorTestFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestDoctorReturnsErrorForInvalidVocabulary(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	root := filepath.Join(repo, "ttmp")

	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".ttmp.yaml"), []byte("root: ttmp\nvocabulary: ttmp/vocabulary.yaml\n"), 0o644); err != nil {
		t.Fatalf("write .ttmp.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "vocabulary.yaml"), []byte("topics:\n  - slug: ok\nstatus:\n  - slug: active\nbroken trailing text\n"), 0o644); err != nil {
		t.Fatalf("write vocabulary.yaml: %v", err)
	}

	oldCwd, _ := os.Getwd()
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("chdir repo: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldCwd) })

	cmd, err := NewDoctorCommand()
	if err != nil {
		t.Fatalf("NewDoctorCommand: %v", err)
	}

	defaultSection, ok := cmd.GetDefaultSection()
	if !ok {
		t.Fatal("doctor command missing default section")
	}
	parsedValues := values.New()
	sectionValues, err := values.NewSectionValues(
		defaultSection,
		values.WithFieldValue("ticket", "SCRAPER-FRONTEND-RUNTIME-EVENTS"),
		values.WithFieldValue("root", "ttmp"),
		values.WithFieldValue("all", false),
		values.WithFieldValue("doc", ""),
		values.WithFieldValue("ignore-dir", []string{}),
		values.WithFieldValue("ignore-glob", []string{}),
		values.WithFieldValue("stale-after", 30),
		values.WithFieldValue("fail-on", "none"),
		values.WithFieldValue("diagnostics-json", ""),
		values.WithFieldValue("print-template-schema", false),
		values.WithFieldValue("schema-format", "json"),
	)
	if err != nil {
		t.Fatalf("NewSectionValues: %v", err)
	}
	parsedValues.Set(schema.DefaultSlug, sectionValues)

	err = cmd.Run(context.Background(), parsedValues)
	if err == nil {
		t.Fatal("expected error for invalid vocabulary")
	}
	if !strings.Contains(err.Error(), "failed to load vocabulary") {
		t.Fatalf("expected vocabulary error, got: %v", err)
	}
}
