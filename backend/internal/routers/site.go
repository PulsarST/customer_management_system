package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func start_message(c gin.Context) {

}

// chat/:msg
func get_user_message(c gin.Context) {
	message := c.Param("msg")
	c.JSON(http.StatusOK, gin.H{
		"message": "got it",
	})

}

// asem/db/...
func get_db_products(c gin.Context) {

}
