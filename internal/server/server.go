package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/muchirisworld/terminal/internal/config"
)

func NewServer(cfg *config.ServerConfig) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: routes(),
	}
}

func Start(s *http.Server) error {
	return s.ListenAndServe()
}

func Stop(ctx context.Context, s *http.Server) error {
	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}
