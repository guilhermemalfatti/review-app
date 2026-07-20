package middleware

import (
	"net/http"

	"github.com/gmalfatti/indica/backend/internal/auth"
)

type ErrorWriter func(w http.ResponseWriter, status int, msg string)

func RequireAuth(sessions *auth.SessionStore, writeErr ErrorWriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(auth.CookieName)
			if err != nil || c.Value == "" {
				writeErr(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			user, err := sessions.GetUser(r.Context(), c.Value)
			if err != nil {
				writeErr(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			ctx := auth.WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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

// OptionalAuth loads the user if a valid session cookie is present; otherwise continues anonymously.
func OptionalAuth(sessions *auth.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(auth.CookieName)
			if err == nil && c.Value != "" {
				if user, err := sessions.GetUser(r.Context(), c.Value); err == nil {
					r = r.WithContext(auth.WithUser(r.Context(), user))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
