package repository

import (
	"gbs-pos-api/internal/model"

	"gorm.io/gorm"
)

type HoldRepository struct {
	db *gorm.DB
}

func NewHoldRepository(db *gorm.DB) *HoldRepository {
	return &HoldRepository{db: db}
}

func (r *HoldRepository) WithTx(tx *gorm.DB) *HoldRepository {
	return &HoldRepository{db: tx}
}

func (r *HoldRepository) Create(session *model.HoldSession) error {
	return r.db.Create(session).Error
}

func (r *HoldRepository) FindByID(id string) (*model.HoldSession, error) {
	var session model.HoldSession

	err := r.db.
		Where("id = ?", id).
		First(&session).Error

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *HoldRepository) FindActiveByTerminal(terminalID string) ([]model.HoldSession, error) {
	var sessions []model.HoldSession

	err := r.db.
		Where("terminal_id = ? AND status = ?", terminalID, model.HoldStatusActive).
		Order("created_at DESC").
		Find(&sessions).Error

	return sessions, err
}

func (r *HoldRepository) UpdateStatus(id string, status model.HoldStatus) error {
	result := r.db.
		Model(&model.HoldSession{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *HoldRepository) Delete(id string) error {
	result := r.db.
		Where("id = ?", id).
		Delete(&model.HoldSession{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
