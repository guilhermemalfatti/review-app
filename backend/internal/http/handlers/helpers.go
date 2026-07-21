package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gmalfatti/indica/backend/internal/logging"
)

const maxJSONBody = 1 << 20 // 1 MiB

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// WriteServerError logs the underlying error (kept off the client response) and returns a 500.
func WriteServerError(w http.ResponseWriter, r *http.Request, publicMsg string, err error) {
	attrs := append(logging.RequestAttrs(r), "err", err)
	slog.ErrorContext(r.Context(), publicMsg, attrs...)
	WriteError(w, http.StatusInternalServerError, publicMsg)
}

// WriteServiceUnavailable logs the error and returns 503.
func WriteServiceUnavailable(w http.ResponseWriter, r *http.Request, publicMsg string, err error) {
	attrs := append(logging.RequestAttrs(r), "err", err)
	slog.ErrorContext(r.Context(), publicMsg, attrs...)
	WriteError(w, http.StatusServiceUnavailable, publicMsg)
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)
	dec := json.NewDecoder(r.Body)
	return dec.Decode(dst)
}

func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	at := strings.Index(email, "@")
	if at < 1 || at == len(email)-1 {
		return false
	}
	dot := strings.LastIndex(email, ".")
	return dot > at+1 && dot < len(email)-1
}

func validScore(score *int) bool {
	if score == nil {
		return true
	}
	return *score >= 1 && *score <= 5
}
