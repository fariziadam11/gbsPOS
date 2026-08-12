package handler

import (
	"net/http"

	"gbs-pos-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CardPaymentHandler struct {
	service *service.CardPaymentService
}

func NewCardPaymentHandler(paymentService *service.CardPaymentService) *CardPaymentHandler {
	return &CardPaymentHandler{service: paymentService}
}

type createCardPaymentRequest struct {
	OrderID    string  `json:"orderId" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	TerminalID string  `json:"terminalId" binding:"required"`
	DeviceID   string  `json:"deviceId" binding:"required"`
}

func (h *CardPaymentHandler) Init(c *gin.Context) {
	var req createCardPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "VALIDATION_ERROR", "message": err.Error()})
		return
	}
	payment, err := h.service.Create(c.Request.Context(), req.OrderID, req.Amount, req.TerminalID, req.DeviceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "PAYMENT_INIT_ERROR", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": payment})
}

func (h *CardPaymentHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "INVALID_PAYMENT_ID"})
		return
	}
	payment, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "PAYMENT_NOT_FOUND"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": payment})
}

func (h *CardPaymentHandler) Pending(c *gin.Context) {
	deviceID := c.Query("deviceId")
	if deviceID == "" {
		deviceID = c.Query("device_id")
	}
	payments, err := h.service.Pending(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "VALIDATION_ERROR", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": payments})
}

func (h *CardPaymentHandler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "INVALID_PAYMENT_ID"})
		return
	}
	payment, err := h.service.Cancel(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "PAYMENT_CANCEL_ERROR", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": payment})
}
