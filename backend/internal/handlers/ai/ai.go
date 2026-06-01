package ai

import (
	"context"
	"customer_managment_system/internal/chat"
	"customer_managment_system/internal/handlers/db"
	"customer_managment_system/internal/models"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

type ParsedRequest struct {
	Intent  string         `json:"intent"`
	Payload RequestPayload `json:"payload"`
}

type RequestPayload struct {
	Text   string         `json:"text"`
	Orders []models.Order `json:"orders"`
}

type AISession struct {
	Messages      []chat.Message
	LastOrderList []models.Order
}

var sessions = map[string]AISession{}

func ConnectToAI(ctx context.Context) *genai.Client {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: "AQ.Ab8RN6KfkTdqdZM4RJ0Dm-ZEsBAMk3En6x3vWUqrWA2siFnk9g",
	})

	if err != nil {
		panic(err)
	}

	return client
}

func PushMessageToHistory(message chat.Message, session *AISession) {
	session.Messages = append(session.Messages, message)
}

func updateOrderList(parsed ParsedRequest, session_id string) {
	if parsed.Intent != "order" {
		return
	}

	history := sessions[session_id]

	history.LastOrderList = parsed.Payload.Orders
}

func getConversationString(session *AISession) string {
	var result string = ""

	for _, message := range session.Messages {
		result += message.Sender + ": " + message.Content + "\n"
	}

	return result
}

func ParseToAI(message chat.Message, client *genai.Client, ctx *context.Context, c *gin.Context) chat.Message {

	session := sessions[message.Sender]

	PushMessageToHistory(message, &session)

	message.Content = strings.ToLower(message.Content)

	var basePrompt = `
		Analyze the user message.

		Return ONLY valid JSON.

		Previous messages: [

	` + getConversationString(&session) +
		`	
		]

		Database contains products, storages and stock, containing info about products in storages

		Tables: [
	` + db.GetSchemaString() +
		`
		]

		You are a store assistant.

		Common Rules:
		- Answer briefly
		- Be friendly
		- Continue conversation logically
		- Remember previous messages.
		- Don't reset dialogue

		Possible intents:
		- chat.
		- order_options. User doesn't send enough info.

			Rules:
			- Give as many options as possible
			- Help the client choose products
			- Don't invent products
			- Show 5 options for broader requests, do not overload user with all the products

		- order. User sent enough info. 
		
			Rules:
			- Check previous user's requests for additional info.
			- Ask if user wants to submit current order.

		Payload for each intent:
			chat: JSON with a string containing a regular reply.
			order_options: JSON with a regular reply string, and a JSON with a list of orders. Order containing product id, product name, product category, storage address, quantity and a total cost being price*quantity.
			order: regular JSON with a reply string, and a JSON with a list of orders.

		Rules:
		- if quantity missing use 1
		- if user writes "a couple", "few", "pair", "пару" -> quantity = 2
		- if user doesn't mention enough data, switch to order_options
		- no explanations
		- json only

		Examples:

		(user's first order)
		User: I need 5 laptops from Almaty Satpayeva 101 and a mouse from Almaty Abaya 55
		Response:
		{
			"intent":"order",
			"payload":{
				"text": "Ваши товары были успешно найдены! Хотите оформить заказ этих товаров?",
				"orders": [
					{
							"product_id": 1, 
							"product_name": "Ноутбук Lenovo ThinkPad E14",
							"category": "Техника",
							"storage_address": "г. Алматы, ул. Сатпаева 101"
							"quantity": 5
							"total_cost": 1574.95
					},
					{
							"product_id": 4, 
							"product_name": "Мышь Logitech MX Master 3",
							"category": "Техника",
							"storage_address": "г. Алматы, пр. Абая 55"
							"quantity": 1
							"total_cost": 54.87
					}
				]
			}
		}

		(context: user sending additional info, user's order is NOT NONE)
		User: давайте штук 5
		Response:
		{
			"intent":"order",
			"payload":{
				"text": "Хорошо, 5 ноутбуков из сатпаева. Подтверждаете заказ этих товаров?",
				"orders": [
					{
							"product_id": 1, 
							"product_name": "Ноутбук Lenovo ThinkPad E14",
							"category": "Техника",
							"storage_address": "г. Алматы, ул. Сатпаева 101"
							"quantity": 5
							"total_cost": 1574.95
					},
					{
							"product_id": 4, 
							"product_name": "Мышь Logitech MX Master 3",
							"category": "Техника",
							"storage_address": "г. Алматы, пр. Абая 55"
							"quantity": 1
							"total_cost": 54.87
					}
				]
			}
		}

		(context: any message. user asked for cost, so they don't need the same product from different storages)
		User: Сколько стоит планшет?
		Response:
		{
			"intent":"order_options",
			"payload":{
				"text": "Есть следующие планшеты:",
				"orders": [
					{
							"product_id": 14, 
							"product_name": "Планшет iPad Air",
							"category": "Техника",
							"storage_address": "г. Караганда, ул. Бухар Жырау 44"
							"quantity": 1
							"total_cost": 65.234
					}
				]
			}
		}

		(context: no pens found in any storage)
		User: есть ручки?
		Response:
		{
			"intent":"order_options",
			"payload":{
				"text": "Извините, но на данный момент ни в одном хранилище не было ручек. Попробуйте спросить в следующий раз.",
				"orders": []
			}
		}

		User: есть аудио девайсы за 2 доллара?
		Response:
		{
			"intent":"order_options",
			"payload":{
				"text": "К сожалению у нас нет звуковых девайсов с такой ценой. Вот самые дешёвые:",
				"orders": [
					{
						"product_id": 11,
						"product_name": "Наушники Sony WH-1000XM4", 
						"category": "Техника",
						"storage_address": "г. Шымкент, ул. Тауке Хана 77"
						"quantity": 1
						"total_cost": 4453.45
					},
					{
						"product_id": 12,
						"product_name": "Колонки Edifier R1280T", 
						"category": "Техника",
						"storage_address": "г. Шымкент, ул. Тауке Хана 77"
						"quantity": 1
						"total_cost": 34.342
					}
				]
			}
		}

		(context: no order, just chat)
		User: привет
		Response:
		{
			"intent":"chat",
			"payload":{"text": "Здравствуйте, что хотите у нас заказать?", "orders": []}
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

	switch parsed.Intent {
	case "order":
		updateOrderList(parsed, message.Sender)
	}

	result_message := chat.Message{
		Timestamp:    time.Now(),
		Message_type: chat.MESSAGE_TYPE_MESSAGE,
		Content:      jsonText,
		Sender:       "server",
		Status:       http.StatusOK,
	}

	PushMessageToHistory(result_message, &session)

	return result_message
}
