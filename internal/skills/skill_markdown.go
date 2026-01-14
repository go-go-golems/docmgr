package skills

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// SkillFrontmatter is the Agent Skills SKILL.md frontmatter.
type SkillFrontmatter struct {
	Name         string                 `yaml:"name"`
	Description  string                 `yaml:"description"`
	License      string                 `yaml:"license"`
	AllowedTools []string               `yaml:"allowed-tools"`
	Metadata     map[string]interface{} `yaml:"metadata"`
}

// ParsedSkillMarkdown captures SKILL.md frontmatter and body.
type ParsedSkillMarkdown struct {
	Frontmatter SkillFrontmatter
	Body        string
	Title       string
	Metadata    SkillExtraMetadata
}

// SkillExtraMetadata captures optional metadata fields.
type SkillExtraMetadata struct {
	Topics        []string `yaml:"topics"`
	WhatFor       string   `yaml:"what_for"`
	WhenToUse     string   `yaml:"when_to_use"`
	Title         string   `yaml:"title"`
	Compatibility string   `yaml:"compatibility"`
}

// RenderSkillMarkdown builds SKILL.md for export.
func RenderSkillMarkdown(plan *Plan, referencePaths []string, appendBodies []string) ([]byte, error) {
	if plan == nil {
		return nil, errors.New("skill plan is required")
	}

	frontmatter := map[string]interface{}{
		"name":        strings.TrimSpace(plan.Skill.Name),
		"description": strings.TrimSpace(plan.Skill.Description),
	}
	if strings.TrimSpace(plan.Skill.License) != "" {
		frontmatter["license"] = strings.TrimSpace(plan.Skill.License)
	}

	metadata := map[string]interface{}{}
	if len(plan.Skill.Topics) > 0 {
		metadata["topics"] = plan.Skill.Topics
	}
	if strings.TrimSpace(plan.Skill.WhatFor) != "" {
		metadata["what_for"] = strings.TrimSpace(plan.Skill.WhatFor)
	}
	if strings.TrimSpace(plan.Skill.WhenToUse) != "" {
		metadata["when_to_use"] = strings.TrimSpace(plan.Skill.WhenToUse)
	}
	if strings.TrimSpace(plan.Skill.Compatibility) != "" {
		metadata["compatibility"] = strings.TrimSpace(plan.Skill.Compatibility)
	}
	if strings.TrimSpace(plan.Skill.Title) != "" {
		metadata["title"] = strings.TrimSpace(plan.Skill.Title)
	}
	if len(metadata) > 0 {
		frontmatter["metadata"] = metadata
	}

	fmData, err := yaml.Marshal(frontmatter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render skill frontmatter")
	}

	cleanedAppendBodies := make([]string, 0, len(appendBodies))
	for _, chunk := range appendBodies {
		trimmed := strings.TrimSpace(chunk)
		if trimmed != "" {
			cleanedAppendBodies = append(cleanedAppendBodies, trimmed)
		}
	}
	appendBodies = cleanedAppendBodies

	includeAutoSections := len(appendBodies) == 0

	includeTitle := true
	if len(appendBodies) > 0 {
		for _, line := range strings.Split(appendBodies[0], "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if strings.HasPrefix(trimmed, "# ") {
				includeTitle = false
			}
			break
		}
	}

	var body strings.Builder
	if includeTitle {
		body.WriteString("# ")
		body.WriteString(plan.DisplayTitle())
		body.WriteString("\n\n")
	}

	if includeAutoSections {
		intro := strings.TrimSpace(plan.Output.SkillMD.Intro)
		if intro != "" {
			body.WriteString(intro)
			body.WriteString("\n\n")
		} else if strings.TrimSpace(plan.Skill.Description) != "" {
			body.WriteString(strings.TrimSpace(plan.Skill.Description))
			body.WriteString("\n\n")
		}

		if strings.TrimSpace(plan.Skill.WhatFor) != "" {
			body.WriteString("## What this skill is for\n\n")
			body.WriteString(strings.TrimSpace(plan.Skill.WhatFor))
			body.WriteString("\n\n")
		}
		if strings.TrimSpace(plan.Skill.WhenToUse) != "" {
			body.WriteString("## When to use\n\n")
			body.WriteString(strings.TrimSpace(plan.Skill.WhenToUse))
			body.WriteString("\n\n")
		}
	}

	for _, chunk := range appendBodies {
		body.WriteString(chunk)
		body.WriteString("\n\n")
	}

	if plan.Output.SkillMD.IncludeIndexDefault() && len(referencePaths) > 0 {
		body.WriteString("## ")
		body.WriteString(plan.Output.SkillMD.IndexTitleDefault())
		body.WriteString("\n\n")
		for _, ref := range referencePaths {
			body.WriteString("- ")
			body.WriteString(ref)
			body.WriteString("\n")
		}
		body.WriteString("\n")
	}

	output := strings.Builder{}
	output.WriteString("---\n")
	output.Write(fmData)
	output.WriteString("---\n\n")
	output.WriteString(body.String())

	return []byte(output.String()), nil
}

