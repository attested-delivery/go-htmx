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
	// Env selects dev or prod asset-serving behavior (AD-9's build-tag
	// split handles the compile-time half; this selects the runtime half
	// where both are compiled in, e.g. for local `go run` without -tags dev).
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
