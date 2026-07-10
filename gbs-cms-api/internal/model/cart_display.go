package model

import (
	"time"
)

// CartDisplay stores the latest JSON payload for a POS terminal customer display.
type CartDisplay struct {
	ID         uint      `gorm:"primaryKey"`
	TerminalID string    `gorm:"uniqueIndex;size:64;not null"`
	Payload    string    `gorm:"type:jsonb;not null"` // stores entire JSON

	// Device information columns (nullable for backward compatibility)
	DeviceModel        *string `gorm:"size:100"`
	DeviceManufacturer *string `gorm:"size:100"`
	DeviceBrand        *string `gorm:"size:100"`
	AndroidVersion     *string `gorm:"size:20"`
	SDKInt             *int    `gorm:"type:integer"`
	AppVersion         *string `gorm:"size:50"`
	AppVersionCode     *int64  `gorm:"type:bigint"`

	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
