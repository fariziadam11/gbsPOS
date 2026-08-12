package repository

import (
	"context"

	"gbs-pos-api/internal/model"

	"gorm.io/gorm"
)

type CompanionDeviceRepository struct {
	db *gorm.DB
}

func NewCompanionDeviceRepository(db *gorm.DB) *CompanionDeviceRepository {
	return &CompanionDeviceRepository{db: db}
}

func (r *CompanionDeviceRepository) Upsert(ctx context.Context, device *model.CompanionDevice) error {
	var existing model.CompanionDevice
	err := r.db.WithContext(ctx).First(&existing, "device_id = ?", device.DeviceID).Error
	if err == nil {
		existing.DeviceName = device.DeviceName
		existing.SDKVersion = device.SDKVersion
		existing.Capabilities = device.Capabilities
		existing.LastSeenAt = device.LastSeenAt
		return r.db.WithContext(ctx).Save(&existing).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return r.db.WithContext(ctx).Create(device).Error
}

func (r *CompanionDeviceRepository) Exists(ctx context.Context, deviceID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.CompanionDevice{}).Where("device_id = ?", deviceID).Count(&count).Error
	return count > 0, err
}
