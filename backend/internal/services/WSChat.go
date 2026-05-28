package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func OnConnect(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		return
	}

	defer conn.Close()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		if err := conn.WriteMessage(msgType, msg); err != nil {
			return
		}

	}
}
