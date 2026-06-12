package handler

import (
	"gbs-common/pkg/response"
	"gbs-pos-api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PosHoldHandler struct {
	service *service.PosHoldService
}

func NewPosHoldHandler(service *service.PosHoldService) *PosHoldHandler {
	return &PosHoldHandler{service: service}
}

func (h *PosHoldHandler) Hold(c *gin.Context) {
	var req struct {
		StoreType  string      `json:"store_type"`
		TerminalID string      `json:"terminal_id"`
		Payload    interface{} `json:"payload"`
		Total      float64     `json:"total"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity,
			response.ValidationError("Invalid request body", nil))
		return
	}

	result, err := h.service.Hold(req.StoreType, req.TerminalID, req.Payload, req.Total)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success(result))
}

func (h *PosHoldHandler) List(c *gin.Context) {
	terminalID := c.Query("terminal_id")

	result, err := h.service.List(terminalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func (h *PosHoldHandler) Get(c *gin.Context) {
	id := c.Param("id")

	result, err := h.service.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound,
			response.Error("HOLD_NOT_FOUND", "Hold session not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func (h *PosHoldHandler) Resume(c *gin.Context) {
	id := c.Param("id")

	_, err := h.service.Resume(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}

	session, err := h.service.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound,
			response.Error("HOLD_NOT_FOUND", "Hold session not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(session))
}

func (h *PosHoldHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.service.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}