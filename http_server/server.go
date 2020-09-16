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
	actionHandlers := action.NewHandlers(sessionStore, dynamicClient, "default", schema.GroupVersionResource{Group: "ke.harmonycloud.io", Version: "v1", Resource: "nodemaintenances"})
	router.Get("/", homeHandlers.Home)
	router.Get("/fetch/{name}", actionHandlers.FrpFetch) // GET /frp_adapter/fetch/xxxxxx
	router.Post("/create", actionHandlers.FrpCreate)     // POST /frp_adapter/create
	router.Post("/update", actionHandlers.FrpUpdate)     // PUT  /frp_adapter/update
	return router
}
