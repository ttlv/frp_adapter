package home

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

type Handlers struct {
	SessionStore *sessions.CookieStore
}

func NewHandlers() Handlers {
	return Handlers{}
}

func (handlers Handlers) Home(c *gin.Context) {
	c.Writer.Write([]byte("Frp Adapter Is Working Now...."))
}
