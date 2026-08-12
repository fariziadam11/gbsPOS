package model

import "time"

type CompanionDevice struct {
	DeviceID     string    `gorm:"primaryKey;size:100" json:"deviceId"`
	DeviceName   string    `gorm:"size:150" json:"deviceName"`
	SDKVersion   string    `gorm:"size:50" json:"sdkVersion"`
	Capabilities string    `gorm:"type:text" json:"capabilities"`
	LastSeenAt   time.Time `json:"lastSeenAt"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
