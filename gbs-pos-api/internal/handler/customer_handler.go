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

type CustomerHandler struct {
	customerService *service.CustomerService
}

func NewCustomerHandler(customerService *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService: customerService}
}

// List godoc
//
//	@Summary		List customers
//	@Description	Search/list customers by query
//	@Tags			Customers
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q	query	string	false	"Search query (name, phone, email)"
//	@Success		200
//	@Failure		401
//	@Router				/v1/customers [get]
func (h *CustomerHandler) List(c *gin.Context) {
	query := c.Query("q")
	customers, err := h.customerService.List(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(customers))
}

// Get godoc
//
//	@Summary		Get customer with order history
//	@Description	Get customer details and purchase history
//	@Tags			Customers
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Customer ID"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		404
//	@Router				/v1/customers/{id} [get]
func (h *CustomerHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid customer ID"))
		return
	}
	customer, orders, err := h.customerService.GetCustomerWithHistory(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.Error("CUSTOMER_NOT_FOUND", "Customer not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	totalSpent := 0.0
	for _, o := range orders {
		totalSpent += o.Total
	}
	c.JSON(http.StatusOK, response.Success(dto.CustomerResponse{
		Customer:     *customer,
		OrderHistory: orders,
		TotalSpent:   totalSpent,
		TotalOrders:  len(orders),
	}))
}

// GetByPhone godoc
//
//	@Summary		Get customer by phone
//	@Description	Find customer by phone number
//	@Tags			Customers
//	@Produce		json
//	@Security		BearerAuth
//	@Param			phone	path	string	true	"Phone number"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		404
//	@Router				/v1/customers/phone/{phone} [get]
func (h *CustomerHandler) GetByPhone(c *gin.Context) {
	phone := c.Param("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Phone number required"))
		return
	}
	customer, err := h.customerService.GetByPhone(phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.Error("CUSTOMER_NOT_FOUND", "Customer not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(customer))
}

// Create godoc
//
//	@Summary		Create customer
//	@Description	Create a new customer
//	@Tags			Customers
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	dto.CreateCustomerRequest	true	"Customer data"
//	@Success		201
//	@Failure		401
//	@Failure		422
//	@Router				/v1/customers [post]
func (h *CustomerHandler) Create(c *gin.Context) {
	var req dto.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	customer := &model.Customer{
		Name:    req.Name,
		Phone:   req.Phone,
		Email:   req.Email,
		Address: req.Address,
	}
	if err := h.customerService.Create(customer); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success(customer))
}

// Update godoc
//
//	@Summary		Update customer
//	@Description	Update customer details
//	@Tags			Customers
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"Customer ID"
//	@Param			request	body	model.Customer	true	"Customer update data"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		404
//	@Failure		422
//	@Router				/v1/customers/{id} [put]
func (h *CustomerHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid customer ID"))
		return
	}
	var updates model.Customer
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	customer, err := h.customerService.Update(uint(id), &updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.Error("CUSTOMER_NOT_FOUND", "Customer not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(customer))
}

// GetOrders godoc
//
//	@Summary		Get customer orders
//	@Description	Get all orders for a customer
//	@Tags			Customers
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Customer ID"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Router				/v1/customers/{id}/orders [get]
func (h *CustomerHandler) GetOrders(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid customer ID"))
		return
	}
	orders, err := h.customerService.GetOrderHistory(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(orders))
}
