package router

import (
	"gbs-common/middleware"
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupDiscountRoutes(
	rg *gin.RouterGroup,
	discountHandler *handler.DiscountHandler,
) {
	if discountHandler == nil {
		return
	}

	rg.GET("/discounts", discountHandler.List)
	rg.POST("/pricing/calculate", discountHandler.Calculate)
	rg.GET("/discounts/:id", discountHandler.Get)

	rg.POST(
		"/discounts",
		middleware.RequireRole("ADMIN"),
		discountHandler.Create,
	)

	rg.PUT(
		"/discounts/:id",
		middleware.RequireRole("ADMIN"),
		discountHandler.Update,
	)

	rg.PATCH(
		"/discounts/:id/stop",
		middleware.RequireRole("ADMIN"),
		discountHandler.Stop,
	)

	rg.PATCH(
		"/discounts/:id/cancel",
		middleware.RequireRole("ADMIN"),
		discountHandler.Cancel,
	)

	rg.DELETE(
		"/discounts/:id",
		middleware.RequireRole("ADMIN"),
		discountHandler.Delete,
	)
}
