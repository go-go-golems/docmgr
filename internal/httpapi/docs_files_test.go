package httpapi

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveFileWithin_DisallowsTraversal(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outside := t.TempDir()
	outsideFile := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("nope"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}

	_, _, _, err := resolveFileWithin(root, filepath.Join("..", filepath.Base(outside), "secret.txt"))
	if err == nil {
		t.Fatalf("expected error")
	}
	var he *HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("expected HTTPError, got %T: %v", err, err)
	}
	if he.Status != http.StatusForbidden {
		t.Fatalf("expected %d, got %d (%s)", http.StatusForbidden, he.Status, he.Code)
	}
}

func TestResolveFileWithin_DisallowsSymlinkEscape(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}

	outsideFile := filepath.Join(parent, "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}

	link := filepath.Join(root, "link.txt")
	if err := os.Symlink(outsideFile, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	_, _, _, err := resolveFileWithin(root, "link.txt")
	if err == nil {
		t.Fatalf("expected error")
	}
	var he *HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("expected HTTPError, got %T: %v", err, err)
	}
	if he.Status != http.StatusForbidden {
		t.Fatalf("expected %d, got %d (%s)", http.StatusForbidden, he.Status, he.Code)
	}
}

func TestReadTextFile_RejectsBinary(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "bin.dat")
	if err := os.WriteFile(path, []byte{0, 1, 2, 3}, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, _, err := readTextFile(path, 100)
	if err == nil {
		t.Fatalf("expected error")
	}
	var he *HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("expected HTTPError, got %T: %v", err, err)
	}
	if he.Status != http.StatusUnsupportedMediaType {
		t.Fatalf("expected %d, got %d (%s)", http.StatusUnsupportedMediaType, he.Status, he.Code)
	}
}

func TestReadTextFile_TruncatesLargeFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "big.txt")
	content := strings.Repeat("a", 50)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, truncated, err := readTextFile(path, 20)
	if err != nil {
		t.Fatalf("readTextFile: %v", err)
	}
	if !truncated {
		t.Fatalf("expected truncated")
	}
	if len(got) != 20 {
		t.Fatalf("expected 20 bytes, got %d", len(got))
	}
}
