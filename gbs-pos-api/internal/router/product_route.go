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
	rg.GET("/products/low-stock", productHandler.GetLowStock)

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

	rg.POST(
		"/products/:id/stock",
		middleware.RequireRole("ADMIN"),
		productHandler.AdjustStock,
	)

	rg.GET("/products/:id/stock-history", productHandler.GetStockHistory)

	rg.POST(
		"/products/import",
		middleware.RequireRole("ADMIN"),
		productHandler.ImportCSV,
	)

	rg.GET(
		"/products/export",
		middleware.RequireRole("ADMIN"),
		productHandler.ExportCSV,
	)
}