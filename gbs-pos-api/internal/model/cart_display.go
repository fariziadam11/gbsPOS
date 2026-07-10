package model

import (
	"time"

	"gorm.io/datatypes"
)

// CartDisplay stores the latest JSON payload for a POS terminal customer display.
// The backend treats Payload as an opaque JSONB document.
type CartDisplay struct {
	ID         uint           `gorm:"primaryKey"`
	TerminalID string         `gorm:"uniqueIndex;size:64;not null"`
	Payload    datatypes.JSON `gorm:"type:jsonb;not null"`

	// Device information columns (nullable for backward compatibility)
	DeviceModel        *string `gorm:"size:100"`
	DeviceManufacturer *string `gorm:"size:100"`
	DeviceBrand        *string `gorm:"size:100"`
	AndroidVersion     *string `gorm:"size:20"`
	SDKInt             *int    `gorm:"type:integer"`
	AppVersion         *string `gorm:"size:50"`
	AppVersionCode     *int64  `gorm:"type:bigint"`

	UpdatedAt time.Time
}
