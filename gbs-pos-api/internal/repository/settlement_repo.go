package repository

import (
	"gbs-pos-api/internal/model"

	"gorm.io/gorm"
)

type SettlementRepository struct {
	db *gorm.DB
}

func NewSettlementRepository(db *gorm.DB) *SettlementRepository {
	return &SettlementRepository{db: db}
}

func (r *SettlementRepository) WithTx(tx *gorm.DB) *SettlementRepository {
	return &SettlementRepository{db: tx}
}

func (r *SettlementRepository) Create(settlement *model.Settlement) error {
	return r.db.Create(settlement).Error
}

func (r *SettlementRepository) FindAll(limit int, storeType string) ([]model.Settlement, error) {
	var settlements []model.Settlement
	query := r.db.Order("timestamp DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if storeType != "" {
		query = query.Where("store_type = ?", storeType)
	}
	if err := query.Find(&settlements).Error; err != nil {
		return nil, err
	}
	return settlements, nil
}

func (r *SettlementRepository) FindByID(id string) (*model.Settlement, error) {
	var settlement model.Settlement
	if err := r.db.Where("id = ?", id).First(&settlement).Error; err != nil {
		return nil, err
	}
	return &settlement, nil
}
