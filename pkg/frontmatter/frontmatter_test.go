package frontmatter

import "testing"

func TestNeedsQuoting(t *testing.T) {
	cases := []struct {
		val  string
		want bool
	}{
		{"simple", false},
		{"with: colon", true},
		{"ends:", true},
		{"@start", true},
		{"hash # inline", true},
		{"tab\tchar", true},
		{"template {{x}}", true},
	}
	for _, c := range cases {
		if got := NeedsQuoting(c.val); got != c.want {
			t.Fatalf("NeedsQuoting(%q)=%v, want %v", c.val, got, c.want)
		}
	}
}

func TestPreprocessYAMLQuotesScalars(t *testing.T) {
	input := []byte("Title: Hello\nSummary: needs: quoting\nList:\n  - item\n")
	out := PreprocessYAML(input)
	s := string(out)
	if !containsLine(s, "Summary: 'needs: quoting'") {
		t.Fatalf("expected Summary to be quoted, got:\n%s", s)
	}
	if !containsLine(s, "Title: Hello") {
		t.Fatalf("expected Title unchanged, got:\n%s", s)
	}
}

func containsLine(s, needle string) bool {
	for _, line := range splitLines(s) {
		if line == needle {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return out
}
