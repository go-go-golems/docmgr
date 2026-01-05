package web

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

type SPAOptions struct {
	APIPrefixes []string
	IndexPath   string
}

func NewSPAHandler(public fs.FS, opts SPAOptions) http.Handler {
	if opts.IndexPath == "" {
		opts.IndexPath = "index.html"
	}
	if len(opts.APIPrefixes) == 0 {
		opts.APIPrefixes = []string{"/api"}
	}

	fileServer := http.FileServer(http.FS(public))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, p := range opts.APIPrefixes {
			if strings.HasPrefix(r.URL.Path, p) {
				http.NotFound(w, r)
				return
			}
		}

		reqPath := path.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
		reqPath = strings.TrimPrefix(reqPath, "/")
		if reqPath == "" {
			reqPath = opts.IndexPath
		}

		f, err := public.Open(reqPath)
		if err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback.
		index, err := public.Open(opts.IndexPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("web ui not available: %v", err), http.StatusNotFound)
			return
		}
		_ = index.Close()

		r2 := *r
		u := *r.URL
		u.Path = "/"
		u.RawPath = ""
		r2.URL = &u
		fileServer.ServeHTTP(w, &r2)
	})
}
