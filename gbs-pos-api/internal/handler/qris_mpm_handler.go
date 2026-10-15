package handler

import (
	"errors"
	"net/http"
	"strings"

	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/service"

	"github.com/gin-gonic/gin"
)

type QrisMPMHandler struct {
	service *service.QrisMPMService
}

func NewQrisMPMHandler(s *service.QrisMPMService) *QrisMPMHandler {
	return &QrisMPMHandler{service: s}
}

// Refund godoc
// @Summary Fully refund a verified BTN QRIS MPM payment
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Param request body dto.BTNMPMRefundRequest true "Refund reason"
// @Success 200 {object} dto.BTNMPMRefundResponse
// @Router /v1/btn-qris/mpm/payments/{transactionId}/refund [post]
func (h *QrisMPMHandler) Refund(c *gin.Context) {
	var req dto.BTNMPMRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "error": "VALIDATION_ERROR", "message": err.Error()})
		return
	}
	result, err := h.service.Refund(c.Request.Context(), c.Param("transactionId"), req.Reason)
	if err != nil {
		writeMPMError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// Report godoc
// @Summary Get a paginated BTN QRIS report
// @Security BearerAuth
// @Param request body dto.BTNMPMReportRequest true "Report filters"
// @Success 200 {object} dto.BtnSnapMPMReportResponse
// @Router /v1/btn-qris/reports [post]
func (h *QrisMPMHandler) Report(c *gin.Context) {
	var req dto.BTNMPMReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "error": "VALIDATION_ERROR", "message": err.Error()})
		return
	}
	result, err := h.service.Report(c.Request.Context(), req)
	if err != nil {
		writeMPMError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func writeMPMError(c *gin.Context, err error) {
	var unknown *service.BtnSnapOutcomeUnknownError
	if errors.As(err, &unknown) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "PAYMENT_OUTCOME_UNKNOWN", "message": "Refund result is unconfirmed; query the payment before retrying."})
		return
	}
	status, code := http.StatusBadGateway, "BTN_MPM_ERROR"
	switch {
	case strings.Contains(err.Error(), "not found"):
		status, code = http.StatusNotFound, "NOT_FOUND"
	case strings.Contains(err.Error(), "disabled"):
		status, code = http.StatusServiceUnavailable, "FEATURE_DISABLED"
	case strings.Contains(err.Error(), "only a verified"), strings.Contains(err.Error(), "must contain"), strings.Contains(err.Error(), "must be"), strings.Contains(err.Error(), "future"):
		status, code = http.StatusBadRequest, "VALIDATION_ERROR"
	}
	c.JSON(status, gin.H{"success": false, "error": code, "message": err.Error()})
}
