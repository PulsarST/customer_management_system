package chat

import (
	"customer_managment_system/internal/models"
	"encoding/json"
	"net/http"
	"sync"
	"time"
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

type Cart struct {
	Orders     []models.Order
	total_cost float32
}

// ChatSession holds the per-user conversation state. Each WebSocket connection
// runs in its own goroutine, so every field that can be touched concurrently is
// guarded by mu.
type ChatSession struct {
	mu         sync.Mutex
	Messages   []Message
	OrdersCart []models.Order
}

var (
	sessionsMu sync.Mutex
	Sessions   = make(map[string]*ChatSession)
)

func RememberMessage(message Message, session *ChatSession) {
	session.mu.Lock()
	defer session.mu.Unlock()
	session.Messages = append(session.Messages, message)
}

func UpdateOrderList(orders []models.Order, session *ChatSession) {
	session.mu.Lock()
	defer session.mu.Unlock()
	session.OrdersCart = orders
}

// CartLen returns the number of items currently in the session's cart.
func CartLen(session *ChatSession) int {
	session.mu.Lock()
	defer session.mu.Unlock()
	return len(session.OrdersCart)
}

// SnapshotCart returns a copy of the current cart, safe to read outside the lock.
func SnapshotCart(session *ChatSession) []models.Order {
	session.mu.Lock()
	defer session.mu.Unlock()
	out := make([]models.Order, len(session.OrdersCart))
	copy(out, session.OrdersCart)
	return out
}

// maxHistoryMessages bounds how much conversation is replayed into each LLM
// prompt. Bot replies are stored verbatim (including large product/cart JSON),
// so an uncapped history would make every request slower than the last.
const maxHistoryMessages = 8

func GetConversationString(session *ChatSession) string {
	session.mu.Lock()
	defer session.mu.Unlock()

	if len(session.Messages) == 0 {
		return "NONE"
	}

	messages := session.Messages
	if len(messages) > maxHistoryMessages {
		messages = messages[len(messages)-maxHistoryMessages:]
	}

	var result string = ""

	for _, message := range messages {
		result += message.Sender + ": " + message.Content + "\n"
	}

	return result
}

func GetCartString(session *ChatSession) string {
	session.mu.Lock()
	defer session.mu.Unlock()

	json_result, _ := json.Marshal(session.OrdersCart)
	return string(json_result)
}

func GetSession(session_id string) *ChatSession {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if user := Sessions[session_id]; user != nil {
		return user
	}
	session := &ChatSession{}
	Sessions[session_id] = session
	return session
}
