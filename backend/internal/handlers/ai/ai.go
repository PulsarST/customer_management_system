package ai

import (
	"context"
	"customer_managment_system/internal/chat"
	"customer_managment_system/internal/handlers/db"
	"customer_managment_system/internal/models"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

type ParsedRequest struct {
	Intent  string  `json:"intent"`
	Text    string  `json:"text"`
	Payload Payload `json:"payload"`
}

type Payload interface{}

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

// ConnectToAI builds the OpenAI-compatible client pointed at the NVIDIA
// endpoint. The .env file is optional: if it is missing we fall back to the real
// process environment instead of crashing the whole server. An empty token is
// reported as an error so the caller can answer the client gracefully.
func ConnectToAI(ctx context.Context) (*openai.Client, error) {
	// .env is a convenience for local dev; a missing file is not fatal.
	loadDotEnv()

	AI_TOKEN := os.Getenv("AI_TOKEN")
	if AI_TOKEN == "" {
		return nil, errors.New("AI_TOKEN is not set (define it in .env or the environment)")
	}

	config := openai.DefaultConfig(AI_TOKEN)
	config.BaseURL = "https://integrate.api.nvidia.com/v1"
	// Never let a stalled upstream block a chat goroutine forever.
	config.HTTPClient = &http.Client{Timeout: 90 * time.Second}

	return openai.NewClientWithConfig(config), nil
}

// defaultModel is used when AI_MODEL is not set. llama-3.1-8b-instruct was
// chosen over qwen3-next-80b after benchmarking on this app's real prompt: it
// is ~2-5x faster (≈1-3s vs ≈5s per turn) while still returning valid JSON and
// correct intents. Override with AI_MODEL to experiment (e.g. revert to
// "qwen/qwen3-next-80b-a3b-instruct" for slightly stronger reasoning).
const defaultModel = "meta/llama-3.1-8b-instruct"

// modelName lets the chat model be swapped without a rebuild via the AI_MODEL
// environment variable (e.g. a smaller, lower-latency model).
func modelName() string {
	if m := os.Getenv("AI_MODEL"); m != "" {
		return m
	}
	return defaultModel
}

// loadDotEnv finds a .env file by walking up from both the executable directory
// and the working directory (checking ./.env and ./backend/.env at each level),
// so the token is picked up no matter which directory the server was started
// from. A missing file is not an error: real process-environment variables and
// an explicit AI_TOKEN still work.
func loadDotEnv() {
	const maxLevels = 6

	var starts []string
	if exe, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}

	for _, start := range starts {
		dir := start
		for i := 0; i < maxLevels; i++ {
			for _, candidate := range []string{
				filepath.Join(dir, ".env"),
				filepath.Join(dir, "backend", ".env"),
			} {
				if _, err := os.Stat(candidate); err == nil {
					if err := godotenv.Load(candidate); err != nil {
						log.Printf("found %s but failed to load it: %v", candidate, err)
					}
					return
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	log.Printf("no .env file found (using process environment)")
}

// chatFallback builds a valid chat-intent message. It is used when the model
// returns malformed/truncated JSON, so the user sees a friendly reply instead of
// a blank bubble.
func chatFallback(text string) chat.Message {
	payload, _ := json.Marshal(map[string]any{
		"intent":  "chat",
		"text":    text,
		"payload": map[string]any{},
	})
	return chat.Message{Status: http.StatusOK, Content: string(payload)}
}

func processParseResult(parseResult string, session *chat.ChatSession) chat.Message {
	jsonText := strings.TrimSpace(parseResult)

	jsonText = strings.ReplaceAll(jsonText, "```json", "")
	jsonText = strings.ReplaceAll(jsonText, "```", "")

	var raw_parsed struct {
		Intent  string          `json:"intent"`
		Text    string          `json:"text"`
		Payload json.RawMessage `json:"payload"`
	}

	var parsed ParsedRequest

	err := json.Unmarshal([]byte(jsonText), &raw_parsed)

	if err != nil {
		log.Printf("failed to parse model JSON (%v); raw=%q", err, jsonText)
		return chatFallback("Извините, не совсем понял. Можете повторить?")
	}

	parsed.Intent = raw_parsed.Intent
	parsed.Text = raw_parsed.Text

	log.Println("PARSEREQUEST: ", parsed)
	log.Println("RAW PARSEREQ: ", jsonText)

	switch parsed.Intent {
	case "chat":
	case "products":
		var products_payload ProductsPayload
		err := json.Unmarshal(raw_parsed.Payload, &products_payload)

		if err != nil {
			log.Printf("failed to parse products payload (%v); raw=%q", err, jsonText)
			return chatFallback("Извините, не получилось показать товары. Попробуйте ещё раз.")
		}

		parsed.Payload = products_payload
	case "cart":
		var cart_payload CartPayload
		err := json.Unmarshal(raw_parsed.Payload, &cart_payload)

		if err != nil {
			log.Printf("failed to parse cart payload (%v); raw=%q", err, jsonText)
			return chatFallback("Извините, не получилось обновить корзину. Попробуйте ещё раз.")
		}

		parsed.Payload = cart_payload

		log.Printf("STRING PAYLOAD: %v", parsed.Payload)
		log.Printf("CART PAYLOAD: %v", cart_payload)

		chat.UpdateOrderList(JSONOrderToModelsOrder(cart_payload.Orders), session)
	case "submit":
		// Guard against the LLM submitting an empty cart: only persist and
		// reset when there is actually something to order.
		if chat.CartLen(session) == 0 {
			log.Println("submit ignored: cart is empty")
			break
		}
		db.LoadOrderToDb(chat.SnapshotCart(session))
		chat.UpdateOrderList([]models.Order{}, session)
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
			"payload": {...}
		}

		intent selects your behaviour (chat / products / cart / submit)
		text is your natural spoken reply to the user, in the user's language (Russian here).
			- Reply like a real shop assistant who already knows the answer.
			- NEVER just repeat or rephrase the user's question as your reply.
			  Bad: user "бумага а4 есть?" -> text "Бумага А4 есть?"
			  Good: user "бумага а4 есть?" -> text "Да, есть! Вот что нашёл:"
			- Confirm and move the dialogue forward.
		payload is a valid json matching the chosen intent

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
		- Reply in the user's language (Russian)
		- "text" must be a real answer, not a copy of the user's message
		- Continue conversation logically
		- Don't reset dialogue

		Possible intents:
		- chat.
			Payload example: {}

		- products. User wants to see the products.
			Payload example: {"products": [
					{"product_name": "Продукт 1", "price": 12.3},
					{"product_name": "Продукт 2", "price": 45.6}
				]}

			Rules:
			- Help the client choose products
			- Don't invent products: only use products from the Tables above
			- Write a short helpful intro in "text" (e.g. "Да, вот что есть:"), never echo the question
			- Show only products that match what the user asked for (stay on topic)
			- Show up to 5 matching options, but FEWER is fine. If only one product matches, show just that one
			- Never repeat the same product twice and never pad the list with unrelated products
			- Keep the full product_name exactly as in the Tables, including any "(Модель N)" suffix
			- If the user asks for more ("ещё", "только эта?", "а другие?"), show DIFFERENT matching products you have not shown yet (same topic)
			- Only output product_name and price for each product, nothing else

		- cart. User wants to make changes to cart. Add, remove or change quantity. You should generate a new cart according to user's action.

		Payload example: {"cart": [
					{"product_id": 1, "product_name": "Продукт 1", "price": 12.3, "storage_address": "г. Алматы, Сатпаева 101", "quantity": 3},
					{"product_id": 2, "product_name": "Продукт 2", "price": 45.6, "storage_address": "г. Алматы, Абая 44", "quantity": 4}
				],
				"total_cost": 219,3}

			Rules:
			- total cost is a sum of quantity*price
			- show current cart
			- set 1 by default
			- if user writes "a couple", "few", "pair", "пару" -> quantity = 2

		- submit. User submits the order and it's getting processed.
			Payload example: {}

			Rules:
			- Check if cart is not empty before switching to "submit" intent.

		Message:
	` + message.Content

	chat.RememberMessage(message, session)

	log.Printf("MESSAGES: %v\n", chat.GetConversationString(session))

	llmStart := time.Now()
	parsedResult, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: modelName(),
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: basePrompt,
				},
			},
			Temperature: 0.3,
			TopP:        0.75,
			// Headroom so a 5-item product list (with full "(Модель N)" names)
			// never gets truncated mid-JSON, which would break parsing.
			MaxTokens: 800,
			Stop:      []string{"```", "\n\n\n"},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	// parseResult, err := client.Models.GenerateContent(
	// 	*ctx,
	// 	"gemini-2.5-flash",
	// 	genai.Text(basePrompt+message.Content),
	// 	nil,
	// )

	if err != nil {
		return chat.Message{Status: http.StatusServiceUnavailable, Content: "503 SERVICE UNAVAILABLE: " + err.Error()}
	}

	log.Printf("LLM latency=%s prompt_tokens=%d completion_tokens=%d",
		time.Since(llmStart).Round(time.Millisecond), parsedResult.Usage.PromptTokens,
		parsedResult.Usage.CompletionTokens)

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

	log.Printf("AI RESPOND: %v\n\nCURRENT CART: %v\n", result_message.Content, chat.SnapshotCart(session))

	return result_message
}
