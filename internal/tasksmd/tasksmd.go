package tasksmd

import (
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Item struct {
	ID      int    `json:"id"`
	Checked bool   `json:"checked"`
	Text    string `json:"text"`
	// StableID is an optional short identifier stored as an invisible HTML
	// comment marker at the end of the task line (e.g. "<!-- t:ab12 -->").
	// Tasks without markers have an empty StableID and are addressed by
	// position (ID) only.
	StableID string `json:"stableId,omitempty"`
}

type Section struct {
	Title string `json:"title"`
	Items []Item `json:"items"`
}

type Parsed struct {
	Sections []Section `json:"sections"`
	Total    int       `json:"total"`
	Done     int       `json:"done"`
}

type parsedTaskLine struct {
	ID        int
	LineIndex int
	Checked   bool
	Text      string
	Section   string
	StableID  string
}

var (
	headingRe = regexp.MustCompile(`^\s{0,3}(#{1,6})\s+(.+?)\s*$`)
	taskRe    = regexp.MustCompile(`^\s{0,3}([-*])\s+\[(?i:[ x])\]\s+(.+?)\s*$`)
	// stableIDMarkerRe matches a trailing invisible task-ID marker, e.g.
	// "Fix resolver <!-- t:a3f2 -->".
	stableIDMarkerRe = regexp.MustCompile(`\s*<!--\s*t:([a-z0-9]{2,12})\s*-->\s*$`)
)

const stableIDAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// ExtractStableID splits a task text into its stable ID (if a trailing
// "<!-- t:xxxx -->" marker is present) and the clean text without the marker.
func ExtractStableID(text string) (string, string) {
	m := stableIDMarkerRe.FindStringSubmatchIndex(text)
	if m == nil {
		return "", strings.TrimSpace(text)
	}
	id := text[m[2]:m[3]]
	clean := strings.TrimSpace(text[:m[0]])
	return id, clean
}

// FormatStableIDMarker renders the invisible marker for a stable task ID.
func FormatStableIDMarker(id string) string {
	return "<!-- t:" + id + " -->"
}

// NewStableID generates a short random task ID (base36, 4+ chars) that does
// not collide with the provided existing IDs.
func NewStableID(existing map[string]struct{}) string {
	for length := 4; length <= 12; length++ {
		for attempt := 0; attempt < 32; attempt++ {
			id, err := randomStableID(length)
			if err != nil {
				continue
			}
			if _, ok := existing[id]; !ok {
				return id
			}
		}
	}
	// Practically unreachable: fall back to a deterministic suffix scan.
	base, _ := randomStableID(8)
	for i := 0; ; i++ {
		id := fmt.Sprintf("%s%d", base, i)
		if _, ok := existing[id]; !ok {
			return id
		}
	}
}

func randomStableID(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := cryptorand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, length)
	for i, b := range buf {
		out[i] = stableIDAlphabet[int(b)%len(stableIDAlphabet)]
	}
	return string(out), nil
}

// CollectStableIDs returns the set of stable IDs already present in the file.
func CollectStableIDs(lines []string) map[string]struct{} {
	out := map[string]struct{}{}
	_, tasks := Parse(lines)
	for _, t := range tasks {
		if t.StableID != "" {
			out[t.StableID] = struct{}{}
		}
	}
	return out
}

