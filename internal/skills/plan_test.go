package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPlanValid(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "skill.yaml")

	data := []byte(`skill:
  name: glaze-help
  title: Glaze Help System
  description: Help topics for Glazed.
  what_for: Provide Glazed help output.
  when_to_use: Use when working with Glazed help.
  topics: [glaze, help]

sources:
  - type: file
    path: glazed/pkg/doc/topics/01-help-system.md
    output: references/help-system.md
`)
	if err := os.WriteFile(planPath, data, 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}

	plan, err := LoadPlan(planPath)
	if err != nil {
		t.Fatalf("load plan: %v", err)
	}
	if plan.Skill.Name != "glaze-help" {
		t.Fatalf("unexpected name: %s", plan.Skill.Name)
	}
	if len(plan.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(plan.Sources))
	}
}

func TestLoadPlanInvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "skill.yaml")

	data := []byte(`skill:
  name: Glaze Help
  description: Help topics.
  what_for: Provide help.
  when_to_use: Use when working with help.
  topics: [glaze]

sources: []
`)
	if err := os.WriteFile(planPath, data, 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}

	_, err := LoadPlan(planPath)
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestLoadPlanAppendBodyWithoutOutput(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := filepath.Join(tmpDir, "skill.yaml")

	data := []byte(`skill:
  name: docmgr
  description: Appendable skill body.
  what_for: Provide guidance.
  when_to_use: Use when working with docs.
  topics: [docs]

sources:
  - type: file
    path: ttmp/skills/docmgr/skill-body.md
    append_to_body: true
`)
	if err := os.WriteFile(planPath, data, 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}

	if _, err := LoadPlan(planPath); err != nil {
		t.Fatalf("expected plan to load, got error: %v", err)
	}
}
