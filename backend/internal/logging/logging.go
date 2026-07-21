package logging

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// Setup configures the process-wide default slog logger.
// Production uses JSON lines; development uses human-readable text with Debug level.
func Setup(appEnv string) {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	var handler slog.Handler
	if appEnv == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}

func RequestID(ctx context.Context) string {
	if id := chimw.GetReqID(ctx); id != "" {
		return id
	}
	return ""
}

// RequestAttrs returns common HTTP fields for structured logs.
func RequestAttrs(r *http.Request) []any {
	return []any{
		"request_id", RequestID(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
		"remote", r.RemoteAddr,
	}
}
