package commands

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgr"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
	"github.com/go-go-golems/docmgr/pkg/frontmatter"
	"github.com/go-go-golems/docmgr/pkg/models"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// ValidateFrontmatterCommand validates YAML frontmatter for a document.
type ValidateFrontmatterCommand struct {
	*cmds.CommandDescription
}

// ValidateFrontmatterSettings holds parameters for validation.
type ValidateFrontmatterSettings struct {
	Doc          string `glazed.parameter:"doc"`
	Root         string `glazed.parameter:"root"`
	SuggestFixes bool   `glazed.parameter:"suggest-fixes"`
	AutoFix      bool   `glazed.parameter:"auto-fix"`
}

func NewValidateFrontmatterCommand() (*ValidateFrontmatterCommand, error) {
	return &ValidateFrontmatterCommand{
		CommandDescription: cmds.NewCommandDescription(
			"frontmatter",
			cmds.WithShort("Validate YAML frontmatter for a document"),
			cmds.WithLong(`Validates YAML frontmatter for a single markdown file.

If parsing fails, the command surfaces a diagnostics taxonomy (line/column/snippet when available).
Use this before running doctor when iterating on frontmatter edits.

Examples:
  docmgr validate frontmatter --doc ttmp/2025/11/29/DOC-1234/index.md
  docmgr validate frontmatter --doc ttmp/.../index.md --suggest-fixes
  docmgr validate frontmatter --doc ttmp/.../index.md --auto-fix
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"doc",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to the markdown document to validate"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Docs root (used when doc is relative)"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"suggest-fixes",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Show suggested fixes when validation fails"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"auto-fix",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Attempt to auto-fix common YAML issues (creates .bak backup)"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *ValidateFrontmatterCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ValidateFrontmatterSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	docPath := settings.Doc
	if !filepath.IsAbs(docPath) {
		root := workspace.ResolveRoot(settings.Root)
		docPath = filepath.Join(root, docPath)
	}

	renderer := docmgr.NewRenderer()
	ctx = docmgr.ContextWithRenderer(ctx, renderer)

	doc, err := validateFrontmatterFile(ctx, docPath, settings.SuggestFixes, settings.AutoFix)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("doc", docPath),
		types.MRP("title", doc.Title),
		types.MRP("ticket", doc.Ticket),
		types.MRP("docType", doc.DocType),
		types.MRP("status", "ok"),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return fmt.Errorf("failed to emit validation result: %w", err)
	}
	return nil
}

func (c *ValidateFrontmatterCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &ValidateFrontmatterSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	docPath := settings.Doc
	if !filepath.IsAbs(docPath) {
		root := workspace.ResolveRoot(settings.Root)
		docPath = filepath.Join(root, docPath)
	}

	renderer := docmgr.NewRenderer()
	ctx = docmgr.ContextWithRenderer(ctx, renderer)

	doc, err := validateFrontmatterFile(ctx, docPath, settings.SuggestFixes, settings.AutoFix)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Frontmatter OK: %s (Ticket=%s DocType=%s)\n", docPath, doc.Ticket, doc.DocType)
	return nil
}

func validateFrontmatterFile(ctx context.Context, path string, suggestFixes bool, autoFix bool) (*models.Document, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	doc, _, err := documents.ReadDocumentWithFrontmatter(path)
	if err == nil {
		return doc, nil
	}

	if tax, ok := core.AsTaxonomy(err); ok && tax != nil {
		applied := false
		if suggestFixes || autoFix {
			applied = tryAttachFixes(tax, raw, path, autoFix)
			if applied && autoFix {
				// Re-parse after auto-fix.
				fixedDoc, _, parseErr := documents.ReadDocumentWithFrontmatter(path)
				if parseErr == nil {
					docmgr.RenderTaxonomy(ctx, tax)
					return fixedDoc, nil
				}
				err = parseErr
			}
		}
		docmgr.RenderTaxonomy(ctx, tax)
		if applied && autoFix {
			return nil, fmt.Errorf("auto-fix applied but re-parse failed: %w", err)
		}
		return nil, err
	}

	return nil, err
}

// tryAttachFixes computes fixes and optionally applies auto-fix. Returns true if a fix was applied.
func tryAttachFixes(tax *core.Taxonomy, raw []byte, path string, autoFix bool) bool {
	fixes, fixedContent, err := generateFixes(raw)
	if err != nil {
		return false
	}
	if ctxPayload, ok := tax.Context.(*docmgrctx.FrontmatterParseContext); ok {
		ctxPayload.Fixes = fixes
	}
	if autoFix && fixedContent != nil {
		if err := applyAutoFix(path, raw, fixedContent); err != nil {
			return false
		}
	}
	return autoFix && fixedContent != nil
}

// generateFixes returns suggested fix lines and a fully rewritten file content with quoted scalars.
func generateFixes(raw []byte) ([]string, []byte, error) {
	fm, body, _, err := documents.SplitFrontmatter(raw)
	if err != nil {
		// Fallback: try to normalize delimiters heuristically.
		if fixed, fixErr := normalizeDelimiters(raw); fixErr == nil {
			fixes := []string{"Normalize frontmatter delimiters (add missing closing ---)"}
			return fixes, fixed, nil
		}
		return nil, nil, err
	}
	cleaned := scrubStrayDelimiters(fm)
	fmLines := bytes.Split(cleaned, []byte("\n"))
	peeledBody := peelTrailingNonKeyLines(&fmLines)
	cleaned = bytes.Join(fmLines, []byte("\n"))
	quoted := frontmatter.PreprocessYAML(cleaned)
	if bytes.Equal(quoted, fm) {
		return nil, nil, fmt.Errorf("no auto-fix available")
	}
	fixes := []string{"Quote unsafe scalars (colons, hashes, special leading chars)"}
	if !bytes.Equal(cleaned, fm) {
		fixes = append(fixes, "Remove stray delimiter lines inside frontmatter")
	}
	if len(peeledBody) > 0 {
		fixes = append(fixes, "Move non key/value lines out of frontmatter")
	}
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(quoted)
	buf.WriteString("\n---\n")
	if len(peeledBody) > 0 {
		if len(body) > 0 {
			buf.Write(append(peeledBody, '\n'))
		} else {
			buf.Write(peeledBody)
			buf.WriteByte('\n')
		}
	}
	buf.Write(body)
	return fixes, buf.Bytes(), nil
}

func applyAutoFix(path string, original []byte, fixed []byte) error {
	if err := backup(path, original); err != nil {
		return err
	}
	return os.WriteFile(path, fixed, 0o644)
}

func backup(path string, data []byte) error {
	bak := path + ".bak"
	if err := os.WriteFile(bak, data, fs.FileMode(0o644)); err != nil {
		return fmt.Errorf("failed to write backup %s: %w", bak, err)
	}
	return nil
}

// normalizeDelimiters attempts to fix missing/invalid closing delimiters by rewriting the file.
func normalizeDelimiters(raw []byte) ([]byte, error) {
	lines := bytes.Split(raw, []byte("\n"))
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty content")
	}
	start := -1
	end := -1
	for i, line := range lines {
		if bytes.Equal(bytes.TrimSpace(line), []byte("---")) && start == -1 {
			start = i
			continue
		}
		trim := bytes.TrimSpace(line)
		if start >= 0 && bytes.HasPrefix(trim, []byte("---")) {
			end = i
			break
		}
	}
	if start == -1 {
		// Treat entire file as frontmatter payload; we'll wrap it.
		start = -1
		end = len(lines)
	}
	if end == -1 {
		// no closing delimiter; assume frontmatter until first blank or end.
		for i := start + 1; i < len(lines); i++ {
			if len(bytes.TrimSpace(lines[i])) == 0 {
				end = i
				break
			}
		}
		if end == -1 {
			end = len(lines)
		}
	}
	fmLines := lines[start+1 : end]
	bodyLines := []byte{}
	if end+1 < len(lines) {
		bodyLines = bytes.Join(lines[end+1:], []byte("\n"))
	}
	// If closing delimiter missing and we captured trailing non key/value lines, peel them into body.
	for len(fmLines) > 0 {
		last := fmLines[len(fmLines)-1]
		if bytes.Contains(last, []byte(":")) {
			break
		}
		bodyLines = append(bodyLines, '\n')
		bodyLines = append(bodyLines, last...)
		fmLines = fmLines[:len(fmLines)-1]
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(bytes.Join(fmLines, []byte("\n")))
	buf.WriteString("\n---\n")
	buf.Write(bodyLines)
	return buf.Bytes(), nil
}

// scrubStrayDelimiters removes lines that look like delimiter variants inside frontmatter content.
func scrubStrayDelimiters(fm []byte) []byte {
	lines := bytes.Split(fm, []byte("\n"))
	out := make([][]byte, 0, len(lines))
	for _, l := range lines {
		trim := bytes.TrimSpace(l)
		if bytes.HasPrefix(trim, []byte("---")) {
			// Skip stray delimiter-like lines.
			continue
		}
		out = append(out, l)
	}
	return bytes.Join(out, []byte("\n"))
}

// peelTrailingNonKeyLines moves trailing lines that do not look like "Key: value" out of the fm.
func peelTrailingNonKeyLines(lines *[][]byte) []byte {
	l := *lines
	var peeled [][]byte
	for len(l) > 0 {
		last := l[len(l)-1]
		if bytes.Contains(last, []byte(":")) {
			break
		}
		peeled = append([][]byte{last}, peeled...)
		l = l[:len(l)-1]
	}
	*lines = l
	if len(peeled) == 0 {
		return nil
	}
	return bytes.Join(peeled, []byte("\n"))
}

var _ cmds.GlazeCommand = &ValidateFrontmatterCommand{}
var _ cmds.BareCommand = &ValidateFrontmatterCommand{}
