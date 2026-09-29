package repository

import (
	"context"
	"time"

	"gbs-pos-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CardPaymentRepository struct {
	db *gorm.DB
}

func NewCardPaymentRepository(db *gorm.DB) *CardPaymentRepository {
	return &CardPaymentRepository{db: db}
}

func (r *CardPaymentRepository) Create(ctx context.Context, payment *model.CardPayment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *CardPaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.CardPayment, error) {
	var payment model.CardPayment
	if err := r.db.WithContext(ctx).First(&payment, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *CardPaymentRepository) Update(ctx context.Context, payment *model.CardPayment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

func (r *CardPaymentRepository) FinalizeSuccess(ctx context.Context, payment *model.CardPayment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(payment).Error; err != nil {
			return err
		}
		result := tx.Model(&model.Order{}).Where("id = ?", payment.OrderID).Updates(map[string]interface{}{
			"payment_method": "CARD",
			"transaction_id": payment.TransactionID,
			"approval_code":  payment.AuthCode,
			"entry_mode":     payment.EntryMode,
			"masked_account": payment.MaskedCard,
			"acq_mid":        payment.AcqMID,
			"acq_tid":        payment.AcqTID,
			"pos_message_id": payment.PosMessageID,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *CardPaymentRepository) FindExpired(ctx context.Context, now time.Time) ([]model.CardPayment, error) {
	var payments []model.CardPayment
	err := r.db.WithContext(ctx).
		Where("status IN (?, ?) AND expires_at <= ?", model.CardPaymentWaiting, model.CardPaymentProcessing, now).
		Find(&payments).Error
	return payments, err
}

func (r *CardPaymentRepository) FindPendingByDevice(ctx context.Context, deviceID string) ([]model.CardPayment, error) {
	var payments []model.CardPayment
	err := r.db.WithContext(ctx).
		Where("device_id = ? AND status IN (?, ?)", deviceID, model.CardPaymentWaiting, model.CardPaymentProcessing).
		Order("created_at ASC").Find(&payments).Error
	return payments, err
}

func (r *CardPaymentRepository) MarkExpired(ctx context.Context, id uuid.UUID, now time.Time) (bool, error) {
	result := r.db.WithContext(ctx).Model(&model.CardPayment{}).
		Where("id = ? AND status IN (?, ?) AND expires_at <= ?", id, model.CardPaymentWaiting, model.CardPaymentProcessing, now).
		Update("status", model.CardPaymentExpired)
	return result.RowsAffected > 0, result.Error
}

// MarkProcessing atomically transitions a WAITING_FOR_CARD payment to PROCESSING and
// refreshes its expiry window. Returns false when the payment is no longer waiting
// (e.g. already SUCCESS/CANCELLED/EXPIRED, or lost a race with Cancel/Expire).
func (r *CardPaymentRepository) MarkProcessing(ctx context.Context, id uuid.UUID, newExpiresAt time.Time) (bool, error) {
	result := r.db.WithContext(ctx).Model(&model.CardPayment{}).
		Where("id = ? AND status = ?", id, model.CardPaymentWaiting).
		Updates(map[string]interface{}{
			"status":     model.CardPaymentProcessing,
			"expires_at": newExpiresAt,
		})
	return result.RowsAffected > 0, result.Error
}

// CancelIfWaiting atomically cancels a payment only while it is still WAITING_FOR_CARD.
// Returns false when the payment already moved on (PROCESSING/SUCCESS/EXPIRED), so the
// caller never voids an order that may have been paid.
func (r *CardPaymentRepository) CancelIfWaiting(ctx context.Context, id uuid.UUID) (bool, error) {
	result := r.db.WithContext(ctx).Model(&model.CardPayment{}).
		Where("id = ? AND status = ?", id, model.CardPaymentWaiting).
		Update("status", model.CardPaymentCancelled)
	return result.RowsAffected > 0, result.Error
}
