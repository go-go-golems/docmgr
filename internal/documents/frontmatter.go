package documents

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
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

	return readDocumentWithFrontmatterBytes(path, raw)
}

// ReadDocumentWithFrontmatterFS is like ReadDocumentWithFrontmatter, but reads from an fs.FS.
// The path must be valid for the provided filesystem (typically a slash-separated relative path).
func ReadDocumentWithFrontmatterFS(fsys fs.FS, path string) (*models.Document, string, error) {
	raw, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, "", err
	}

	return readDocumentWithFrontmatterBytes(path, raw)
}

func readDocumentWithFrontmatterBytes(path string, raw []byte) (*models.Document, string, error) {
	fm, body, fmStartLine, err := extractFrontmatter(raw)
	if err != nil {
		tax := docmgrctx.NewFrontmatterParseTaxonomy(path, 0, 0, "", err.Error(), err)
		return nil, "", core.WrapWithCause(err, tax)
	}

	// Preserve valid YAML semantics (including anchors). Repair risky scalars
	// only when parsing fails; preprocessing valid YAML would turn aliases into text.

	lines := strings.Split(string(raw), "\n")

	var node yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(fm))
	decodeErr := dec.Decode(&node)
	if decodeErr != nil {
		dec = yaml.NewDecoder(bytes.NewReader(frontmatter.PreprocessYAML(fm)))
		decodeErr = dec.Decode(&node)
	}
	if err := decodeErr; err != nil {
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

// WriteDocumentWithFrontmatter preserves body bytes exactly. Framing uses LF;
// creation callers supply their initial separator with CreationBody.
func WriteDocumentWithFrontmatter(path string, doc *models.Document, body string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}

	data, err := SerializeDocument(doc, body)
	if err != nil {
		return err
	}
	_, err = WriteFileIfChanged(path, data)
	return err
}

// CreationBody supplies one initial blank separator, without trimming authored text.
func CreationBody(body string) string {
	if body == "" || strings.HasPrefix(body, "\n") || strings.HasPrefix(body, "\r\n") {
		return body
	}
	return "\n" + body
}

// SerializeDocument canonicalizes typed YAML metadata, not Markdown body bytes.
// Unknown YAML fields survive; YAML comments and original scalar styling do not.
func SerializeDocument(doc *models.Document, body string) ([]byte, error) {
	if doc == nil {
		return nil, fmt.Errorf("nil document")
	}
	budget := 10000
	copyDoc := *doc
	copyDoc.Extra = make(map[string]yaml.Node, len(doc.Extra))
	for key, node := range doc.Extra {
		expanded, err := expandMetadataNode(&node, 0, &budget)
		if err != nil {
			return nil, fmt.Errorf("metadata %s: %w", key, err)
		}
		copyDoc.Extra[key] = *expanded
	}
	var fmBuf bytes.Buffer
	enc := yaml.NewEncoder(&fmBuf)
	if err := enc.Encode(&copyDoc); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	fmBytes := frontmatter.PreprocessYAML(fmBuf.Bytes())
	return []byte("---\n" + string(fmBytes) + "---\n" + body), nil
}

// Expand aliases because their anchor may belong to a known field that is
// re-encoded without the original YAML node. Recursive aliases fail before writing.
func expandMetadataNode(node *yaml.Node, depth int, budget *int) (*yaml.Node, error) {
	if node == nil || depth > 64 || *budget <= 0 {
		return nil, fmt.Errorf("invalid alias or metadata expansion exceeds depth 64 / 10000 nodes")
	}
	*budget -= 1
	if node.Kind == yaml.AliasNode {
		return expandMetadataNode(node.Alias, depth+1, budget)
	}
	copyNode := *node
	copyNode.Anchor = ""
	// Comments on unknown metadata nodes would be reinterpreted as scalar
	// content by PreprocessYAML; canonical serialization deliberately drops
	// YAML comments while preserving values and structure.
	copyNode.HeadComment = ""
	copyNode.LineComment = ""
	copyNode.FootComment = ""
	copyNode.Content = nil
	for _, child := range node.Content {
		expanded, err := expandMetadataNode(child, depth+1, budget)
		if err != nil {
			return nil, err
		}
		copyNode.Content = append(copyNode.Content, expanded)
	}
	return &copyNode, nil
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
		fmt.Fprintf(&b, "%4d | %s\n", i, lines[i-1])
		if i == line && col > 0 {
			fmt.Fprintf(&b, "     | %s^\n", strings.Repeat(" ", col-1))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
