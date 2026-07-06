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

func newMetaUpdateValues(t *testing.T, cmd *MetaUpdateCommand, doc, field, value string) *values.Values {
	t.Helper()
	defaultSection, ok := cmd.GetDefaultSection()
	if !ok {
		t.Fatal("meta update command missing default section")
	}
	parsedValues := values.New()
	sectionValues, err := values.NewSectionValues(
		defaultSection,
		values.WithFieldValue("doc", doc),
		values.WithFieldValue("ticket", ""),
		values.WithFieldValue("doc-type", ""),
		values.WithFieldValue("field", field),
		values.WithFieldValue("value", value),
		values.WithFieldValue("root", "ttmp"),
	)
	if err != nil {
		t.Fatalf("NewSectionValues: %v", err)
	}
	parsedValues.Set(schema.DefaultSlug, sectionValues)
	return parsedValues
}

func TestMetaUpdateRunReturnsErrorWhenUpdateFails(t *testing.T) {
	tmp := t.TempDir()
	missingDoc := filepath.Join(tmp, "does-not-exist.md")

	cmd, err := NewMetaUpdateCommand()
	if err != nil {
		t.Fatalf("NewMetaUpdateCommand: %v", err)
	}

	err = cmd.Run(context.Background(), newMetaUpdateValues(t, cmd, missingDoc, "Status", "active"))
	if err == nil {
		t.Fatal("expected error when updating a missing document")
	}
	if !strings.Contains(err.Error(), "doc not found") {
		t.Fatalf("expected doc-not-found error, got: %v", err)
	}
}

func TestMetaUpdateRunSucceedsForValidDoc(t *testing.T) {
	tmp := t.TempDir()
	docPath := filepath.Join(tmp, "doc.md")
	content := `---
Title: Test doc
Status: draft
---

# Test doc
`
	if err := os.WriteFile(docPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write doc: %v", err)
	}

	cmd, err := NewMetaUpdateCommand()
	if err != nil {
		t.Fatalf("NewMetaUpdateCommand: %v", err)
	}

	if err := cmd.Run(context.Background(), newMetaUpdateValues(t, cmd, docPath, "Status", "active")); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}
