package handler

import (
	"errors"
	"gbs-common/pkg/response"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	orderService      *service.OrderService
	settlementService *service.SettlementService
}

func NewOrderHandler(
	orderService *service.OrderService,
	settlementService *service.SettlementService,
) *OrderHandler {
	return &OrderHandler{orderService: orderService, settlementService: settlementService}
}

func (h *OrderHandler) List(c *gin.Context) {
	storeType := c.Query("storeType")
	paymentMethod := c.Query("paymentMethod")
	terminalID := c.Query("terminalId")
	startDate, _ := strconv.ParseInt(c.Query("startDate"), 10, 64)
	endDate, _ := strconv.ParseInt(c.Query("endDate"), 10, 64)
	var isVoided, isSettled *bool
	if v := c.Query("isVoided"); v != "" {
		b := v == "true"
		isVoided = &b
	}
	if v := c.Query("isSettled"); v != "" {
		b := v == "true"
		isSettled = &b
	}
	orders, err := h.orderService.List(
		storeType,
		startDate,
		endDate,
		isVoided,
		isSettled,
		paymentMethod,
		terminalID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(orders))
}

func (h *OrderHandler) Get(c *gin.Context) {
	id := c.Param("id")
	order, err := h.orderService.Get(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(
				http.StatusNotFound,
				response.Error("ORDER_NOT_FOUND", "Order with ID "+id+" not found"),
			)
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(order))
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req dto.CreateOrderRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusUnprocessableEntity,
			response.ValidationError("Invalid request body", nil),
		)
		return
	}
	items := make([]model.OrderItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = model.OrderItem{
			ProductID:    it.ProductID,
			ProductName:  it.ProductName,
			ProductPrice: it.ProductPrice,
			Qty:          it.Qty,
			Subtotal:     it.Subtotal,
		}
	}
	newOrder := &model.Order{
		ID:            req.ID,
		Items:         items,
		Subtotal:      req.Subtotal,
		Tax:           req.Tax,
		Total:         req.Total,
		PaymentMethod: req.PaymentMethod,
		CashReceived:  req.CashReceived,
		ChangeAmount:  req.ChangeAmount,
		Timestamp:     req.Timestamp,
		StoreType:     req.StoreType,
		TerminalID:    req.TerminalID,
		TransactionID: req.TransactionID,
		ApprovalCode:  req.ApprovalCode,
		EntryMode:     req.EntryMode,
		MaskedAccount: req.MaskedAccount,
		AcqMid:        req.AcqMid,
		AcqTid:        req.AcqTid,
		PosMessageID:  req.PosMessageID,
		BankName:      req.BankName,
		CustomerID:    req.CustomerID,
		CustomerPhone: req.CustomerPhone,
		CustomerName:  req.CustomerName,
	}
	if err := service.ValidateOrder(newOrder); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError(err.Error(), nil))
		return
	}
	result, idempotent, err := h.orderService.Create(newOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	if idempotent {
		c.JSON(http.StatusOK, response.SuccessIdempotent(result))
		return
	}
	c.JSON(http.StatusCreated, response.Success(result))
}

func (h *OrderHandler) Void(c *gin.Context) {
	id := c.Param("id")
	var req dto.VoidOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusUnprocessableEntity,
			response.ValidationError("Invalid request body", nil),
		)
		return
	}
	voidedBy := c.GetString("username")
	order, err := h.orderService.Void(id, req.Reason, voidedBy)
	if err != nil {
		switch err.Error() {
		case "ORDER_NOT_FOUND":
			c.JSON(
				http.StatusNotFound,
				response.Error("ORDER_NOT_FOUND", "Order with ID "+id+" not found"),
			)
		case "ORDER_ALREADY_VOIDED":
			c.JSON(
				http.StatusConflict,
				response.Error("ORDER_ALREADY_VOIDED", "Order "+id+" has already been voided"),
			)
		case "ORDER_ALREADY_SETTLED":
			c.JSON(
				http.StatusConflict,
				response.Error("ORDER_ALREADY_SETTLED", "Cannot void a settled order"),
			)
		default:
			c.JSON(
				http.StatusInternalServerError,
				response.Error("INTERNAL_SERVER_ERROR", err.Error()),
			)
		}
		return
	}
	c.JSON(http.StatusOK, response.Success(order))
}

func (h *OrderHandler) UnsettledSummary(c *gin.Context) {
	storeType := c.Query("storeType")
	terminalID := c.Query("terminalId")
	summary, err := h.settlementService.GetUnsettledSummary(storeType, terminalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(summary))
}

func (h *OrderHandler) BulkSync(c *gin.Context) {
	var req dto.BulkSyncOrderRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusUnprocessableEntity,
			response.ValidationError("Invalid request body", nil),
		)
		return
	}
	for i := range req.Orders {
		if req.Orders[i].TerminalID == "" {
			req.Orders[i].TerminalID = req.TerminalID
		}
	}
	result, err := h.orderService.BulkCreate(req.Orders)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

func (h *OrderHandler) Settle(c *gin.Context) {
	var req dto.SettleOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusUnprocessableEntity,
			response.ValidationError("Invalid request body", nil),
		)
		return
	}
	settlement, err := h.settlementService.Settle(
		req.SettlementID,
		req.Timestamp,
		req.StoreType,
		req.TerminalID,
	)
	if err != nil {
		if err.Error() == "NO_UNSETTLED_ORDERS" {
			c.JSON(
				http.StatusConflict,
				response.Error("NO_UNSETTLED_ORDERS", "There are no unsettled orders to settle"),
			)
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(settlement))
}
