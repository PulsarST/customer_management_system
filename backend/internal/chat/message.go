package chat

import (
	"customer_managment_system/internal/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const MESSAGE_TYPE_MESSAGE string = "wh_message"
const MESSAGE_TYPE_REQUEST string = "wh_request"

type Message struct {
	Timestamp    time.Time `json:"timestamp"`
	Message_type string    `json:"messageType"`
	Content      string    `json:"content"`
	Sender       string    `json:"sender"`
	Status       int       `json:"status"`
}

var OKAY_MESSAGE Message = Message{Status: http.StatusOK}

func MessageToJson(message Message) gin.H {
	return gin.H{
		"timestamp":    message.Timestamp,
		"message_type": message.Message_type,
		"content":      message.Content,
		"sender":       message.Sender,
		"status":       message.Status,
	}
}

func JsonToMessage(json gin.H) Message {

	return Message{
		Timestamp:    json["timestamp"].(time.Time),
		Message_type: json["message_type"].(string),
		Content:      json["content"].(string),
		Sender:       json["sender"].(string),
		Status:       json["status"].(int),
	}
}

type Cart struct {
	Orders     []models.Order
	total_cost float32
}

type ChatSession struct {
	Messages   []Message
	OrdersCart []models.Order
}

var Sessions = make(map[string]*ChatSession)

func RememberMessage(message Message, session *ChatSession) {
	session.Messages = append(session.Messages, message)
}

func UpdateOrderList(orders []models.Order, session *ChatSession) {
	session.OrdersCart = orders
}

func GetConversationString(session *ChatSession) string {
	if len(session.Messages) == 0 {
		return "NONE"
	}

	var result string = ""

	for _, message := range session.Messages {
		result += message.Sender + ": " + message.Content + "\n"
	}

	return result
}

func GetCartString(session *ChatSession) string {
	var result string = ""

	json_result, _ := json.Marshal(session.OrdersCart)
	result = string(json_result)

	return result
}

func GetSession(session_id string) *ChatSession {
	if user := Sessions[session_id]; user != nil {
		return user
	}
	Sessions[session_id] = &ChatSession{}
	return Sessions[session_id]
}
