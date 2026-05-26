package handlers

import (
	"customer_managment_system/internal/models"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
)

func CreateOrder(ordersJson gin.H) {
	var orders []models.Order

	bytes, err := json.Marshal(ordersJson["orders"])
	if err != nil {
		fmt.Println("Serialization error:", err)
		return
	}

	if err := json.Unmarshal(bytes, &orders); err != nil {
		fmt.Println("Deserialization error:", err)
		return
	}

	fmt.Printf("Successful holding orders: %d\n", len(orders))
}

func GenerateJson(msg string) (gin.H, error) {
	return nil, fmt.Errorf("Cannot create JSON")
}
