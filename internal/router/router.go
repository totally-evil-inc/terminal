package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"

	"github.com/muchirisworld/terminal/internal/auth"
)

type Router struct {
	db       *sqlx.DB
	verifier auth.Verifier
}

func NewRouter(db *sqlx.DB, verifier auth.Verifier) *Router {
	return &Router{
		db:       db,
		verifier: verifier,
	}
}

func (rtr *Router) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))

	health := NewHealthHandler(rtr.db)
	r.Route("/health", func(r chi.Router) {
		r.Get("/livez", health.live)
		r.Get("/readyz", health.ready)
	})

	test := NewTestHandler()
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(auth.Middleware(rtr.verifier))
		r.Get("/test", test.test)
	})

	return r
}
