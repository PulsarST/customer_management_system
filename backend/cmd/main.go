package main

import (
	"log"
	_ "net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	if err := r.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
