package router

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
)

type Router struct {
	db *sqlx.DB
}

func NewRouter(db *sqlx.DB) *Router {
	return &Router{
		db: db,
	}
}

func (rtr *Router) Routes() http.Handler {
	r := chi.NewRouter()
	
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/health", func(r chi.Router) {
		r.Get("/readyz", rtr.ready)
	})

	return r
}

func (rtr *Router) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
	defer cancel()

	if err := rtr.db.PingContext(ctx); err != nil {
		http.Error(w, "Database service not ready", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ok\n"))
}