package chat

import (
	"github.com/gin-gonic/gin"
)

const MESSAGE_TYPE_MESSAGE string = "wh_message"
const MESSAGE_TYPE_REQUEST string = "wh_request"

type Message struct {
	Timestamp    string
	Message_type string
	Content      string
	Sender       string
	Status       int
}

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
		Timestamp:    json["timestamp"].(string),
		Message_type: json["message_type"].(string),
		Content:      json["content"].(string),
		Sender:       json["sender"].(string),
		Status:       json["status"].(int),
	}
}
