package repository

import (
	"gbs-cms-api/internal/model"
	"gorm.io/gorm"
)

type AdPlayLogRepository struct {
	db *gorm.DB
}

func NewAdPlayLogRepository(db *gorm.DB) *AdPlayLogRepository {
	return &AdPlayLogRepository{db: db}
}

func (r *AdPlayLogRepository) Create(log *model.AdPlayLog) error {
	return r.db.Create(log).Error
}
