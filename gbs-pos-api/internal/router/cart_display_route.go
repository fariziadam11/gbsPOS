package router

import (
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupCartDisplayRoutes(
	rg *gin.RouterGroup,
	h *handler.CartDisplayHandler,
	authMiddleware gin.HandlerFunc,
) {
	if h == nil {
		return
	}

	// Public endpoint for customer/browser displays.
	public := rg.Group("")
	{
		public.GET("/display/cart", h.Get)
	}

	// Protected endpoints used by Android POS and admin/testing.
	auth := rg.Group("", authMiddleware)
	{
		auth.POST("/display/cart", h.Save)
		auth.DELETE("/display/cart/:terminalId", h.Delete)
	}
}
