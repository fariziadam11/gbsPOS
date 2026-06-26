package handler

import (
	"gbs-common/pkg/response"
	"gbs-pos-api/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

func (h *DashboardHandler) Summary(c *gin.Context) {
	storeType := c.Query("storeType")
	startDate, endDate := parseDashboardDateRange(c)

	summary, err := h.dashboardService.GetSummary(storeType, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(summary))
}

func parseDashboardDateRange(c *gin.Context) (startDate, endDate *time.Time) {
	if startStr := c.Query("startDate"); startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = &parsed
		}
	}
	if endStr := c.Query("endDate"); endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = &parsed
		}
	}
	return
}

func (h *DashboardHandler) Revenue(c *gin.Context) {
	storeType := c.Query("storeType")
	startDate, endDate := parseDashboardDateRange(c)
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	points, err := h.dashboardService.GetRevenueTrend(storeType, startDate, endDate, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(points))
}

func (h *DashboardHandler) TopProducts(c *gin.Context) {
	storeType := c.Query("storeType")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	products, err := h.dashboardService.GetTopProducts(storeType, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_SERVER_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(products))
}
