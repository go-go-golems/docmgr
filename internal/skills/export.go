package skills

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/pkg/errors"
)

// ExportOptions controls skill export behavior.
type ExportOptions struct {
	OutDir   string
	SkillDir string
	Force    bool
}

// ExportResult captures export outputs.
type ExportResult struct {
	SkillDir    string
	PackagePath string
}

// ExportPlan resolves a plan and packages it into a .skill file.
func ExportPlan(ctx context.Context, ws *workspace.Workspace, handle PlanHandle, opts ExportOptions) (ExportResult, error) {
	if ws == nil {
		return ExportResult{}, errors.New("workspace is required")
	}
	if handle.Plan == nil {
		return ExportResult{}, errors.New("skill plan is required")
	}

	skillName := handle.Plan.SkillDirName()
	if strings.TrimSpace(skillName) == "" {
		return ExportResult{}, errors.New("skill name is required")
	}

	baseDir := strings.TrimSpace(opts.SkillDir)
	var skillDir string
	if baseDir != "" {
		skillDir = filepath.Join(baseDir, skillName)
	} else {
		tmpDir, err := os.MkdirTemp("", "docmgr-skill-")
		if err != nil {
			return ExportResult{}, errors.Wrap(err, "failed to create temp directory")
		}
		skillDir = filepath.Join(tmpDir, skillName)
	}

	if err := ensureEmptyDir(skillDir, opts.Force); err != nil {
		return ExportResult{}, err
	}
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return ExportResult{}, errors.Wrap(err, "failed to create skill directory")
	}

	resolved, err := ResolvePlan(ctx, ws, handle, ResolveOptions{AllowBinary: true})
	if err != nil {
		return ExportResult{}, err
	}

	var referencePaths []string
	for _, res := range resolved {
		outputPath := cleanOutputPath(res.OutputPath)
		if strings.TrimSpace(outputPath) == "" {
			return ExportResult{}, errors.New("resolved output path is empty")
		}
		absPath := filepath.Join(skillDir, filepath.FromSlash(outputPath))
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			return ExportResult{}, errors.Wrap(err, "failed to create output directory")
		}
		if err := os.WriteFile(absPath, res.Content, 0o644); err != nil {
			return ExportResult{}, errors.Wrap(err, "failed to write resolved content")
		}
		if strings.HasPrefix(outputPath, "references/") {
			referencePaths = append(referencePaths, outputPath)
		}
	}

	if len(referencePaths) == 0 {
		for _, res := range resolved {
			referencePaths = append(referencePaths, cleanOutputPath(res.OutputPath))
		}
	}

	skillMD, err := RenderSkillMarkdown(handle.Plan, referencePaths)
	if err != nil {
		return ExportResult{}, err
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), skillMD, 0o644); err != nil {
		return ExportResult{}, errors.Wrap(err, "failed to write SKILL.md")
	}

	parsed, err := ParseSkillMarkdown(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		return ExportResult{}, err
	}
	if err := ValidateSkillMarkdown(parsed.Frontmatter); err != nil {
		return ExportResult{}, err
	}

	outDir := strings.TrimSpace(opts.OutDir)
	if outDir == "" {
		outDir = "."
	}
	packagePath, err := PackageSkillDir(skillDir, outDir, opts.Force)
	if err != nil {
		return ExportResult{}, err
	}

	return ExportResult{SkillDir: skillDir, PackagePath: packagePath}, nil
}

func ensureEmptyDir(path string, force bool) error {
	if info, err := os.Stat(path); err == nil {
		if !info.IsDir() {
			return errors.Errorf("path exists and is not a directory: %s", path)
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return errors.Wrap(err, "failed to read skill directory")
		}
		if len(entries) > 0 && !force {
			return errors.Errorf("skill directory is not empty: %s", path)
		}
	} else if !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to stat skill directory")
	}
	return nil
}
