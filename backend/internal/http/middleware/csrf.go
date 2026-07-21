package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gmalfatti/indica/backend/internal/logging"
)

const (
	CSRFCookieName = "csrf"
	CSRFHeaderName = "X-CSRF-Token"
)

func CSRFTokenHandler(cookieSecure bool, writeErr ErrorWriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := ensureCSRFCookie(w, r, cookieSecure)
		if err != nil {
			attrs := append(logging.RequestAttrs(r), "err", err)
			slog.Error("failed to issue csrf token", attrs...)
			writeErr(w, http.StatusInternalServerError, "failed to issue csrf token")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"csrf_token": token})
	}
}

// CSRF protects mutating /api requests.
//
// Preferred check (works when cross-site cookies are blocked, e.g. Safari on
// GitHub Pages → Render): Origin/Referer must match allowedOrigin.
// Fallback: classic double-submit cookie (csrf cookie == X-CSRF-Token header).
func CSRF(allowedOrigin string, writeErr ErrorWriter) func(http.Handler) http.Handler {
	allowedOrigin = strings.TrimRight(strings.TrimSpace(allowedOrigin), "/")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				next.ServeHTTP(w, r)
				return
			}

			if !strings.HasPrefix(r.URL.Path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}

			if originAllowed(r, allowedOrigin) {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(CSRFCookieName)
			header := r.Header.Get(CSRFHeaderName)
			if err != nil || cookie.Value == "" || header == "" ||
				subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(header)) != 1 {
				writeErr(w, http.StatusForbidden, "csrf validation failed")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func originAllowed(r *http.Request, allowedOrigin string) bool {
	if allowedOrigin == "" {
		return false
	}
	if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
		return origin == allowedOrigin
	}
	ref := strings.TrimSpace(r.Header.Get("Referer"))
	if ref == "" {
		return false
	}
	u, err := url.Parse(ref)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return u.Scheme+"://"+u.Host == allowedOrigin
}

func ensureCSRFCookie(w http.ResponseWriter, r *http.Request, cookieSecure bool) (string, error) {
	if c, err := r.Cookie(CSRFCookieName); err == nil && c.Value != "" {
		return c.Value, nil
	}

	token, err := newCSRFToken()
	if err != nil {
		return "", err
	}
	setCSRFCookie(w, token, cookieSecure)
	return token, nil
}

func newCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func setCSRFCookie(w http.ResponseWriter, token string, cookieSecure bool) {
	sameSite := http.SameSiteLaxMode
	if cookieSecure {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		SameSite: sameSite,
		Secure:   cookieSecure,
	})
}
