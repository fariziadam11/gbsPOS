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

		// Confirm payment (kasir confirms customer has paid)
		qrisDirect.POST("/transactions/:transactionId/confirm", qrisDirectHandler.ConfirmPayment)
		qrisDirect.POST("/confirm", qrisDirectHandler.ConfirmPayment)

		// Cancel payment
		qrisDirect.POST("/transactions/:transactionId/cancel", qrisDirectHandler.CancelPayment)
		qrisDirect.POST("/cancel", qrisDirectHandler.CancelPayment)
	}
}
