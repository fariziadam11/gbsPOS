package repository

import (
	"gbs-pos-api/internal/model"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CartDisplayRepository struct {
	db *gorm.DB
}

func NewCartDisplayRepository(db *gorm.DB) *CartDisplayRepository {
	return &CartDisplayRepository{db: db}
}

func (r *CartDisplayRepository) WithTx(tx *gorm.DB) *CartDisplayRepository {
	return &CartDisplayRepository{db: tx}
}

// Upsert inserts a new CartDisplay or updates the payload for an existing terminal.
func (r *CartDisplayRepository) Upsert(terminalID string, payload datatypes.JSON, deviceInfo *DeviceInfo) error {
	display := model.CartDisplay{
		TerminalID: terminalID,
		Payload:    payload,
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

// DeviceInfo represents device information from the Android POS app.
type DeviceInfo struct {
	DeviceModel        string
	DeviceManufacturer string
	DeviceBrand        string
	AndroidVersion     string
	SDKInt             int
	AppVersion         string
	AppVersionCode     int64
}

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

func (r *CartDisplayRepository) Delete(terminalID string) error {
	result := r.db.
		Where("terminal_id = ?", terminalID).
		Delete(&model.CartDisplay{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
