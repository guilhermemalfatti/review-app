package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ResetDB truncates all application data. Schema and migrations stay intact.
// Call before Seed / SeedDemo when RESET_DB=true.
func ResetDB(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE
			reviews,
			providers,
			sessions,
			users,
			condos
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		return fmt.Errorf("reset db: %w", err)
	}
	return nil
}
