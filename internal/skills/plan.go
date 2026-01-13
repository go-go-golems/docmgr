package skills

import "strings"

// Plan is the root skill.yaml document.
type Plan struct {
	Skill   SkillMetadata `yaml:"skill"`
	Sources []Source      `yaml:"sources"`
	Output  OutputConfig  `yaml:"output"`
}

// SkillMetadata holds metadata used by list/show and export.
type SkillMetadata struct {
	Name          string   `yaml:"name"`
	Title         string   `yaml:"title"`
	Description   string   `yaml:"description"`
	WhatFor       string   `yaml:"what_for"`
	WhenToUse     string   `yaml:"when_to_use"`
	Topics        []string `yaml:"topics"`
	License       string   `yaml:"license"`
	Compatibility string   `yaml:"compatibility"`
}

// Source describes a content source for building an Agent Skill.
type Source struct {
	Type             string `yaml:"type"`
	Path             string `yaml:"path"`
	StripFrontmatter bool   `yaml:"strip-frontmatter"`
	Output           string `yaml:"output"`
	Binary           string `yaml:"binary"`
	Topic            string `yaml:"topic"`
	Wrap             string `yaml:"wrap"`
	TimeoutSeconds   int    `yaml:"timeout_seconds"`
}

// OutputConfig controls export layout.
type OutputConfig struct {
	SkillDirName string        `yaml:"skill_dir_name"`
	SkillMD      SkillMDConfig `yaml:"skill_md"`
}

// SkillMDConfig controls SKILL.md generation.
type SkillMDConfig struct {
	Intro        string `yaml:"intro"`
	IncludeIndex *bool  `yaml:"include_index"`
	IndexTitle   string `yaml:"index_title"`
}

// DisplayTitle returns a human-friendly title for list/show output.
func (p *Plan) DisplayTitle() string {
	if p == nil {
		return ""
	}
	if strings.TrimSpace(p.Skill.Title) != "" {
		return strings.TrimSpace(p.Skill.Title)
	}
	return strings.TrimSpace(p.Skill.Name)
}

// SkillDirName returns the output directory name for export.
func (p *Plan) SkillDirName() string {
	if p == nil {
		return ""
	}
	if strings.TrimSpace(p.Output.SkillDirName) != "" {
		return strings.TrimSpace(p.Output.SkillDirName)
	}
	return strings.TrimSpace(p.Skill.Name)
}

// IncludeIndex reports whether export should add a references index.
func (cfg SkillMDConfig) IncludeIndexDefault() bool {
	if cfg.IncludeIndex == nil {
		return true
	}
	return *cfg.IncludeIndex
}

// IndexTitleDefault returns the default index title for SKILL.md.
func (cfg SkillMDConfig) IndexTitleDefault() string {
	if strings.TrimSpace(cfg.IndexTitle) != "" {
		return strings.TrimSpace(cfg.IndexTitle)
	}
	return "References"
}
