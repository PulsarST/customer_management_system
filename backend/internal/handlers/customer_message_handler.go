package handlers

import (
	"customer_managment_system/internal/models"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
)

func CreateOrder(ordersJson gin.H) ([]models.Order, error) {
	var orders []models.Order

	bytes, err := json.Marshal(ordersJson["orders"])
	if err != nil {
		fmt.Errorf("Serialization error:", err)
	}

	if err := json.Unmarshal(bytes, &orders); err != nil {
		fmt.Errorf("Deserialization error:", err)
	}

	return orders, nil
}

func GenerateJson(msg string) (gin.H, error) {
	return nil, fmt.Errorf("Cannot create JSON")
}
