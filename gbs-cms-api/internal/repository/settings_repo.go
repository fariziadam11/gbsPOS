package repository

import (
	"gbs-cms-api/internal/model"
	"gorm.io/gorm"
)

type SettingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) GetAll() ([]model.Setting, error) {
	var settings []model.Setting
	if err := r.db.Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *SettingsRepository) GetByKey(key string) (*model.Setting, error) {
	var setting model.Setting
	if err := r.db.Where("key = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *SettingsRepository) Upsert(key, value string) error {
	setting := model.Setting{Key: key, Value: value}
	return r.db.Where(model.Setting{Key: key}).
		Assign(model.Setting{Value: value}).
		FirstOrCreate(&setting).Error
}

func (r *SettingsRepository) UpsertAll(settings map[string]string) error {
	for key, value := range settings {
		if err := r.Upsert(key, value); err != nil {
			return err
		}
	}
	return nil
}
