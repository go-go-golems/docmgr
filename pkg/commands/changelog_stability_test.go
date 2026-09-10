package commands

import (
	"strings"
	"testing"
)

func TestChangelogCanonicalBoundaries(t *testing.T) {
	for _, old := range []string{"", "# Changelog\n\n", "# Changelog\n\n## 2000-01-01\n\nAuthored  \n\nparagraph\n\n"} {
		for _, files := range []map[string]string{nil, {"z.go": "last", "a.go": "first"}} {
			result := BuildChangelogEntry(old, "2026-09-10", "Title", "\nLine  \n\nSecond\n\n", files)
			if !strings.HasSuffix(result, "\n") || strings.HasSuffix(result, "\n\n") {
				t.Fatalf("bad EOF: %q", result)
			}
			if !strings.Contains(result, "Line  \n\nSecond") || !strings.Contains(result, "## 2026-09-10 - Title\n\n") {
				t.Fatalf("lost content: %q", result)
			}
			if old != "" && !strings.HasPrefix(result, strings.TrimRight(old, "\n")) {
				t.Fatal("old content changed")
			}
		}
	}
}
