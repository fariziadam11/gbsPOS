package router

import (
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

// setupQrisDirectRoutes configures QRIS direct payment routes
// These endpoints handle static to dynamic QRIS conversion without payment gateway
// The static QRIS string is configured via QRIS_DIRECT_STATIC_QRIS env var
func setupQrisDirectRoutes(r *gin.RouterGroup, qrisDirectHandler *handler.QrisDirectHandler) {
	// QRIS Direct endpoints (static to dynamic)
	// Protected routes - require JWT or Keycloak authentication
	qrisDirect := r.Group("/qris-direct")
	{
		// Convert static to dynamic QRIS (uses configured static QRIS)
		qrisDirect.POST("/convert", qrisDirectHandler.ConvertQRIS)

		// Get transaction status
		qrisDirect.GET("/transactions/:transactionId", qrisDirectHandler.GetTransactionStatus)

		// Confirm payment (kasir confirms customer has paid) - immediate
		qrisDirect.POST("/transactions/:transactionId/confirm", qrisDirectHandler.ConfirmPayment)
		qrisDirect.POST("/confirm", qrisDirectHandler.ConfirmPayment)

		// Confirm pending (trigger auto-confirm with delay)
		// Kasir clicks this after verifying customer scanned, then auto-confirm runs after delay
		qrisDirect.POST("/transactions/:transactionId/confirm-pending", qrisDirectHandler.ConfirmPending)
		qrisDirect.POST("/confirm-pending", qrisDirectHandler.ConfirmPending)

		// Cancel payment
		qrisDirect.POST("/transactions/:transactionId/cancel", qrisDirectHandler.CancelPayment)
		qrisDirect.POST("/cancel", qrisDirectHandler.CancelPayment)
	}
}
