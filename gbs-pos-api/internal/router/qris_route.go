package router

import (
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

// setupQrisRoutes configures QRIS payment routes
// authMiddleware supports both JWT and Keycloak authentication
func setupQrisRoutes(r *gin.Engine, qrisHandler *handler.QrisHandler) {
	// Protected routes - require JWT or Keycloak authentication
	r.POST("/v1/qris/payments", qrisHandler.InitPayment)
	r.GET("/v1/qris/payments/:orderId/status", qrisHandler.GetPaymentStatus)

	// Webhook route (public, uses webhook token auth via X-Webhook-Token header)
	// Placed outside v1 group to avoid auth middleware
	r.POST("/v1/qris/webhook", qrisHandler.HandleWebhook)
}
