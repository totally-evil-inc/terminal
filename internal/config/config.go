package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
)

// Config is the top-level configuration container. It composes all subsystem
// configs and is the single value passed into the application at startup.
// Load it once in main via Load(), then hand the relevant sub-config to each subsystem.
type Config struct {
	App      *AppConfig      `validate:"required"`
	Server   *ServerConfig   `validate:"required"`
	Database *DatabaseConfig `validate:"required"`
}

// AppConfig holds application-wide settings that don't belong to a specific
// subsystem, such as the runtime environment (development, staging, production).
type AppConfig struct {
	Env string `validate:"required"`
}

// ServerConfig holds everything the HTTP layer needs to bind and operate:
// the port, read/write/idle timeouts, and the graceful-shutdown drain
// period. It must not contain business logic, database credentials, or
// any value that changes at runtime — those belong elsewhere.
type ServerConfig struct {
	Port            int           `validate:"gte=3000,lte=9999"`
	ShutdownTimeout time.Duration `validate:"required"`
}

// DatabaseConfig carries the credentials and tuning knobs for the
// PostgreSQL connection pool: DSN/URL, max open/idle connections, and
// connection lifetime limits. It must not bleed into HTTP concerns; the
// server layer receives a *sqlx.DB, never this struct directly.
type DatabaseConfig struct {
	URL             string        `validate:"required"`
	MaxOpenConns    int           `validate:"gte=1,lte=50"`
	MaxIdleConns    int           `validate:"gte=1,lte=50"`
	ConnMaxLifetime time.Duration `validate:"required"`
	ConnMaxIdleTime time.Duration `validate:"required"`
}

var validate *validator.Validate

func Load() (*Config, error) {
	var validationErrs []error
	validate = validator.New(validator.WithRequiredStructEnabled())

	appConfig := AppConfig{
		Env: getString("APP_ENV", "development"),
	}
	serverConfig := ServerConfig{
		Port: getInt("PORT", 8080),
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 5*time.Minute),
	}
	dbURL, err := getRequiredString("DATABASE_URL")
	if err != nil {
		validationErrs = append(validationErrs, err)
	}
	databaseConfig := DatabaseConfig{
		URL:             dbURL,
		MaxOpenConns:    getInt("DATABASE_MAX_OPEN_CONNS", 10),
		MaxIdleConns:    getInt("DATABASE_MAX_IDLE_CONNS", 10),
		ConnMaxLifetime: getDuration("DATABASE_CONN_MAX_LIFETIME", 30*time.Minute),
		ConnMaxIdleTime: getDuration("DATABASE_CONN_MAX_IDLE_TIME", 5*time.Minute),
	}

	if err := validate.Struct(appConfig); err != nil {
		validationErrs = append(validationErrs, err)
	}
	if err := validate.Struct(serverConfig); err != nil {
		validationErrs = append(validationErrs, err)
	}
	if err := validate.Struct(databaseConfig); err != nil {
		validationErrs = append(validationErrs, err)
	}

	cfg := Config{
		App:      &appConfig,
		Server:   &serverConfig,
		Database: &databaseConfig,
	}

	if len(validationErrs) > 0 {
		return nil, errors.Join(validationErrs...)
	}

	return &cfg, nil
}

func getString(key string, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
	}

	return val
}

func getRequiredString(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return "", fmt.Errorf("%s is required", key)
	}

	return val, nil
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
