package repository

import (
	"gbs-pos-api/internal/dto"
	"time"

	"gorm.io/gorm"
)

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

// dateExpr returns a SQL expression that extracts the date as a YYYY-MM-DD string
// from the order's business timestamp (milliseconds since epoch).
func (r *DashboardRepository) dateExpr() string {
	if r.db.Dialector.Name() == "sqlite" {
		return "DATE(timestamp / 1000, 'unixepoch')"
	}
	return "TO_CHAR(TO_TIMESTAMP(timestamp / 1000)::DATE, 'YYYY-MM-DD')"
}

func dayBounds(startDate, endDate time.Time) (startMs, endMs int64) {
	start := startDate.UTC().Truncate(24 * time.Hour)
	end := endDate.UTC().Truncate(24 * time.Hour).Add(24*time.Hour - time.Millisecond)
	return start.UnixMilli(), end.UnixMilli()
}

func (r *DashboardRepository) GetSummary(storeType string, startDate, endDate time.Time) (*dto.DashboardSummary, error) {
	summary := &dto.DashboardSummary{}
	startMs, endMs := dayBounds(startDate, endDate)

	q := r.db.Table("orders").
		Where("is_voided = ?", false).
		Where("is_settled = ?", true).
		Where("timestamp >= ?", startMs).
		Where("timestamp <= ?", endMs)

	if storeType != "" {
		q = q.Where("store_type = ?", storeType)
	}

	row := q.Select(
		"COALESCE(COUNT(*), 0) as total_orders",
		"COALESCE(SUM(total), 0) as total_revenue",
		"CASE WHEN COUNT(*) > 0 THEN COALESCE(SUM(total), 0) / COUNT(*) ELSE 0 END as avg_order_value",
		"COALESCE(SUM(CASE WHEN payment_method = 'CASH' THEN total ELSE 0 END), 0) as cash_total",
		"COALESCE(SUM(CASE WHEN payment_method = 'CARD' THEN total ELSE 0 END), 0) as card_total",
		"COALESCE(SUM(CASE WHEN payment_method = 'QRIS' THEN total ELSE 0 END), 0) as qris_total",
	).Row()

	if err := row.Scan(&summary.TotalOrders, &summary.TotalRevenue, &summary.AvgOrderValue,
		&summary.CashTotal, &summary.CardTotal, &summary.QrisTotal); err != nil {
		return nil, err
	}

	var voidedCount int64
	vq := r.db.Table("orders").
		Where("is_voided = ?", true).
		Where("timestamp >= ?", startMs).
		Where("timestamp <= ?", endMs)
	if storeType != "" {
		vq = vq.Where("store_type = ?", storeType)
	}
	vq.Count(&voidedCount)
	summary.VoidedCount = int(voidedCount)

	return summary, nil
}

func (r *DashboardRepository) GetRevenueTrend(storeType string, startDate, endDate time.Time) ([]dto.RevenuePoint, error) {
	startMs, endMs := dayBounds(startDate, endDate)
	dateExpr := r.dateExpr()

	var results []struct {
		Date    string
		Revenue float64
		Orders  int
	}

	q := r.db.Table("orders").
		Select(dateExpr+" as date, COALESCE(SUM(total), 0) as revenue, COUNT(*) as orders").
		Where("is_voided = ?", false).
		Where("timestamp >= ?", startMs).
		Where("timestamp <= ?", endMs)

	if storeType != "" {
		q = q.Where("store_type = ?", storeType)
	}

	if err := q.Group(dateExpr).Order("date ASC").Scan(&results).Error; err != nil {
		return nil, err
	}

	dataMap := make(map[string]dto.RevenuePoint, len(results))
	for _, res := range results {
		dataMap[res.Date] = dto.RevenuePoint{
			Date:    res.Date,
			Revenue: res.Revenue,
			Orders:  res.Orders,
		}
	}

	days := int(endDate.Sub(startDate).Hours()/24) + 1
	points := make([]dto.RevenuePoint, 0, days)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateKey := d.Format("2006-01-02")
		if point, ok := dataMap[dateKey]; ok {
			points = append(points, point)
		} else {
			points = append(points, dto.RevenuePoint{Date: dateKey, Revenue: 0, Orders: 0})
		}
	}

	return points, nil
}

func (r *DashboardRepository) GetTopProducts(storeType string, limit int) ([]dto.TopProduct, error) {
	var results []dto.TopProduct

	q := r.db.Table("order_items").
		Select("order_items.product_id, order_items.product_name, SUM(order_items.qty) as total_sold, SUM(order_items.subtotal) as revenue").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.is_voided = ?", false)

	if storeType != "" {
		q = q.Where("orders.store_type = ?", storeType)
	}

	if err := q.Group("order_items.product_id, order_items.product_name").
		Order("revenue DESC").
		Limit(limit).
		Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
