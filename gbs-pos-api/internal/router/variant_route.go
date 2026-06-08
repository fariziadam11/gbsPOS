package router

import (
	"gbs-common/middleware"
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupVariantRoutes(
	rg *gin.RouterGroup,
	variantHandler *handler.ProductVariantHandler,
) {
	rg.GET("/products/:id/variants", variantHandler.List)
	rg.POST("/products/:id/variants", middleware.RequireRole("ADMIN"), variantHandler.Create)
	rg.PUT("/variants/:id", middleware.RequireRole("ADMIN"), variantHandler.Update)
	rg.DELETE("/variants/:id", middleware.RequireRole("ADMIN"), variantHandler.Delete)
}
