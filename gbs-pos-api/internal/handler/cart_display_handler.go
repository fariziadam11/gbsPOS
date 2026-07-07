package handler

import (
	"errors"
	"net/http"

	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/service"

	"gbs-common/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CartDisplayHandler struct {
	service *service.CartDisplayService
}

func NewCartDisplayHandler(service *service.CartDisplayService) *CartDisplayHandler {
	return &CartDisplayHandler{service: service}
}

// Save godoc
//
//	@Summary		Upload cart display JSON
//	@Description	Receives cart display JSON from Android POS and stores it. Requires Bearer token authentication.
//	@Tags			Cart Display
//	@Accept		json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body	dto.SaveCartDisplayRequest	true	"Cart display JSON (must include terminalId field)"
//	@Success		200
//	@Failure		400
//	@Failure		401
//	@Failure		422
//	@Router				/v1/display/cart [post]
func (h *CartDisplayHandler) Save(c *gin.Context) {
	var req dto.SaveCartDisplayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity,
			response.ValidationError("Invalid request body", nil))
		return
	}

	payload, err := req.MarshalPayload()
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}

	if err := h.service.Save(req.TerminalID, datatypes.JSON(payload)); err != nil {
		c.JSON(http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]string{"terminalId": req.TerminalID}))
}

// Get godoc
//
//	@Summary		Get cart display JSON
//	@Description	Returns the latest stored cart display JSON for a terminal. Public endpoint - no authentication required.
//	@Tags			Cart Display
//	@Produce		json
//	@Param			terminalId	query	string	false	"Android terminal ID (ANDROID_ID)"
//	@Success		200		{string}	string	"JSON cart display data"
//	@Failure		400
//	@Failure		500
//	@Router				/v1/display/cart [get]
func (h *CartDisplayHandler) Get(c *gin.Context) {
	terminalID := c.Query("terminalId")
	if terminalID == "" {
		c.JSON(http.StatusBadRequest,
			response.Error("VALIDATION_ERROR", "terminalId is required"))
		return
	}

	payload, err := h.service.Get(terminalID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Data(http.StatusOK, "application/json", dto.DefaultCartDisplay)
			return
		}
		c.JSON(http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}

	c.Data(http.StatusOK, "application/json", payload)
}

// Delete godoc
//
//	@Summary		Delete cart display
//	@Description	Remove stored cart display for a terminal
//	@Tags			Cart Display
//	@Security		BearerAuth
//	@Param			terminalId	path	string	true	"Terminal ID"
//	@Success		204
//	@Failure		401
//	@Failure		404
//	@Failure		500
//	@Router				/v1/display/cart/{terminalId} [delete]
func (h *CartDisplayHandler) Delete(c *gin.Context) {
	terminalID := c.Param("terminalId")

	err := h.service.Delete(terminalID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound,
				response.Error("NOT_FOUND", "Cart display not found"))
			return
		}
		c.JSON(http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}
