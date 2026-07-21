package httpserver

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// MountSPA serves the Vite build from staticDir for non-/api routes.
// Missing files fall back to index.html (client-side routing).
// No-op when staticDir is empty or missing (API-only / local Go runs).
func MountSPA(mux interface {
	Get(pattern string, handlerFn http.HandlerFunc)
	Head(pattern string, handlerFn http.HandlerFunc)
}, staticDir string) {
	staticDir = strings.TrimSpace(staticDir)
	if staticDir == "" {
		return
	}
	abs, err := filepath.Abs(staticDir)
	if err != nil {
		slog.Warn("static dir invalid, SPA disabled", "static_dir", staticDir, "err", err)
		return
	}
	if st, err := os.Stat(abs); err != nil || !st.IsDir() {
		slog.Warn("static dir missing, SPA disabled", "static_dir", abs, "err", err)
		return
	}
	slog.Info("serving SPA", "static_dir", abs)
	h := spaHandler(abs)
	mux.Get("/", h)
	mux.Head("/", h)
	mux.Get("/*", h)
	mux.Head("/*", h)
}

func spaHandler(root string) http.HandlerFunc {
	index := filepath.Join(root, "index.html")
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			http.NotFound(w, r)
			return
		}

		rel := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
		if rel == "." || rel == "" {
			http.ServeFile(w, r, index)
			return
		}

		full := filepath.Join(root, rel)
		if !strings.HasPrefix(full, root+string(os.PathSeparator)) && full != root {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		st, err := os.Stat(full)
		if err == nil && !st.IsDir() {
			if strings.HasPrefix(rel, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			http.ServeFile(w, r, full)
			return
		}

		http.ServeFile(w, r, index)
	}
}
