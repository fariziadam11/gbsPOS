package handler

import (
	"net/http"
	"strconv"
	"gbs-pos-api/internal/service"
	"gbs-pos-api/pkg/response"
	"github.com/gin-gonic/gin"
)

type SettlementHandler struct {
	settlementService *service.SettlementService
}

func NewSettlementHandler(settlementService *service.SettlementService) *SettlementHandler {
	return &SettlementHandler{settlementService: settlementService}
}

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

func (h *SettlementHandler) Get(c *gin.Context) {
	id := c.Param("id")
	settlement, err := h.settlementService.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("SETTLEMENT_NOT_FOUND", "Settlement with ID "+id+" not found"))
		return
	}
	c.JSON(http.StatusOK, response.Success(settlement))
}
