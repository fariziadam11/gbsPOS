package router

import (
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupDashboardRoutes(
	rg *gin.RouterGroup,
	dashboardHandler *handler.DashboardHandler,
) {
	rg.GET("/dashboard/summary", dashboardHandler.Summary)
	rg.GET("/dashboard/revenue", dashboardHandler.Revenue)
	rg.GET("/dashboard/top-products", dashboardHandler.TopProducts)
}
