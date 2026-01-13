package skills

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/go-go-golems/docmgr/internal/workspace"
)

func TestResolveBinaryHelp(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script not supported on windows")
	}

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "stub-help")
	script := "#!/bin/sh\nif [ \"$1\" = \"help\" ]; then echo \"help for $2\"; exit 0; fi\nexit 1\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}

	docsRoot := filepath.Join(tmpDir, "ttmp")
	if err := os.MkdirAll(docsRoot, 0o755); err != nil {
		t.Fatalf("mkdir docs root: %v", err)
	}

	ws, err := workspace.NewWorkspaceFromContext(workspace.WorkspaceContext{
		Root:      docsRoot,
		ConfigDir: tmpDir,
		RepoRoot:  tmpDir,
	})
	if err != nil {
		t.Fatalf("workspace: %v", err)
	}

	plan := &Plan{
		Skill: SkillMetadata{
			Name:        "stub-help",
			Description: "Stub help",
			WhatFor:     "Provide help",
			WhenToUse:   "Use when testing",
			Topics:      []string{"test"},
		},
		Sources: []Source{
			{
				Type:   "binary-help",
				Binary: scriptPath,
				Topic:  "topic",
				Output: "references/help.txt",
				Wrap:   "none",
			},
		},
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("validate plan: %v", err)
	}

	handle := PlanHandle{Plan: plan, Path: filepath.Join(tmpDir, "skill.yaml")}
	resolved, err := ResolvePlan(context.Background(), ws, handle, ResolveOptions{AllowBinary: true})
	if err != nil {
		t.Fatalf("resolve plan: %v", err)
	}
	if len(resolved) != 1 {
		t.Fatalf("expected 1 resolved source, got %d", len(resolved))
	}
	if !strings.Contains(string(resolved[0].Content), "help for topic") {
		t.Fatalf("unexpected output: %s", string(resolved[0].Content))
	}
}
