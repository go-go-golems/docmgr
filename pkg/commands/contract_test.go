package commands_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// maxBareSuccessOutput is the agent-contract budget for bare-mode success
// output of mutating/representative commands (design D2/D4: terse success).
const maxBareSuccessOutput = 400

var (
	buildOnce sync.Once
	buildErr  error
	binPath   string
)

// docmgrBinary builds the docmgr binary once per test run.
//
// The contract is exercised through a real subprocess because the glazed cobra
// wrapper exits the process on command errors (cobra.CheckErr), which would
// kill an in-process test binary on the failure cases.
func docmgrBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
		if err != nil {
			buildErr = err
			return
		}
		dir, err := os.MkdirTemp("", "docmgr-contract-*")
		if err != nil {
			buildErr = err
			return
		}
		binPath = filepath.Join(dir, "docmgr")
		cmd := exec.Command("go", "build", "-o", binPath, "./cmd/docmgr")
		cmd.Dir = repoRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			buildErr = err
			t.Logf("go build output:\n%s", out)
		}
	})
	if buildErr != nil {
		t.Fatalf("failed to build docmgr binary: %v", buildErr)
	}
	return binPath
}

// runDocmgr executes docmgr in the given workspace dir, returning stdout and
// the exec error (non-nil means non-zero exit).
func runDocmgr(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(docmgrBinary(t), args...)
	cmd.Dir = dir
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil && stderr.Len() > 0 {
		t.Logf("docmgr %s stderr: %s", strings.Join(args, " "), stderr.String())
	}
	return stdout.String(), err
}

func mustSucceed(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := runDocmgr(t, dir, args...)
	if err != nil {
		t.Fatalf("docmgr %s failed: %v\noutput:\n%s", strings.Join(args, " "), err, out)
	}
	return out
}

func assertTerse(t *testing.T, out string, args ...string) {
	t.Helper()
	if len(out) > maxBareSuccessOutput {
		t.Errorf("docmgr %s: bare success output is %d bytes (budget %d):\n%s",
			strings.Join(args, " "), len(out), maxBareSuccessOutput, out)
	}
}

