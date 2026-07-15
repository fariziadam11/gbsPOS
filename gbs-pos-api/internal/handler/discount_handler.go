package handler

import (
	"errors"
	"gbs-common/pkg/response"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DiscountHandler struct {
	discountService *service.DiscountService
	pricingService  *service.PricingService
}

func NewDiscountHandler(
	discountService *service.DiscountService,
	pricingService *service.PricingService,
) *DiscountHandler {
	return &DiscountHandler{discountService: discountService, pricingService: pricingService}
}

func (h *DiscountHandler) List(c *gin.Context) {
	productID, err := parseOptionalUintQuery(c.Query("productId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid productId"))
		return
	}

	discounts, err := h.discountService.List(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(discounts))
}

func (h *DiscountHandler) Get(c *gin.Context) {
	id, ok := parseDiscountID(c)
	if !ok {
		return
	}

	discount, err := h.discountService.Get(id)
	if err != nil {
		h.writeDiscountError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(discount))
}

func (h *DiscountHandler) Calculate(c *gin.Context) {
	if h.pricingService == nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", "Pricing service is not configured"))
		return
	}

	var req dto.PricingCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}

	result, err := h.pricingService.Calculate(req)
	if err != nil {
		h.writeDiscountError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

func (h *DiscountHandler) Create(c *gin.Context) {
	var req dto.CreateDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}

	discount, err := h.discountService.Create(req)
	if err != nil {
		h.writeDiscountError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success(discount))
}

func (h *DiscountHandler) Update(c *gin.Context) {
	id, ok := parseDiscountID(c)
	if !ok {
		return
	}

	var req dto.UpdateDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}

	discount, err := h.discountService.Update(id, req)
	if err != nil {
		h.writeDiscountError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(discount))
}

func (h *DiscountHandler) Stop(c *gin.Context) {
	id, ok := parseDiscountID(c)
	if !ok {
		return
	}

	discount, err := h.discountService.Stop(id)
	if err != nil {
		h.writeDiscountError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(discount))
}

func (h *DiscountHandler) Cancel(c *gin.Context) {
	id, ok := parseDiscountID(c)
	if !ok {
		return
	}

	discount, err := h.discountService.Cancel(id)
	if err != nil {
		h.writeDiscountError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(discount))
}

func (h *DiscountHandler) Delete(c *gin.Context) {
	id, ok := parseDiscountID(c)
	if !ok {
		return
	}

	if err := h.discountService.Delete(id); err != nil {
		h.writeDiscountError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *DiscountHandler) writeDiscountError(c *gin.Context, err error) {
	var validationErr *service.DiscountValidationError
	switch {
	case errors.As(err, &validationErr):
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", validationErr.Message))
	case errors.Is(err, service.ErrDiscountNotFound):
		c.JSON(http.StatusNotFound, response.Error("DISCOUNT_NOT_FOUND", "Discount not found"))
	case errors.Is(err, service.ErrDiscountProductNotFound):
		c.JSON(http.StatusNotFound, response.Error("PRODUCT_NOT_FOUND", "Product not found"))
	case errors.Is(err, service.ErrDiscountOverlap):
		c.JSON(http.StatusConflict, response.Error("DISCOUNT_PERIOD_OVERLAP", "Discount period overlaps with another active or pending discount for this product"))
	case errors.Is(err, service.ErrDiscountInvalidStatus):
		c.JSON(http.StatusConflict, response.Error("INVALID_DISCOUNT_STATUS", "Discount cannot be changed to the requested status"))
	case errors.Is(err, service.ErrVoucherNotFound):
		c.JSON(http.StatusNotFound, response.Error("VOUCHER_NOT_FOUND", "Voucher not found"))
	case errors.Is(err, service.ErrVoucherInvalid):
		c.JSON(http.StatusConflict, response.Error("VOUCHER_INVALID", "Voucher is not active"))
	case errors.Is(err, service.ErrVoucherMinimumNotMet):
		c.JSON(http.StatusConflict, response.Error("VOUCHER_MIN_TRANSACTION_NOT_MET", "Voucher minimum transaction is not met"))
	case errors.Is(err, service.ErrPricingEmptyCart):
		c.JSON(http.StatusBadRequest, response.Error("PRICING_EMPTY_CART", "Cart items cannot be empty"))
	case errors.Is(err, service.ErrPricingInvalidQuantity):
		c.JSON(http.StatusBadRequest, response.Error("PRICING_INVALID_QUANTITY", "Item quantity must be greater than zero"))
	default:
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
	}
}

func parseDiscountID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid discount ID"))
		return 0, false
	}
	return uint(id), true
}

func parseOptionalUintQuery(value string) (uint, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}
