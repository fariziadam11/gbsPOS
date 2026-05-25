package repository

import (
	"gbs-cms-api/internal/model"
	"time"

	"gorm.io/gorm"
)

type AdRepository struct {
	db *gorm.DB
}

func NewAdRepository(db *gorm.DB) *AdRepository {
	return &AdRepository{db: db}
}

func (r *AdRepository) FindAll(page, limit int) ([]model.Ad, int64, error) {
	var ads []model.Ad
	var total int64
	offset := (page - 1) * limit
	if err := r.db.Model(&model.Ad{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Order("playlist_order ASC, id ASC").Limit(limit).Offset(offset).Find(&ads).Error; err != nil {
		return nil, 0, err
	}
	return ads, total, nil
}

func (r *AdRepository) FindByID(id uint) (*model.Ad, error) {
	var ad model.Ad
	if err := r.db.First(&ad, id).Error; err != nil {
		return nil, err
	}
	return &ad, nil
}

func (r *AdRepository) Create(ad *model.Ad) error {
	return r.db.Create(ad).Error
}

func (r *AdRepository) Update(ad *model.Ad) error {
	return r.db.Save(ad).Error
}

func (r *AdRepository) Delete(id uint) error {
	return r.db.Delete(&model.Ad{}, id).Error
}

func (r *AdRepository) FindActiveByStoreType(storeType string) ([]model.Ad, error) {
	var ads []model.Ad
	now := time.Now()
	query := r.db.Where("is_active = ?", true).
		Where("store_types LIKE ?", "%"+storeType+"%").
		Where("(start_date IS NULL OR start_date <= ?)", now).
		Where("(end_date IS NULL OR end_date >= ?)", now).
		Where("(start_time IS NULL OR start_time <= ?)", now).
		Where("(end_time IS NULL OR end_time >= ?)", now).
		Order("playlist_order ASC")
	if err := query.Find(&ads).Error; err != nil {
		return nil, err
	}
	return ads, nil
}
