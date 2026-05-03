package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Application struct {
	server *http.Server
	config *Config
}

type Config struct {
	Port            int
	ShutdownTimeout time.Duration
	// DbUrl string
}

func NewApplication(cfg *Config) *Application {
	return &Application{
		server: NewServer(cfg),
		config: cfg,
	}
}

func NewConfig() *Config {
	return &Config{
		Port:            8080,
		ShutdownTimeout: time.Duration(time.Second * 5),
	}
}

func NewServer(cfg *Config) *http.Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: routes(),
	}

	return srv
}

func (a *Application) Start() error {
	return a.server.ListenAndServe()
}

func (a *Application) Stop(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	// TODO: close connecions

	return nil
}
