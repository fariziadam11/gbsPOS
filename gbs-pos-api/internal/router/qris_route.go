package router

import (
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

// setupQrisRoutes configures QRIS payment routes
func setupQrisRoutes(v1 *gin.RouterGroup, qrisHandler *handler.QrisHandler) {
	// Protected routes require authentication
	protected := v1.Group("")
	protected.Use(func(c *gin.Context) {
		// Auth middleware is applied at parent level in Setup()
		c.Next()
	})
	{
		protected.POST("/qris/payments", qrisHandler.InitPayment)
		protected.GET("/qris/payments/:orderId/status", qrisHandler.GetPaymentStatus)
	}

	// Webhook route (public, uses token auth via X-Webhook-Token header)
	v1.POST("/qris/webhook", qrisHandler.HandleWebhook)
}
