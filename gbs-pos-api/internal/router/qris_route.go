package router

import (
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

// setupQrisRoutes configures QRIS payment routes
// authMiddleware supports both JWT and Keycloak authentication
func setupQrisRoutes(auth *gin.RouterGroup, r *gin.Engine, qrisHandler *handler.QrisHandler) {
	// Protected routes - require JWT or Keycloak authentication
	auth.POST("/qris/payments", qrisHandler.InitPayment)
	auth.GET("/qris/payments/:orderId/status", qrisHandler.GetPaymentStatus)

	// Webhook route (public, uses webhook token auth via X-Webhook-Token header)
	// Placed outside v1 group to avoid auth middleware
	r.POST("/v1/qris/webhook", qrisHandler.HandleWebhook)
}
