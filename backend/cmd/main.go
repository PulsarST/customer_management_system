package main

import (
	"customer_managment_system/internal/handlers/db"
	"customer_managment_system/internal/models"
	"log"
	_ "net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	db.LoadOrderToDb([]models.Order{
		{Name: "RUCHKA", Address: "ALMATY", Quantity: 12, Cost: 34.434},
		{Name: "SDFDSDSFSDF", Address: "asdas", Quantity: 34, Cost: 3.431},
	})
	r := gin.Default()
	if err := r.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
