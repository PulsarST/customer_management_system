package ai

import (
	"context"
	"customer_managment_system/internal/chat"
	"customer_managment_system/internal/handlers/db"
	"customer_managment_system/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type ParsedRequest struct {
	Intent  string `json:"intent"`
	Text    string `json:"text"`
	Payload string `json:"payload"`
}

type EmptyPayload struct{}

type ProductsPayload struct {
	Products []models.Product `json:"products"`
}

type JSONOrder struct {
	ProductID      int     `json:"product_id"`
	ProductName    string  `json:"product_name"`
	Price          float32 `json:"price"`
	StorageAddress string  `json:"storage_address"`
	Quantity       int     `json:"quantity"`
}

type CartPayload struct {
	Orders    []JSONOrder `json:"cart"`
	TotalCost float32     `json:"total_cost"`
}

func JSONOrderToModelsOrder(json_orders []JSONOrder) []models.Order {
	var result []models.Order
	for _, json_order := range json_orders {
		result = append(
			result,
			models.Order{
				Name:     json_order.ProductName,
				Address:  json_order.StorageAddress,
				Quantity: json_order.Quantity,
				Cost:     float32(json_order.Quantity) * json_order.Price})
	}
	return result
}
func ConnectToAI(ctx context.Context) *openai.Client {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	AI_TOKEN := os.Getenv("AI_TOKEN")

	config := openai.DefaultConfig(AI_TOKEN)
	config.BaseURL = "https://integrate.api.nvidia.com/v1"

	client := openai.NewClientWithConfig(config)
	return client
}

func processPayload(parsed *ParsedRequest, session *chat.ChatSession) error {
	switch parsed.Intent {
	case "chat":
	case "products":
	case "cart":
		var cart_payload CartPayload
		err := json.Unmarshal([]byte(parsed.Payload), &cart_payload)

		if err != nil {
			return err
		}

		log.Println("STRING PAYLOAD: %v", parsed.Payload)
		log.Println("CART PAYLOAD: %v", cart_payload)

		chat.UpdateOrderList(JSONOrderToModelsOrder(cart_payload.Orders), session)
	case "submit":
		db.LoadOrderToDb(session.OrdersCart)
		chat.UpdateOrderList([]models.Order{}, session)
	}
	return nil
}

func processParseResult(parseResult string, session *chat.ChatSession) chat.Message {
	jsonText := strings.TrimSpace(parseResult)

	jsonText = strings.ReplaceAll(jsonText, "```json", "")
	jsonText = strings.ReplaceAll(jsonText, "```", "")

	var parsed ParsedRequest

	err := json.Unmarshal([]byte(jsonText), &parsed)

	if err != nil {
		return chat.Message{Status: http.StatusInternalServerError, Content: "500 INTERNAL SERVER ERROR: " + err.Error() + "\n FAULTY ASS JSON: " + jsonText}
	}

	log.Println("PARSEREQUEST: %v", parsed)

	err = processPayload(&parsed, session)
	if err != nil {
		return chat.Message{Status: http.StatusInternalServerError, Content: "500 INTERNAL SERVER ERROR: " + err.Error() + "\n FAULTY ASS JSON: " + jsonText}
	}

	return chat.Message{Content: jsonText, Status: http.StatusOK}
}

func ParseToAI(message chat.Message, client *openai.Client, ctx *context.Context, c *gin.Context) chat.Message {

	session := chat.GetSession(message.Sender)

	log.Printf("%p\n", session)

	message.Content = strings.ToLower(message.Content)

	log.Printf("%v\n", chat.GetConversationString(session))

	var basePrompt = `
		Analyze the user message.

		Return ONLY valid JSON.
		JSON has this format

		{
			"intent": "chat",
			"text": "Hey, it's me!",
			"payload": "{...}"
		}

		intent corrects your behaviour
		text is your regular reply
		payload is a string of a valid json

		Database contains products, storages and stock, containing info about products in storages

		Tables:
	` + db.GetSchemaString() +
		`

		User's cart:
		
	` + chat.GetCartString(session) +
		`

		Your conversation with the user: [

	` + chat.GetConversationString(session) +
		`
		]
		Use JSONs from this conversation to get more context.

		You are a store assistant.

		Common Rules:
		- Answer briefly
		- Be friendly
		- Continue conversation logically
		- Don't reset dialogue

		Possible intents:
		- chat.
			Payload example: "{}"

		- products. User wants to see the products.
			Payload example: "{"products": [
					{"product_id": 1, "product_name": "Продукт 1", category: "Категория 1", "price": 12.3},
					{"product_id": 2, "product_name": "Продукт 2", category: "Категория 2", "price": 45.6}
				]}"

			Rules:
			- Help the client choose products
			- Don't invent products
			- Show 5 options for broader requests, do not overload user with all the products

		- cart. User wants to make changes to cart. Add, remove or change quantity. You should generate a new cart according to user's action.

		Payload example: "{"cart": [
					{"product_id": 1, "product_name": "Продукт 1", "price": 12.3, "storage_address": "г. Алматы, Сатпаева 101", "quantity": 3},
					{"product_id": 2, "product_name": "Продукт 2", "price": 45.6, "storage_address": "г. Алматы, Абая 44", "quantity": 4}
				],
				"total_cost": 219,3}"

			Rules:
			- total cost is a sum of quantity*price
			- show current cart
			- set 1 by default
			- if user writes "a couple", "few", "pair", "пару" -> quantity = 2

		- submit. User submits the order.
			Payload example: "{}"

			Rules:
			- Check if cart is not empty before switching to "submit" intent.

		Message:
	` + message.Content

	chat.RememberMessage(message, session)

	log.Printf("MESSAGES: %v\n", session.Messages)

	parsedResult, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "qwen/qwen3-coder-480b-a35b-instruct",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: basePrompt,
				},
			},
			Temperature: 0.6,
			TopP:        0.75,
			MaxTokens:   8192,
		},
	)

	if err != nil {
		return chat.Message{Status: http.StatusServiceUnavailable, Content: "503 SERVICE UNAVAILABLE: " + err.Error()}
	}

	result := processParseResult(parsedResult.Choices[0].Message.Content, session)
	if result.Status != http.StatusOK {
		return result
	}

	result_message := chat.Message{
		Timestamp:    time.Now(),
		Message_type: chat.MESSAGE_TYPE_MESSAGE,
		Content:      result.Content,
		Sender:       "you",
		Status:       http.StatusOK,
	}

	chat.RememberMessage(result_message, session)

	log.Printf("AI RESPOND: %v\n\nCURRENT CART: %v\n", result_message.Content, session.OrdersCart)

	return result_message
}
