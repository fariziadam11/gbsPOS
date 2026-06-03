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

func (h *ProductHandler) GetLowStock(c *gin.Context) {
	products, err := h.productService.GetLowStockProducts(0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(products))
}

func (h *ProductHandler) List(c *gin.Context) {
	storeType := c.Query("storeType")
	category := c.Query("category")
	lastSync, _ := strconv.ParseInt(c.Query("lastSync"), 10, 64)
	products, err := h.productService.List(storeType, category, lastSync)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.Header("X-Last-Sync", strconv.FormatInt(time.Now().UnixMilli(), 10))
	c.JSON(http.StatusOK, response.Success(products))
}

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

func (h *ProductHandler) ExportCSV(c *gin.Context) {
	storeType := c.Query("storeType")

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=products_export.csv")

	if err := h.productService.ExportCSV(c.Writer, storeType); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("EXPORT_ERROR", err.Error()))
		return
	}
}