// ParseSkillMarkdown reads and parses SKILL.md from a skill directory.
func ParseSkillMarkdown(path string) (*ParsedSkillMarkdown, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read SKILL.md")
	}
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return nil, errors.New("SKILL.md missing YAML frontmatter")
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, errors.New("SKILL.md frontmatter must start with ---")
	}

	frontmatterEnd := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			frontmatterEnd = i
			break
		}
	}
	if frontmatterEnd == -1 {
		return nil, errors.New("SKILL.md frontmatter not terminated")
	}

	frontmatterText := strings.Join(lines[1:frontmatterEnd], "\n")
	body := strings.Join(lines[frontmatterEnd+1:], "\n")

	var fm SkillFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatterText), &fm); err != nil {
		return nil, errors.Wrap(err, "failed to parse SKILL.md frontmatter")
	}

	parsed := &ParsedSkillMarkdown{
		Frontmatter: fm,
		Body:        body,
		Title:       extractTitle(body),
		Metadata:    extractSkillMetadata(fm.Metadata),
	}

	return parsed, nil
}

func extractTitle(body string) string {
	for _, line := range strings.Split(body, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(trim, "# "))
		}
	}
	return ""
}

func extractSkillMetadata(metadata map[string]interface{}) SkillExtraMetadata {
	if metadata == nil {
		return SkillExtraMetadata{}
	}
	data, err := yaml.Marshal(metadata)
	if err != nil {
		return SkillExtraMetadata{}
	}
	var out SkillExtraMetadata
	if err := yaml.Unmarshal(data, &out); err != nil {
		return SkillExtraMetadata{}
	}
	return out
}

// ValidateSkillMarkdown applies Agent Skills frontmatter rules.
func ValidateSkillMarkdown(frontmatter SkillFrontmatter) error {
	name := strings.TrimSpace(frontmatter.Name)
	description := strings.TrimSpace(frontmatter.Description)
	if name == "" {
		return errors.New("SKILL.md frontmatter missing name")
	}
	if description == "" {
		return errors.New("SKILL.md frontmatter missing description")
	}
	if err := validateSkillName(name); err != nil {
		return err
	}
	if err := validateSkillDescription(description); err != nil {
		return err
	}
	return nil
}

// FrontmatterMap returns a map for new plan synthesis.
func (p *ParsedSkillMarkdown) FrontmatterMap() map[string]interface{} {
	m := map[string]interface{}{
		"name":        strings.TrimSpace(p.Frontmatter.Name),
		"description": strings.TrimSpace(p.Frontmatter.Description),
	}
	if strings.TrimSpace(p.Frontmatter.License) != "" {
		m["license"] = strings.TrimSpace(p.Frontmatter.License)
	}
	if len(p.Frontmatter.AllowedTools) > 0 {
		m["allowed-tools"] = p.Frontmatter.AllowedTools
	}
	if len(p.Frontmatter.Metadata) > 0 {
		m["metadata"] = p.Frontmatter.Metadata
	}
	return m
}

// EnsureSkillFrontmatterAllowed is a strict check against unexpected keys.
func EnsureSkillFrontmatterAllowed(frontmatter map[string]interface{}) error {
	allowed := map[string]bool{
		"name":          true,
		"description":   true,
		"license":       true,
		"allowed-tools": true,
		"metadata":      true,
	}
	for key := range frontmatter {
		if !allowed[key] {
			return fmt.Errorf("unexpected frontmatter key %q", key)
		}
	}
	return nil
}
