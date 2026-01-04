package searchsvc

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/docmgr/internal/workspace"
)

type FileSuggestion struct {
	File   string
	Source string
	Reason string
}

type SuggestFilesQuery struct {
	Ticket string
	Topics []string
	Query  string
}

func SuggestFiles(ctx context.Context, ws *workspace.Workspace, q SuggestFilesQuery) ([]FileSuggestion, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil context")
	}
	if ws == nil {
		return nil, fmt.Errorf("nil workspace")
	}

	// Find ticket directory if specified (for git/ripgrep heuristics).
	ticketDir := ws.Context().Root
	if strings.TrimSpace(q.Ticket) != "" {
		idxRes, err := ws.QueryDocs(ctx, workspace.DocQuery{
			Scope:   workspace.Scope{Kind: workspace.ScopeTicket, TicketID: strings.TrimSpace(q.Ticket)},
			Filters: workspace.DocFilters{DocType: "index"},
			Options: workspace.DocQueryOptions{
				IncludeErrors:       false,
				IncludeDiagnostics:  false,
				IncludeArchivedPath: true,
				IncludeScriptsPath:  true,
				IncludeControlDocs:  true,
				OrderBy:             workspace.OrderByPath,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to resolve ticket directory: %w", err)
		}
		if len(idxRes.Docs) != 1 || strings.TrimSpace(idxRes.Docs[0].Path) == "" {
			return nil, fmt.Errorf("ticket not found or ambiguous: %s", strings.TrimSpace(q.Ticket))
		}
		ticketDir = filepath.Dir(filepath.FromSlash(idxRes.Docs[0].Path))
	}

	// Collect search terms from query and topics.
	searchTerms := []string{}
	if strings.TrimSpace(q.Query) != "" {
		searchTerms = append(searchTerms, strings.TrimSpace(q.Query))
	}
	searchTerms = append(searchTerms, q.Topics...)

	suggested := map[string]bool{}
	var out []FileSuggestion

	emit := func(file, source, reason string) {
		file = strings.TrimSpace(file)
		if file == "" {
			return
		}
		if suggested[file] {
			return
		}
		suggested[file] = true
		out = append(out, FileSuggestion{File: file, Source: source, Reason: reason})
	}

	// 1) Existing RelatedFiles across documents.
	scope := workspace.Scope{Kind: workspace.ScopeRepo}
	if strings.TrimSpace(q.Ticket) != "" {
		scope = workspace.Scope{Kind: workspace.ScopeTicket, TicketID: strings.TrimSpace(q.Ticket)}
	}
	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: scope,
		Filters: workspace.DocFilters{
			TopicsAny: q.Topics,
		},
		Options: workspace.DocQueryOptions{
			IncludeBody:         false,
			IncludeErrors:       false,
			IncludeDiagnostics:  false,
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByPath,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query docs for related_files suggestions: %w", err)
	}
	for _, h := range res.Docs {
		if h.Doc == nil {
			continue
		}
		for _, rf := range h.Doc.RelatedFiles {
			if strings.TrimSpace(rf.Path) == "" {
				continue
			}
			emit(rf.Path, "related_files", "referenced by documents")
		}
	}

	// 2) Git history.
	if gitFiles, err := SuggestFilesFromGit(ticketDir); err == nil {
		for _, file := range gitFiles {
			emit(file, "git_history", "recent commit activity")
		}
	}

	// 3) Git status.
	if modified, staged, untracked, err := SuggestFilesFromGitStatus(ticketDir); err == nil {
		for _, file := range modified {
			emit(file, "git_modified", "working tree modified")
		}
		for _, file := range staged {
			emit(file, "git_staged", "staged for commit")
		}
		for _, file := range untracked {
			emit(file, "git_untracked", "untracked new file")
		}
	}

	// 4) Ripgrep.
	if strings.TrimSpace(q.Query) != "" || len(q.Topics) > 0 {
		if rgFiles, err := SuggestFilesFromRipgrep(ticketDir, searchTerms); err == nil {
			reason := fmt.Sprintf("content match: %s", FirstTerm(searchTerms))
			for _, file := range rgFiles {
				emit(file, "ripgrep", reason)
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Source == out[j].Source {
			return out[i].File < out[j].File
		}
		return out[i].Source < out[j].Source
	})

	return out, nil
}

// ---- Heuristics helpers (ported from the CLI command) ----

func SuggestFilesFromGit(repoPath string) ([]string, error) {
	// Check if we're in a git repository.
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("not a git repository")
	}

	// Get recently modified files (last 30 commits).
	cmd = exec.Command("git", "log", "--name-only", "--pretty=format:", "-30")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %w", err)
	}

	files := make(map[string]bool)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, " ") {
			continue
		}
		if isCodeFile(line) {
			files[line] = true
		}
	}

	result := make([]string, 0, len(files))
	for file := range files {
		result = append(result, file)
	}
	sort.Strings(result)
	return result, nil
}

func SuggestFilesFromGitStatus(repoPath string) ([]string, []string, []string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return nil, nil, nil, fmt.Errorf("not a git repository")
	}

	cmd = exec.Command("git", "diff", "--name-only")
	cmd.Dir = repoPath
	outMod, _ := cmd.Output()
	modified := nonEmptyLines(string(outMod))

	cmd = exec.Command("git", "diff", "--name-only", "--cached")
	cmd.Dir = repoPath
	outStaged, _ := cmd.Output()
	staged := nonEmptyLines(string(outStaged))

	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = repoPath
	outUntracked, _ := cmd.Output()
	untracked := nonEmptyLines(string(outUntracked))

	modified = filterCodeFiles(modified)
	staged = filterCodeFiles(staged)
	untracked = filterCodeFiles(untracked)

	sort.Strings(modified)
	sort.Strings(staged)
	sort.Strings(untracked)

	return modified, staged, untracked, nil
}

func SuggestFilesFromRipgrep(searchPath string, searchTerms []string) ([]string, error) {
	if len(searchTerms) == 0 {
		return nil, nil
	}
	rgPath, err := exec.LookPath("rg")
	if err != nil {
		return SuggestFilesFromGrep(searchPath, searchTerms)
	}

	args := []string{
		"--files-with-matches",
		"--type", "go",
		"--type", "typescript",
		"--type", "javascript",
		"--type", "python",
		"--type", "rust",
		"--type", "java",
		"--type", "kotlin",
		"--type", "scala",
	}
	args = append(args, searchTerms[0])

	cmd := exec.Command(rgPath, args...)
	cmd.Dir = searchPath
	output, err := cmd.Output()
	if err != nil {
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 1 {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to run ripgrep: %w", err)
	}

	var files []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	sort.Strings(files)
	return files, nil
}

func SuggestFilesFromGrep(searchPath string, searchTerms []string) ([]string, error) {
	if len(searchTerms) == 0 {
		return nil, nil
	}

	cmd := exec.Command("grep", "-rl", "--include=*.go", "--include=*.ts", "--include=*.tsx", "--include=*.js", "--include=*.py", "--include=*.rs", "--include=*.java", "--include=*.kt", "--include=*.scala", searchTerms[0], ".")
	cmd.Dir = searchPath
	output, err := cmd.Output()
	if err != nil {
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 1 {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to run grep: %w", err)
	}

	var files []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	sort.Strings(files)
	return files, nil
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func filterCodeFiles(files []string) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		if isCodeFile(f) {
			out = append(out, f)
		}
	}
	return out
}

func FirstTerm(terms []string) string {
	if len(terms) == 0 {
		return ""
	}
	return terms[0]
}

func isCodeFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".rs", ".java", ".kt", ".scala":
		return true
	default:
		return false
	}
}
