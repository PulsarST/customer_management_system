package ai

import (
	"context"
	"customer_managment_system/internal/chat"
	"customer_managment_system/internal/handlers/db"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

type ParsedRequest struct {
	Intent  string                 `json:"intent"`
	Payload map[string]interface{} `json:"payload"`
}

func ConnectToAI(ctx context.Context) *genai.Client {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: "AQ.Ab8RN6KTPybFvoiZcc2gSBdB1McRYYgSn4CYMazF-124PvgBbw",
	})

	if err != nil {
		panic(err)
	}

	return client
}

var sessions = map[string][]string{}

func PushMessageToHistory(message chat.Message, session_id string) {
	history := sessions[session_id]

	history = append(history, message.Sender+": "+message.Content)

	sessions[session_id] = history
}

func ParseToAI(message chat.Message, client *genai.Client, ctx *context.Context, c *gin.Context) chat.Message {

	PushMessageToHistory(message, message.Sender)

	message.Content = strings.ToLower(message.Content)

	var basePrompt = `
		Analyze the user message.

		Return ONLY valid JSON.

		Database contains products, storages and stock, containing info about products in storages

		Tables:
	` + db.GetSchemaString() +
		`

		Possible intents:
		- chat. You are a store AI assistant.

			Rules:
			- Answer briefly
			- Be friendly
			- Continue conversation logically
			- Remember previous messages
			- Don't reset dialogue

		- order_options. User doesn't send enough info.

			Rules:
			- Give as many options as possible
			- Help the client choose products
			- Don't invent products

		- order. User does send enough info.

		- user_submits_order. User is 100 percent sure about his last order.

		Payload for each intent:
			chat: JSON with a string containing a regular reply.
			order_options: JSON with a regular reply string, and a JSON with a list of orders. Order containing product id, product name, storage address, quantity and a total cost being price*quantity.
			order: regular JSON with a reply string, and a JSON with a list of orders.
			user_submits_order: JSON with a string containing a regular reply, notifying user that the order was submitted.

		Rules:
		- product must be in english
		- quantity must be number
		- if quantity missing use 1
		- if user writes "a couple", "few", "pair", "пару" -> quantity = 2
		- if user doesn't mention enough data, switch to order_options
		- no explanations
		- json only

		Examples:

		User: I need 5 laptops from Almaty Satpayeva 101 and a mouse from Almaty Abaya 55
		Response:
		{
			"intent":"order",
			"payload":{
				"text": "Ваши товары были успешно найдены!",
				"orders": [
					{
							"product_id": 1, 
							"product_name": "Ноутбук Lenovo ThinkPad E14",
							"storage_address": "г. Алматы, ул. Сатпаева 101"
							"quantity": 5
							"total_cost": 1574.95
					},
					{
							"product_id": 4, 
							"product_name": "Мышь Logitech MX Master 3",
							"storage_address": "г. Алматы, пр. Абая 55"
							"quantity": 1
							"total_cost": 54.87
					}
				]
			}
		}

		User: Сколько стоит планшет?
		Response:
		{
			"intent":"order_options",
			"payload":{
				"text": "",
				"orders": [
					{
							"product_id": 14, 
							"product_name": "Планшет iPad Air",
							"storage_address": "г. Караганда, ул. Бухар Жырау 44"
							"quantity": 1
							"total_cost": 65.234
					},
					{
							"product_id": 14, 
							"product_name": "Планшет iPad Air",
							"storage_address": "г. Алматы, пр. Абая 55"
							"quantity": 1
							"total_cost": 65.234
					}
				]
			}
		}

		User: есть ручки?
		Response:
		{
			"intent":"order_options",
			"payload":{
				"text": "Извините, но на данный момент ни в одном хранилище не было ручек. Попробуйте спросить в следующий раз.",
				"orders": []
			}
		}

		User: есть наушники или колонки за 2 доллара?
		Response:
		{
			"intent":"order_options",
			"payload":{
				"text": "К сожалению у нас нет звуковых девайсов с такой ценой. Однако может вас заинтересует это?",
				"orders": [
					{
						"product_id": 11,
						"product_name": "Наушники Sony WH-1000XM4", 
						"storage_address": "г. Шымкент, ул. Тауке Хана 77"
						"quantity": 1
						"total_cost": 4453.45
					},
					{
						"product_id": 12,
						"product_name": "Колонки Edifier R1280T", 
						"storage_address": "г. Шымкент, ул. Тауке Хана 77"
						"quantity": 1
						"total_cost": 34.342
					}
				]
			}
		}

		User: привет
		Response:
		{
			"intent":"chat",
			"payload":{"text": "Здравствуйте, что хотите у нас заказать?"}
		}

		User: подверждаю заказ
		Response:
		{
			"intent":"user_submits_order",
			"payload":{"text": "Хорошо. Ваш последний заказ обрабатывается."}
		}

		Message:
	`

	parseResult, err := client.Models.GenerateContent(
		*ctx,
		"gemini-2.5-flash",
		genai.Text(basePrompt+message.Content),
		nil,
	)

	if err != nil {
		return chat.Message{Status: http.StatusServiceUnavailable, Content: "503 SERVICE UNAVAILABLE: " + err.Error()}
	}

	jsonText := strings.TrimSpace(parseResult.Text())

	jsonText = strings.ReplaceAll(jsonText, "```json", "")
	jsonText = strings.ReplaceAll(jsonText, "```", "")

	var parsed ParsedRequest

	err = json.Unmarshal([]byte(jsonText), &parsed)

	if err != nil {
		return chat.Message{Status: http.StatusInternalServerError, Content: "500 INTERNAL SERVER ERROR: " + err.Error() + "\n FAULTY ASS JSON: " + jsonText}
	}

	result_message := chat.Message{
		Timestamp:    time.Now(),
		Message_type: chat.MESSAGE_TYPE_MESSAGE,
		Content:      jsonText,
		Sender:       "server",
		Status:       http.StatusOK,
	}

	PushMessageToHistory(result_message, message.Sender)

	return result_message
}
