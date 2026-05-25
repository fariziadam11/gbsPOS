package router

import (
	"gbs-common/middleware"
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupOrderRoutes(
	rg *gin.RouterGroup,
	orderHandler *handler.OrderHandler,
) {

	rg.GET("/orders", orderHandler.List)
	rg.GET("/orders/:id", orderHandler.Get)

	rg.POST("/orders", orderHandler.Create)
	rg.POST("/sync/orders", orderHandler.BulkSync)

	rg.GET(
		"/orders/unsettled/summary",
		orderHandler.UnsettledSummary,
	)

	rg.PATCH(
		"/orders/:id/void",
		middleware.RequireRole("ADMIN"),
		orderHandler.Void,
	)

	rg.POST(
		"/orders/settle",
		middleware.RequireRole("ADMIN"),
		orderHandler.Settle,
	)
}