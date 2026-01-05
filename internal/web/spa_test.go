package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestSPAHandler_FallbackIndex(t *testing.T) {
	t.Parallel()

	public := fstest.MapFS{
		"index.html":    &fstest.MapFile{Data: []byte("<html>ok</html>")},
		"assets/app.js": &fstest.MapFile{Data: []byte("console.log('ok')")},
	}

	h := NewSPAHandler(fs.FS(public), SPAOptions{APIPrefixes: []string{"/api"}})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/some/route", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "<html>ok</html>" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}
