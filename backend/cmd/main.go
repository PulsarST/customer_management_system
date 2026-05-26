package main

import (
	"customer_managment_system/internal/routers"
	"log"
	_ "net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	routers.RegisterRoutes(r)

	if err := r.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
