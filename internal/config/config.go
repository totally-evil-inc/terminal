package config

import (
	"os"
	"strconv"
	"time"
)

// Config is the top-level configuration container. It composes all subsystem
// configs and is the single value passed into the application at startup.
// Load it once in main via Load(), then hand the relevant sub-config to each subsystem.
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
}

// AppConfig holds application-wide settings that don't belong to a specific
// subsystem, such as the runtime environment (development, staging, production).
type AppConfig struct {
	Env string
}

// ServerConfig holds everything the HTTP layer needs to bind and operate:
// the port, read/write/idle timeouts, and the graceful-shutdown drain
// period. It must not contain business logic, database credentials, or
// any value that changes at runtime — those belong elsewhere.
type ServerConfig struct {
	Port            string
	ShutdownTimeout time.Duration
}

// DatabaseConfig carries the credentials and tuning knobs for the
// PostgreSQL connection pool: DSN/URL, max open/idle connections, and
// connection lifetime limits. It must not bleed into HTTP concerns; the
// server layer receives a *sqlx.DB, never this struct directly.
type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		App:      AppConfig{},
		Server:   ServerConfig{},
		Database: DatabaseConfig{},
	}

	return cfg, nil
}

func getString(key string, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
	}

	return val
}

func getRequiredString(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return ""
	}

	return val
}

func getInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
	}

	n, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}

	return n
}

func getDuration(key string, fallback time.Duration) time.Duration {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
	}

	d, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}

	return d
}
