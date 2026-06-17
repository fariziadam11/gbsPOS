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

type PosHoldHandler struct {
	service *service.PosHoldService
}

func NewPosHoldHandler(service *service.PosHoldService) *PosHoldHandler {
	return &PosHoldHandler{service: service}
}

func (h *PosHoldHandler) Hold(c *gin.Context) {
	var req dto.CreatePosHoldRequest

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

	c.JSON(http.StatusCreated, response.Success(toPosHoldResponse(result)))
}

func (h *PosHoldHandler) List(c *gin.Context) {
	terminalID := c.Query("terminalId")

	result, err := h.service.List(terminalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(toPosHoldResponses(result)))
}

func (h *PosHoldHandler) Get(c *gin.Context) {
	id := c.Param("id")

	result, err := h.service.Get(id)
	if err != nil {
		h.writeHoldError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(toPosHoldResponse(result)))
}

func (h *PosHoldHandler) Resume(c *gin.Context) {
	id := c.Param("id")

	session, err := h.service.Resume(id)
	if err != nil {
		h.writeHoldError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(toPosHoldResponse(session)))
}

func (h *PosHoldHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.service.Delete(id)
	if err != nil {
		h.writeHoldError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PosHoldHandler) writeHoldError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, response.Error("HOLD_NOT_FOUND", "Hold session not found"))
	case err.Error() == "CANNOT_RESUME_NON_ACTIVE_HOLD":
		c.JSON(http.StatusConflict, response.Error("CANNOT_RESUME_NON_ACTIVE_HOLD", "Only ACTIVE hold can be resumed"))
	case err.Error() == "CANNOT_DELETE_NON_ACTIVE_HOLD":
		c.JSON(http.StatusConflict, response.Error("CANNOT_DELETE_NON_ACTIVE_HOLD", "Only ACTIVE hold can be deleted"))
	default:
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
	}
}

func toPosHoldResponse(session *model.PosHoldSession) dto.PosHoldResponse {
	return dto.PosHoldResponse{
		ID:         session.ID,
		StoreType:  session.StoreType,
		TerminalID: session.TerminalID,
		Payload:    []byte(session.Payload),
		Total:      session.Total,
		Status:     session.Status,
		CreatedAt:  session.CreatedAt,
		UpdatedAt:  session.UpdatedAt,
	}
}

func toPosHoldResponses(sessions []model.PosHoldSession) []dto.PosHoldResponse {
	responses := make([]dto.PosHoldResponse, 0, len(sessions))
	for i := range sessions {
		responses = append(responses, toPosHoldResponse(&sessions[i]))
	}
	return responses
}
