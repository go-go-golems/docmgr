// Run from repository root: go run -tags sqlite_fts5 ./ttmp/.../scripts/01-reproduce-writes.go
// Research-only reproducer: creates and deletes its own temporary directory.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/docmgr/pkg/models"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	dir, err := os.MkdirTemp("", "docmgr-friction-")
	must(err)
	defer func() { must(os.RemoveAll(dir)) }()
	path := filepath.Join(dir, "doc.md")
	doc := &models.Document{Title: "Fixture", Ticket: "FIXTURE", Status: "active", DocType: "reference", LastUpdated: time.Unix(0, 0).UTC()}
	must(documents.WriteDocumentWithFrontmatter(path, doc, "# Body\n\nKeep  two spaces.\n", true))
	gaps := []int{}
	for i := 0; i < 4; i++ {
		metadata, body, err := documents.ReadDocumentWithFrontmatter(path)
		must(err)
		gaps = append(gaps, len(body)-len(bytes.TrimLeft([]byte(body), "\n")))
		must(documents.WriteDocumentWithFrontmatter(path, metadata, body, true))
	}
	log := filepath.Join(dir, "changelog.md")
	_, err = commands.AppendChangelogEntry(log, "", "Fixture milestone", nil)
	must(err)
	raw, err := os.ReadFile(log)
	must(err)
	must(json.NewEncoder(os.Stdout).Encode(map[string]any{"body_leading_newlines_after_round_trips": gaps, "changelog_trailing_newlines": len(raw) - len(bytes.TrimRight(raw, "\n")), "changelog": string(raw)}))
}
