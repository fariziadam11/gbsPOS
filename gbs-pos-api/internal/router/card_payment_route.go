package router

import (
	"gbs-pos-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func setupCardPaymentRoutes(auth *gin.RouterGroup, h *handler.CardPaymentHandler) {
	auth.POST("/card-payment/init", h.Init)
	auth.GET("/card-payment/pending", h.Pending)
	auth.GET("/card-payment/:id", h.Get)
	auth.POST("/card-payment/:id/cancel", h.Cancel)
}
