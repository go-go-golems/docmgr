package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCursorRoundTrip(t *testing.T) {
	t.Parallel()

	c, err := encodeCursor(200)
	if err != nil {
		t.Fatalf("encodeCursor: %v", err)
	}
	got, err := decodeCursor(c)
	if err != nil {
		t.Fatalf("decodeCursor: %v", err)
	}
	if got != 200 {
		t.Fatalf("expected 200, got %d", got)
	}
}

func TestServer_IndexNotReady(t *testing.T) {
	t.Parallel()

	mgr := NewIndexManager("ttmp")
	s := NewServer(mgr, ServerOptions{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/docs?query=test", nil)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected %d, got %d (%s)", http.StatusServiceUnavailable, rr.Code, rr.Body.String())
	}
}

func TestServer_SearchPathFiltersRejectUnsafeValuesBeforeIndex(t *testing.T) {
	t.Parallel()

	mgr := NewIndexManager("ttmp")
	s := NewServer(mgr, ServerOptions{})

	cases := []string{
		"/api/v1/search/docs?file=../secret.txt",
		"/api/v1/search/docs?dir=/etc",
		"/api/v1/search/docs?file=repo://pkg/foo.go",
		"/api/v1/search/docs?file=C:/Windows/win.ini",
	}
	for _, path := range cases {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected %d, got %d (%s)", path, http.StatusBadRequest, rr.Code, rr.Body.String())
		}
	}
}
