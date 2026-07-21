package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL      string
	InviteCode       string
	CORSOrigin       string
	Port             string
	AdminEmail       string
	AdminPassword    string
	AdminDisplayName string
	SessionDays      int
	SeedDemo         bool
	ResetDB          bool
	AppEnv           string
	CookieSecure     bool
}

func Load() (*Config, error) {
	appEnv := strings.ToLower(strings.TrimSpace(getEnv("APP_ENV", "development")))

	cfg := &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://indica:indica@localhost:5432/indica?sslmode=disable"),
		InviteCode:       getEnv("INVITE_CODE", "CANTEGRIL2026"),
		CORSOrigin:       getEnv("CORS_ORIGIN", "http://localhost:5173"),
		Port:             getEnv("PORT", "8080"),
		AdminEmail:       getEnv("ADMIN_EMAIL", "admin@cantegril.local"),
		AdminPassword:    getEnv("ADMIN_PASSWORD", "admin123"),
		AdminDisplayName: getEnv("ADMIN_DISPLAY_NAME", "Admin"),
		SessionDays:      30,
		SeedDemo:         getEnv("SEED_DEMO", "false") == "true",
		ResetDB:          getEnv("RESET_DB", "false") == "true",
		AppEnv:           appEnv,
	}

	if days := os.Getenv("SESSION_DAYS"); days != "" {
		n, err := strconv.Atoi(days)
		if err != nil {
			return nil, fmt.Errorf("SESSION_DAYS: %w", err)
		}
		cfg.SessionDays = n
	}

	if v, ok := os.LookupEnv("COOKIE_SECURE"); ok && v != "" {
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1", "yes":
			cfg.CookieSecure = true
		case "false", "0", "no":
			cfg.CookieSecure = false
		default:
			return nil, fmt.Errorf("COOKIE_SECURE must be true or false")
		}
	} else {
		cfg.CookieSecure = cfg.AppEnv == "production"
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.AppEnv == "production" {
		if cfg.AdminPassword == "" || cfg.AdminPassword == "admin123" {
			return nil, fmt.Errorf("ADMIN_PASSWORD must be set to a non-default value in production")
		}
		if cfg.InviteCode == "" || cfg.InviteCode == "CANTEGRIL2026" {
			return nil, fmt.Errorf("INVITE_CODE must be set to a non-default value in production")
		}
		if !cfg.CookieSecure {
			return nil, fmt.Errorf("COOKIE_SECURE must be true in production")
		}
		if cfg.ResetDB {
			return nil, fmt.Errorf("RESET_DB is not allowed in production")
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
