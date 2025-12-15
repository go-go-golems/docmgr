package scenariolog

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExecStepCapturesArtifactsAndExitCode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	root := t.TempDir()
	dbPath := filepath.Join(root, "run.db")

	db, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	runID := "run-1"
	// StartRun with deterministic timestamps.
	startedAt := mustNowUTC()
	if err := StartRun(ctx, db, runID, root, "suite", startedAt); err != nil {
		t.Fatalf("StartRun: %v", err)
	}

	// Use logDir "." to avoid needing to create directories.
	res, err := ExecStep(ctx, db, ExecStepSpec{
		RunID:    runID,
		RootDir:  root,
		LogDir:   ".",
		StepNum:  1,
		StepName: "test-step",
		// Use --noprofile/--norc to avoid user shell init influencing stderr.
		Command: []string{"bash", "--noprofile", "--norc", "-c", "echo out; echo err 1>&2; exit 3"},
	})
	if err != nil {
		t.Fatalf("ExecStep: %v", err)
	}
	if res.ExitCode != 3 {
		t.Fatalf("ExitCode=%d, want 3", res.ExitCode)
	}

	stdoutAbs := filepath.Join(root, filepath.FromSlash(res.StdoutPath))
	stderrAbs := filepath.Join(root, filepath.FromSlash(res.StderrPath))
	stdoutBytes, err := os.ReadFile(stdoutAbs)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	stderrBytes, err := os.ReadFile(stderrAbs)
	if err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	if string(stdoutBytes) != "out\n" {
		t.Fatalf("stdout=%q, want %q", string(stdoutBytes), "out\n")
	}
	if string(stderrBytes) != "err\n" {
		t.Fatalf("stderr=%q, want %q", string(stderrBytes), "err\n")
	}
}

func mustNowUTC() (t0 time.Time) {
	return time.Now().UTC()
}


