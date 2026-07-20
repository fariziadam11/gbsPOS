package handler

import (
	"net/http"
	"strings"

	"gbs-pos-api/internal/config"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// QrisHandler handles QRIS payment endpoints
type QrisHandler struct {
	qrisService *service.QrisService
	cfg        *config.Config
}

// NewQrisHandler creates a new QRIS handler
func NewQrisHandler(qrisService *service.QrisService, cfg *config.Config) *QrisHandler {
	return &QrisHandler{
		qrisService: qrisService,
		cfg:        cfg,
	}
}

// InitPayment godoc
// @Summary Initialize QRIS payment
// @Description Create a QRIS payment link for an order via SumoPod
// @Tags QRIS
// @Accept json
// @Produce json
// @Param request body dto.CreateQrisPaymentRequest true "QRIS Payment Request"
// @Success 200 {object} response.Response{data=dto.QrisInitResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /v1/qris/payments [post]
func (h *QrisHandler) InitPayment(c *gin.Context) {
	var req dto.CreateQrisPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"success": false,
			"error":   "VALIDATION_ERROR",
			"message":  err.Error(),
		})
		return
	}

	// Validate request
	if req.OrderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "VALIDATION_ERROR",
			"message":  "orderId is required",
		})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "VALIDATION_ERROR",
			"message":  "amount must be greater than 0",
		})
		return
	}

	result, err := h.qrisService.InitPayment(c.Request.Context(), req)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "ORDER_NOT_FOUND"):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "ORDER_NOT_FOUND",
				"message":  err.Error(),
			})
		case strings.Contains(err.Error(), "INVALID_PAYMENT_METHOD"):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "INVALID_PAYMENT_METHOD",
				"message":  err.Error(),
			})
		case strings.Contains(err.Error(), "SUMOPOD_API_ERROR"):
			log.Error().Err(err).Msg("SumoPod API error")
			c.JSON(http.StatusBadGateway, gin.H{
				"success": false,
				"error":   "PAYMENT_GATEWAY_ERROR",
				"message":  "Failed to connect to payment gateway",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "INTERNAL_SERVER_ERROR",
				"message":  err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetPaymentStatus godoc
// @Summary Get QRIS payment status
// @Description Get the current status of a QRIS payment for an order
// @Tags QRIS
// @Produce json
// @Param orderId path string true "Order ID"
// @Success 200 {object} response.Response{data=dto.QrisPaymentStatusResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /v1/qris/payments/{orderId}/status [get]
func (h *QrisHandler) GetPaymentStatus(c *gin.Context) {
	orderID := c.Param("orderId")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "VALIDATION_ERROR",
			"message":  "orderId is required",
		})
		return
	}

	result, err := h.qrisService.GetPaymentStatus(c.Request.Context(), orderID)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "ORDER_NOT_FOUND"):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "ORDER_NOT_FOUND",
				"message":  err.Error(),
			})
		case strings.Contains(err.Error(), "INVALID_PAYMENT_METHOD"):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "INVALID_PAYMENT_METHOD",
				"message":  err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "INTERNAL_SERVER_ERROR",
				"message":  err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// HandleWebhook godoc
// @Summary Handle SumoPod webhook
// @Description Receive and process webhook notifications from SumoPod
// @Tags QRIS
// @Accept json
// @Produce json
// @Param X-Webhook-Token header string true "Webhook Token"
// @Param payload body dto.SumoPodWebhookPayload true "Webhook Payload"
// @Success 200 {object} map[string]string
// @Failure 401 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /v1/qris/webhook [post]
func (h *QrisHandler) HandleWebhook(c *gin.Context) {
	// Verify webhook token (simple auth)
	webhookToken := c.GetHeader("X-Webhook-Token")
	if webhookToken == "" || webhookToken != h.cfg.SumoPodWebhookToken {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "UNAUTHORIZED",
			"message":  "Invalid webhook token",
		})
		return
	}

	var payload dto.SumoPodWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "VALIDATION_ERROR",
			"message":  err.Error(),
		})
		return
	}

	log.Info().
		Str("event_type", payload.EventType).
		Str("payment_id", payload.Data.PaymentID).
		Str("order_id", payload.Data.OrderID).
		Str("status", payload.Data.Status).
		Msg("Received QRIS webhook")

	if err := h.qrisService.HandleWebhook(c.Request.Context(), payload); err != nil {
		log.Error().Err(err).Msg("Failed to handle QRIS webhook")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_SERVER_ERROR",
			"message":  "Failed to process webhook",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message":  "Webhook processed successfully",
	})
}
