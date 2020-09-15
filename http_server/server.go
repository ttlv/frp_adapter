package http_server

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/sessions"
	"github.com/ttlv/frp_adapter/app/action"
	"github.com/ttlv/frp_adapter/home"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func New(sessionStore *sessions.CookieStore, dynamicClient dynamic.Interface) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	homeHandlers := home.NewHandlers(sessionStore)
	router.Get("/", homeHandlers.Home)
	actionHandlers := action.NewHandlers(sessionStore, dynamicClient, "default", schema.GroupVersionResource{Group: "ke.harmonycloud.io", Version: "v1", Resource: "nodemaintenances"})
	router.Post("/frp_create", actionHandlers.FrpCreate)
	router.Post("/frp_update", actionHandlers.FrpUpdate)
	return router
}
