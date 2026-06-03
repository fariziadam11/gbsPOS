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

func (r *DashboardRepository) GetSummary(storeType string, since time.Time) (*dto.DashboardSummary, error) {
	summary := &dto.DashboardSummary{}

	q := r.db.Table("orders").
		Where("is_voided = ?", false).
		Where("is_settled = ?", true).
		Where("created_at >= ?", since)

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
		Where("created_at >= ?", since)
	if storeType != "" {
		vq = vq.Where("store_type = ?", storeType)
	}
	vq.Count(&voidedCount)
	summary.VoidedCount = int(voidedCount)

	return summary, nil
}

func (r *DashboardRepository) GetRevenueTrend(storeType string, days int) ([]dto.RevenuePoint, error) {
	since := time.Now().AddDate(0, 0, -days)
	var results []struct {
		Date    string
		Revenue float64
		Orders  int
	}

	q := r.db.Table("orders").
		Select("TO_CHAR(DATE(created_at), 'YYYY-MM-DD') as date, COALESCE(SUM(total), 0) as revenue, COUNT(*) as orders").
		Where("is_voided = ?", false).
		Where("created_at >= ?", since)

	if storeType != "" {
		q = q.Where("store_type = ?", storeType)
	}

	if err := q.Group("DATE(created_at)").Order("date ASC").Scan(&results).Error; err != nil {
		return nil, err
	}

	points := make([]dto.RevenuePoint, len(results))
	for i, r := range results {
		points[i] = dto.RevenuePoint{
			Date:    r.Date,
			Revenue: r.Revenue,
			Orders:  r.Orders,
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
