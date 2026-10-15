package router

import (
	"gbs-common/middleware"
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupBTNCPMRoutes(r *gin.RouterGroup, h *handler.QrisCPMHandler) {
	cpm := r.Group("/btn-qris/cpm", middleware.RequireRole("ADMIN", "CASHIER"))
	cpm.POST("/payments", h.StartPayment)
	cpm.GET("/payments/:transactionId", h.GetStatus)
	cpm.GET("/orders/:orderId", h.GetByOrder)
	cpm.POST("/payments/:transactionId/cancel", h.Cancel)
	cpm.POST("/payments/:transactionId/refund", middleware.RequireRole("ADMIN"), h.Refund)
}

func setupBTNMPMRoutes(r *gin.RouterGroup, h *handler.QrisMPMHandler) {
	mpm := r.Group("/btn-qris/mpm", middleware.RequireRole("ADMIN"))
	mpm.POST("/payments/:transactionId/refund", h.Refund)
	r.POST("/btn-qris/reports", middleware.RequireRole("ADMIN"), h.Report)
}

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
		qrisDirect.GET("/orders/:orderId", qrisDirectHandler.GetLatestBTNStatusForOrder)

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
