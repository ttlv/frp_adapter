package helpers

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/ttlv/frp_adapter/app/entries"
	"time"
)

func RenderFailureJSON(c *gin.Context, code int, message string) {
	c.JSON(code, entries.Error{
		Error: entries.ErrorDetail{
			Message: message,
		},
	})
}

func RenderSuccessJSON(c *gin.Context, code int, data interface{}) {
	c.JSON(code, entries.Success{
		Data: data,
	})
}

func WsErrorHandle(ws *websocket.Conn, err error) bool {
	if err != nil {
		logrus.WithError(err).Error("handler ws ERROR:")
		dt := time.Now().Add(time.Second)
		if err := ws.WriteControl(websocket.CloseMessage, []byte(err.Error()), dt); err != nil {
			logrus.WithError(err).Error("websocket writes control message failed:")
		}
		return true
	}
	return false
}
