package scenariolog

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type ExecStepSpec struct {
	RunID      string
	RootDir    string // used for making artifact paths portable (store root-rel when possible)
	LogDir     string // relative to RootDir unless absolute
	StepNum    int
	StepName   string
	ScriptPath string
	Command    []string // argv
}

type ExecStepResult struct {
	StepID       string
	ExitCode     int
	DurationMs   int64
	StdoutPath   string // stored path (root-rel when possible)
	StderrPath   string // stored path (root-rel when possible)
	StdoutSHA256 string
	StderrSHA256 string
}

func ExecStep(ctx context.Context, db *sql.DB, spec ExecStepSpec) (*ExecStepResult, error) {
	if spec.RunID == "" {
		return nil, errors.New("ExecStep: RunID is required")
	}
	if spec.StepNum < 0 {
		return nil, errors.New("ExecStep: StepNum must be >= 0")
	}
	if spec.StepName == "" {
		return nil, errors.New("ExecStep: StepName is required")
	}
	if len(spec.Command) == 0 {
		return nil, errors.New("ExecStep: Command is required")
	}

	startedAt := time.Now()
	stepID := fmt.Sprintf("%s-step-%02d", spec.RunID, spec.StepNum)

	// Record step start
	if _, err := db.ExecContext(ctx,
		`INSERT INTO steps (step_id, run_id, step_num, step_name, script_path, started_at)
		 VALUES (?, ?, ?, ?, ?, ?);`,
		stepID,
		spec.RunID,
		spec.StepNum,
		spec.StepName,
		nullIfEmpty(spec.ScriptPath),
		startedAt.UTC().Format(time.RFC3339Nano),
	); err != nil {
		return nil, errors.Wrap(err, "insert steps")
	}

	// Best-effort metadata for debugging/repro.
	_ = SetKV(ctx, db, spec.RunID, stepID, "", "step.name", spec.StepName)
	_ = SetKV(ctx, db, spec.RunID, stepID, "", "step.num", fmt.Sprintf("%d", spec.StepNum))
	_ = SetKV(ctx, db, spec.RunID, stepID, "", "step.script_path", spec.ScriptPath)
	_ = SetKV(ctx, db, spec.RunID, stepID, "", "cmd.argv0", spec.Command[0])
	if b, err := json.Marshal(spec.Command[1:]); err == nil {
		_ = SetKV(ctx, db, spec.RunID, stepID, "", "cmd.args_json", string(b))
	}

	stdoutAbs, stderrAbs, err := stepLogPaths(spec)
	if err != nil {
		_ = finalizeStep(ctx, db, stepID, startedAt, time.Now(), 127)
		return nil, err
	}

	stdoutFile, err := os.Create(stdoutAbs)
	if err != nil {
		_ = finalizeStep(ctx, db, stepID, startedAt, time.Now(), 127)
		return nil, errors.Wrap(err, "create stdout log file")
	}
	defer func() { _ = stdoutFile.Close() }()

	stderrFile, err := os.Create(stderrAbs)
	if err != nil {
		_ = finalizeStep(ctx, db, stepID, startedAt, time.Now(), 127)
		return nil, errors.Wrap(err, "create stderr log file")
	}
	defer func() { _ = stderrFile.Close() }()

	cmd := exec.CommandContext(ctx, spec.Command[0], spec.Command[1:]...)
	if spec.RootDir != "" {
		cmd.Dir = spec.RootDir
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		_ = finalizeStep(ctx, db, stepID, startedAt, time.Now(), 127)
		return nil, errors.Wrap(err, "StdoutPipe")
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		_ = finalizeStep(ctx, db, stepID, startedAt, time.Now(), 127)
		return nil, errors.Wrap(err, "StderrPipe")
	}

	if err := cmd.Start(); err != nil {
		_ = finalizeStep(ctx, db, stepID, startedAt, time.Now(), 127)
		return nil, errors.Wrap(err, "start command")
	}

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		_, err := io.Copy(stdoutFile, stdoutPipe)
		return errors.Wrap(err, "copy stdout")
	})
	eg.Go(func() error {
		_, err := io.Copy(stderrFile, stderrPipe)
		return errors.Wrap(err, "copy stderr")
	})
	waitErr := cmd.Wait()
	copyErr := eg.Wait()

	completedAt := time.Now()
	exitCode := exitCodeFromWaitErr(waitErr)

	// Flush + close files so hashing sees full content.
	_ = stdoutFile.Sync()
	_ = stderrFile.Sync()
	_ = stdoutFile.Close()
	_ = stderrFile.Close()

	if copyErr != nil && egCtx.Err() == nil {
		// Copy errors are unexpected; still finalize the step and return the error.
		_ = finalizeStep(ctx, db, stepID, startedAt, completedAt, exitCode)
		return nil, copyErr
	}

	if err := finalizeStep(ctx, db, stepID, startedAt, completedAt, exitCode); err != nil {
		return nil, err
	}

	stdoutSHA, stdoutSize, err := fileSHA256AndSize(stdoutAbs)
	if err != nil {
		return nil, err
	}
	stderrSHA, stderrSize, err := fileSHA256AndSize(stderrAbs)
	if err != nil {
		return nil, err
	}

	stdoutStored := storePathBestEffort(spec.RootDir, stdoutAbs)
	stderrStored := storePathBestEffort(spec.RootDir, stderrAbs)

	stdoutArtifactID, err := insertArtifact(ctx, db, spec.RunID, stepID, "", "stdout", stdoutStored, true, stdoutSize, stdoutSHA)
	if err != nil {
		return nil, err
	}
	stderrArtifactID, err := insertArtifact(ctx, db, spec.RunID, stepID, "", "stderr", stderrStored, true, stderrSize, stderrSHA)
	if err != nil {
		return nil, err
	}

	// Best-effort: index into FTS if table exists (degraded mode is allowed).
	_ = indexArtifactLinesFTS(ctx, db, spec.RunID, stdoutArtifactID, stdoutAbs)
	_ = indexArtifactLinesFTS(ctx, db, spec.RunID, stderrArtifactID, stderrAbs)

	return &ExecStepResult{
		StepID:       stepID,
		ExitCode:     exitCode,
		DurationMs:   completedAt.Sub(startedAt).Milliseconds(),
		StdoutPath:   stdoutStored,
		StderrPath:   stderrStored,
		StdoutSHA256: stdoutSHA,
		StderrSHA256: stderrSHA,
	}, nil
}

