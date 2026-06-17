package handler

import (
	"errors"
	"gbs-common/pkg/response"
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"
	"gbs-pos-api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HoldHandler struct {
	service *service.HoldService
}

func NewHoldHandler(service *service.HoldService) *HoldHandler {
	return &HoldHandler{service: service}
}

func (h *HoldHandler) Create(c *gin.Context) {
	var req dto.CreateHoldRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity,
			response.ValidationError("Invalid request body", nil))
		return
	}

	result, err := h.service.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success(toHoldResponse(result)))
}

func (h *HoldHandler) List(c *gin.Context) {
	terminalID := c.Query("terminalId")

	result, err := h.service.List(terminalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(toHoldResponses(result)))
}

func (h *HoldHandler) Get(c *gin.Context) {
	id := c.Param("id")

	result, err := h.service.Get(id)
	if err != nil {
		h.writeHoldError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(toHoldResponse(result)))
}

func (h *HoldHandler) Resume(c *gin.Context) {
	id := c.Param("id")

	session, err := h.service.Resume(id)
	if err != nil {
		h.writeHoldError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(toHoldResponse(session)))
}

func (h *HoldHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.service.Delete(id)
	if err != nil {
		h.writeHoldError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *HoldHandler) writeHoldError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, response.Error("HOLD_NOT_FOUND", "Hold session not found"))
	case errors.Is(err, service.ErrHoldCannotResumeNonActive):
		c.JSON(http.StatusConflict, response.Error("CANNOT_RESUME_NON_ACTIVE_HOLD", "Only ACTIVE hold can be resumed"))
	case errors.Is(err, service.ErrHoldCannotDeleteNonActive):
		c.JSON(http.StatusConflict, response.Error("CANNOT_DELETE_NON_ACTIVE_HOLD", "Only ACTIVE hold can be deleted"))
	default:
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
	}
}

func toHoldResponse(session *model.HoldSession) dto.HoldResponse {
	return dto.HoldResponse{
		ID:         session.ID,
		StoreType:  session.StoreType,
		TerminalID: session.TerminalID,
		Payload:    []byte(session.Payload),
		Total:      session.Total,
		Status:     string(session.Status),
		CreatedAt:  session.CreatedAt,
		UpdatedAt:  session.UpdatedAt,
	}
}

func toHoldResponses(sessions []model.HoldSession) []dto.HoldResponse {
	responses := make([]dto.HoldResponse, 0, len(sessions))
	for i := range sessions {
		responses = append(responses, toHoldResponse(&sessions[i]))
	}
	return responses
}
