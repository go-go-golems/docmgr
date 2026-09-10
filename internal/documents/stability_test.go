package documents

import (
	"github.com/go-go-golems/docmgr/pkg/models"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func mustTest(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func TestBodyStability(t *testing.T) {
	for _, body := range []string{"", "# Heading", "\n# Heading\n", "\n\n```go\nx  \n```\n\n", "\r\n# Heading\r\nline  \r\n"} {
		t.Run(body, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "doc.md")
			mustTest(t, WriteDocumentWithFrontmatter(path, &models.Document{Title: "Title"}, body, true))
			mustTest(t, os.Chmod(path, 0640))
			stamp := time.Unix(1000, 0)
			mustTest(t, os.Chtimes(path, stamp, stamp))
			original, err := os.ReadFile(path)
			mustTest(t, err)
			for i := 0; i < 10; i++ {
				doc, got, err := ReadDocumentWithFrontmatter(path)
				mustTest(t, err)
				if got != body {
					t.Fatalf("body changed: %q != %q", got, body)
				}
				mustTest(t, WriteDocumentWithFrontmatter(path, doc, got, true))
				current, err := os.ReadFile(path)
				mustTest(t, err)
				if string(original) != string(current) {
					t.Fatal("round trip changed bytes")
				}
				info, err := os.Stat(path)
				mustTest(t, err)
				if !stamp.Equal(info.ModTime()) || info.Mode().Perm() != 0640 {
					t.Fatal("no-op changed mtime/mode")
				}
			}
		})
	}
}
func TestUnknownMetadataAndCRLF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.md")
	mustTest(t, os.WriteFile(path, []byte("---\r\nTitle: old\r\nCustom: {nested: [one, two]}\r\n---\r\n\r\nBody  \r\n"), 0644))
	doc, body, err := ReadDocumentWithFrontmatter(path)
	mustTest(t, err)
	if body != "\r\nBody  \r\n" {
		t.Fatalf("body: %q", body)
	}
	doc.Title = "new"
	mustTest(t, WriteDocumentWithFrontmatter(path, doc, body, true))
	next, got, err := ReadDocumentWithFrontmatter(path)
	mustTest(t, err)
	if body != got {
		t.Fatal("body changed")
	}
	var custom map[string][]string
	node := next.Extra["Custom"]
	mustTest(t, node.Decode(&custom))
	if !reflect.DeepEqual(custom["nested"], []string{"one", "two"}) {
		t.Fatalf("lost unknown key: %v", custom)
	}
}
func TestCreationAndPermissionPreservation(t *testing.T) {
	if CreationBody("# Body") != "\n# Body" || CreationBody("\n\nBody") != "\n\nBody" {
		t.Fatal("creation separator")
	}
	path := filepath.Join(t.TempDir(), "doc.md")
	mustTest(t, os.WriteFile(path, []byte("old"), 0640))
	changed, err := WriteFileIfChanged(path, []byte("new"))
	mustTest(t, err)
	if !changed {
		t.Fatal("expected change")
	}
	info, err := os.Stat(path)
	mustTest(t, err)
	if info.Mode().Perm() != 0640 {
		t.Fatal("mode changed")
	}
	link := filepath.Join(t.TempDir(), "link")
	mustTest(t, os.Symlink(path, link))
	if _, err = WriteFileIfChanged(link, []byte("bad")); err == nil {
		t.Fatal("accepted symlink")
	}
}
