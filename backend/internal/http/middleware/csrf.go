package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
)

const (
	CSRFCookieName = "csrf"
	CSRFHeaderName = "X-CSRF-Token"
)

func CSRFTokenHandler(cookieSecure bool, writeErr ErrorWriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := ensureCSRFCookie(w, r, cookieSecure)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "failed to issue csrf token")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"csrf_token": token})
	}
}

func CSRF(writeErr ErrorWriter) func(http.Handler) http.Handler {
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
