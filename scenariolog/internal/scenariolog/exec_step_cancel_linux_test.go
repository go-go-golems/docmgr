//go:build linux

package scenariolog

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestExecStepCancelKillsChildProcessGroup(t *testing.T) {
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

	runID := "run-cancel-1"
	if err := StartRun(ctx, db, runID, root, "suite", time.Now().UTC()); err != nil {
		t.Fatalf("StartRun: %v", err)
	}

	stepCtx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()

	// This command spawns a child and records its pid in root/child.pid, then waits.
	// If cancellation only kills the parent, the child would continue running.
	_ = os.Remove(filepath.Join(root, "child.pid"))

	_, _ = ExecStep(stepCtx, db, ExecStepSpec{
		RunID:    runID,
		RootDir:  root,
		WorkDir:  root,
		LogDir:   ".",
		StepNum:  1,
		StepName: "cancel-test",
		Command:  []string{"bash", "--noprofile", "--norc", "-c", "sleep 30 & echo $! > child.pid; wait"},
	})

	pidBytes, err := os.ReadFile(filepath.Join(root, "child.pid"))
	if err != nil {
		t.Fatalf("expected child.pid to be written before cancel; read error: %v", err)
	}
	pidStr := strings.TrimSpace(string(pidBytes))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		t.Fatalf("parse child pid %q: %v", pidStr, err)
	}

	// Poll /proc to ensure the child is gone (give it a moment to die).
	deadline := time.Now().Add(2 * time.Second)
	for {
		_, err := os.Stat(filepath.Join("/proc", strconv.Itoa(pid)))
		if err != nil {
			break // gone (or unreadable); good enough for this test
		}
		if time.Now().After(deadline) {
			t.Fatalf("child pid %d still appears alive after cancel (proc entry exists)", pid)
		}
		time.Sleep(50 * time.Millisecond)
	}
}


