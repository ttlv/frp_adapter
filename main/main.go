package main

import (
	"github.com/gorilla/sessions"
	"github.com/rs/cors"
	"github.com/ttlv/frp_adapter/config"
	"github.com/ttlv/frp_adapter/http_server"
	"github.com/ttlv/frp_adapter/kubeconfig_init"
	"k8s.io/client-go/dynamic"
	"log"
	"net/http"
)

func main() {
	var (
		cfg           = config.MustGetConfig()
		dynamicClient dynamic.Interface
		err           interface{}
	)
	cs := cors.New(cors.Options{
		//AllowedOrigins:   []string{"http://localhost:3002"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
		Debug:            true,
	})
	dynamicClient, err = kubeconfig_init.NewDynamicClient()
	defer func() {
		if err = recover(); err != nil {
			log.Println("Frp Adapter has been recovered")
		}
	}()
	sessionStore := sessions.NewCookieStore([]byte("GbeVMHok6yjFXTgDkwUzVMj"))

	router := http_server.New(sessionStore, dynamicClient)
	handler := cs.Handler(router)

	log.Printf("========== Visit http://localhost%v ==========\n", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.Port, handler))
}
