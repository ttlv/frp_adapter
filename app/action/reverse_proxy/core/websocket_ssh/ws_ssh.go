package websocket_ssh

import "github.com/gorilla/websocket"

func WsSsh(upGrader websocket.Upgrader){
	wsConn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if handleError(c, err) {
		return
	}
	defer wsConn.Close()

}
