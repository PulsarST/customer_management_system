package services

import (
	"context"
	"customer_managment_system/internal/chat"
	"customer_managment_system/internal/handlers/ai"
	"log"
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

	ctx := context.Background()

	client := ai.ConnectToAI(ctx)

	for {
		var incomingMessage chat.Message

		err := conn.ReadJSON(&incomingMessage)
		if err != nil {
			log.Fatalf("Error reading JSON: %v", err.Error())
			return
		}

		log.Printf("%v", incomingMessage)

		serverMsg := ai.ParseToAI(incomingMessage, client, &ctx, c)

		log.Printf("%v", serverMsg)

		if err := conn.WriteJSON(serverMsg); err != nil {
			log.Fatalf("Error writing JSON: %v", err.Error())
			return
		}

	}
}
