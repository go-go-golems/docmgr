package skills

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	maxSkillNameLength        = 64
	maxSkillDescriptionLength = 1024
)

var skillNamePattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// Validate checks the plan for required fields and constraints.
func (p *Plan) Validate() error {
	if p == nil {
		return errors.New("skill plan is nil")
	}

	var issues []string

	name := strings.TrimSpace(p.Skill.Name)
	description := strings.TrimSpace(p.Skill.Description)
	whatFor := strings.TrimSpace(p.Skill.WhatFor)
	whenToUse := strings.TrimSpace(p.Skill.WhenToUse)

	if name == "" {
		issues = append(issues, "skill.name is required")
	} else if err := validateSkillName(name); err != nil {
		issues = append(issues, err.Error())
	}

	if description == "" {
		issues = append(issues, "skill.description is required")
	} else if err := validateSkillDescription(description); err != nil {
		issues = append(issues, err.Error())
	}

	if whatFor == "" {
		issues = append(issues, "skill.what_for is required")
	}
	if whenToUse == "" {
		issues = append(issues, "skill.when_to_use is required")
	}

	if len(p.Skill.Topics) == 0 {
		issues = append(issues, "skill.topics must include at least one topic")
	}

	if strings.TrimSpace(p.Output.SkillDirName) != "" && strings.TrimSpace(p.Output.SkillDirName) != name {
		issues = append(issues, fmt.Sprintf("output.skill_dir_name must match skill.name (%q)", name))
	}

	for i, source := range p.Sources {
		issuePrefix := fmt.Sprintf("sources[%d]", i)
		if strings.TrimSpace(source.Type) == "" {
			issues = append(issues, fmt.Sprintf("%s.type is required", issuePrefix))
			continue
		}

		switch strings.ToLower(strings.TrimSpace(source.Type)) {
		case "file":
			if strings.TrimSpace(source.Path) == "" {
				issues = append(issues, fmt.Sprintf("%s.path is required for file source", issuePrefix))
			}
			if strings.TrimSpace(source.Output) == "" {
				if !source.AppendToBody {
					issues = append(issues, fmt.Sprintf("%s.output is required for file source", issuePrefix))
				}
			} else if err := validateOutputPath(source.Output); err != nil {
				issues = append(issues, fmt.Sprintf("%s.output invalid: %s", issuePrefix, err))
			}
		case "binary-help":
			if strings.TrimSpace(source.Binary) == "" {
				issues = append(issues, fmt.Sprintf("%s.binary is required for binary-help source", issuePrefix))
			}
			if strings.TrimSpace(source.Topic) == "" {
				issues = append(issues, fmt.Sprintf("%s.topic is required for binary-help source", issuePrefix))
			}
			if strings.TrimSpace(source.Output) == "" {
				if !source.AppendToBody {
					issues = append(issues, fmt.Sprintf("%s.output is required for binary-help source", issuePrefix))
				}
			} else if err := validateOutputPath(source.Output); err != nil {
				issues = append(issues, fmt.Sprintf("%s.output invalid: %s", issuePrefix, err))
			}
			if strings.TrimSpace(source.Wrap) != "" {
				wrap := strings.ToLower(strings.TrimSpace(source.Wrap))
				if wrap != "markdown" && wrap != "none" {
					issues = append(issues, fmt.Sprintf("%s.wrap must be 'markdown' or 'none'", issuePrefix))
				}
			}
		default:
			issues = append(issues, fmt.Sprintf("%s.type %q is not supported", issuePrefix, source.Type))
		}
	}

	if len(issues) > 0 {
		return errors.Errorf("invalid skill plan:\n- %s", strings.Join(issues, "\n- "))
	}

	return nil
}

func validateSkillName(name string) error {
	if !skillNamePattern.MatchString(name) {
		return fmt.Errorf("skill.name must be lowercase letters, digits, and hyphens only")
	}
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") || strings.Contains(name, "--") {
		return fmt.Errorf("skill.name cannot start/end with '-' or contain consecutive hyphens")
	}
	if len(name) > maxSkillNameLength {
		return fmt.Errorf("skill.name is too long (%d characters), max is %d", len(name), maxSkillNameLength)
	}
	return nil
}

func validateSkillDescription(description string) error {
	if strings.Contains(description, "<") || strings.Contains(description, ">") {
		return fmt.Errorf("skill.description cannot contain angle brackets")
	}
	if len(description) > maxSkillDescriptionLength {
		return fmt.Errorf("skill.description is too long (%d characters), max is %d", len(description), maxSkillDescriptionLength)
	}
	return nil
}

func validateOutputPath(output string) error {
	clean := filepath.ToSlash(filepath.Clean(strings.TrimSpace(output)))
	if clean == "." || clean == "" {
		return fmt.Errorf("output path is empty")
	}
	if strings.HasPrefix(clean, "/") {
		return fmt.Errorf("output path must be relative")
	}
	if strings.HasPrefix(clean, "../") || clean == ".." || strings.Contains(clean, "/../") {
		return fmt.Errorf("output path must not escape the skill directory")
	}
	return nil
}
