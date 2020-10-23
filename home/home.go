package home

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/ttlv/frp_adapter/app/helpers"
	"net/http"
)

type Handlers struct {
	SessionStore *sessions.CookieStore
}

func NewHandlers() Handlers {
	return Handlers{}
}

func (handlers Handlers) Home(c *gin.Context) {
	helpers.RenderSuccessJSON(c, http.StatusOK, "Frp Adapter Is Working Now....")
}
