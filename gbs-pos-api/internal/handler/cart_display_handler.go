package handler

import (
	"errors"
	"net/http"

	"gbs-common/pkg/response"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/service"

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

// Save receives a cart display JSON from Android and stores it as JSONB.
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

// Get returns the latest stored cart display JSON for a terminal.
// This endpoint is public and returns the raw JSON document without the standard envelope.
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

// Delete removes the stored cart display for a terminal.
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
