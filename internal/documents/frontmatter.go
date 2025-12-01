package documents

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
	"github.com/go-go-golems/docmgr/pkg/frontmatter"
	"github.com/go-go-golems/docmgr/pkg/models"
	"gopkg.in/yaml.v3"
)

var yamlLineRe = regexp.MustCompile(`line ([0-9]+)`)

// ReadDocumentWithFrontmatter reads a markdown file that contains YAML frontmatter.
// It returns the parsed Document metadata along with the markdown body content.
func ReadDocumentWithFrontmatter(path string) (*models.Document, string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}

	fm, body, fmStartLine, err := extractFrontmatter(raw)
	if err != nil {
		tax := docmgrctx.NewFrontmatterParseTaxonomy(path, 0, 0, "", err.Error(), err)
		return nil, "", core.WrapWithCause(err, tax)
	}

	// Auto-quote risky scalars before decode to reduce YAML failures.
	fm = frontmatter.PreprocessYAML(fm)

	lines := strings.Split(string(raw), "\n")

	var node yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(fm))
	if err := dec.Decode(&node); err != nil {
		line, col := extractLineCol(err.Error(), fmStartLine)
		snippet := buildSnippet(lines, line, col)
		problem := classifyYAMLError(err.Error())
		tax := docmgrctx.NewFrontmatterParse(path, line, col, snippet, problem, err)
		return nil, "", core.WrapWithCause(err, tax)
	}

	var doc models.Document
	if err := node.Decode(&doc); err != nil {
		// Decode errors often lack line numbers; still surface problem text.
		line, col := extractLineCol(err.Error(), fmStartLine)
		snippet := buildSnippet(lines, line, col)
		problem := classifyYAMLError(err.Error())
		tax := docmgrctx.NewFrontmatterParse(path, line, col, snippet, problem, err)
		return nil, "", core.WrapWithCause(err, tax)
	}

	return &doc, string(body), nil
}

// WriteDocumentWithFrontmatter writes the provided document metadata and body
// to the target markdown path using YAML frontmatter.
func WriteDocumentWithFrontmatter(path string, doc *models.Document, body string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".docmgr-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()

	var fmBuf bytes.Buffer
	enc := yaml.NewEncoder(&fmBuf)
	if err := enc.Encode(doc); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := enc.Close(); err != nil {
		_ = tmp.Close()
		return err
	}
	fmBytes := frontmatter.PreprocessYAML(fmBuf.Bytes())

	if _, err := tmp.WriteString("---\n"); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(fmBytes); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.WriteString("---\n\n"); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.WriteString(body); err != nil {
		_ = tmp.Close()
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), path)
}

// extractFrontmatter returns the frontmatter bytes, body bytes, and the starting line number (1-based) of the YAML block.
func extractFrontmatter(raw []byte) ([]byte, []byte, int, error) {
	lines := bytes.Split(raw, []byte("\n"))
	if len(lines) == 0 {
		return nil, nil, 0, fmt.Errorf("empty file")
	}

	start := -1
	end := -1
	for i, line := range lines {
		if i == 0 && bytes.Equal(bytes.TrimSpace(line), []byte("---")) {
			start = i
			continue
		}
		if start >= 0 && bytes.Equal(bytes.TrimSpace(line), []byte("---")) {
			end = i
			break
		}
	}

	if start != 0 || end <= start {
		return nil, nil, 0, fmt.Errorf("frontmatter delimiters '---' not found")
	}

	fmLines := lines[start+1 : end]
	bodyLines := []byte{}
	if end+1 < len(lines) {
		bodyLines = bytes.Join(lines[end+1:], []byte("\n"))
	}

	// YAML parser line numbers start at 1 for the frontmatter content (first line after initial ---).
	fmStartLine := start + 2
	return bytes.Join(fmLines, []byte("\n")), bodyLines, fmStartLine, nil
}

// SplitFrontmatter exposes frontmatter/body split to other internal consumers (e.g., fixers).
func SplitFrontmatter(raw []byte) ([]byte, []byte, int, error) {
	return extractFrontmatter(raw)
}

// extractLineCol best-effort extracts line/col and maps to absolute file line using the start line offset.
func extractLineCol(msg string, fmStartLine int) (int, int) {
	line := 0
	if m := yamlLineRe.FindStringSubmatch(msg); len(m) == 2 {
		if v, err := strconv.Atoi(m[1]); err == nil {
			line = fmStartLine + v - 1
		}
	}
	return line, 0
}

// classifyYAMLError returns a user-friendly problem summary.
func classifyYAMLError(msg string) string {
	l := strings.ToLower(msg)
	switch {
	case strings.Contains(l, "mapping values are not allowed"):
		return "mapping values are not allowed (missing quotes before ':' or bad indentation)"
	case strings.Contains(l, "did not find expected key"):
		return "did not find expected key (check colons and indentation)"
	case strings.Contains(l, "cannot unmarshal"):
		return msg
	case strings.Contains(l, "found character that cannot start any token"):
		return "invalid character (likely needs quoting or escaping)"
	default:
		return msg
	}
}

// buildSnippet returns a small line-context snippet with an optional caret.
func buildSnippet(lines []string, line, col int) string {
	if line <= 0 || line > len(lines) {
		return ""
	}
	start := line - 1
	if start < 1 {
		start = 1
	}
	end := line + 1
	if end > len(lines) {
		end = len(lines)
	}
	var b strings.Builder
	for i := start; i <= end; i++ {
		b.WriteString(fmt.Sprintf("%4d | %s\n", i, lines[i-1]))
		if i == line && col > 0 {
			b.WriteString(fmt.Sprintf("     | %s^\n", strings.Repeat(" ", col-1)))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
