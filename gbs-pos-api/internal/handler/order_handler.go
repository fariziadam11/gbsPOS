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

// List godoc
//
//	@Summary		List orders
//	@Description	Get orders with optional filters
//	@Tags			Orders
//	@Produce		json
//	@Security		BearerAuth
//	@Param			storeType		query		string	false	"Store type filter"
//	@Param			paymentMethod	query		string	false	"Payment method filter"
//	@Param			terminalId	query		string	false	"Terminal ID filter"
//	@Param			startDate		query		int64		false	"Start date Unix timestamp"
//	@Param			endDate		query		int64		false	"End date Unix timestamp"
//	@Param			isVoided		query		bool		false	"Filter voided orders"
//	@Param			isSettled		query		bool		false	"Filter settled orders"
//	@Success		200
//	@Failure		401
//	@Router				/v1/orders [get]
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

// Get godoc
//
//	@Summary		Get order by ID
//	@Description	Get a single order by its ID
//	@Tags			Orders
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Order ID"
//	@Success		200
//	@Failure		401
//	@Failure		404
//	@Router				/v1/orders/{id} [get]
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

// Create godoc
//
//	@Summary		Create order
//	@Description	Create a new order
//	@Tags			Orders
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	dto.CreateOrderRequest	true	"Order data"
//	@Success		201
//	@Success		200		"Idempotent response (order already exists"
//	@Failure		401
//	@Failure		422
//	@Router				/v1/orders [post]
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
			VariantID:    it.VariantID,
			VariantName:  it.VariantName,
			SKU:          it.SKU,
		}
	}
	newOrder := &model.Order{
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
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
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

// Void godoc
//
//	@Summary		Void order
//	@Description	Void/cancel an order (ADMIN only)
//	@Tags			Orders
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path				string				true	"Order ID"
//	@Param			request	body	dto.VoidOrderRequest	true	"Void reason"
//	@Success		200
//	@Failure		401
//	@Failure		403
//	@Failure		404
//	@Failure		409
//	@Router				/v1/orders/{id}/void [patch]
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
				response.Error("ORDER_NOT_FOUND", "Order "+id+" not found"),
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

// UnsettledSummary godoc
//
//	@Summary		Get unsettled orders summary
//	@Description	Get summary of unsettled orders for terminal
//	@Tags			Orders
//	@Produce		json
//	@Security		BearerAuth
//	@Param			storeType	query	string	false	"Store type filter"
//	@Param			terminalId	query	string	false	"Terminal ID filter"
//	@Success		200
//	@Failure		401
//	@Router				/v1/orders/unsettled/summary [get]
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

// BulkSync godoc
//
//	@Summary		Bulk sync orders
//	@Description	Sync multiple orders at once (offline support)
//	@Tags			Orders
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	dto.BulkSyncOrderRequest	true	"Orders to sync"
//	@Success		200
//	@Failure		401
//	@Failure		422
//	@Router				/v1/orders/bulk [post]
//	@Router				/v1/sync/orders [post]
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

// Settle godoc
//
//	@Summary		Settle orders
//	@Description	Settle/close orders for a terminal (ADMIN only)
//	@Tags			Orders
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	dto.SettleOrderRequest	true	"Settlement data"
//	@Success		200
//	@Failure		401
//	@Failure		403
//	@Failure		409
//	@Failure		422
//	@Router				/v1/orders/settle [post]
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
