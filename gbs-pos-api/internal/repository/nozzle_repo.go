package repository

import (
	"gbs-pos-api/internal/model"

	"gorm.io/gorm"
)

type NozzleRepository struct {
	db *gorm.DB
}

func NewNozzleRepository(db *gorm.DB) *NozzleRepository {
	return &NozzleRepository{db: db}
}

func (r *NozzleRepository) FindAll() ([]model.Nozzle, error) {
	var nozzles []model.Nozzle
	if err := r.db.Order("id").Find(&nozzles).Error; err != nil {
		return nil, err
	}
	return nozzles, nil
}

func (r *NozzleRepository) FindByPumpID(pumpID string) ([]model.Nozzle, error) {
	var nozzles []model.Nozzle
	if err := r.db.Where("pump_id = ?", pumpID).Order("id").Find(&nozzles).Error; err != nil {
		return nil, err
	}
	return nozzles, nil
}

func (r *NozzleRepository) FindByID(id string) (*model.Nozzle, error) {
	var nozzle model.Nozzle
	if err := r.db.First(&nozzle, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &nozzle, nil
}

func (r *NozzleRepository) Create(nozzle *model.Nozzle) error {
	return r.db.Create(nozzle).Error
}

func (r *NozzleRepository) Update(nozzle *model.Nozzle) error {
	return r.db.Save(nozzle).Error
}

func (r *NozzleRepository) Delete(id string) error {
	return r.db.Delete(&model.Nozzle{}, "id = ?", id).Error
}
