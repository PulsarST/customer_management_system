package services

import (
	"context"
	"customer_managment_system/internal/chat"
	"customer_managment_system/internal/handlers/ai"
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

	ctx := context.Background()

	client, err := ai.ConnectToAI(ctx)
	if err != nil {
		log.Printf("AI client unavailable: %v", err)
		_ = conn.WriteJSON(chat.Message{
			Status:  http.StatusServiceUnavailable,
			Content: "AI assistant is not configured: " + err.Error(),
		})
		return
	}

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

		incomingMessage.Timestamp = time.Now()

		serverMsg := ai.ParseToAI(incomingMessage, client, &ctx, c)
		// serverMsg := chat.Message{Content: "{\"intent\": \"cart\", \"text\": \"YOOFDGODFOG\", \"payload\": \"{\\\"cart\\\": [{\\\"product_id\\\": 1, \\\"product_name\\\": \\\"sdfsdf\\\", \\\"price\\\": 123.32, \\\"storage_address\\\": \\\"dsfsdf\\\", \\\"quantity\\\": 4}], \\\"total_cost\\\": 54.34}\"}", Status: 200}
		// serverMsg := chat.Message{Content: "{\"intent\": \"products\", \"text\": \"YOOFDGODFOG\", \"payload\": \"{\\\"products\\\": [{\\\"product_id\\\": 1, \\\"product_name\\\": \\\"sdfsdf\\\", \\\"category\\\": \\\"CATEGORY\\\", \\\"price\\\": 123.32}]}\"}", Status: 200}

		if err := conn.WriteJSON(serverMsg); err != nil {
			log.Printf("Ошибка записи JSON: %v", err)
			return
		}
	}

}
