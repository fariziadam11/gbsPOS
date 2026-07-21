package repository

import (
	"context"
	"fmt"

	"gbs-pos-api/internal/model"

	"gorm.io/gorm"
)

// QrisTransactionRepository handles QRIS transaction data access
type QrisTransactionRepository struct {
	db *gorm.DB
}

// NewQrisTransactionRepository creates a new QRIS transaction repository
func NewQrisTransactionRepository(db *gorm.DB) *QrisTransactionRepository {
	return &QrisTransactionRepository{db: db}
}

// Create creates a new QRIS transaction
func (r *QrisTransactionRepository) Create(ctx context.Context, tx *model.QrisTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

// FindByID finds a QRIS transaction by ID
func (r *QrisTransactionRepository) FindByID(ctx context.Context, id string) (*model.QrisTransaction, error) {
	var tx model.QrisTransaction
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&tx).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("QRIS transaction not found")
		}
		return nil, err
	}
	return &tx, nil
}

// FindByOrderID finds QRIS transactions by order ID
func (r *QrisTransactionRepository) FindByOrderID(ctx context.Context, orderID string) ([]model.QrisTransaction, error) {
	var txs []model.QrisTransaction
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at DESC").Find(&txs).Error; err != nil {
		return nil, err
	}
	return txs, nil
}

// FindPendingByTerminalID finds all pending QRIS transactions for a terminal
func (r *QrisTransactionRepository) FindPendingByTerminalID(ctx context.Context, terminalID string) ([]model.QrisTransaction, error) {
	var txs []model.QrisTransaction
	if err := r.db.WithContext(ctx).
		Where("terminal_id = ? AND status = ?", terminalID, model.QrisTransactionStatusPending).
		Order("created_at DESC").
		Find(&txs).Error; err != nil {
		return nil, err
	}
	return txs, nil
}

// UpdateStatus updates the status of a QRIS transaction
func (r *QrisTransactionRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == model.QrisTransactionStatusPaid {
		updates["paid_at"] = gorm.Expr("NOW()")
	}

	return r.db.WithContext(ctx).Model(&model.QrisTransaction{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// MarkAsPaid marks a QRIS transaction as paid
func (r *QrisTransactionRepository) MarkAsPaid(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&model.QrisTransaction{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":   model.QrisTransactionStatusPaid,
			"paid_at":  gorm.Expr("NOW()"),
		}).Error
}

// Cancel cancels a QRIS transaction
func (r *QrisTransactionRepository) Cancel(ctx context.Context, id string, cancelledBy string, reason string) error {
	return r.db.WithContext(ctx).Model(&model.QrisTransaction{}).
		Where("id = ? AND status = ?", id, model.QrisTransactionStatusPending).
		Updates(map[string]interface{}{
			"status":        model.QrisTransactionStatusCancelled,
			"cancelled_at":  gorm.Expr("NOW()"),
			"cancelled_by":  cancelledBy,
			"cancel_reason": reason,
		}).Error
}

// FindExpired finds and marks expired QRIS transactions
func (r *QrisTransactionRepository) FindExpired(ctx context.Context) ([]model.QrisTransaction, error) {
	var txs []model.QrisTransaction
	if err := r.db.WithContext(ctx).
		Where("status = ? AND expires_at < NOW()", model.QrisTransactionStatusPending).
		Find(&txs).Error; err != nil {
		return nil, err
	}
	return txs, nil
}

// MarkAsExpired marks a QRIS transaction as expired
func (r *QrisTransactionRepository) MarkAsExpired(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&model.QrisTransaction{}).
		Where("id = ?", id).
		Update("status", model.QrisTransactionStatusExpired).Error
}
