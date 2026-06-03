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
			if websocket.IsCloseError(err,
				websocket.CloseNormalClosure,    // Код 1000: нормальное закрытие
				websocket.CloseGoingAway,        // Код 1001: клиент закрыл вкладку/обновил страницу
				websocket.CloseNoStatusReceived, // Код 1005: закрыто без статуса
			) {
				log.Printf("Клиент отключился (onclose).")
			} else {
				log.Printf("Ошибка чтения JSON или обрыв связи: %v", err)
			}

			return
		}

		serverMsg := ai.ParseToAI(incomingMessage, client, &ctx, c)

		if err := conn.WriteJSON(serverMsg); err != nil {
			log.Printf("Ошибка записи JSON: %v", err)
			return
		}
	}

}
