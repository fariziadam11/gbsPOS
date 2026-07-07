package handler

import (
	"errors"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gbs-common/pkg/response"
)

type ProductVariantHandler struct {
	variantService *service.ProductVariantService
}

func NewProductVariantHandler(variantService *service.ProductVariantService) *ProductVariantHandler {
	return &ProductVariantHandler{variantService: variantService}
}

// List godoc
//
//	@Summary		List product variants
//	@Description	Get all variants for a product
//	@Tags			Product Variants
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Product ID"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Router				/v1/products/{id}/variants [get]
func (h *ProductVariantHandler) List(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid product ID"))
		return
	}
	variants, err := h.variantService.ListByProduct(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(variants))
}

// Create godoc
//
//	@Summary		Create variant
//	@Description	Create a new product variant
//	@Tags			Product Variants
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int							true	"Product ID"
//	@Param			request	body	dto.CreateVariantRequest	true	"Variant data"
//	@Success		201
//	@Failure		400
//	@Failure		401
//	@Failure		422
//	@Router				/v1/products/{id}/variants [post]
func (h *ProductVariantHandler) Create(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid product ID"))
		return
	}
	var req dto.CreateVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	variant, err := h.variantService.Create(productID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, response.Success(variant))
}

// Update godoc
//
//	@Summary		Update variant
//	@Description	Update variant details
//	@Tags			Product Variants
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	int							true	"Variant ID"
//	@Param			request	body	dto.UpdateVariantRequest	true	"Update data"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		404
//	@Failure		422
//	@Router				/v1/products/{id}/variants/{id} [patch]
func (h *ProductVariantHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid variant ID"))
		return
	}
	var req dto.UpdateVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.ValidationError("Invalid request body", nil))
		return
	}
	variant, err := h.variantService.Update(id, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.Error("VARIANT_NOT_FOUND", "Variant not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(variant))
}

// Delete godoc
//
//	@Summary		Delete variant
//	@Description	Delete a product variant
//	Tags			Product Variants
//	Security		BearerAuth
//	Param			id	path	int	true	"Variant ID"
//	Success		204
//	Failure		400
//	Failure		401
//	Router				/v1/products/{id}/variants/{id} [delete]
func (h *ProductVariantHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("VALIDATION_ERROR", "Invalid variant ID"))
		return
	}
	if err := h.variantService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}
