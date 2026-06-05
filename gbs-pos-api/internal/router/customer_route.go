package router

import (
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupCustomerRoutes(
	rg *gin.RouterGroup,
	customerHandler *handler.CustomerHandler,
) {
	rg.GET("/customers", customerHandler.List)
	rg.GET("/customers/:id", customerHandler.Get)
	rg.GET("/customers/phone/:phone", customerHandler.GetByPhone)
	rg.POST("/customers", customerHandler.Create)
	rg.PUT("/customers/:id", customerHandler.Update)
	rg.GET("/customers/:id/orders", customerHandler.GetOrders)
}
