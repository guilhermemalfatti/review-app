package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gmalfatti/indica/backend/internal/auth"
	"github.com/gmalfatti/indica/backend/internal/config"
	"github.com/gmalfatti/indica/backend/internal/db"
	httpserver "github.com/gmalfatti/indica/backend/internal/http"
	"github.com/gmalfatti/indica/backend/internal/logging"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	cfg, err := config.Load()
	if err != nil {
		// logging not set up yet
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	logging.Setup(cfg.AppEnv)

	ctx := context.Background()

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = findMigrations()
	}

	if err := db.Migrate(cfg.DatabaseURL, migrationsDir); err != nil {
		slog.Error("migrate", "err", err)
		os.Exit(1)
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if cfg.ResetDB {
		if cfg.AppEnv == "production" {
			slog.Error("RESET_DB is not allowed in production")
			os.Exit(1)
		}
		slog.Warn("RESET_DB=true — wiping all data, then re-seeding")
		if err := db.ResetDB(ctx, pool); err != nil {
			slog.Error("reset db", "err", err)
			os.Exit(1)
		}
	}

	condoID, err := db.Seed(ctx, pool, db.SeedConfig{
		InviteCode:       cfg.InviteCode,
		AdminEmail:       cfg.AdminEmail,
		AdminPassword:    cfg.AdminPassword,
		AdminDisplayName: cfg.AdminDisplayName,
	})
	if err != nil {
		slog.Error("seed", "err", err)
		os.Exit(1)
	}
	slog.Info("seeded condo", "slug", db.CondoSlug, "condo_id", condoID)

	if cfg.SeedDemo || cfg.ResetDB {
		if err := db.SeedDemo(ctx, pool, condoID); err != nil {
			slog.Error("seed demo", "err", err)
			os.Exit(1)
		}
	}

	sessions := auth.NewSessionStore(pool, cfg.SessionDays)
	if err := sessions.DeleteExpired(ctx); err != nil {
		slog.Warn("delete expired sessions on startup", "err", err)
	}

	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := sessions.DeleteExpired(context.Background()); err != nil {
				slog.Error("delete expired sessions", "err", err)
			}
		}
	}()

	router := httpserver.NewRouter(httpserver.Deps{
		Pool:         pool,
		Sessions:     sessions,
		CondoID:      condoID,
		InviteCode:   cfg.InviteCode,
		CORSOrigin:   cfg.CORSOrigin,
		CookieSecure: cfg.CookieSecure,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	go func() {
		slog.Info("listening", "addr", ":"+cfg.Port, "env", cfg.AppEnv, "cookie_secure", cfg.CookieSecure)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "err", err)
	}
}

func findMigrations() string {
	candidates := []string{
		"migrations",
		filepath.Join("backend", "migrations"),
		filepath.Join("..", "migrations"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			abs, err := filepath.Abs(c)
			if err == nil {
				return abs
			}
			return c
		}
	}
	return "migrations"
}
