package http_server

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/sessions"
	"github.com/ttlv/frp_adapter/app/action"
	"github.com/ttlv/frp_adapter/home"
)

func New(sessionStore *sessions.CookieStore) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	homeHandlers := home.NewHandlers(sessionStore)
	router.Get("/", homeHandlers.Home)
	actionHandlers := action.NewHandlers(sessionStore)
	router.Post("/frp_create", actionHandlers.FrpCreate)
	return router
}
