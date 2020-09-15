package home

import (
	"github.com/gorilla/sessions"
	"net/http"
)

type Handlers struct {
	SessionStore *sessions.CookieStore
}

func NewHandlers(sessionStore *sessions.CookieStore) Handlers {
	return Handlers{SessionStore: sessionStore}
}

func (handlers Handlers) Home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Frp Adapter Is Working Now...."))
}
