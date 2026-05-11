package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/muchirisworld/terminal/database"
	"github.com/muchirisworld/terminal/internal/config"
	"github.com/muchirisworld/terminal/internal/router"
	"github.com/muchirisworld/terminal/internal/server"
)

type Application struct {
	server *http.Server
	db     *sqlx.DB
	cfg    *config.Config
}

func main() {
	ctx := signalContext()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("App failed to start up: %v", err)
	}

	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	r := router.NewRouter(db)

	app := &Application{
		server: server.NewServer(cfg.Server, r.Routes()),
		db:     db,
		cfg:    cfg,
	}

	if err := app.run(ctx); err != nil {
		log.Fatal(err)
	}
}

func (a *Application) run(ctx context.Context) error {
	serverErr := make(chan error, 1)
	go func() {
		if err := server.Start(a.server); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	return server.Stop(shutdownCtx, a.server)
}

func signalContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()
	return ctx
}
