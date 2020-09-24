package http_server

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/ttlv/frp_adapter/app/action/frp"
	//"github.com/ttlv/frp_adapter/app/action/reverse_proxy"
	"github.com/ttlv/frp_adapter/config"
	"github.com/ttlv/frp_adapter/home"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func New(dynamicClient dynamic.Interface, frpsConfig *config.FrpsConfig, gvr schema.GroupVersionResource) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	homeHandlers := home.NewHandlers()
	actionHandlers := frp.NewHandlers(dynamicClient, gvr)
	//reverseProxyHandlers := reverse_proxy.NewHandlers()
	router.Get("/", homeHandlers.Home)
	router.Get("/frp_fetch/{name}", actionHandlers.FrpFetch) // GET /frp_adapter/fetch/xxxxxx
	router.Post("/frp_create", actionHandlers.FrpCreate)     // POST /frp_adapter/create
	router.Put("/frp_update", actionHandlers.FrpUpdate)      // PUT  /frp_adapter/update

	// reserve proxy
	//router.Get("/reverse_proxy", reverseProxyHandlers.ReverseProxy)
	return router
}
