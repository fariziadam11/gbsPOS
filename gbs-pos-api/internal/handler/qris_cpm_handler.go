package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/service"

	"github.com/gin-gonic/gin"
)

type QrisCPMHandler struct {
	service *service.QrisCPMService
}

func NewQrisCPMHandler(s *service.QrisCPMService) *QrisCPMHandler {
	return &QrisCPMHandler{service: s}
}

// StartPayment godoc
// @Summary Start BTN QRIS CPM payment
// @Security BearerAuth
// @Param request body dto.StartBTNCPMPaymentRequest true "CPM checkout and scanned QR"
// @Success 200 {object} dto.BTNCPMPaymentResponse
// @Router /v1/btn-qris/cpm/payments [post]
func (h *QrisCPMHandler) StartPayment(c *gin.Context) {
	var req dto.StartBTNCPMPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "error": "VALIDATION_ERROR", "message": err.Error()})
		return
	}
	result, err := h.service.StartPayment(c.Request.Context(), req)
	if err != nil {
		writeCPMError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetStatus godoc
// @Summary Query a BTN QRIS CPM payment
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Success 200 {object} dto.BTNCPMPaymentResponse
// @Router /v1/btn-qris/cpm/payments/{transactionId} [get]
func (h *QrisCPMHandler) GetStatus(c *gin.Context) {
	result, err := h.service.GetStatus(c.Request.Context(), c.Param("transactionId"))
	if err != nil {
		writeCPMError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetByOrder godoc
// @Summary Recover the CPM payment for an order
// @Security BearerAuth
// @Param orderId path string true "Order ID"
// @Success 200 {object} dto.BTNCPMPaymentResponse
// @Router /v1/btn-qris/cpm/orders/{orderId} [get]
func (h *QrisCPMHandler) GetByOrder(c *gin.Context) {
	result, err := h.service.GetByOrderID(c.Request.Context(), c.Param("orderId"))
	if err != nil {
		writeCPMError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// Cancel godoc
// @Summary Cancel a pending BTN QRIS CPM payment
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Param request body dto.BTNCPMCancelRequest false "Cancel reason"
// @Success 200 {object} dto.BTNCPMPaymentResponse
// @Router /v1/btn-qris/cpm/payments/{transactionId}/cancel [post]
func (h *QrisCPMHandler) Cancel(c *gin.Context) {
	reason, ok := bindOptionalCPMReason(c)
	if !ok {
		return
	}
	result, err := h.service.Cancel(c.Request.Context(), c.Param("transactionId"), reason)
	if err != nil {
		writeCPMError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// Refund godoc
// @Summary Fully refund a verified BTN QRIS CPM payment
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Param request body dto.BTNCPMCancelRequest false "Refund reason"
// @Success 200 {object} dto.BTNCPMRefundResponse
// @Router /v1/btn-qris/cpm/payments/{transactionId}/refund [post]
func (h *QrisCPMHandler) Refund(c *gin.Context) {
	reason, ok := bindOptionalCPMReason(c)
	if !ok {
		return
	}
	result, err := h.service.Refund(c.Request.Context(), c.Param("transactionId"), reason)
	if err != nil {
		writeCPMError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func bindOptionalCPMReason(c *gin.Context) (string, bool) {
	var req dto.BTNCPMCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "error": "VALIDATION_ERROR", "message": err.Error()})
		return "", false
	}
	return req.Reason, true
}

func writeCPMError(c *gin.Context, err error) {
	var unknown *service.BtnSnapOutcomeUnknownError
	if errors.As(err, &unknown) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "PAYMENT_OUTCOME_UNKNOWN", "message": "Payment result is not confirmed; query the existing transaction before retrying."})
		return
	}
	var rejected *service.BtnSnapCPMRejectedError
	if errors.As(err, &rejected) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "BTN_REJECTED", "message": err.Error()})
		return
	}
	status, code := http.StatusInternalServerError, "CPM_ERROR"
	switch {
	case strings.Contains(err.Error(), "not found"):
		status, code = http.StatusNotFound, "NOT_FOUND"
	case strings.Contains(err.Error(), "not pending"), strings.Contains(err.Error(), "only a verified"), strings.Contains(err.Error(), "already exists"):
		status, code = http.StatusConflict, "INVALID_STATUS"
	case strings.Contains(err.Error(), "disabled"):
		status, code = http.StatusServiceUnavailable, "FEATURE_DISABLED"
	case strings.Contains(err.Error(), "required"), strings.Contains(err.Error(), "invalid"), strings.Contains(err.Error(), "does not match"), strings.Contains(err.Error(), "no longer available"):
		status, code = http.StatusBadRequest, "VALIDATION_ERROR"
	}
	c.JSON(status, gin.H{"success": false, "error": code, "message": err.Error()})
}
