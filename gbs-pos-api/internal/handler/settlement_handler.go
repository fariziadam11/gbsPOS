package handler

import (
	"errors"
	"gbs-common/pkg/response"
	"gbs-pos-api/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SettlementHandler struct {
	settlementService *service.SettlementService
}

func NewSettlementHandler(settlementService *service.SettlementService) *SettlementHandler {
	return &SettlementHandler{settlementService: settlementService}
}

// List godoc
//
//	@Summary		List settlements
//	@Description	Get paginated list of settlements
//	@Tags			Settlements
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query	int		false	"Limit (default: 20)"
//	@Param			storeType	query	string	false	"Store type filter"
//	@Success		200
//	@Failure		401
//	@Router				/v1/settlements [get]
func (h *SettlementHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 20
	}
	storeType := c.Query("storeType")
	settlements, err := h.settlementService.List(limit, storeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(settlements))
}

// Get godoc
//
//	@Summary		Get settlement by ID
//	@Description	Get a single settlement by ID
//	@Tags			Settlements
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Settlement ID"
//	@Success		200
//	@Failure		401
//	@Failure		404
//	@Router				/v1/settlements/{id} [get]
func (h *SettlementHandler) Get(c *gin.Context) {
	id := c.Param("id")
	settlement, err := h.settlementService.Get(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(
				http.StatusNotFound,
				response.Error("SETTLEMENT_NOT_FOUND", "Settlement with ID "+id+" not found"),
			)
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(settlement))
}
