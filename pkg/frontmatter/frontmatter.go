package frontmatter

import (
	"bytes"
	"strings"
)

// NeedsQuoting returns true if the scalar value should be quoted to be YAML-safe.
func NeedsQuoting(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	// Leading special characters often need quoting.
	switch trimmed[0] {
	case '@', '`', '#', '&', '*', '!', '|', '>', '%', '?':
		return true
	}
	// Colon patterns.
	if strings.Contains(value, ": ") || strings.HasSuffix(trimmed, ":") {
		return true
	}
	// Inline comment or tabs.
	if strings.Contains(value, " #") || strings.Contains(value, "\t") {
		return true
	}
	// Template-like.
	if strings.Contains(value, "{{") || strings.Contains(value, "}}") {
		return true
	}
	return false
}

// QuoteValue single-quotes a scalar, escaping internal single quotes.
func QuoteValue(value string) string {
	escaped := strings.ReplaceAll(value, "'", "''")
	return "'" + escaped + "'"
}

// PreprocessYAML adds quotes to top-level key scalars that look unsafe.
// It expects only the YAML frontmatter (no --- delimiters).
func PreprocessYAML(fm []byte) []byte {
	lines := bytes.Split(fm, []byte("\n"))
	var out [][]byte
	for _, line := range lines {
		// Preserve blank/comment lines.
		if len(bytes.TrimSpace(line)) == 0 || bytes.HasPrefix(bytes.TrimSpace(line), []byte("#")) {
			out = append(out, line)
			continue
		}
		trimmed := bytes.TrimLeft(line, " \t")
		indentLen := len(line) - len(trimmed)
		indent := line[:indentLen]

		// Skip lists / nested structures (heuristic).
		if bytes.HasPrefix(trimmed, []byte("- ")) || bytes.HasPrefix(trimmed, []byte("  ")) {
			out = append(out, line)
			continue
		}

		// Split on first colon.
		parts := bytes.SplitN(trimmed, []byte(":"), 2)
		if len(parts) != 2 {
			out = append(out, line)
			continue
		}
		key := strings.TrimSpace(string(parts[0]))
		val := strings.TrimSpace(string(parts[1]))

		// Already quoted or complex types; leave alone.
		if val == "" || val[0] == '"' || val[0] == '\'' || val[0] == '[' || val[0] == '{' || val[0] == '|' || val[0] == '>' {
			out = append(out, line)
			continue
		}

		if NeedsQuoting(val) {
			quoted := QuoteValue(val)
			newLine := append([]byte{}, indent...)
			newLine = append(newLine, []byte(key)...)
			newLine = append(newLine, []byte(": ")...)
			newLine = append(newLine, []byte(quoted)...)
			out = append(out, newLine)
			continue
		}

		out = append(out, line)
	}
	return bytes.Join(out, []byte("\n"))
}
