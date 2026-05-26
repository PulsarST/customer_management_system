package routers

import (
	"customer_managment_system/internal/handlers"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func start_message(c gin.Context) {

}

// chat/:msg
func GetUserMessage(c gin.Context) {
	message := c.Param("msg")
	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})

	_, err := handlers.CreateOrders(message)
	if err != nil {
		log.Fatalf(err.Error())
	}

}

// asem/db/...
func GetDbProducts(c gin.Context) {

}
