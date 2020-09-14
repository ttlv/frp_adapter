package home

import (
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"net/http"
)

type Handlers struct {
	DB           *gorm.DB
	SessionStore *sessions.CookieStore
}

func NewHandlers(sessionStore *sessions.CookieStore) Handlers {
	return Handlers{SessionStore: sessionStore}
}

func (handlers Handlers) Home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Frp Adapter Is Working Now...."))
}
