package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Config is the top-level configuration container. It composes all subsystem
// configs and is the single value passed into the application at startup.
// Load it once in main via Load(), then hand the relevant sub-config to each subsystem.
type Config struct {
	App      *AppConfig
	Server   *ServerConfig
	Database *DatabaseConfig
	Auth     *AuthConfig
}

// AppConfig holds application-wide settings that don't belong to a specific
// subsystem, such as the runtime environment (development, staging, production).
type AppConfig struct {
	Env string `env:"APP_ENV"`
}

// ServerConfig holds everything the HTTP layer needs to bind and operate:
// the port, read/write/idle timeouts, and the graceful-shutdown drain
// period. It must not contain business logic, database credentials, or
// any value that changes at runtime — those belong elsewhere.
type ServerConfig struct {
	Port            int           `env:"PORT" validate:"gte=3000,lte=9999"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" validate:"required"`
}

// DatabaseConfig carries the credentials and tuning knobs for the
// PostgreSQL connection pool: DSN/URL, max open/idle connections, and
// connection lifetime limits. It must not bleed into HTTP concerns; the
// server layer receives a *sqlx.DB, never this struct directly.
type DatabaseConfig struct {
	URL             string        `env:"DATABASE_URL" validate:"required"`
	MaxOpenConns    int           `env:"DATABASE_MAX_OPEN_CONNS" validate:"gte=1,lte=50"`
	MaxIdleConns    int           `env:"DATABASE_MAX_IDLE_CONNS" validate:"gte=1,lte=50"`
	ConnMaxLifetime time.Duration `env:"DATABASE_CONN_MAX_LIFETIME" validate:"required"`
	ConnMaxIdleTime time.Duration `env:"DATABASE_CONN_MAX_IDLE_TIME" validate:"required"`
}

// AuthConfig carries the settings required to verify JWTs issued by the
// external auth server.
type AuthConfig struct {
	ServerURL string `env:"AUTH_SERVER_URL" validate:"required,url"`
	Issuer    string `env:"AUTH_ISSUER" validate:"required"`
	Audience  string `env:"AUTH_AUDIENCE" validate:"required"`
}

func (c *AuthConfig) JWKSURL() string {
	return strings.TrimRight(c.ServerURL, "/") + "/api/auth/jwks"
}

var validate = validator.New(validator.WithRequiredStructEnabled())

func init() {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("env")
		if name == "" {
			return fld.Name
		}
		return name
	})
}

func Load() (*Config, error) {
	var errs []error

	appConfig := AppConfig{
		Env: getString("APP_ENV", "development"),
	}
	serverConfig := ServerConfig{
		Port:            getInt("PORT", 8080),
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 5*time.Minute),
	}
	databaseConfig := DatabaseConfig{
		URL:             getString("DATABASE_URL", ""),
		MaxOpenConns:    getInt("DATABASE_MAX_OPEN_CONNS", 10),
		MaxIdleConns:    getInt("DATABASE_MAX_IDLE_CONNS", 10),
		ConnMaxLifetime: getDuration("DATABASE_CONN_MAX_LIFETIME", 30*time.Minute),
		ConnMaxIdleTime: getDuration("DATABASE_CONN_MAX_IDLE_TIME", 5*time.Minute),
	}
	authServerURL := getString("AUTH_SERVER_URL", "http://localhost:4000/")
	authConfig := AuthConfig{
		ServerURL: authServerURL,
		Issuer:    getString("AUTH_ISSUER", authServerURL),
		Audience:  getString("AUTH_AUDIENCE", "http://localhost:8080"),
	}

	for _, s := range []any{appConfig, serverConfig, databaseConfig, authConfig} {
		if err := validate.Struct(s); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				for _, e := range ve {
					errs = append(errs, formatValidationError(e))
				}
			} else {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return &Config{
		App:      &appConfig,
		Server:   &serverConfig,
		Database: &databaseConfig,
		Auth:     &authConfig,
	}, nil
}

func formatValidationError(e validator.FieldError) error {
	switch e.Tag() {
	case "required":
		return fmt.Errorf("%s is required", e.Field())
	case "gte":
		return fmt.Errorf("%s must be at least %s", e.Field(), e.Param())
	case "lte":
		return fmt.Errorf("%s must be at most %s", e.Field(), e.Param())
	default:
		return fmt.Errorf("%s is invalid", e.Field())
	}
}

func getString(key string, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
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
