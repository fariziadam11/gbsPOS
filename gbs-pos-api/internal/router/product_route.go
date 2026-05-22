package router

import (
	"gbs-common/middleware"
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupProductRoutes(
	rg *gin.RouterGroup,
	productHandler *handler.ProductHandler,
) {

	rg.GET("/products", productHandler.List)

	rg.POST(
		"/products",
		middleware.RequireRole("ADMIN"),
		productHandler.Create,
	)

	rg.PUT(
		"/products/:id",
		middleware.RequireRole("ADMIN"),
		productHandler.Update,
	)

	rg.DELETE(
		"/products/:id",
		middleware.RequireRole("ADMIN"),
		productHandler.Delete,
	)
}