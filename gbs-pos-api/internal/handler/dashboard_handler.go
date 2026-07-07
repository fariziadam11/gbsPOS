package handler

import (
	"gbs-pos-api/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"gbs-common/pkg/response"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// Summary godoc
//
//	@Summary		Get dashboard summary
//	@Description	Get summary stats for dashboard
//	@Tags			Dashboard
//	@Produce		json
//	@Security		BearerAuth
//	@Param			storeType	query	string	false	"Store type filter"
//	@Param			startDate	query	string	false	"Start date (YYYY-MM-DD)"
//	@Param			endDate	query	string	false	"End date (YYYY-MM-DD)"
//	@Success		200
//	@Failure		401
//	@Router				/v1/dashboard/summary [get]
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

// Revenue godoc
//
//	@Summary		Get revenue trend
//	@Description	Get revenue trend over time period
//	@Tags			Dashboard
//	@Produce		json
//	@Security		BearerAuth
//	@Param			storeType	query	string	false	"Store type filter"
//	@Param			startDate	query	string	false	"Start date (YYYY-MM-DD)"
//	@Param			endDate	query	string	false	"End date (YYYY-MM-DD)"
//	@Param			days		query	int		false	"Number of days (default: 7)"
//	@Success		200
//	@Failure		401
//	@Router				/v1/dashboard/revenue [get]
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

// TopProducts godoc
//
//	@Summary		Get top selling products
//	@Description	Get top selling products by revenue
//	@Tags			Dashboard
//	@Produce		json
//	@Security		BearerAuth
//	@Param			storeType	query	string	false	"Store type filter"
//	@Param			limit		query	int		false	"Limit (default: 10)"
//	@Success		200
//	@Failure		401
//	@Router				/v1/dashboard/top-products [get]
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