func Parse(lines []string) (Parsed, map[int]parsedTaskLine) {
	sections := []Section{}
	taskByID := map[int]parsedTaskLine{}

	sectionTitle := "Tasks"
	secIdx := -1
	ensureSection := func(title string) {
		title = strings.TrimSpace(title)
		if title == "" {
			title = "Tasks"
		}
		if secIdx >= 0 && sections[secIdx].Title == title {
			return
		}
		sections = append(sections, Section{Title: title})
		secIdx = len(sections) - 1
	}
	total := 0
	done := 0

	for i, raw := range lines {
		if m := headingRe.FindStringSubmatch(raw); len(m) == 3 {
			if len(m[1]) == 1 {
				// Treat H1 as the document title, not a tasks section.
				continue
			}
			sectionTitle = strings.TrimSpace(m[2])
			ensureSection(sectionTitle)
			continue
		}

		m := taskRe.FindStringSubmatch(raw)
		if len(m) != 3 {
			continue
		}
		if secIdx == -1 {
			ensureSection(sectionTitle)
		}

		trimmed := strings.TrimSpace(raw)
		checked := strings.HasPrefix(strings.ToLower(trimmed), "- [x]") || strings.HasPrefix(strings.ToLower(trimmed), "* [x]")

		stableID, text := ExtractStableID(strings.TrimSpace(m[2]))
		total++
		if checked {
			done++
		}

		item := Item{ID: total, Checked: checked, Text: text, StableID: stableID}
		sections[secIdx].Items = append(sections[secIdx].Items, item)
		taskByID[item.ID] = parsedTaskLine{
			ID:        item.ID,
			LineIndex: i,
			Checked:   checked,
			Text:      text,
			Section:   sections[secIdx].Title,
			StableID:  stableID,
		}
	}

	return Parsed{Sections: sections, Total: total, Done: done}, taskByID
}

func ReadFile(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := strings.ReplaceAll(string(b), "\r\n", "\n")
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return []string{}, nil
	}
	return strings.Split(s, "\n"), nil
}

func WriteFile(path string, lines []string) error {
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

func ToggleChecked(lines []string, ids []int, checked bool) ([]string, error) {
	_, tasks := Parse(lines)

	idSet := map[int]struct{}{}
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		idSet[id] = struct{}{}
	}
	if len(idSet) == 0 {
		return nil, errors.New("no valid ids")
	}

	out := append([]string{}, lines...)
	for id := range idSet {
		t, ok := tasks[id]
		if !ok {
			return nil, fmt.Errorf("task id not found: %d", id)
		}
		out[t.LineIndex] = setTaskLineChecked(out[t.LineIndex], checked)
	}
	return out, nil
}

func AppendTask(lines []string, section string, text string) ([]string, error) {
	section = strings.TrimSpace(section)
	if section == "" {
		section = "TODO"
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("missing task text")
	}

	// Find a heading matching the section title.
	headingLine := -1
	nextHeading := -1
	for i, raw := range lines {
		m := headingRe.FindStringSubmatch(raw)
		if len(m) != 3 {
			continue
		}
		title := strings.TrimSpace(m[2])
		if headingLine == -1 && strings.EqualFold(title, section) {
			headingLine = i
			continue
		}
		if headingLine != -1 && i > headingLine {
			nextHeading = i
			break
		}
	}

	newTaskLine := "- [ ] " + text + " " + FormatStableIDMarker(NewStableID(CollectStableIDs(lines)))

	out := append([]string{}, lines...)
	switch {
	case headingLine == -1:
		// Append a new section at end.
		if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		out = append(out, "## "+section, "", newTaskLine)
		return out, nil
	case nextHeading == -1:
		// Append to end of file (after existing section content).
		if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		out = append(out, newTaskLine)
		return out, nil
	default:
		// Insert before next heading; prefer after last task line in the section.
		insertAt := nextHeading
		for i := headingLine + 1; i < nextHeading; i++ {
			if taskRe.MatchString(out[i]) {
				insertAt = i + 1
			}
		}
		out = append(out[:insertAt], append([]string{newTaskLine}, out[insertAt:]...)...)
		return out, nil
	}
}

func setTaskLineChecked(raw string, checked bool) string {
	trimmed := strings.TrimLeft(raw, " \t")
	prefixLen := len(raw) - len(trimmed)
	prefix := raw[:prefixLen]

	bullet := "-"
	rest := trimmed
	if strings.HasPrefix(rest, "*") {
		bullet = "*"
		rest = strings.TrimSpace(rest[1:])
	} else if strings.HasPrefix(rest, "-") {
		bullet = "-"
		rest = strings.TrimSpace(rest[1:])
	}

	// Expect rest starts with [x] or [ ].
	open := strings.Index(rest, "[")
	closeIdx := strings.Index(rest, "]")
	if open == -1 || closeIdx == -1 || closeIdx <= open {
		return raw
	}

	mark := " "
	if checked {
		mark = "x"
	}
	rest = rest[:open+1] + mark + rest[open+2:]
	return prefix + bullet + " " + rest
}
