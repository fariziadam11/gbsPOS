package repository

import (
	"gbs-pos-api/internal/model"

	"gorm.io/gorm"
)

type FuelPriceRepository struct {
	db *gorm.DB
}

func NewFuelPriceRepository(db *gorm.DB) *FuelPriceRepository {
	return &FuelPriceRepository{db: db}
}

func (r *FuelPriceRepository) FindAll() ([]model.FuelPrice, error) {
	var prices []model.FuelPrice
	if err := r.db.Order("code").Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

func (r *FuelPriceRepository) FindByCode(code string) (*model.FuelPrice, error) {
	var price model.FuelPrice
	if err := r.db.First(&price, "code = ?", code).Error; err != nil {
		return nil, err
	}
	return &price, nil
}

func (r *FuelPriceRepository) Update(price *model.FuelPrice) error {
	return r.db.Save(price).Error
}

func (r *FuelPriceRepository) Upsert(prices []model.FuelPrice) error {
	for i := range prices {
		if err := r.db.Save(&prices[i]).Error; err != nil {
			return err
		}
	}
	return nil
}
