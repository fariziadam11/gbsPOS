package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (r *QrisTransactionRepository) FindLatestByOrderIDAndProvider(ctx context.Context, orderID, provider string) (*model.QrisTransaction, error) {
	var tx model.QrisTransaction
	if err := r.db.WithContext(ctx).
		Where("order_id = ? AND provider = ?", orderID, provider).
		Order("created_at DESC").First(&tx).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tx, nil
}

func (r *QrisTransactionRepository) FindByPartnerReferenceNo(ctx context.Context, reference string) (*model.QrisTransaction, error) {
	var tx model.QrisTransaction
	if err := r.db.WithContext(ctx).Where("partner_reference_no = ?", reference).First(&tx).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("QRIS transaction not found")
		}
		return nil, err
	}
	return &tx, nil
}

func (r *QrisTransactionRepository) UpdateBTNGenerated(ctx context.Context, id, bankReferenceNo, qrContent, merchantName, terminalID string) error {
	result := r.db.WithContext(ctx).Model(&model.QrisTransaction{}).
		Where("id = ? AND provider = ? AND status IN ?", id, model.QrisProviderBTNSnap, []string{model.QrisTransactionStatusPending, model.QrisTransactionStatusUnknown}).
		Updates(map[string]interface{}{
			"bank_reference_no":   bankReferenceNo,
			"dynamic_qris_string": qrContent,
			"merchant_name":       merchantName,
			"terminal_id":         terminalID,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("BTN QR transaction is no longer pending")
	}
	return nil
}

func (r *QrisTransactionRepository) ApplyBTNStatus(
	ctx context.Context,
	id, status, bankStatus, description, reference, paymentReference string,
	bankPaidAt *time.Time,
) (bool, error) {
	query := r.db.WithContext(ctx).Model(&model.QrisTransaction{}).Where("id = ? AND provider = ?", id, model.QrisProviderBTNSnap)
	updates := map[string]interface{}{
		"status":                  status,
		"bank_status":             bankStatus,
		"bank_status_description": description,
	}
	if reference != "" {
		updates["bank_reference_no"] = reference
	}
	if paymentReference != "" {
		updates["bank_payment_reference"] = paymentReference
	}
	if status == model.QrisTransactionStatusPaid {
		if bankPaidAt == nil {
			return false, fmt.Errorf("bank paid time is required for BTN success")
		}
		updates["paid_at"] = bankPaidAt
		updates["bank_paid_at"] = bankPaidAt
		query = query.Where("status NOT IN ?", []string{model.QrisTransactionStatusRefunded})
	} else if status == model.QrisTransactionStatusRefunded {
		query = query.Where("status IN ?", []string{model.QrisTransactionStatusPaid, model.QrisTransactionStatusRefunded})
	} else {
		query = query.Where("status NOT IN ?", []string{model.QrisTransactionStatusPaid, model.QrisTransactionStatusRefunded})
	}
	result := query.Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
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
	result := r.db.WithContext(ctx).Model(&model.QrisTransaction{}).
		Where("id = ? AND status IN ?", id, []string{model.QrisTransactionStatusPending, model.QrisTransactionStatusAwaitingConfirmation}).
		Updates(map[string]interface{}{
			"status":  model.QrisTransactionStatusPaid,
			"paid_at": gorm.Expr("NOW()"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("QRIS transaction status changed")
	}
	return nil
}

// Cancel cancels a QRIS transaction
func (r *QrisTransactionRepository) Cancel(ctx context.Context, id string, cancelledBy string, reason string) error {
	return r.db.WithContext(ctx).Model(&model.QrisTransaction{}).
		Where("id = ? AND status IN (?, ?)", id, model.QrisTransactionStatusPending, model.QrisTransactionStatusAwaitingConfirmation).
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
		Where("status = ? AND provider <> ? AND expires_at < NOW()", model.QrisTransactionStatusPending, model.QrisProviderBTNSnap).
		Find(&txs).Error; err != nil {
		return nil, err
	}
	return txs, nil
}

// MarkAsExpired marks a QRIS transaction as expired
func (r *QrisTransactionRepository) MarkAsExpired(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&model.QrisTransaction{}).
		Where("id = ? AND status = ?", id, model.QrisTransactionStatusPending).
		Update("status", model.QrisTransactionStatusExpired).Error
}

// AtomicConfirmIfStatus atomically confirms a transaction only if it has the expected status.
// Returns (updated bool, error) - updated is true if the status was changed.
func (r *QrisTransactionRepository) AtomicConfirmIfStatus(ctx context.Context, id string, expectedStatus string) (bool, error) {
	result := r.db.WithContext(ctx).Model(&model.QrisTransaction{}).
		Where("id = ? AND status = ?", id, expectedStatus).
		Updates(map[string]interface{}{
			"status":  model.QrisTransactionStatusPaid,
			"paid_at": gorm.Expr("NOW()"),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
