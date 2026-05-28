package repository

import (
	"gbs-pos-api/internal/dto"
	"gbs-pos-api/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) WithTx(tx *gorm.DB) *OrderRepository {
	return &OrderRepository{db: tx}
}

func (r *OrderRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *OrderRepository) FindAll(
	storeType string,
	startDate, endDate int64,
	isVoided, isSettled *bool,
	paymentMethod, terminalID string,
) ([]model.Order, error) {
	var orders []model.Order
	query := r.db.Order("timestamp DESC")
	if storeType != "" {
		query = query.Where("store_type = ?", storeType)
	}
	if startDate > 0 {
		query = query.Where("timestamp >= ?", startDate)
	}
	if endDate > 0 {
		query = query.Where("timestamp <= ?", endDate)
	}
	if isVoided != nil {
		query = query.Where("is_voided = ?", *isVoided)
	}
	if isSettled != nil {
		query = query.Where("is_settled = ?", *isSettled)
	}
	if paymentMethod != "" {
		query = query.Where("payment_method = ?", paymentMethod)
	}
	if terminalID != "" {
		query = query.Where("terminal_id = ?", terminalID)
	}
	if err := query.Preload("Items").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepository) FindByID(id string) (*model.Order, error) {
	var order model.Order
	if err := r.db.Where("id = ?", id).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) FindByIDWithItems(id string) (*model.Order, error) {
	var order model.Order
	if err := r.db.Where("id = ?", id).Preload("Items").First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepository) UpdateVoid(order *model.Order) error {
	return r.db.Save(order).Error
}

func (r *OrderRepository) FindUnsettledSummary(
	storeType, terminalID string,
) (count int, total float64, summary map[string]dto.PaymentSummary, err error) {
	var results []dto.PaymentMethodQueryResult

	query := r.db.Model(&model.Order{}).
		Select("payment_method, COUNT(*) as count, SUM(total) as total").
		Where("is_settled = ? AND is_voided = ?", false, false)
	if storeType != "" {
		query = query.Where("store_type = ?", storeType)
	}
	if terminalID != "" {
		query = query.Where("terminal_id = ?", terminalID)
	}
	if err = query.Group("payment_method").Scan(&results).Error; err != nil {
		return 0, 0, nil, err
	}

	summary = make(map[string]dto.PaymentSummary)
	for _, res := range results {
		count += res.Count
		total += res.Total
		summary[res.PaymentMethod] = dto.PaymentSummary{Count: res.Count, Total: res.Total}
	}
	for _, pm := range []string{"CASH", "CARD", "QRIS"} {
		if _, ok := summary[pm]; !ok {
			summary[pm] = dto.PaymentSummary{Count: 0, Total: 0}
		}
	}
	return count, total, summary, nil
}

func (r *OrderRepository) FindUnsettledOrders(
	storeType, terminalID string,
	forUpdate bool,
) ([]model.Order, error) {
	var orders []model.Order

	query := r.db.Where("is_settled = ? AND is_voided = ?", false, false)
	if storeType != "" {
		query = query.Where("store_type = ?", storeType)
	}
	if terminalID != "" {
		query = query.Where("terminal_id = ?", terminalID)
	}
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepository) MarkSettled(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Model(&model.Order{}).Where("id IN ?", ids).Update("is_settled", true).Error
}
