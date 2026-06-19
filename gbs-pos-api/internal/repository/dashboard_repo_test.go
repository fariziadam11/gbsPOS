package repository

import (
	"testing"
	"time"

	"gbs-pos-api/internal/database"
	"gbs-pos-api/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardRepository_GetRevenueTrend_FillsGaps(t *testing.T) {
	db, err := database.NewTestDB()
	require.NoError(t, err)

	repo := NewDashboardRepository(db)

	// Create an order 2 days ago
	orderDate := time.Now().AddDate(0, 0, -2)
	order := model.Order{
		ID:            "ORD-001",
		Subtotal:      100000,
		Tax:           10000,
		Total:         110000,
		PaymentMethod: "CASH",
		Timestamp:     orderDate.UnixMilli(),
		StoreType:     "RETAIL",
		IsVoided:      false,
		Items: []model.OrderItem{
			{ProductID: 1, ProductName: "Test Product", ProductPrice: 100000, Qty: 1, Subtotal: 100000},
		},
	}
	require.NoError(t, db.Create(&order).Error)

	endDate := time.Now().UTC().Truncate(24 * time.Hour)
	startDate := endDate.AddDate(0, 0, -6)
	points, err := repo.GetRevenueTrend("", startDate, endDate)
	require.NoError(t, err)
	require.Len(t, points, 7)

	// All points should have valid YYYY-MM-DD dates
	for _, p := range points {
		assert.Regexp(t, `^\d{4}-\d{2}-\d{2}$`, p.Date)
	}

	// Find the point matching our order date
	orderDateStr := orderDate.Format("2006-01-02")
	var found bool
	for _, p := range points {
		if p.Date == orderDateStr {
			found = true
			assert.Equal(t, 110000.0, p.Revenue)
			assert.Equal(t, 1, p.Orders)
		}
	}
	assert.True(t, found, "expected to find revenue point for %s", orderDateStr)

	// Other days should be zero
	for _, p := range points {
		if p.Date != orderDateStr {
			assert.Equal(t, 0.0, p.Revenue)
			assert.Equal(t, 0, p.Orders)
		}
	}
}

func TestDashboardRepository_GetRevenueTrend_IncludesLateDayOrder(t *testing.T) {
	db, err := database.NewTestDB()
	require.NoError(t, err)

	repo := NewDashboardRepository(db)

	// Order at 23:59:59 UTC on the current day should still be counted today
	now := time.Now().UTC()
	orderTime := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	order := model.Order{
		ID:            "ORD-LATE",
		Subtotal:      50000,
		Tax:           5000,
		Total:         55000,
		PaymentMethod: "QRIS",
		Timestamp:     orderTime.UnixMilli(),
		StoreType:     "OUTFIT",
		IsVoided:      false,
		Items: []model.OrderItem{
			{ProductID: 3, ProductName: "Test Product 3", ProductPrice: 50000, Qty: 1, Subtotal: 50000},
		},
	}
	require.NoError(t, db.Create(&order).Error)

	endDate := orderTime.UTC().Truncate(24 * time.Hour)
	startDate := endDate
	points, err := repo.GetRevenueTrend("", startDate, endDate)
	require.NoError(t, err)
	require.Len(t, points, 1)

	assert.Equal(t, formatDate(orderTime), points[0].Date)
	assert.Equal(t, 55000.0, points[0].Revenue)
	assert.Equal(t, 1, points[0].Orders)
}

func formatDate(d time.Time) string {
	return d.UTC().Format("2006-01-02")
}

func TestDashboardRepository_GetSummary_DateRange(t *testing.T) {
	db, err := database.NewTestDB()
	require.NoError(t, err)

	repo := NewDashboardRepository(db)

	orderDate := time.Now().AddDate(0, 0, -3)
	order := model.Order{
		ID:            "ORD-002",
		Subtotal:      200000,
		Tax:           20000,
		Total:         220000,
		PaymentMethod: "CARD",
		Timestamp:     orderDate.UnixMilli(),
		StoreType:     "FNB",
		IsVoided:      false,
		IsSettled:     true,
		Items: []model.OrderItem{
			{ProductID: 2, ProductName: "Test Product 2", ProductPrice: 200000, Qty: 1, Subtotal: 200000},
		},
	}
	require.NoError(t, db.Create(&order).Error)

	start := orderDate.Truncate(24 * time.Hour)
	end := start
	summary, err := repo.GetSummary("", start, end)
	require.NoError(t, err)

	assert.Equal(t, 1, summary.TotalOrders)
	assert.Equal(t, 220000.0, summary.TotalRevenue)
	assert.Equal(t, 220000.0, summary.AvgOrderValue)
	assert.Equal(t, 0.0, summary.CashTotal)
	assert.Equal(t, 220000.0, summary.CardTotal)
	assert.Equal(t, 0.0, summary.QrisTotal)

	// Different date range with no orders
	emptyStart := time.Now().AddDate(0, 0, -10)
	emptyEnd := time.Now().AddDate(0, 0, -9)
	emptySummary, err := repo.GetSummary("", emptyStart, emptyEnd)
	require.NoError(t, err)
	assert.Equal(t, 0, emptySummary.TotalOrders)
}
