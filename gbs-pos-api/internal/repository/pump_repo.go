package repository

import (
	"gbs-pos-api/internal/model"

	"gorm.io/gorm"
)

type PumpRepository struct {
	db *gorm.DB
}

func NewPumpRepository(db *gorm.DB) *PumpRepository {
	return &PumpRepository{db: db}
}

func (r *PumpRepository) FindAll() ([]model.Pump, error) {
	var pumps []model.Pump
	if err := r.db.Order("id").Find(&pumps).Error; err != nil {
		return nil, err
	}
	return pumps, nil
}

func (r *PumpRepository) FindByID(id string) (*model.Pump, error) {
	var pump model.Pump
	if err := r.db.First(&pump, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &pump, nil
}

func (r *PumpRepository) Create(pump *model.Pump) error {
	return r.db.Create(pump).Error
}

func (r *PumpRepository) Update(pump *model.Pump) error {
	return r.db.Save(pump).Error
}

func (r *PumpRepository) Delete(id string) error {
	return r.db.Delete(&model.Pump{}, "id = ?", id).Error
}
