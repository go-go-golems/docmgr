package commands

import (
	"fmt"
	"github.com/go-go-golems/docmgr/internal/documents"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ChangelogEntry is one dated section of a ticket's changelog.md
// ("## YYYY-MM-DD[ - Title]" followed by free-form markdown).
type ChangelogEntry struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Heading string `json:"heading"`
	Body    string `json:"body"`
}

var changelogHeadingRe = regexp.MustCompile(`^##\s+(.*)$`)
var changelogDateRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})\s*(?:[-–—]\s*)?(.*)$`)

// ParseChangelogEntries splits changelog.md content into its "## " sections,
// in file order (appends put the newest entry last). Content before the first
// "## " heading (e.g. the "# Changelog" header) is skipped.
func ParseChangelogEntries(content string) []ChangelogEntry {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	entries := []ChangelogEntry{}
	var cur *ChangelogEntry
	var body []string

	flush := func() {
		if cur == nil {
			return
		}
		cur.Body = strings.Trim(strings.Join(body, "\n"), "\n")
		entries = append(entries, *cur)
		cur = nil
		body = nil
	}

	for _, line := range lines {
		if m := changelogHeadingRe.FindStringSubmatch(line); m != nil {
			flush()
			heading := strings.TrimSpace(m[1])
			e := ChangelogEntry{Heading: heading}
			if dm := changelogDateRe.FindStringSubmatch(heading); dm != nil {
				e.Date = dm[1]
				e.Title = strings.TrimSpace(dm[2])
			} else {
				e.Title = heading
			}
			cur = &e
			continue
		}
		if cur != nil {
			body = append(body, line)
		}
	}
	flush()
	return entries
}

// AppendChangelogEntry appends a dated entry (with optional title and
// optional related-files map path->note) to changelog.md, creating the file
// with a "# Changelog" header when missing. It returns the entry date
// (YYYY-MM-DD). This is the shared write primitive behind
// 'docmgr changelog update' and the HTTP API's POST /tickets/changelog.
func AppendChangelogEntry(changelogPath string, title string, entry string, files map[string]string) (string, error) {
	if strings.TrimSpace(entry) == "" {
		return "", fmt.Errorf("entry must not be empty")
	}

	old, err := os.ReadFile(changelogPath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	today := time.Now().Format("2006-01-02")
	next := BuildChangelogEntry(string(old), today, title, entry, files)
	_, err = documents.WriteFileIfChanged(changelogPath, []byte(next))
	return today, err
}

// BuildChangelogEntry is a pure projection with explicit date and canonical boundaries.
// Internal Markdown whitespace is preserved; only boundary newline bytes are trimmed.
func BuildChangelogEntry(old, date, title, entry string, files map[string]string) string {
	if old == "" {
		old = "# Changelog"
	}
	heading := "## " + date
	if strings.TrimSpace(title) != "" {
		heading += " - " + title
	}

	var sb strings.Builder
	sb.WriteString(strings.TrimRight(old, "\r\n"))
	sb.WriteString("\n\n")
	sb.WriteString(heading)
	sb.WriteString("\n\n")
	sb.WriteString(strings.Trim(entry, "\r\n"))
	sb.WriteString("\n\n")
	if len(files) > 0 {
		sb.WriteString("### Related Files\n\n")
		var names []string
		for f := range files {
			names = append(names, f)
		}
		sort.Strings(names)
		for _, f := range names {
			note := strings.TrimSpace(files[f])
			if note != "" {
				sb.WriteString("- " + f + " — " + note + "\n")
			} else {
				sb.WriteString("- " + f + "\n")
			}
		}
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n") + "\n"
}
