package commands

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/docmgr/internal/skills"
	"github.com/go-go-golems/docmgr/internal/tickets"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// SkillImportCommand imports an Agent Skills package into a skill plan.
type SkillImportCommand struct {
	*cmds.CommandDescription
}

// SkillImportSettings holds the parameters for skill import.
type SkillImportSettings struct {
	Root      string   `glazed.parameter:"root"`
	Ticket    string   `glazed.parameter:"ticket"`
	Input     string   `glazed.parameter:"input"`
	Topics    []string `glazed.parameter:"topics"`
	Title     string   `glazed.parameter:"title"`
	WhatFor   string   `glazed.parameter:"what-for"`
	WhenToUse string   `glazed.parameter:"when-to-use"`
	Force     bool     `glazed.parameter:"force"`
}

func NewSkillImportCommand() (*SkillImportCommand, error) {
	return &SkillImportCommand{
		CommandDescription: cmds.NewCommandDescription(
			"import",
			cmds.WithShort("Import an Agent Skills package into skill.yaml"),
			cmds.WithLong(`Imports a .skill archive or skill directory into a skill.yaml plan.

This unpacks SKILL.md and references/ into a plan directory under ttmp/skills/
(or <ticket>/skills/ when --ticket is provided) and generates a skill.yaml file.

Examples:
  docmgr skill import ./dist/glaze-help.skill
  docmgr skill import ./dist/glaze-help.skill --ticket MEN-4242
  docmgr skill import ./my-skill-dir --topics tooling,docs
`),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"input",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to .skill archive or skill directory"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Write plan into a ticket's skills folder"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"topics",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Override topics for the imported plan"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"title",
					parameters.ParameterTypeString,
					parameters.WithHelp("Override title for the imported plan"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"what-for",
					parameters.ParameterTypeString,
					parameters.WithHelp("Override what_for for the imported plan"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"when-to-use",
					parameters.ParameterTypeString,
					parameters.WithHelp("Override when_to_use for the imported plan"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"force",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Overwrite existing plan files"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

// Run implements BareCommand.
func (c *SkillImportCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &SkillImportSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	inputPath := strings.TrimSpace(settings.Input)
	if inputPath == "" {
		return errors.New("input path is required")
	}

	settings.Root = workspace.ResolveRoot(settings.Root)
	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return errors.Wrap(err, "failed to discover workspace")
	}
	settings.Root = ws.Context().Root

	if strings.TrimSpace(settings.Ticket) != "" {
		if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
			return errors.Wrap(err, "failed to initialize workspace index")
		}
	}

	skillDir, cleanup, err := resolveSkillImportInput(inputPath)
	if err != nil {
		return err
	}
	defer cleanup()

	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	parsed, err := skills.ParseSkillMarkdown(skillMDPath)
	if err != nil {
		return err
	}
	if err := skills.ValidateSkillMarkdown(parsed.Frontmatter); err != nil {
		return err
	}

	skillName := strings.TrimSpace(parsed.Frontmatter.Name)
	if skillName == "" {
		return errors.New("skill name missing from SKILL.md")
	}

	title := firstNonEmpty(strings.TrimSpace(settings.Title), strings.TrimSpace(parsed.Metadata.Title), strings.TrimSpace(parsed.Title), skillName)
	whatFor := firstNonEmpty(strings.TrimSpace(settings.WhatFor), strings.TrimSpace(parsed.Metadata.WhatFor), strings.TrimSpace(parsed.Frontmatter.Description))
	whenToUse := firstNonEmpty(strings.TrimSpace(settings.WhenToUse), strings.TrimSpace(parsed.Metadata.WhenToUse))
	if whenToUse == "" {
		whenToUse = fmt.Sprintf("Use when working with %s.", title)
	}

	topics := settings.Topics
	if len(topics) == 0 {
		topics = parsed.Metadata.Topics
	}
	if len(topics) == 0 {
		topics = []string{"imported-skill"}
	}

	planDir, err := resolvePlanDir(ctx, ws, skillName, strings.TrimSpace(settings.Ticket))
	if err != nil {
		return err
	}
	if err := ensurePlanDir(planDir, settings.Force); err != nil {
		return err
	}

	referenceFiles, err := collectReferenceFiles(filepath.Join(skillDir, "references"))
	if err != nil {
		return err
	}

	planRel, err := filepath.Rel(ws.Context().RepoRoot, planDir)
	if err != nil {
		planRel = planDir
	}
	planRel = filepath.ToSlash(planRel)

	plan := skills.Plan{
		Skill: skills.SkillMetadata{
			Name:          skillName,
			Title:         title,
			Description:   strings.TrimSpace(parsed.Frontmatter.Description),
			WhatFor:       whatFor,
			WhenToUse:     whenToUse,
			Topics:        topics,
			License:       strings.TrimSpace(parsed.Frontmatter.License),
			Compatibility: strings.TrimSpace(parsed.Metadata.Compatibility),
		},
		Output: skills.OutputConfig{
			SkillDirName: skillName,
		},
	}

	body := strings.TrimSpace(parsed.Body)
	if body != "" {
		bodyPath := filepath.Join(planDir, "skill-body.md")
		if err := os.WriteFile(bodyPath, []byte(body+"\n"), 0o644); err != nil {
			return errors.Wrap(err, "failed to write skill body")
		}
		sourcePath := filepath.ToSlash(filepath.Join(planRel, "skill-body.md"))
		plan.Sources = append(plan.Sources, skills.Source{
			Type:         "file",
			Path:         sourcePath,
			AppendToBody: true,
		})
	}

	for _, src := range referenceFiles {
		rel := filepath.ToSlash(src.RelPath)
		if rel == "" {
			continue
		}
		destPath := filepath.Join(planDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return errors.Wrap(err, "failed to create plan references directory")
		}
		if err := copyFile(src.AbsPath, destPath); err != nil {
			return err
		}

		sourcePath := filepath.ToSlash(filepath.Join(planRel, rel))
		plan.Sources = append(plan.Sources, skills.Source{
			Type:   "file",
			Path:   sourcePath,
			Output: rel,
		})
	}

	if err := plan.Validate(); err != nil {
		return err
	}

	data, err := yaml.Marshal(&plan)
	if err != nil {
		return errors.Wrap(err, "failed to serialize skill plan")
	}
	planPath := filepath.Join(planDir, "skill.yaml")
	if err := os.WriteFile(planPath, data, 0o644); err != nil {
		return errors.Wrap(err, "failed to write skill plan")
	}

	fmt.Fprintf(os.Stdout, "Imported skill plan to %s\n", planPath)
	return nil
}

func resolvePlanDir(ctx context.Context, ws *workspace.Workspace, skillName string, ticketID string) (string, error) {
	if strings.TrimSpace(ticketID) == "" {
		return filepath.Join(ws.Context().Root, "skills", skillName), nil
	}

	res, err := tickets.Resolve(ctx, ws, ticketID)
	if err != nil {
		return "", errors.Wrap(err, "failed to resolve ticket")
	}
	return filepath.Join(res.TicketDirAbs, "skills", skillName), nil
}

func ensurePlanDir(path string, force bool) error {
	if info, err := os.Stat(path); err == nil {
		if !info.IsDir() {
			return errors.Errorf("plan path exists and is not a directory: %s", path)
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return errors.Wrap(err, "failed to read plan directory")
		}
		if len(entries) > 0 && !force {
			return errors.Errorf("plan directory is not empty: %s", path)
		}
	} else if !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to stat plan directory")
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return errors.Wrap(err, "failed to create plan directory")
	}
	return nil
}

type referenceFile struct {
	RelPath string
	AbsPath string
}

func collectReferenceFiles(base string) ([]referenceFile, error) {
	info, err := os.Stat(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to stat references directory")
	}
	if !info.IsDir() {
		return nil, errors.New("references path is not a directory")
	}

	var files []referenceFile
	err = filepath.WalkDir(base, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		files = append(files, referenceFile{RelPath: rel, AbsPath: path})
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan references")
	}

	sort.Slice(files, func(i, j int) bool { return files[i].RelPath < files[j].RelPath })
	return files, nil
}

func resolveSkillImportInput(path string) (string, func(), error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", func() {}, errors.Wrap(err, "failed to stat input")
	}
	if info.IsDir() {
		return path, func() {}, nil
	}
	if !strings.HasSuffix(strings.ToLower(path), ".skill") {
		return "", func() {}, errors.New("input must be a .skill file or directory")
	}

	tmpDir, err := os.MkdirTemp("", "docmgr-skill-import-")
	if err != nil {
		return "", func() {}, errors.Wrap(err, "failed to create temp directory")
	}
	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	if err := unzipSkillArchive(path, tmpDir); err != nil {
		cleanup()
		return "", func() {}, err
	}

	skillDir, err := findSkillDir(tmpDir)
	if err != nil {
		cleanup()
		return "", func() {}, err
	}

	return skillDir, cleanup, nil
}

func unzipSkillArchive(archivePath string, destDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return errors.Wrap(err, "failed to open .skill archive")
	}
	defer func() {
		_ = reader.Close()
	}()

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		cleanName := filepath.Clean(file.Name)
		destPath := filepath.Join(destDir, filepath.FromSlash(cleanName))
		if rel, err := filepath.Rel(destDir, destPath); err != nil || strings.HasPrefix(rel, "..") {
			return errors.Errorf("archive entry escapes destination: %s", file.Name)
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return errors.Wrap(err, "failed to create archive output dir")
		}

		rc, err := file.Open()
		if err != nil {
			return errors.Wrap(err, "failed to read archive file")
		}

		out, err := os.Create(destPath)
		if err != nil {
			_ = rc.Close()
			return errors.Wrap(err, "failed to write archive file")
		}

		if _, err := io.Copy(out, rc); err != nil {
			_ = rc.Close()
			_ = out.Close()
			return errors.Wrap(err, "failed to copy archive file")
		}
		_ = rc.Close()
		_ = out.Close()
	}

	return nil
}

func findSkillDir(root string) (string, error) {
	var matches []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), "SKILL.md") {
			matches = append(matches, filepath.Dir(path))
		}
		return nil
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to locate SKILL.md")
	}
	if len(matches) == 0 {
		return "", errors.New("no SKILL.md found in archive")
	}
	if len(matches) > 1 {
		return "", errors.New("multiple SKILL.md files found in archive")
	}
	return matches[0], nil
}

func copyFile(src string, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "failed to open source file")
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.Create(dest)
	if err != nil {
		return errors.Wrap(err, "failed to create destination file")
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return errors.Wrap(err, "failed to copy file")
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

var _ cmds.BareCommand = &SkillImportCommand{}
