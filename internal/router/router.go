package router

import (
	"net/http"

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

	health := NewHealthHandler(rtr.db)
	r.Route("/health", func(r chi.Router) {
		r.Get("/readyz", health.ready)
	})

	return r
}