package handlers

import (
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
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cookieSecure,
		Expires:  expiresAt,
	})
}

func (h *AuthHandler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cookieSecure,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := DecodeJSON(r, &req); err != nil {
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
	if req.InviteCode != h.inviteCode {
		WriteError(w, http.StatusForbidden, "invalid invite code")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	var user auth.User
	err = h.pool.QueryRow(r.Context(), `
		INSERT INTO users (condo_id, email, password_hash, display_name, role)
		VALUES ($1, $2, $3, $4, 'resident')
		RETURNING id, email, display_name, role, condo_id
	`, h.condoID, req.Email, hash, req.DisplayName).Scan(
		&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.CondoID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			WriteError(w, http.StatusConflict, "email already registered")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	token, expiresAt, err := h.sessions.Create(r.Context(), user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	h.setSessionCookie(w, token, expiresAt)
	WriteJSON(w, http.StatusCreated, userResponse{User: &user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := DecodeJSON(r, &req); err != nil {
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
		SELECT id, email, display_name, role, condo_id, password_hash
		FROM users WHERE email = $1
	`, req.Email).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.CondoID, &passwordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to lookup user")
		return
	}

	if !auth.CheckPassword(passwordHash, req.Password) {
		WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, expiresAt, err := h.sessions.Create(r.Context(), user.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	h.setSessionCookie(w, token, expiresAt)
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
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	WriteJSON(w, http.StatusOK, userResponse{User: user})
}
