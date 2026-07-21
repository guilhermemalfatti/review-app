package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const CookieName = "session"

// DummyPasswordHash is a valid bcrypt hash used to mitigate timing leaks when a user is missing.
const DummyPasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

type User struct {
	ID                 uuid.UUID `json:"id"`
	Email              string    `json:"email"`
	DisplayName        string    `json:"display_name"`
	Role               string    `json:"role"`
	CondoID            uuid.UUID `json:"condo_id"`
	MustChangePassword bool      `json:"must_change_password"`
}

type SessionStore struct {
	pool        *pgxpool.Pool
	sessionDays int
}

func NewSessionStore(pool *pgxpool.Pool, sessionDays int) *SessionStore {
	return &SessionStore{pool: pool, sessionDays: sessionDays}
}

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *SessionStore) Create(ctx context.Context, userID uuid.UUID) (token string, expiresAt time.Time, err error) {
	token, err = newToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt = time.Now().UTC().Add(time.Duration(s.sessionDays) * 24 * time.Hour)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, hashToken(token), expiresAt)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("create session: %w", err)
	}
	return token, expiresAt, nil
}

func (s *SessionStore) GetUser(ctx context.Context, token string) (*User, error) {
	if token == "" {
		return nil, pgx.ErrNoRows
	}
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.email, u.display_name, u.role, u.condo_id, u.must_change_password
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > now()
	`, hashToken(token)).Scan(&u.ID, &u.Email, &u.DisplayName, &u.Role, &u.CondoID, &u.MustChangePassword)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *SessionStore) Delete(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, hashToken(token))
	return err
}

func (s *SessionStore) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

func GenerateTemporaryPassword() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	const n = len(alphabet)
	const maxUnbiased = 256 - (256 % n)

	out := make([]byte, 12)
	for i := 0; i < len(out); {
		var b [1]byte
		if _, err := rand.Read(b[:]); err != nil {
			return "", err
		}
		if int(b[0]) >= maxUnbiased {
			continue
		}
		out[i] = alphabet[int(b[0])%n]
		i++
	}
	return string(out), nil
}

func (s *SessionStore) DeleteExpired(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= now()`)
	return err
}

type contextKey string

const userContextKey contextKey = "user"

func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

func UserFromContext(ctx context.Context) *User {
	u, _ := ctx.Value(userContextKey).(*User)
	return u
}
