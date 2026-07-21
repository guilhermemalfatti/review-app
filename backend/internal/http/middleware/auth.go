package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gmalfatti/indica/backend/internal/auth"
	"github.com/gmalfatti/indica/backend/internal/logging"
	"github.com/jackc/pgx/v5"
)

type ErrorWriter func(w http.ResponseWriter, status int, msg string)

func RequireAuth(sessions *auth.SessionStore, writeErr ErrorWriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(auth.CookieName)
			if err != nil || c.Value == "" {
				reason := "session_cookie_missing"
				if err == nil && c.Value == "" {
					reason = "session_cookie_empty"
				} else if err != nil && !errors.Is(err, http.ErrNoCookie) {
					reason = "session_cookie_read_error"
				}
				logUnauthorized(r, reason, err)
				writeErr(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			user, err := sessions.GetUser(r.Context(), c.Value)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					logUnauthorized(r, "session_not_found_or_expired", err)
					writeErr(w, http.StatusUnauthorized, "unauthorized")
					return
				}
				attrs := append(logging.RequestAttrs(r), "err", err)
				slog.Error("session lookup failed", attrs...)
				writeErr(w, http.StatusServiceUnavailable, "service unavailable")
				return
			}
			ctx := auth.WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func logUnauthorized(r *http.Request, reason string, err error) {
	cookieHeader := r.Header.Get("Cookie")
	attrs := append(logging.RequestAttrs(r),
		"reason", reason,
		"origin", r.Header.Get("Origin"),
		"referer", r.Header.Get("Referer"),
		"user_agent", r.UserAgent(),
		"has_cookie_header", cookieHeader != "",
		"cookie_names", cookieNames(cookieHeader),
	)
	if err != nil {
		attrs = append(attrs, "err", err)
	}
	slog.Warn("unauthorized", attrs...)
}

func cookieNames(header string) []string {
	if header == "" {
		return nil
	}
	parts := strings.Split(header, ";")
	names := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		name, _, _ := strings.Cut(p, "=")
		if name = strings.TrimSpace(name); name != "" {
			names = append(names, name)
		}
	}
	return names
}

func RequireAdmin(writeErr ErrorWriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := auth.UserFromContext(r.Context())
			if user == nil || user.Role != "admin" {
				writeErr(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequirePasswordChanged blocks write actions until the user sets a new password after an admin reset.
func RequirePasswordChanged(writeErr ErrorWriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := auth.UserFromContext(r.Context())
			if user != nil && user.MustChangePassword {
				writeErr(w, http.StatusForbidden, "password change required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuth loads the user if a valid session cookie is present; otherwise continues anonymously.
func OptionalAuth(sessions *auth.SessionStore, writeErr ErrorWriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(auth.CookieName)
			if err == nil && c.Value != "" {
				user, err := sessions.GetUser(r.Context(), c.Value)
				if err != nil {
					if !errors.Is(err, pgx.ErrNoRows) {
						attrs := append(logging.RequestAttrs(r), "err", err)
						slog.Error("optional auth session lookup failed", attrs...)
						writeErr(w, http.StatusServiceUnavailable, "service unavailable")
						return
					}
				} else {
					r = r.WithContext(auth.WithUser(r.Context(), user))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
