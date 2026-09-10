package operations

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func checkTest(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func fixture(t *testing.T) (string, func() ([]Change, error)) {
	t.Helper()
	dir := t.TempDir()
	checkTest(t, os.WriteFile(filepath.Join(dir, "tasks.md"), []byte("old tasks"), 0644))
	checkTest(t, os.WriteFile(filepath.Join(dir, "changelog.md"), []byte("old log"), 0644))
	return dir, func() ([]Change, error) {
		return []Change{{Path: "tasks.md", Before: Hash([]byte("old tasks")), After: []byte("new tasks")}, {Path: "changelog.md", Before: Hash([]byte("old log")), After: []byte("new log")}}, nil
	}
}
func TestDryRunRetryAndConflict(t *testing.T) {
	dir, plan := fixture(t)
	ctx := context.Background()
	s := Store{}
	r, err := s.Execute(ctx, dir, "test", map[string]string{"summary": "a"}, true, plan)
	checkTest(t, err)
	if r.State != "dry-run" {
		t.Fatal(r)
	}
	if _, err := os.Stat(filepath.Join(dir, directory)); !os.IsNotExist(err) {
		t.Fatal("dry-run wrote journal")
	}
	first, err := s.Execute(ctx, dir, "test", map[string]string{"summary": "a"}, false, plan)
	checkTest(t, err)
	second, err := s.Execute(ctx, dir, "test", map[string]string{"summary": "a"}, false, func() ([]Change, error) { t.Fatal("replayed plan"); return nil, nil })
	checkTest(t, err)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("receipts differ: %#v %#v", first, second)
	}
	if _, err := s.Execute(ctx, dir, "test", map[string]string{"summary": "b"}, false, plan); err == nil {
		t.Fatal("ID collision accepted")
	}
}
func TestRecoverEveryWriteBoundary(t *testing.T) {
	for _, boundary := range []string{"journal:prepared", "tasks.md", "changelog.md", "journal:committed"} {
		t.Run(boundary, func(t *testing.T) {
			dir, plan := fixture(t)
			ctx := context.Background()
			s := Store{BeforeWrite: func(path string) error {
				if path == boundary {
					return errors.New("injected " + path)
				}
				return nil
			}}
			_, err := s.Execute(ctx, dir, "test", map[string]string{"summary": "x"}, false, plan)
			if err == nil {
				t.Fatal("fault not injected")
			}
			if boundary != "journal:prepared" {
				var applied *ApplyError
				if !errors.As(err, &applied) {
					t.Fatalf("missing receipt: %v", err)
				}
			}
			r, err := (Store{}).Execute(ctx, dir, "test", map[string]string{"summary": "x"}, false, plan)
			checkTest(t, err)
			if r.State != "committed" {
				t.Fatal(r)
			}
			for path, want := range map[string]string{"tasks.md": "new tasks", "changelog.md": "new log"} {
				got, _, err := Read(dir, path)
				checkTest(t, err)
				if string(got) != want {
					t.Fatal(path)
				}
			}
		})
	}
}
func TestRecoveryDoesNotClobberHumanEdit(t *testing.T) {
	dir, plan := fixture(t)
	ctx := context.Background()
	_, err := (Store{BeforeWrite: func(path string) error {
		if path == "changelog.md" {
			return errors.New("fault")
		}
		return nil
	}}).Execute(ctx, dir, "test", true, false, plan)
	if err == nil {
		t.Fatal("expected fault")
	}
	checkTest(t, os.WriteFile(filepath.Join(dir, "changelog.md"), []byte("human edit"), 0644))
	if _, err := (Store{}).Execute(ctx, dir, "test", true, false, plan); err == nil || !strings.Contains(err.Error(), "revision conflict") {
		t.Fatalf("missing conflict: %v", err)
	}
	if _, err := (Store{}).Execute(ctx, dir, "other", true, false, plan); err == nil {
		t.Fatal("pending operation bypassed")
	}
	got, _, err := Read(dir, "changelog.md")
	checkTest(t, err)
	if string(got) != "human edit" {
		t.Fatal("human edit lost")
	}
}
func TestRejectPathsAndPreflight(t *testing.T) {
	ctx := context.Background()
	for _, name := range []string{"../outside", "/tmp/outside", "sources/file", "tasks.md"} {
		t.Run(name, func(t *testing.T) {
			dir, _ := fixture(t)
			if name == "tasks.md" {
				checkTest(t, os.Remove(filepath.Join(dir, name)))
				checkTest(t, os.Symlink(filepath.Join(t.TempDir(), "out"), filepath.Join(dir, name)))
			}
			_, err := (Store{}).Execute(ctx, dir, "test", true, false, func() ([]Change, error) { return []Change{{Path: name, Before: "missing", After: []byte("bad")}}, nil })
			if err == nil {
				t.Fatal("unsafe path accepted")
			}
			if _, err := os.Stat(filepath.Join(dir, directory)); !os.IsNotExist(err) {
				t.Fatal("preflight wrote journal")
			}
		})
	}
}
func TestLockCancellationAndConcurrentIdempotency(t *testing.T) {
	dir, plan := fixture(t)
	ctx := context.Background()
	entered := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan error, 1)
	go func() { finished <- WithLock(ctx, dir, func() error { close(entered); <-release; return nil }) }()
	<-entered
	wait, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	if err := WithLock(wait, dir, func() error { return nil }); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lock ignored cancellation: %v", err)
	}
	close(release)
	checkTest(t, <-finished)
	results := make(chan error, 8)
	for i := 0; i < 8; i++ {
		go func() { _, err := (Store{}).Execute(ctx, dir, "same", true, false, plan); results <- err }()
	}
	for i := 0; i < 8; i++ {
		checkTest(t, <-results)
	}
	records, err := Records(dir)
	checkTest(t, err)
	if len(records) != 1 {
		t.Fatal("duplicate records")
	}
}
