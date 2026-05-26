package routers

import (
	"customer_managment_system/internal/services"

	"github.com/gin-gonic/gin"
)

func RegisterSiteRoutes(rg *gin.RouterGroup) {
	database_routes := rg.Group("/data")
	{
		database_routes.GET("/products", services.GetProducts)
		database_routes.GET("/storages", services.GetStorages)
		database_routes.GET("/stocks", services.GetStocks)
	}
}
