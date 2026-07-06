package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

var pngMagic = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 13, 'I', 'H', 'D', 'R'}

func newRawTestServer(t *testing.T) (*Server, string) {
	t.Helper()

	root := filepath.Join(t.TempDir(), "ttmp")
	ticketDir := filepath.Join(root, "2026", "01", "03", "RAW-1--assets")
	mustMkdirAll(t, filepath.Join(ticketDir, "images"))

	mustWriteFile(t, filepath.Join(ticketDir, "index.md"), `---
Title: Raw Assets
Ticket: RAW-1
Status: active
DocType: index
---

# Raw Assets
`)
	if err := os.WriteFile(filepath.Join(ticketDir, "images", "pic.png"), pngMagic, 0o644); err != nil {
		t.Fatalf("write png: %v", err)
	}

	mgr := NewIndexManager(root)
	if _, err := mgr.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	return NewServer(mgr, ServerOptions{}), root
}

func TestFilesRaw_ServesBytesWithContentType(t *testing.T) {
	t.Parallel()

	s, _ := newRawTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/raw?root=docs&path=2026/01/03/RAW-1--assets/images/pic.png", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("expected image/png, got %q", ct)
	}
	if !bytes.Equal(rr.Body.Bytes(), pngMagic) {
		t.Fatalf("body mismatch: got %v", rr.Body.Bytes())
	}
}

func TestFilesRaw_SniffsContentTypeWithoutExtension(t *testing.T) {
	t.Parallel()

	s, root := newRawTestServer(t)
	if err := os.WriteFile(filepath.Join(root, "2026", "01", "03", "RAW-1--assets", "images", "noext"), pngMagic, 0o644); err != nil {
		t.Fatalf("write noext: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/raw?root=docs&path=2026/01/03/RAW-1--assets/images/noext", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("expected sniffed image/png, got %q", ct)
	}
}

func TestFilesRaw_ForbidsTraversal(t *testing.T) {
	t.Parallel()

	s, root := newRawTestServer(t)
	// A real file one level above the docs root that must not be reachable
	// with root=docs.
	if err := os.WriteFile(filepath.Join(filepath.Dir(root), "secret.txt"), []byte("nope"), 0o644); err != nil {
		t.Fatalf("write secret: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/raw?root=docs&path=../secret.txt", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d (%s)", http.StatusForbidden, rr.Code, rr.Body.String())
	}
}

func TestFilesRaw_ForbidsSymlinkEscape(t *testing.T) {
	t.Parallel()

	s, root := newRawTestServer(t)

	outside := filepath.Join(filepath.Dir(root), "outside.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o644); err != nil {
		t.Fatalf("write outside: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link.txt")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/raw?root=docs&path=link.txt", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d (%s)", http.StatusForbidden, rr.Code, rr.Body.String())
	}
}

func TestFilesRaw_RejectsOversizedFiles(t *testing.T) {
	t.Parallel()

	s, root := newRawTestServer(t)

	big := filepath.Join(root, "2026", "01", "03", "RAW-1--assets", "images", "big.bin")
	f, err := os.Create(big) // #nosec G304 -- test fixture path
	if err != nil {
		t.Fatalf("create big: %v", err)
	}
	// Sparse file just over the cap.
	if err := f.Truncate(maxRawFileBytes + 1); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/raw?root=docs&path=2026/01/03/RAW-1--assets/images/big.bin", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected %d, got %d (%s)", http.StatusRequestEntityTooLarge, rr.Code, rr.Body.String())
	}
}
