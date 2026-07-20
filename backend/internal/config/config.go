package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL     string
	SessionSecret   string
	InviteCode      string
	CORSOrigin      string
	Port            string
	AdminEmail      string
	AdminPassword   string
	AdminDisplayName string
	SessionDays     int
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://indica:indica@localhost:5432/indica?sslmode=disable"),
		SessionSecret:    getEnv("SESSION_SECRET", "dev-secret-change-me"),
		InviteCode:       getEnv("INVITE_CODE", "CANTEGRIL2026"),
		CORSOrigin:       getEnv("CORS_ORIGIN", "http://localhost:5173"),
		Port:             getEnv("PORT", "8080"),
		AdminEmail:       getEnv("ADMIN_EMAIL", "admin@cantegril.local"),
		AdminPassword:    getEnv("ADMIN_PASSWORD", "admin123"),
		AdminDisplayName: getEnv("ADMIN_DISPLAY_NAME", "Admin"),
		SessionDays:      30,
	}

	if days := os.Getenv("SESSION_DAYS"); days != "" {
		n, err := strconv.Atoi(days)
		if err != nil {
			return nil, fmt.Errorf("SESSION_DAYS: %w", err)
		}
		cfg.SessionDays = n
	}

	if cfg.SessionSecret == "" {
		return nil, fmt.Errorf("SESSION_SECRET is required")
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
