package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gmalfatti/indica/backend/internal/logging"
)

// RequestLogger logs each request with method, path, status, duration, and request id.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)

		level := slog.LevelInfo
		if ww.status >= 500 {
			level = slog.LevelError
		} else if ww.status >= 400 {
			level = slog.LevelWarn
		}

		attrs := append(logging.RequestAttrs(r),
			"status", ww.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"bytes", ww.bytes,
		)
		slog.Log(r.Context(), level, "request", attrs...)
	})
}

// Recoverer logs panics with a stack trace and returns 500.
func Recoverer(writeErr ErrorWriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					attrs := append(logging.RequestAttrs(r),
						"panic", rec,
						"stack", string(debug.Stack()),
					)
					slog.Error("panic recovered", attrs...)
					writeErr(w, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}
