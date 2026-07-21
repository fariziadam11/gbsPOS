package handler

import (
	"net/http"
	"strings"

	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// QrisDirectHandler handles QRIS direct payment endpoints (static to dynamic)
type QrisDirectHandler struct {
	qrisDirectService *service.QrisDirectService
}

// NewQrisDirectHandler creates a new QRIS direct handler
func NewQrisDirectHandler(qrisDirectService *service.QrisDirectService) *QrisDirectHandler {
	return &QrisDirectHandler{
		qrisDirectService: qrisDirectService,
	}
}

// ParseQRIS godoc
// @Summary Parse QRIS string
// @Description Parse a QRIS string and return structured information
// @Tags QRIS Direct
// @Accept json
// @Produce json
// @Param request body dto.ParseQRISRequest true "QRIS String"
// @Success 200 {object} response.Response{data=dto.ParseQRISResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /v1/qris-direct/parse [post]
func (h *QrisDirectHandler) ParseQRIS(c *gin.Context) {
	var req dto.ParseQRISRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"success": false,
			"error":   "VALIDATION_ERROR",
			"message":  err.Error(),
		})
		return
	}

	if req.QrisString == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "VALIDATION_ERROR",
			"message":  "qrisString is required",
		})
		return
	}

	result, err := h.qrisDirectService.ParseQRIS(c.Request.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse QRIS")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "PARSE_ERROR",
			"message":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ConvertQRIS godoc
// @Summary Convert static QRIS to dynamic
// @Description Convert a static QRIS to dynamic by injecting amount
// @Tags QRIS Direct
// @Accept json
// @Produce json
// @Param request body dto.ConvertQRISRequest true "Conversion Request"
// @Success 200 {object} response.Response{data=dto.ConvertQRISResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /v1/qris-direct/convert [post]
func (h *QrisDirectHandler) ConvertQRIS(c *gin.Context) {
	var req dto.ConvertQRISRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"success": false,
			"error":   "VALIDATION_ERROR",
			"message":  err.Error(),
		})
		return
	}

	// Validate fee type if provided
	if req.FeeType != "" && req.FeeType != "fixed" && req.FeeType != "percentage" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "VALIDATION_ERROR",
			"message":  "feeType must be 'fixed' or 'percentage'",
		})
		return
	}

	result, err := h.qrisDirectService.ConvertQRIS(c.Request.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "order not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "ORDER_NOT_FOUND",
				"message":  err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "not a QRIS payment") {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "INVALID_PAYMENT_METHOD",
				"message":  err.Error(),
			})
			return
		}
		log.Error().Err(err).Msg("Failed to convert QRIS")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "CONVERSION_ERROR",
			"message":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetTransactionStatus godoc
// @Summary Get QRIS transaction status
// @Description Get the status of a QRIS direct transaction
// @Tags QRIS Direct
// @Produce json
// @Param transactionId path string true "Transaction ID"
// @Success 200 {object} response.Response{data=dto.GetQRISStatusResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /v1/qris-direct/transactions/{transactionId} [get]
func (h *QrisDirectHandler) GetTransactionStatus(c *gin.Context) {
	transactionID := c.Param("transactionId")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "VALIDATION_ERROR",
			"message":  "transactionId is required",
		})
		return
	}

	result, err := h.qrisDirectService.GetTransactionStatus(c.Request.Context(), transactionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "NOT_FOUND",
				"message":  err.Error(),
			})
			return
		}
		log.Error().Err(err).Msg("Failed to get transaction status")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_ERROR",
			"message":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ConfirmPayment godoc
// @Summary Confirm QRIS payment
// @Description Confirm that a QRIS payment has been completed by the customer
// @Tags QRIS Direct
// @Accept json
// @Produce json
// @Param request body dto.ConfirmQRISPaymentRequest true "Confirm Request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /v1/qris-direct/transactions/{transactionId}/confirm [post]
// @Router /v1/qris-direct/confirm [post]
func (h *QrisDirectHandler) ConfirmPayment(c *gin.Context) {
	var req dto.ConfirmQRISPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"success": false,
			"error":   "VALIDATION_ERROR",
			"message":  err.Error(),
		})
		return
	}

	if req.TransactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "VALIDATION_ERROR",
			"message":  "transactionId is required",
		})
		return
	}

	err := h.qrisDirectService.ConfirmPayment(c.Request.Context(), req.TransactionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "NOT_FOUND",
				"message":  err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "not pending") || strings.Contains(err.Error(), "expired") {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "INVALID_STATUS",
				"message":  err.Error(),
			})
			return
		}
		log.Error().Err(err).Msg("Failed to confirm payment")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_ERROR",
			"message":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message":  "Payment confirmed successfully",
	})
}

// CancelPayment godoc
// @Summary Cancel QRIS payment
// @Description Cancel a pending QRIS payment
// @Tags QRIS Direct
// @Accept json
// @Produce json
// @Param request body dto.CancelQRISPaymentRequest true "Cancel Request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /v1/qris-direct/transactions/{transactionId}/cancel [post]
// @Router /v1/qris-direct/cancel [post]
func (h *QrisDirectHandler) CancelPayment(c *gin.Context) {
	var req dto.CancelQRISPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"success": false,
			"error":   "VALIDATION_ERROR",
			"message":  err.Error(),
		})
		return
	}

	if req.TransactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "VALIDATION_ERROR",
			"message":  "transactionId is required",
		})
		return
	}

	// Get user from context (set by auth middleware)
	cancelledBy := "system"
	if userName, exists := c.Get("username"); exists {
		cancelledBy = userName.(string)
	}

	err := h.qrisDirectService.CancelPayment(c.Request.Context(), req.TransactionID, req.Reason, cancelledBy)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "NOT_FOUND",
				"message":  err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "not pending") {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "INVALID_STATUS",
				"message":  err.Error(),
			})
			return
		}
		log.Error().Err(err).Msg("Failed to cancel payment")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_ERROR",
			"message":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message":  "Payment cancelled successfully",
	})
}
