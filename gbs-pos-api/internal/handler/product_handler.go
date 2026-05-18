package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/service"
	"gbs-pos-api/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
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
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
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
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	product, err := h.productService.Update(uint(id), &updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.Error("PRODUCT_NOT_FOUND", "Product with ID "+idStr+" not found"))
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
			c.JSON(http.StatusNotFound, response.Error("PRODUCT_NOT_FOUND", "Product with ID "+idStr+" not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}
