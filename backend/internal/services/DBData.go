package services

import (
	"customer_managment_system/internal/handlers/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetProducts(c *gin.Context) {
	c.JSON(http.StatusOK, db.GetProducts())
}

func GetStocks(c *gin.Context) {
	c.JSON(http.StatusOK, db.GetStocks())
}

func GetStorages(c *gin.Context) {
	c.JSON(http.StatusOK, db.GetStorages())
}
