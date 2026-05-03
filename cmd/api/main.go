package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/muchirisworld/terminal/internal/server"
)

func main() {
	ctx := gracefulShutdown()
	cfg := server.NewConfig()

	if err := run(ctx, cfg); err != nil {
		log.Fatal(err.Error())
	}
}

func run(ctx context.Context, cfg *server.Config) error {
	a := server.NewApplication(cfg)

	serverErr := make(chan error, 1)
	go func() {
		if err := a.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

	return a.Stop(shutdownCtx)
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