func stepLogPaths(spec ExecStepSpec) (stdoutAbs string, stderrAbs string, _ error) {
	logDir := spec.LogDir
	if logDir == "" {
		return "", "", errors.New("ExecStep: LogDir is required (caller must create it)")
	}

	if !filepath.IsAbs(logDir) {
		if spec.RootDir == "" {
			return "", "", errors.New("ExecStep: RootDir is required when LogDir is relative")
		}
		logDir = filepath.Join(spec.RootDir, logDir)
	}

	stdoutAbs = filepath.Join(logDir, fmt.Sprintf("step-%02d-stdout.txt", spec.StepNum))
	stderrAbs = filepath.Join(logDir, fmt.Sprintf("step-%02d-stderr.txt", spec.StepNum))
	return stdoutAbs, stderrAbs, nil
}

func finalizeStep(ctx context.Context, db *sql.DB, stepID string, startedAt, completedAt time.Time, exitCode int) error {
	durationMs := int64(completedAt.Sub(startedAt).Milliseconds())
	if durationMs < 0 {
		durationMs = 0
	}
	_, err := db.ExecContext(ctx,
		`UPDATE steps
		 SET completed_at = ?, exit_code = ?, duration_ms = ?
		 WHERE step_id = ?;`,
		completedAt.UTC().Format(time.RFC3339Nano),
		exitCode,
		durationMs,
		stepID,
	)
	if err != nil {
		return errors.Wrap(err, "update steps completion")
	}
	return nil
}

func exitCodeFromWaitErr(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	// Something else (signal, context cancellation, start error already handled).
	return 127
}

func fileSHA256AndSize(path string) (sha string, sizeBytes int64, _ error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, errors.Wrap(err, "open file for sha256")
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, errors.Wrap(err, "hash file")
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func storePathBestEffort(rootDir, absPath string) string {
	if rootDir == "" {
		return absPath
	}
	rel, err := filepath.Rel(rootDir, absPath)
	if err != nil {
		return absPath
	}
	return filepath.ToSlash(rel)
}


