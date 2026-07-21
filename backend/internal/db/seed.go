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

	_, err = tx.Exec(ctx, `
		INSERT INTO condos (name, slug, invite_code)
		VALUES ('Cantegril', $1, $2)
		ON CONFLICT (slug) DO NOTHING
	`, CondoSlug, cfg.InviteCode)
	if err != nil {
		return uuid.Nil, fmt.Errorf("seed condo: %w", err)
	}

	err = tx.QueryRow(ctx, `SELECT id FROM condos WHERE slug = $1`, CondoSlug).Scan(&condoID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("lookup condo: %w", err)
	}

	var existingID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT id FROM users WHERE condo_id = $1 AND email = $2
	`, condoID, cfg.AdminEmail).Scan(&existingID)
	if err == pgx.ErrNoRows {
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
		if hashErr != nil {
			return uuid.Nil, fmt.Errorf("hash admin password: %w", hashErr)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO users (condo_id, email, password_hash, display_name, role)
			VALUES ($1, $2, $3, $4, 'admin')
		`, condoID, cfg.AdminEmail, string(hash), cfg.AdminDisplayName)
		if err != nil {
			return uuid.Nil, fmt.Errorf("seed admin: %w", err)
		}
	} else if err != nil {
		return uuid.Nil, fmt.Errorf("lookup admin: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit seed: %w", err)
	}
	return condoID, nil
}
