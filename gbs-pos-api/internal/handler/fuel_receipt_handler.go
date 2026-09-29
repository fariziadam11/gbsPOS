package handler

import (
	"gbs-common/pkg/response"
	"gbs-pos-api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetReceiptJSON godoc
//
//	@Summary		Public receipt (JSON)
//	@Description	Return a safe JSON subset of a fuel sale by its public receipt token
//	@Tags			Fuel Receipt
//	@Produce		json
//	@Param			token	path	string	true	"Public receipt token"
//	@Success		200
//	@Failure		404
//	@Router			/v1/public/receipts/{token}/data [get]
func (h *FuelHandler) GetReceiptJSON(c *gin.Context) {
	receipt, err := h.fuelService.GetPublicReceipt(c.Param("token"))
	if err != nil {
		if service.IsFuelNotFound(err) {
			c.JSON(http.StatusNotFound, response.Error("NOT_FOUND", "Receipt not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(receipt))
}

// GetReceiptHTML godoc
//
//	@Summary		Public receipt page (HTML)
//	@Description	Render a mobile-friendly receipt page with an embedded authorization QR code
//	@Tags			Fuel Receipt
//	@Produce		html
//	@Param			token	path	string	true	"Public receipt token"
//	@Success		200
//	@Failure		404
//	@Router			/v1/public/receipts/{token} [get]
func (h *FuelHandler) GetReceiptHTML(c *gin.Context) {
	page, err := h.fuelService.RenderReceiptPage(c.Param("token"))
	if err != nil {
		if service.IsFuelNotFound(err) {
			c.JSON(http.StatusNotFound, response.Error("NOT_FOUND", "Receipt not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, page)
}

// ValidateAuthorization godoc
//
//	@Summary		Validate dispenser authorization token
//	@Description	Check a dispenser authorization token without side effects (demo / machine integration)
//	@Tags			Fuel Authorization
//	@Produce		json
//	@Param			token	path	string	true	"Authorization token"
//	@Success		200
//	@Failure		404
//	@Router			/v1/public/authorizations/{token} [get]
func (h *FuelHandler) ValidateAuthorization(c *gin.Context) {
	res, err := h.fuelService.ValidateAuthorization(c.Param("token"))
	if err != nil {
		if service.IsFuelNotFound(err) {
			c.JSON(http.StatusNotFound, response.Error("NOT_FOUND", "Authorization token not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(res))
}

// ScanAuthorize godoc
//
//	@Summary		Authorize dispenser from scan URL (GET, demo-friendly)
//	@Description	Single-use PAID → AUTHORIZED. Opening this URL in a browser
//	@Description	(phone camera scan of the receipt QR) performs the authorize and
//	@Description	returns a mobile result page. Same single-use semantics as POST
//	@Description	/authorize: a second scan shows an "already used" page.
//	@Tags			Fuel Authorization
//	@Produce		html
//	@Param			token	path	string	true	"Authorization token"
//	@Success		200
//	@Router			/v1/public/authorizations/{token}/scan [get]
func (h *FuelHandler) ScanAuthorize(c *gin.Context) {
	page, err := h.fuelService.AuthorizeScanPage(c.Param("token"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, page)
}

// Authorize godoc
//
//	@Summary		Authorize dispenser (single-use)
//	@Description	Transitions a PAID sale to AUTHORIZED. Calling twice returns 409 (already used).
//	@Tags			Fuel Authorization
//	@Produce		json
//	@Param			token	path	string	true	"Authorization token"
//	@Success		200
//	@Failure		404
//	@Failure		409
//	@Router			/v1/public/authorizations/{token}/authorize [post]
func (h *FuelHandler) Authorize(c *gin.Context) {
	res, alreadyUsed, err := h.fuelService.Authorize(c.Param("token"))
	if err != nil {
		if service.IsFuelNotFound(err) {
			c.JSON(http.StatusNotFound, response.Error("NOT_FOUND", "Authorization token not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	if alreadyUsed {
		c.JSON(http.StatusConflict, response.Error("ALREADY_USED", "Authorization token has already been used"))
		return
	}
	c.JSON(http.StatusOK, response.Success(res))
}
