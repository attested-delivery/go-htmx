// Package config loads runtime configuration from the environment.
package config

import (
	"fmt"
	"os"
)

// Config holds the application's runtime settings, sourced from
// environment variables so the single embedded binary (AD-9) needs no
// external config file to start.
type Config struct {
	// Addr is the address the HTTP server listens on, e.g. ":8080".
	Addr string
	// Env is informational only (logged at startup as "env"). The actual
	// dev/prod asset-serving split is a compile-time decision made by the
	// `dev` build tag (see internal/web/assets) — only one implementation
	// is ever compiled in, so this value doesn't select between them.
	Env string
	// DBPath is the SQLite database file path.
	DBPath string
}

// Load reads Config from the environment, applying defaults for anything
// unset.
func Load() (Config, error) {
	cfg := Config{
		Addr:   getEnv("GO_HTMX_ADDR", ":8080"),
		Env:    getEnv("GO_HTMX_ENV", "prod"),
		DBPath: getEnv("GO_HTMX_DB_PATH", "go-htmx.db"),
	}

	if cfg.Env != "dev" && cfg.Env != "prod" {
		return Config{}, fmt.Errorf("config: GO_HTMX_ENV must be %q or %q, got %q", "dev", "prod", cfg.Env)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
