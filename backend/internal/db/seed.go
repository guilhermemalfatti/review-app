package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const CondoSlug = "cantegril"

type SeedConfig struct {
	InviteCode       string
	AdminEmail       string
	AdminPassword    string
	AdminDisplayName string
}

func Seed(ctx context.Context, pool *pgxpool.Pool, cfg SeedConfig) (condoID uuid.UUID, err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin seed tx: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO condos (name, slug, invite_code)
		VALUES ('Cantegril', $1, $2)
		ON CONFLICT (slug) DO UPDATE SET invite_code = EXCLUDED.invite_code
		RETURNING id
	`, CondoSlug, cfg.InviteCode).Scan(&condoID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("seed condo: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash admin password: %w", err)
	}

	var existingID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, cfg.AdminEmail).Scan(&existingID)
	if err == pgx.ErrNoRows {
		_, err = tx.Exec(ctx, `
			INSERT INTO users (condo_id, email, password_hash, display_name, role)
			VALUES ($1, $2, $3, $4, 'admin')
		`, condoID, cfg.AdminEmail, string(hash), cfg.AdminDisplayName)
		if err != nil {
			return uuid.Nil, fmt.Errorf("seed admin: %w", err)
		}
	} else if err != nil {
		return uuid.Nil, fmt.Errorf("lookup admin: %w", err)
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE users
			SET password_hash = $1, display_name = $2, role = 'admin', condo_id = $3
			WHERE id = $4
		`, string(hash), cfg.AdminDisplayName, condoID, existingID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("update admin: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit seed: %w", err)
	}
	return condoID, nil
}
