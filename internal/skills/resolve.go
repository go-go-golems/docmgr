package skills

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/pkg/errors"
)

const defaultBinaryHelpTimeout = 10 * time.Second

// ResolvedSource is the materialized output of a plan source.
type ResolvedSource struct {
	Source     Source
	OutputPath string
	Content    []byte
}

// ResolveOptions controls plan resolution.
type ResolveOptions struct {
	AllowBinary    bool
	DefaultTimeout time.Duration
}

// ResolvePlan materializes plan sources without writing files.
func ResolvePlan(ctx context.Context, ws *workspace.Workspace, handle PlanHandle, opts ResolveOptions) ([]ResolvedSource, error) {
	if ws == nil {
		return nil, errors.New("workspace is required")
	}
	if handle.Plan == nil {
		return nil, errors.New("skill plan is required")
	}

	if opts.DefaultTimeout <= 0 {
		opts.DefaultTimeout = defaultBinaryHelpTimeout
	}

	planDir := filepath.Dir(handle.Path)
	repoRoot := ws.Context().RepoRoot

	var resolved []ResolvedSource
	for _, source := range handle.Plan.Sources {
		switch strings.ToLower(strings.TrimSpace(source.Type)) {
		case "file":
			content, err := resolveFileSource(repoRoot, planDir, source)
			if err != nil {
				return nil, err
			}
			resolved = append(resolved, ResolvedSource{
				Source:     source,
				OutputPath: normalizeOutputPath(source.Output),
				Content:    content,
			})
		case "binary-help":
			if !opts.AllowBinary {
				return nil, errors.Errorf("binary help source %q requires --resolve or export", source.Topic)
			}
			content, err := resolveBinaryHelp(ctx, source, opts.DefaultTimeout)
			if err != nil {
				return nil, err
			}
			resolved = append(resolved, ResolvedSource{
				Source:     source,
				OutputPath: normalizeOutputPath(source.Output),
				Content:    content,
			})
		default:
			return nil, errors.Errorf("unsupported source type %q", source.Type)
		}
	}

	return resolved, nil
}

func resolveFileSource(repoRoot string, planDir string, source Source) ([]byte, error) {
	absPath, err := resolveSourcePath(repoRoot, planDir, source.Path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read source file %s", absPath)
	}
	if source.StripFrontmatter {
		data = stripFrontmatter(data)
	}
	return data, nil
}

func resolveBinaryHelp(ctx context.Context, source Source, defaultTimeout time.Duration) ([]byte, error) {
	timeout := defaultTimeout
	if source.TimeoutSeconds > 0 {
		timeout = time.Duration(source.TimeoutSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := []string{"help"}
	if strings.TrimSpace(source.Topic) != "" {
		args = append(args, strings.TrimSpace(source.Topic))
	}

	cmd := exec.CommandContext(ctx, strings.TrimSpace(source.Binary), args...)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, errors.Errorf("binary help timed out after %s for %s", timeout, source.Binary)
	}
	if err != nil {
		if stderrors.Is(err, exec.ErrNotFound) {
			return nil, errors.Errorf("binary not found on PATH: %s", strings.TrimSpace(source.Binary))
		}
		if execErr, ok := err.(*exec.Error); ok && execErr.Err == exec.ErrNotFound {
			return nil, errors.Errorf("binary not found on PATH: %s", strings.TrimSpace(source.Binary))
		}
		return nil, errors.Wrapf(err, "binary help failed: %s", strings.TrimSpace(string(output)))
	}

	wrapMode := strings.ToLower(strings.TrimSpace(source.Wrap))
	if wrapMode == "" {
		wrapMode = "markdown"
	}
	if wrapMode == "markdown" {
		return wrapMarkdown(output), nil
	}
	return output, nil
}

func wrapMarkdown(content []byte) []byte {
	trimmed := strings.TrimRight(string(content), "\n")
	return []byte(fmt.Sprintf("```text\n%s\n```\n", trimmed))
}

func resolveSourcePath(repoRoot string, planDir string, raw string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(strings.TrimSpace(raw)))
	if clean == "." || clean == "" {
		return "", errors.New("source path is empty")
	}
	if filepath.IsAbs(clean) {
		return clean, nil
	}
	if repoRoot == "" {
		return "", errors.New("repository root is required to resolve relative paths")
	}
	abs := filepath.Join(repoRoot, clean)
	if rel, err := filepath.Rel(repoRoot, abs); err != nil || strings.HasPrefix(rel, "..") {
		return "", errors.Errorf("source path escapes repository: %s", raw)
	}
	return abs, nil
}

func cleanOutputPath(path string) string {
	return filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
}

func normalizeOutputPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	return cleanOutputPath(trimmed)
}

func stripFrontmatter(data []byte) []byte {
	content := string(data)
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return data
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return []byte(strings.Join(lines[i+1:], "\n"))
		}
	}
	return data
}
