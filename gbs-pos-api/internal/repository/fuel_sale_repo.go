package repository

import (
	"gbs-pos-api/internal/model"
	"time"

	"gorm.io/gorm"
)

type FuelSaleRepository struct {
	db *gorm.DB
}

func NewFuelSaleRepository(db *gorm.DB) *FuelSaleRepository {
	return &FuelSaleRepository{db: db}
}

func (r *FuelSaleRepository) Create(sale *model.FuelSale) error {
	return r.db.Create(sale).Error
}

func (r *FuelSaleRepository) FindByID(id string) (*model.FuelSale, error) {
	var sale model.FuelSale
	if err := r.db.First(&sale, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &sale, nil
}

func (r *FuelSaleRepository) Report(from, to time.Time) (*model.FuelSalesReport, error) {
	var summary []model.FuelSalesReportItem
	if err := r.db.Raw(`
		SELECT fuel_code, COALESCE(SUM(liters), 0) as liters, COALESCE(SUM(total_amount), 0) as total_amount
		FROM fuel_sales
		WHERE timestamp BETWEEN ? AND ?
		GROUP BY fuel_code
	`, from, to).Scan(&summary).Error; err != nil {
		return nil, err
	}

	var pumpTotals []model.FuelSalesPumpReportItem
	if err := r.db.Raw(`
		SELECT pump_id, COALESCE(SUM(liters), 0) as liters, COALESCE(SUM(total_amount), 0) as total_amount
		FROM fuel_sales
		WHERE timestamp BETWEEN ? AND ?
		GROUP BY pump_id
	`, from, to).Scan(&pumpTotals).Error; err != nil {
		return nil, err
	}

	return &model.FuelSalesReport{
		Summary:    summary,
		PumpTotals: pumpTotals,
	}, nil
}
