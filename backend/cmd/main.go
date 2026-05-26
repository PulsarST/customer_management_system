package main

import (
	"customer_managment_system/internal/handlers/db"
	"log"
	_ "net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println(db.GetProducts())
	log.Println(db.GetStorages())
	log.Println(db.GetStocks())
	r := gin.Default()
	if err := r.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
