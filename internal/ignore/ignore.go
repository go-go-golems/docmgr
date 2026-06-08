package ignore

import (
	"context"
	"path/filepath"
	"strings"

	gitignore "github.com/denormal/go-gitignore"
	"github.com/pkg/errors"
)

const FileName = ".docmgrignore"

var BuiltinPatterns = []string{
	".git/",
	"**/.git/**",
	"node_modules/",
	"**/node_modules/**",
	".pnpm/",
	"**/.pnpm/**",
	"dist/",
	"**/dist/**",
	"build/",
	"**/build/**",
	"coverage/",
	"**/coverage/**",
	".venv/",
	"**/.venv/**",
	"__pycache__/",
	"**/__pycache__/**",
}

type SourceKind string

const (
	SourceBuiltin    SourceKind = "builtin"
	SourceRepository SourceKind = "repository"
	SourceDocsRoot   SourceKind = "docs-root"
)

type LoadOptions struct {
	RepoRoot string
	DocsRoot string

	IncludeBuiltin bool
	IncludeNested  bool // Kept for API clarity; go-gitignore repository matching handles nested files.
}

type Matcher struct {
	repoRoot string
	docsRoot string
	sources  []sourceMatcher
}

type sourceMatcher struct {
	kind SourceKind
	base string
	name string
	gi   gitignore.GitIgnore
}

type Decision struct {
	Path          string
	IsDir         bool
	Ignored       bool
	Matched       bool
	SourceKind    SourceKind
	SourceName    string
	Pattern       string
	PatternFile   string
	PatternLine   int
	PatternColumn int
	Trace         []TraceStep
}

type TraceStep struct {
	SourceKind SourceKind
	SourceName string
	Matched    bool
	Ignored    bool
	Pattern    string
	File       string
	Line       int
	Column     int
}

func Load(ctx context.Context, opts LoadOptions) (*Matcher, error) {
	if ctx == nil {
		return nil, errors.New("nil context")
	}
	docsRoot, err := cleanAbs(opts.DocsRoot)
	if err != nil {
		return nil, errors.Wrap(err, "resolve docs root")
	}
	repoRoot := ""
	if strings.TrimSpace(opts.RepoRoot) != "" {
		repoRoot, err = cleanAbs(opts.RepoRoot)
		if err != nil {
			return nil, errors.Wrap(err, "resolve repo root")
		}
	}

	m := &Matcher{repoRoot: repoRoot, docsRoot: docsRoot}

	if opts.IncludeBuiltin {
		builtin := gitignore.New(strings.NewReader(strings.Join(BuiltinPatterns, "\n")+"\n"), docsRoot, nil)
		m.sources = append(m.sources, sourceMatcher{kind: SourceBuiltin, base: docsRoot, name: "built-in docmgr ignores", gi: builtin})
	}

	if repoRoot != "" {
		gi, err := gitignore.NewRepositoryWithFile(repoRoot, FileName)
		if err != nil {
			return nil, errors.Wrap(err, "load repository .docmgrignore hierarchy")
		}
		m.sources = append(m.sources, sourceMatcher{kind: SourceRepository, base: repoRoot, name: filepath.Join(repoRoot, FileName), gi: gi})
	}

	// If the docs root is not below the repository root, or no repository root was
	// available, load a docs-root hierarchy explicitly. When docsRoot is under
	// repoRoot, the repository matcher already observes nested .docmgrignore files
	// under docsRoot.
	if repoRoot == "" || !isWithin(repoRoot, docsRoot) {
		gi, err := gitignore.NewRepositoryWithFile(docsRoot, FileName)
		if err != nil {
			return nil, errors.Wrap(err, "load docs-root .docmgrignore hierarchy")
		}
		m.sources = append(m.sources, sourceMatcher{kind: SourceDocsRoot, base: docsRoot, name: filepath.Join(docsRoot, FileName), gi: gi})
	}

	return m, nil
}

func (m *Matcher) Match(path string, isDir bool) Decision {
	if m == nil {
		return Decision{Path: path, IsDir: isDir}
	}
	abs := m.resolvePath(path)
	decision := Decision{Path: filepath.ToSlash(abs), IsDir: isDir}
	for _, src := range m.sources {
		if src.gi == nil || !isWithin(src.base, abs) || filepath.Clean(abs) == filepath.Clean(src.base) {
			decision.Trace = append(decision.Trace, TraceStep{SourceKind: src.kind, SourceName: src.name})
			continue
		}
		match := src.gi.Absolute(abs, isDir)
		if match == nil {
			decision.Trace = append(decision.Trace, TraceStep{SourceKind: src.kind, SourceName: src.name})
			continue
		}
		pos := match.Position()
		step := TraceStep{
			SourceKind: src.kind,
			SourceName: src.name,
			Matched:    true,
			Ignored:    match.Ignore(),
			Pattern:    match.String(),
			File:       pos.File,
			Line:       pos.Line,
			Column:     pos.Column,
		}
		decision.Trace = append(decision.Trace, step)
		decision.Matched = true
		decision.Ignored = match.Ignore()
		decision.SourceKind = src.kind
		decision.SourceName = src.name
		decision.Pattern = match.String()
		decision.PatternFile = pos.File
		decision.PatternLine = pos.Line
		decision.PatternColumn = pos.Column
	}
	return decision
}

func (m *Matcher) Ignore(path string, isDir bool) bool {
	return m.Match(path, isDir).Ignored
}

func (m *Matcher) DocsRoot() string { return m.docsRoot }
func (m *Matcher) RepoRoot() string { return m.repoRoot }

func (m *Matcher) resolvePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err == nil {
			return filepath.Clean(abs)
		}
		return filepath.Clean(path)
	}
	cleanRel := filepath.Clean(path)
	if m.repoRoot != "" && m.docsRoot != "" {
		docsBase := filepath.Base(m.docsRoot)
		if cleanRel == docsBase || strings.HasPrefix(cleanRel, docsBase+string(filepath.Separator)) {
			return filepath.Clean(filepath.Join(m.repoRoot, cleanRel))
		}
	}
	if m.docsRoot != "" {
		candidate := filepath.Join(m.docsRoot, cleanRel)
		if isWithin(m.docsRoot, candidate) {
			return filepath.Clean(candidate)
		}
	}
	abs, err := filepath.Abs(cleanRel)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(abs)
}

func cleanAbs(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("empty path")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func isWithin(base, path string) bool {
	base = filepath.Clean(base)
	path = filepath.Clean(path)
	if base == "" || path == "" {
		return false
	}
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
