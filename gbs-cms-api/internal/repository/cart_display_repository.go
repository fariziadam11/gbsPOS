package repository

import (
	"gbs-cms-api/internal/model"
	"gbs-cms-api/internal/dto"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CartDisplayRepository struct {
	db *gorm.DB
}

func NewCartDisplayRepository(db *gorm.DB) *CartDisplayRepository {
	return &CartDisplayRepository{db: db}
}

// Upsert inserts or updates the cart display payload for a terminal.
func (r *CartDisplayRepository) Upsert(terminalID string, payload string, deviceInfo *dto.DeviceInfo) error {
	display := model.CartDisplay{
		TerminalID: terminalID,
		Payload:   payload,
	}

	// Set device info if provided
	if deviceInfo != nil {
		display.DeviceModel = &deviceInfo.DeviceModel
		display.DeviceManufacturer = &deviceInfo.DeviceManufacturer
		display.DeviceBrand = &deviceInfo.DeviceBrand
		display.AndroidVersion = &deviceInfo.AndroidVersion
		display.SDKInt = &deviceInfo.SDKInt
		display.AppVersion = &deviceInfo.AppVersion
		display.AppVersionCode = &deviceInfo.AppVersionCode
	}

	return r.db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "terminal_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"payload", "updated_at", "device_model", "device_manufacturer", "device_brand", "android_version", "sdk_int", "app_version", "app_version_code"}),
		}).
		Create(&display).Error
}

// GetByTerminalID retrieves the cart display for a terminal.
func (r *CartDisplayRepository) GetByTerminalID(terminalID string) (*model.CartDisplay, error) {
	var display model.CartDisplay

	err := r.db.
		Where("terminal_id = ?", terminalID).
		First(&display).Error

	if err != nil {
		return nil, err
	}

	return &display, nil
}

// GetAll retrieves all cart displays ordered by most recent first.
func (r *CartDisplayRepository) GetAll() ([]model.CartDisplay, error) {
	var displays []model.CartDisplay

	err := r.db.
		Order("updated_at DESC").
		Find(&displays).Error

	if err != nil {
		return nil, err
	}

	return displays, nil
}
