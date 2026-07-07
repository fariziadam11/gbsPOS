package handler

import (
	"errors"
	"gbs-common/pkg/response"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// AdjustStock godoc
//
//	@Summary		Adjust product stock
//	@Description	Adjust stock quantity (IN/OUT/ADJUSTMENT) for a product
//	@Tags			Products
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"Product ID"
//	@Param			request	body	dto.AdjustStockRequest	true	"Stock adjustment data"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		404
//	@Failure		409
//	@Router			/v1/products/{id}/stock [post]
func (h *ProductHandler) AdjustStock(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid product ID"))
		return
	}
	var req dto.AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	user := c.GetString("username")
	if err := h.productService.AdjustStock(uint(id), req.Type, req.Quantity, req.Reason, user); err != nil {
		switch err.Error() {
		case "PRODUCT_NOT_FOUND":
			c.JSON(http.StatusNotFound, response.Error("PRODUCT_NOT_FOUND", "Product not found"))
		case "INSUFFICIENT_STOCK":
			c.JSON(http.StatusConflict, response.Error("INSUFFICIENT_STOCK", "Not enough stock for this adjustment"))
		case "INVALID_ADJUSTMENT_TYPE":
			c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid adjustment type"))
		default:
			c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		}
		return
	}
	c.JSON(http.StatusOK, response.Success(map[string]string{"status": "ok"}))
}

// GetStockHistory godoc
//
//	@Summary		Get product stock history
//	@Description	Get stock movement history for a product
//	@Tags			Products
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Product ID"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Router			/v1/products/{id}/history [get]
func (h *ProductHandler) GetStockHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid product ID"))
		return
	}
	movements, err := h.productService.GetStockHistory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(movements))
}

// GetLowStock godoc
//
//	@Summary		Get low stock products
//	@Description	Get list of products with stock below threshold
//	@Tags			Products
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200
//	@Failure		401
//	@Router			/v1/products/low-stock [get]
func (h *ProductHandler) GetLowStock(c *gin.Context) {
	products, err := h.productService.GetLowStockProducts(0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(products))
}

// List godoc
//
//	@Summary		List products
//	@Description	Get all products with optional filtering and sync timestamp
//	@Tags			Products
//	@Produce		json
//	@Security		BearerAuth
//	@Param			storeType	query	string	false	"Store type filter (RETAIL, FOOD, OUTFIT)"
//	@Param			category	query	string	false	"Category filter"
//	@Param			lastSync	query	int64		false	"Unix timestamp for incremental sync"
//	@Success		200		{header}	X-Last-Sync
//	@Failure		401
//	@Router			/v1/products [get]
func (h *ProductHandler) List(c *gin.Context) {
	storeType := c.Query("storeType")
	category := c.Query("category")
	lastSync, _ := strconv.ParseInt(c.Query("lastSync"), 10, 64)
	products, err := h.productService.ListWithDiscounts(storeType, category, lastSync)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.Header("X-Last-Sync", strconv.FormatInt(time.Now().UnixMilli(), 10))
	c.JSON(http.StatusOK, response.Success(products))
}

// Get godoc
//
//	@Summary		Get product by ID
//	@Description	Get a single product with active discounts
//	@Tags			Products
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Product ID"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		404
//	@Router			/v1/products/{id} [get]
func (h *ProductHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid product ID"))
		return
	}

	product, err := h.productService.GetWithDiscount(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(
				http.StatusNotFound,
				response.Error("PRODUCT_NOT_FOUND", "Product with ID "+idStr+" not found"),
			)
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(product))
}

// Create godoc
//
//	@Summary		Create product
//	@Description	Create a new product (ADMIN only)
//	@Tags			Products
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	model.Product	true	"Product data"
//	@Success		201
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		422
//	@Router			/v1/products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var product model.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(
			http.StatusUnprocessableEntity,
			response.ValidationError("Invalid request body", nil),
		)
		return
	}
	if err := h.productService.Create(&product); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success(product))
}

// Update godoc
//
//	@Summary		Update product
//	@Description	Update an existing product (ADMIN only)
//	@Tags			Products
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int			true	"Product ID"
//	@Param			request	body	model.Product	true	"Product update data"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		404
//	@Failure		422
//	@Router			/v1/products/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid product ID"))
		return
	}
	var updates model.Product
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(
			http.StatusUnprocessableEntity,
			response.ValidationError("Invalid request body", nil),
		)
		return
	}
	product, err := h.productService.Update(uint(id), &updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(
				http.StatusNotFound,
				response.Error("PRODUCT_NOT_FOUND", "Product with ID "+idStr+" not found"),
			)
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(product))
}

// Delete godoc
//
//	@Summary		Delete product
//	@Description	Delete a product by ID (ADMIN only)
//	@Tags			Products
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Product ID"
//	@Success		204
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		404
//	@Router			/v1/products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid product ID"))
		return
	}
	if err := h.productService.Delete(uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(
				http.StatusNotFound,
				response.Error("PRODUCT_NOT_FOUND", "Product with ID "+idStr+" not found"),
			)
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}

// ImportCSV godoc
//
//	@Summary		Import products from CSV
//	@Description	Import products from CSV file (ADMIN only)
//	@Tags			Products
//	@Accept		multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file		formData	file		true	"CSV file"
//	@Param			storeType	formData	string	false	"Store type"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Router			/v1/products/import [post]
func (h *ProductHandler) ImportCSV(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "File is required"))
		return
	}
	defer file.Close()

	storeType := c.DefaultPostForm("storeType", "")

	result, err := h.productService.ImportCSV(file, header.Filename, storeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("IMPORT_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

// ExportCSV godoc
//
//	@Summary		Export products to CSV
//	@Description	Export all products to CSV file (ADMIN only)
//	@Tags			Products
//	@Produce		text/csv
//	@Security		BearerAuth
//	@Param			storeType	query	string	false	"Store type filter"
//	@Success		200		{file}	text/csv
//	@Failure		401
//	@Failure		403
//	@Router			/v1/products/export [get]
func (h *ProductHandler) ExportCSV(c *gin.Context) {
	storeType := c.Query("storeType")

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=products_export.csv")

	if err := h.productService.ExportCSV(c.Writer, storeType); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("EXPORT_ERROR", err.Error()))
		return
	}
}
