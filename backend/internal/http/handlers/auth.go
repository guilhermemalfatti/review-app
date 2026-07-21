package handlers

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gmalfatti/indica/backend/internal/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	pool         *pgxpool.Pool
	sessions     *auth.SessionStore
	condoID      uuid.UUID
	inviteCode   string
	cookieSecure bool
}

func NewAuthHandler(pool *pgxpool.Pool, sessions *auth.SessionStore, condoID uuid.UUID, inviteCode string, cookieSecure bool) *AuthHandler {
	return &AuthHandler{
		pool:         pool,
		sessions:     sessions,
		condoID:      condoID,
		inviteCode:   inviteCode,
		cookieSecure: cookieSecure,
	}
}

type signupRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	InviteCode  string `json:"invite_code"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	User *auth.User `json:"user"`
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	sameSite := http.SameSiteLaxMode
	if h.cookieSecure {
		// Cross-site SPA (e.g. GitHub Pages → Render) needs None; Secure.
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   h.cookieSecure,
		Expires:  expiresAt,
	})
}

func (h *AuthHandler) clearSessionCookie(w http.ResponseWriter) {
	sameSite := http.SameSiteLaxMode
	if h.cookieSecure {
		sameSite = http.SameSiteNoneMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   h.cookieSecure,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func inviteCodesEqual(a, b string) bool {
	ha := sha256.Sum256([]byte(a))
	hb := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(ha[:], hb[:]) == 1
}

func (h *AuthHandler) createExclusiveSession(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	if err := h.sessions.DeleteAllForUser(r.Context(), userID); err != nil {
		return err
	}
	token, expiresAt, err := h.sessions.Create(r.Context(), userID)
	if err != nil {
		return err
	}
	h.setSessionCookie(w, token, expiresAt)
	return nil
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := DecodeJSON(w, r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.InviteCode = strings.TrimSpace(req.InviteCode)

	if !isValidEmail(req.Email) {
		WriteError(w, http.StatusBadRequest, "invalid email")
		return
	}
	if len(req.Password) < 8 {
		WriteError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if req.DisplayName == "" {
		WriteError(w, http.StatusBadRequest, "display_name is required")
		return
	}
	if !inviteCodesEqual(req.InviteCode, h.inviteCode) {
		WriteError(w, http.StatusForbidden, "invalid invite code")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		WriteServerError(w, r, "failed to hash password", err)
		return
	}

	var user auth.User
	err = h.pool.QueryRow(r.Context(), `
		INSERT INTO users (condo_id, email, password_hash, display_name, role)
		VALUES ($1, $2, $3, $4, 'resident')
		RETURNING id, email, display_name, role, condo_id, must_change_password
	`, h.condoID, req.Email, hash, req.DisplayName).Scan(
		&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.CondoID, &user.MustChangePassword,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			WriteError(w, http.StatusConflict, "email already registered")
			return
		}
		WriteServerError(w, r, "failed to create user", err)
		return
	}

	if err := h.createExclusiveSession(w, r, user.ID); err != nil {
		WriteServerError(w, r, "failed to create session", err)
		return
	}
	WriteJSON(w, http.StatusCreated, userResponse{User: &user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := DecodeJSON(w, r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	var user auth.User
	var passwordHash string
	err := h.pool.QueryRow(r.Context(), `
		SELECT id, email, display_name, role, condo_id, must_change_password, password_hash
		FROM users WHERE condo_id = $1 AND email = $2
	`, h.condoID, req.Email).Scan(
		&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.CondoID, &user.MustChangePassword, &passwordHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_ = auth.CheckPassword(auth.DummyPasswordHash, req.Password)
			WriteError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		WriteServerError(w, r, "failed to lookup user", err)
		return
	}

	if !auth.CheckPassword(passwordHash, req.Password) {
		WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := h.createExclusiveSession(w, r, user.ID); err != nil {
		WriteServerError(w, r, "failed to create session", err)
		return
	}
	WriteJSON(w, http.StatusOK, userResponse{User: &user})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(auth.CookieName); err == nil && c.Value != "" {
		_ = h.sessions.Delete(r.Context(), c.Value)
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(auth.CookieName)
	if err != nil || c.Value == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.sessions.GetUser(r.Context(), c.Value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		WriteServiceUnavailable(w, r, "service unavailable", err)
		return
	}
	WriteJSON(w, http.StatusOK, userResponse{User: user})
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req changePasswordRequest
	if err := DecodeJSON(w, r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.CurrentPassword == "" || len(req.NewPassword) < 8 {
		WriteError(w, http.StatusBadRequest, "current_password and new_password (min 8) are required")
		return
	}
	if req.CurrentPassword == req.NewPassword {
		WriteError(w, http.StatusBadRequest, "new password must be different")
		return
	}

	var passwordHash string
	err := h.pool.QueryRow(r.Context(), `
		SELECT password_hash FROM users WHERE id = $1
	`, user.ID).Scan(&passwordHash)
	if err != nil {
		WriteServerError(w, r, "failed to lookup user", err)
		return
	}
	if !auth.CheckPassword(passwordHash, req.CurrentPassword) {
		WriteError(w, http.StatusUnauthorized, "invalid current password")
		return
	}

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		WriteServerError(w, r, "failed to hash password", err)
		return
	}

	err = h.pool.QueryRow(r.Context(), `
		UPDATE users
		SET password_hash = $1, must_change_password = false
		WHERE id = $2
		RETURNING id, email, display_name, role, condo_id, must_change_password
	`, hash, user.ID).Scan(
		&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.CondoID, &user.MustChangePassword,
	)
	if err != nil {
		WriteServerError(w, r, "failed to update password", err)
		return
	}

	if err := h.createExclusiveSession(w, r, user.ID); err != nil {
		WriteServerError(w, r, "failed to create session", err)
		return
	}
	WriteJSON(w, http.StatusOK, userResponse{User: user})
}
