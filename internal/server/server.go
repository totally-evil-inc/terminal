package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/muchirisworld/terminal/internal/config"
)

type Server struct {
	server          *http.Server
	port            int
	shutdownTimeout time.Duration
}

func NewServer(cfg *config.ServerConfig) *http.Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: routes(),
	}

	return srv
}

func Start(s *http.Server) error {
	return s.ListenAndServe()
}

func Stop(s *http.Server, ctx context.Context) error {
	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	// TODO: close connecions

	return nil
}
