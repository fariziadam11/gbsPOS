package router

import (
	"gbs-common/middleware"
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupFuelRoutes(rg *gin.RouterGroup, h *handler.FuelHandler) {
	// Public kiosk endpoints — no authentication required for customer self-service
	public := rg.Group("")
	{
		public.GET("/fuel-prices", h.ListPrices)
		public.GET("/pumps", h.ListPumps)
		public.GET("/nozzles", h.ListNozzles)
		public.POST("/fuel-sales", h.CreateSale)
	}

	// Admin endpoints — require ADMIN role
	admin := rg.Group("", middleware.RequireRole("ADMIN"))
	{
		admin.PUT("/fuel-prices/:code", h.UpdatePrice)

		admin.POST("/pumps", h.CreatePump)
		admin.PUT("/pumps/:id", h.UpdatePump)
		admin.DELETE("/pumps/:id", h.DeletePump)

		admin.POST("/nozzles", h.CreateNozzle)
		admin.PUT("/nozzles/:id", h.UpdateNozzle)
		admin.DELETE("/nozzles/:id", h.DeleteNozzle)

		admin.GET("/fuel-sales/report", h.Report)
	}
}
