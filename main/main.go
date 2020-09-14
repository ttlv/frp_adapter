package main

import (
	"github.com/gorilla/sessions"
	"github.com/rs/cors"
	"github.com/ttlv/frp_adapter/config"
	"github.com/ttlv/frp_adapter/http_server"
	"log"
	"net/http"
)

func main() {
	cfg := config.MustGetConfig()
	cs := cors.New(cors.Options{
		//AllowedOrigins:   []string{"http://localhost:3002"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
		Debug:            true,
	})

	sessionStore := sessions.NewCookieStore([]byte("GbeVMHok6yjFXTgDkwUzVMj"))

	router := http_server.New(sessionStore)
	handler := cs.Handler(router)

	log.Printf("========== Visit http://localhost%v ==========\n", cfg.Port)
	log.Fatal(http.ListenAndServe(cfg.Port, handler))
}
