package commands

import (
	"fmt"
	"os"
	"path/filepath"
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

	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(changelogPath), 0o755); err != nil {
			return "", fmt.Errorf("failed to create changelog directory: %w", err)
		}
		if err := os.WriteFile(changelogPath, []byte("# Changelog\n\n"), 0o644); err != nil {
			return "", fmt.Errorf("failed to create changelog.md: %w", err)
		}
	}

	today := time.Now().Format("2006-01-02")
	heading := "## " + today
	if strings.TrimSpace(title) != "" {
		heading += " - " + title
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(heading)
	sb.WriteString("\n\n")
	sb.WriteString(entry)
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

	fp, err := os.OpenFile(changelogPath, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return "", fmt.Errorf("failed to open changelog.md: %w", err)
	}
	defer func() { _ = fp.Close() }()
	if _, err := fp.WriteString(sb.String()); err != nil {
		return "", fmt.Errorf("failed to write changelog entry: %w", err)
	}
	return today, nil
}
