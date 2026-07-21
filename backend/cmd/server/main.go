package main

import (
	"context"
	"log"
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
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = findMigrations()
	}

	if err := db.Migrate(cfg.DatabaseURL, migrationsDir); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	if cfg.ResetDB {
		if cfg.AppEnv == "production" {
			log.Fatalf("RESET_DB is not allowed in production")
		}
		log.Printf("RESET_DB=true — wiping all data, then re-seeding")
		if err := db.ResetDB(ctx, pool); err != nil {
			log.Fatalf("reset db: %v", err)
		}
	}

	condoID, err := db.Seed(ctx, pool, db.SeedConfig{
		InviteCode:       cfg.InviteCode,
		AdminEmail:       cfg.AdminEmail,
		AdminPassword:    cfg.AdminPassword,
		AdminDisplayName: cfg.AdminDisplayName,
	})
	if err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Printf("seeded condo %s (%s)", db.CondoSlug, condoID)

	if cfg.SeedDemo || cfg.ResetDB {
		if err := db.SeedDemo(ctx, pool, condoID); err != nil {
			log.Fatalf("seed demo: %v", err)
		}
	}

	sessions := auth.NewSessionStore(pool, cfg.SessionDays)
	_ = sessions.DeleteExpired(ctx)

	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := sessions.DeleteExpired(context.Background()); err != nil {
				log.Printf("delete expired sessions: %v", err)
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
		log.Printf("listening on :%s (env=%s)", cfg.Port, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
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
