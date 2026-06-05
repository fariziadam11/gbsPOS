package repository

import (
	"gbs-pos-api/internal/model"

	"gorm.io/gorm"
)

type StockMovementRepository struct {
	db *gorm.DB
}

func NewStockMovementRepository(db *gorm.DB) *StockMovementRepository {
	return &StockMovementRepository{db: db}
}

func (r *StockMovementRepository) Create(movement *model.StockMovement) error {
	return r.db.Create(movement).Error
}

func (r *StockMovementRepository) FindByProductID(productID int) ([]model.StockMovement, error) {
	var movements []model.StockMovement
	if err := r.db.Where("product_id = ?", productID).Order("created_at DESC").Find(&movements).Error; err != nil {
		return nil, err
	}
	return movements, nil
}

func (r *StockMovementRepository) FindAll(limit int) ([]model.StockMovement, error) {
	var movements []model.StockMovement
	query := r.db.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&movements).Error; err != nil {
		return nil, err
	}
	return movements, nil
}
