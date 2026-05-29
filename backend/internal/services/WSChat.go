package services

import (
	"customer_managment_system/internal/chat"
	"log"
	"net/http"
	"time"

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
		var incomingMessage chat.Message

		err := conn.ReadJSON(&incomingMessage)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			return
		}

		serverMsg := chat.Message{
			Timestamp:    time.Now(),
			Message_type: chat.MESSAGE_TYPE_MESSAGE,
			Content:      "hi sender from server !",
			Sender:       "server",
			Status:       200,
		}

		if err := conn.WriteJSON(serverMsg); err != nil {
			log.Printf("Error writing JSON: %v", err)
			return
		}

	}
}
