package routers

import "github.com/gin-gonic/gin"

func RegisterSiteRoutes(rg *gin.RouterGroup) {
	database_routes := rg.Group("data")
	{
		database_routes.GET("/products")
		database_routes.GET("/products")
		database_routes.GET("/products")
	}
}
