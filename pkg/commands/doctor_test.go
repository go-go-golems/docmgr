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

func TestDoctorIgnorePatternMatchesNestedDirectorySegments(t *testing.T) {
	patterns := []string{".git/", "node_modules/", "dist/"}

	cases := []struct {
		name string
		path string
		want bool
	}{
		{name: "direct directory", path: "node_modules", want: true},
		{name: "nested package path", path: "2026/06/08/TICKET/scripts/node_modules/playwright/README.md", want: true},
		{name: "absolute nested path", path: filepath.Join(string(filepath.Separator), "repo", "ttmp", "ticket", "scripts", "node_modules", ".pnpm", "pkg", "README.md"), want: true},
		{name: "dist segment", path: "ticket/scripts/dist/bundle.js", want: true},
		{name: "not a segment substring", path: "ticket/scripts/my-node_modules-cache/README.md", want: false},
		{name: "normal doc", path: "ticket/design-doc/01-plan.md", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := matchesAnyGlob(patterns, tc.path)
			if got != tc.want {
				t.Fatalf("matchesAnyGlob(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
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
