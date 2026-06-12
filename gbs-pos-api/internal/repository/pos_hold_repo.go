package repository

import (
	"gbs-pos-api/internal/model"

	"gorm.io/gorm"
)

type PosHoldRepository struct {
	db *gorm.DB
}

func NewPosHoldRepository(db *gorm.DB) *PosHoldRepository {
	return &PosHoldRepository{db: db}
}

func (r *PosHoldRepository) WithTx(tx *gorm.DB) *PosHoldRepository {
	return &PosHoldRepository{db: tx}
}

func (r *PosHoldRepository) Create(session *model.PosHoldSession) error {
	return r.db.Create(session).Error
}

func (r *PosHoldRepository) FindByID(id string) (*model.PosHoldSession, error) {
	var session model.PosHoldSession

	err := r.db.
		Where("id = ?", id).
		First(&session).Error

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *PosHoldRepository) FindActiveByTerminal(terminalID string) ([]model.PosHoldSession, error) {
	var sessions []model.PosHoldSession

	err := r.db.
		Where("terminal_id = ? AND status = ?", terminalID, model.PosHoldStatusActive).
		Order("created_at DESC").
		Find(&sessions).Error

	return sessions, err
}

func (r *PosHoldRepository) UpdateStatus(id string, status string) error {
	return r.db.
		Model(&model.PosHoldSession{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *PosHoldRepository) Delete(id string) error {
	return r.db.
		Where("id = ?", id).
		Delete(&model.PosHoldSession{}).Error
}