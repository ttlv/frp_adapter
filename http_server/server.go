package http_server

import (
	"flag"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/sessions"
	"github.com/ttlv/frp_adapter/app/action"
	"github.com/ttlv/frp_adapter/home"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

var Kubeconfig *string

func init() {
	// init kubeconfig
	if home := homedir.HomeDir(); home != "" {
		Kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		Kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
}

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
