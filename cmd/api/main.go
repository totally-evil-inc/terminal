package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/muchirisworld/terminal/internal/config"
	"github.com/muchirisworld/terminal/internal/server"
)

func main() {
	ctx := gracefulShutdown()
	config, err := config.Load()
	if err != nil {
		log.Fatalf("App failed to start up: %v", err)
	}
	serverCfg := config.Server

	if err := run(ctx, serverCfg); err != nil {
		log.Fatal(err.Error())
	}
}

func run(ctx context.Context, cfg *config.ServerConfig) error {
	s := server.NewServer(cfg)

	serverErr := make(chan error, 1)
	go func() {
		if err := server.Start(s); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	return server.Stop(s, shutdownCtx)
}

func gracefulShutdown() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()
	return ctx
}
