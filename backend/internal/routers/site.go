package routers

import (
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
}

// asem/db/...
func GetDbProducts(c gin.Context) {

}
