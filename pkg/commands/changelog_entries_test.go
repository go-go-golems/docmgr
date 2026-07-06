package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseChangelogEntries(t *testing.T) {
	t.Parallel()

	content := `# Changelog

Intro text that belongs to no entry.

## 2026-07-01 - First pass

Did the first thing.

### Related Files

- pkg/foo.go — main change

## 2026-07-03

Second entry without a title.
`

	entries := ParseChangelogEntries(content)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d: %+v", len(entries), entries)
	}

	if entries[0].Date != "2026-07-01" || entries[0].Title != "First pass" {
		t.Fatalf("unexpected first entry: %+v", entries[0])
	}
	if !strings.Contains(entries[0].Body, "Did the first thing.") ||
		!strings.Contains(entries[0].Body, "### Related Files") {
		t.Fatalf("unexpected first body: %q", entries[0].Body)
	}

	if entries[1].Date != "2026-07-03" || entries[1].Title != "" {
		t.Fatalf("unexpected second entry: %+v", entries[1])
	}
	if !strings.Contains(entries[1].Body, "Second entry without a title.") {
		t.Fatalf("unexpected second body: %q", entries[1].Body)
	}
}

func TestAppendChangelogEntry_RoundTrip(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "sub", "changelog.md")
	date, err := AppendChangelogEntry(path, "Round trip", "Something happened.", map[string]string{
		"pkg/foo.go": "the change",
	})
	if err != nil {
		t.Fatalf("AppendChangelogEntry: %v", err)
	}
	if date == "" {
		t.Fatal("expected non-empty date")
	}

	raw, err := os.ReadFile(path) // #nosec G304 -- test fixture path
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	entries := ParseChangelogEntries(string(raw))
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d (%s)", len(entries), string(raw))
	}
	if entries[0].Date != date || entries[0].Title != "Round trip" {
		t.Fatalf("unexpected entry: %+v", entries[0])
	}
	if !strings.Contains(entries[0].Body, "Something happened.") ||
		!strings.Contains(entries[0].Body, "pkg/foo.go — the change") {
		t.Fatalf("unexpected body: %q", entries[0].Body)
	}

	if _, err := AppendChangelogEntry(path, "", "   ", nil); err == nil {
		t.Fatal("expected error for empty entry")
	}
}