// TestAgentContract sets up one workspace fixture and asserts the P0+P1 agent
// CLI contract: terse success output, non-zero exit on failures, and
// --verbose restoring the banner.
func TestAgentContract(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary-based contract test in -short mode")
	}

	tmp := t.TempDir()
	// Mark the temp dir as a repository root so path resolution is anchored.
	if err := os.MkdirAll(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}

	// Fixture: init + one ticket.
	mustSucceed(t, tmp, "init", "--seed-vocabulary")
	createOut := mustSucceed(t, tmp, "ticket", "create", "--ticket", "TEST-1", "--title", "Contract fixture", "--topics", "chat")
	assertTerse(t, createOut, "ticket create")
	if !strings.Contains(createOut, "created TEST-1 at ") {
		t.Errorf("ticket create output missing terse summary: %q", createOut)
	}

	ticketDir := ""
	for _, f := range strings.Fields(createOut) {
		if strings.HasPrefix(f, "ttmp/") {
			ticketDir = f
		}
	}
	if ticketDir == "" {
		t.Fatalf("could not extract ticket dir from create output: %q", createOut)
	}

	t.Run("terse success outputs", func(t *testing.T) {
		cases := [][]string{
			{"doc", "add", "--ticket", "TEST-1", "--doc-type", "design-doc", "--title", "Contract design"},
			{"ticket", "list"},
			{"ticket", "show", "TEST-1"},
			{"task", "add", "--ticket", "TEST-1", "--text", "contract task"},
			{"task", "check", "--ticket", "TEST-1", "--id", "1"},
			{"task", "uncheck", "--ticket", "TEST-1", "--id", "1"},
			{"changelog", "update", "--ticket", "TEST-1", "--entry", "contract entry"},
			{"doc", "relate", "--ticket", "TEST-1", "--file-note", "main.go:entry point"},
			{"meta", "update", "--ticket", "TEST-1", "--field", "Status", "--value", "review"},
			{"task", "remove", "--ticket", "TEST-1", "--id", "1"},
		}
		for _, args := range cases {
			out := mustSucceed(t, tmp, args...)
			assertTerse(t, out, args...)
			if strings.Contains(out, "Docs root:") {
				t.Errorf("docmgr %s: banner printed without --verbose:\n%s", strings.Join(args, " "), out)
			}
			if strings.Contains(out, "Reminder:") {
				t.Errorf("docmgr %s: reminder nag printed without --verbose:\n%s", strings.Join(args, " "), out)
			}
		}
	})

	t.Run("forgiving references", func(t *testing.T) {
		// Unique ID prefix.
		mustSucceed(t, tmp, "ticket", "show", "--ticket", "TEST")
		// Pasted directory slug.
		mustSucceed(t, tmp, "ticket", "show", filepath.Base(ticketDir))
		// Repo-relative --doc (the historical ttmp/ttmp double-join).
		mustSucceed(t, tmp, "meta", "update", "--doc", ticketDir+"/index.md", "--field", "Status", "--value", "active")
		// Docs-root-relative --doc.
		mustSucceed(t, tmp, "meta", "update", "--doc", strings.TrimPrefix(ticketDir, "ttmp/")+"/index.md", "--field", "Status", "--value", "active")

		// Mutating commands must persist canonical ticket IDs, not the short ref.
		mustSucceed(t, tmp, "doc", "add", "--ticket", "TEST", "--doc-type", "analysis", "--title", "Short Ref Doc")
		docPath := filepath.Join(tmp, ticketDir, "analysis", "01-short-ref-doc.md")
		b, err := os.ReadFile(docPath)
		if err != nil {
			t.Fatalf("read short-ref doc: %v", err)
		}
		if !strings.Contains(string(b), "Ticket: TEST-1") {
			t.Fatalf("doc add persisted non-canonical ticket frontmatter:\n%s", b)
		}
		showOut := mustSucceed(t, tmp, "ticket", "show", "TEST")
		if !strings.Contains(showOut, "analysis/01-short-ref-doc.md") {
			t.Fatalf("ticket show did not include doc added with forgiving ref:\n%s", showOut)
		}

		// Doctor should accept the same forgiving ticket refs as ticket show.
		doctorOut := mustSucceed(t, tmp, "doctor", "--ticket", "TEST")
		if strings.Contains(doctorOut, "No tickets checked") {
			t.Fatalf("doctor checked zero tickets for forgiving ref:\n%s", doctorOut)
		}
	})

	t.Run("failures exit non-zero", func(t *testing.T) {
		cases := [][]string{
			{"ticket", "show", "NOPE-1"},
			{"doc", "relate", "--doc", "nope/missing.md", "--file-note", "a.go:b"},
			{"doc", "relate", "--ticket", "NOPE-1", "--file-note", "a.go:b"},
			{"changelog", "update", "--ticket", "TEST-1"},                                // empty --entry
			{"task", "check", "--ticket", "TEST-1", "--id", "99"},                        // unknown task id
			{"meta", "update", "--ticket", "TEST-1", "--field", "Bogus", "--value", "x"}, // unknown field
			{"meta", "update", "--doc", "does-not-exist.md", "--field", "Status", "--value", "x"},
			{"doc", "relate", "--ticket", "TEST-1", "--file-note", "malformed-no-note"}, // malformed file-note
			{"meta", "update", "--ticket", "TEST-1", "--field", "Bogus", "--value", "x", "--with-glaze-output"},
		}
		for _, args := range cases {
			out, err := runDocmgr(t, tmp, args...)
			if err == nil {
				t.Errorf("docmgr %s: expected non-zero exit, got success:\n%s", strings.Join(args, " "), out)
			}
		}
	})

	t.Run("verbose restores banner", func(t *testing.T) {
		out := mustSucceed(t, tmp, "--verbose", "meta", "update", "--ticket", "TEST-1", "--field", "Status", "--value", "active")
		if !strings.Contains(out, "Docs root:") {
			t.Errorf("--verbose did not restore the banner:\n%s", out)
		}
	})
}
