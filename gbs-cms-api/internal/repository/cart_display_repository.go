package repository

import (
	"gbs-cms-api/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CartDisplayRepository struct {
	db *gorm.DB
}

func NewCartDisplayRepository(db *gorm.DB) *CartDisplayRepository {
	return &CartDisplayRepository{db: db}
}

// Upsert inserts or updates the cart display payload for a terminal.
func (r *CartDisplayRepository) Upsert(terminalID string, payload string) error {
	display := model.CartDisplay{
		TerminalID: terminalID,
		Payload:   payload,
	}

	return r.db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "terminal_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"payload", "updated_at"}),
		}).
		Create(&display).Error
}

// GetByTerminalID retrieves the cart display for a terminal.
func (r *CartDisplayRepository) GetByTerminalID(terminalID string) (*model.CartDisplay, error) {
	var display model.CartDisplay

	err := r.db.
		Where("terminal_id = ?", terminalID).
		First(&display).Error

	if err != nil {
		return nil, err
	}

	return &display, nil
}
