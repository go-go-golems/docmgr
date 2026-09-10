package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestScaffoldCanonicalEOFAndNoop(t *testing.T) {
	root := t.TempDir()
	if err := scaffoldTemplatesAndGuidelines(root, false); err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(root, "_*", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("missing scaffold")
	}
	stamp := time.Unix(1000, 0)
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasSuffix(string(data), "\n") || strings.HasSuffix(string(data), "\n\n") {
			t.Fatalf("noncanonical EOF: %s", path)
		}
		if err := os.Chtimes(path, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
	if err := scaffoldTemplatesAndGuidelines(root, true); err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if !info.ModTime().Equal(stamp) {
			t.Fatalf("no-op replaced %s", path)
		}
	}
}
